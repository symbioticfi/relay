package key_registerer

import (
	"context"
	"encoding/json"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	bls12381fp "github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	relaybls12381 "github.com/symbioticfi/relay/symbiotic/usecase/crypto/bls12381"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
)

type evmClient interface {
	GetCurrentEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetConfig(ctx context.Context, timestamp symbiotic.Timestamp, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error)
	GetEip712Domain(ctx context.Context, addr symbiotic.CrossChainAddress) (symbiotic.Eip712Domain, error)
	GetEpochStart(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.Timestamp, error)
	RegisterKey(ctx context.Context, addr symbiotic.CrossChainAddress, keyTag symbiotic.KeyTag, key symbiotic.CompactPublicKey, signature symbiotic.RawSignature, extraData []byte) (symbiotic.TxResult, error)
}

type Config struct {
	EVMClient evmClient `validate:"required"`
}

type Registerer struct {
	evmClient evmClient
}

type RegistrationArtifact struct {
	KeyTag    symbiotic.KeyTag
	Key       symbiotic.CompactPublicKey
	Signature symbiotic.RawSignature
	ExtraData []byte
}

type registrationContext struct {
	keysProvider symbiotic.CrossChainAddress
	eip712Domain symbiotic.Eip712Domain
}

func NewRegisterer(cfg Config) (*Registerer, error) {
	if err := validate.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("invalid registerer config: %w", err)
	}
	return &Registerer{
		evmClient: cfg.EVMClient,
	}, nil
}

func (r *Registerer) Register(
	ctx context.Context,
	pk symbioticCrypto.PrivateKey,
	kt symbiotic.KeyTag,
	operatorAddress common.Address,
) (symbiotic.TxResult, error) {
	regContext, err := r.registrationContext(ctx)
	if err != nil {
		return symbiotic.TxResult{}, err
	}

	artifact, err := buildRegistrationArtifact(pk, kt, regContext.eip712Domain, operatorAddress)
	if err != nil {
		return symbiotic.TxResult{}, err
	}

	txResult, err := r.evmClient.RegisterKey(
		ctx,
		regContext.keysProvider,
		kt,
		artifact.Key,
		artifact.Signature,
		artifact.ExtraData,
	)
	if err != nil {
		return txResult, errors.Errorf("failed to register key: %w", err)
	}

	return txResult, nil
}

func (r *Registerer) BuildArtifact(
	ctx context.Context,
	pk symbioticCrypto.PrivateKey,
	kt symbiotic.KeyTag,
	operatorAddress common.Address,
) (RegistrationArtifact, error) {
	regContext, err := r.registrationContext(ctx)
	if err != nil {
		return RegistrationArtifact{}, err
	}
	return buildRegistrationArtifact(pk, kt, regContext.eip712Domain, operatorAddress)
}

func (r *Registerer) registrationContext(
	ctx context.Context,
) (registrationContext, error) {
	currentOnchainEpoch, err := r.evmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return registrationContext{}, errors.Errorf("failed to get current epoch: %w", err)
	}

	captureTimestamp, err := r.evmClient.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return registrationContext{}, errors.Errorf("failed to get capture timestamp: %w", err)
	}

	networkConfig, err := r.evmClient.GetConfig(ctx, captureTimestamp, currentOnchainEpoch)
	if err != nil {
		return registrationContext{}, errors.Errorf("failed to get config: %w", err)
	}

	eip712Domain, err := r.evmClient.GetEip712Domain(ctx, networkConfig.KeysProvider)
	if err != nil {
		return registrationContext{}, errors.Errorf("failed to get eip712 domain: %w", err)
	}

	return registrationContext{
		keysProvider: networkConfig.KeysProvider,
		eip712Domain: eip712Domain,
	}, nil
}

func buildRegistrationArtifact(
	pk symbioticCrypto.PrivateKey,
	kt symbiotic.KeyTag,
	eip712Domain symbiotic.Eip712Domain,
	operatorAddress common.Address,
) (RegistrationArtifact, error) {
	key := pk.PublicKey().OnChain()

	commitmentData, err := keyCommitmentData(eip712Domain, operatorAddress, key)
	if err != nil {
		return RegistrationArtifact{}, errors.Errorf("failed to get commitment data: %w", err)
	}

	signature, _, err := pk.Sign(commitmentData)
	if err != nil {
		return RegistrationArtifact{}, errors.Errorf("failed to sign commitment data: %w", err)
	}

	extraData, err := registrationExtraData(pk, kt)
	if err != nil {
		return RegistrationArtifact{}, err
	}

	signature, err = normalizeSignature(kt, signature)
	if err != nil {
		return RegistrationArtifact{}, err
	}

	return RegistrationArtifact{
		KeyTag:    kt,
		Key:       key,
		Signature: signature,
		ExtraData: extraData,
	}, nil
}

func keyCommitmentData(
	eip712Domain symbiotic.Eip712Domain,
	operator common.Address,
	keyBytes []byte,
) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"KeyOwnership": []apitypes.Type{
				{Name: "operator", Type: "address"},
				{Name: "key", Type: "bytes"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:              eip712Domain.Name,
			Version:           eip712Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(eip712Domain.ChainId),
			VerifyingContract: eip712Domain.VerifyingContract.Hex(),
		},
		PrimaryType: "KeyOwnership",
		Message: map[string]interface{}{
			"operator": operator.Hex(),
			"key":      keyBytes,
		},
	}

	_, preHashedData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	return []byte(preHashedData), nil
}

func normalizeSignature(kt symbiotic.KeyTag, signature symbiotic.RawSignature) (symbiotic.RawSignature, error) {
	switch kt.Type() {
	case symbiotic.KeyTypeBlsBn254:
		return signature, nil
	case symbiotic.KeyTypeEcdsaSecp256k1:
		if len(signature) == 65 {
			signature = append(symbiotic.RawSignature(nil), signature...)
			signature[64] += 27
		}
		return signature, nil
	case symbiotic.KeyTypeBls12381:
		formatted, err := formatBls12381G1OnChain(signature)
		if err != nil {
			return nil, err
		}
		return symbiotic.RawSignature(formatted), nil
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("invalid key type")
	default:
		return nil, errors.New("invalid key type")
	}
}

func registrationExtraData(pk symbioticCrypto.PrivateKey, kt symbiotic.KeyTag) ([]byte, error) {
	switch kt.Type() {
	case symbiotic.KeyTypeBlsBn254:
		blsKey, err := blsBn254.FromRaw(pk.PublicKey().Raw())
		if err != nil {
			return nil, errors.Errorf("failed to parse BLS public key: %w", err)
		}
		rawBytes := blsKey.G2().RawBytes()
		return rawBytes[:], nil
	case symbiotic.KeyTypeBls12381:
		return bls12381ExtraData(pk)
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return nil, nil
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("invalid key type")
	default:
		return nil, errors.New("invalid key type")
	}
}

func bls12381ExtraData(pk symbioticCrypto.PrivateKey) ([]byte, error) {
	blsKey, err := relaybls12381.FromRaw(pk.PublicKey().Raw())
	if err != nil {
		return nil, errors.Errorf("failed to parse BLS public key: %w", err)
	}
	return formatBls12381G2OnChain(blsKey.G2())
}

func formatBls12381G1OnChain(raw []byte) (symbiotic.CompactPublicKey, error) {
	if len(raw) != 96 {
		return nil, errors.Errorf("invalid bls12381 G1 length: %d", len(raw))
	}

	paddedPk := make([]byte, 128)
	copy(paddedPk[16:64], raw[0:48])
	copy(paddedPk[80:128], raw[48:96])
	return paddedPk, nil
}

func formatBls12381G2OnChain(point *bls12381.G2Affine) ([]byte, error) {
	encoded := make([]byte, 256)

	if err := putBls12381FieldElementOnChain(encoded[0:64], point.X.A0); err != nil {
		return nil, err
	}
	if err := putBls12381FieldElementOnChain(encoded[64:128], point.X.A1); err != nil {
		return nil, err
	}
	if err := putBls12381FieldElementOnChain(encoded[128:192], point.Y.A0); err != nil {
		return nil, err
	}
	if err := putBls12381FieldElementOnChain(encoded[192:256], point.Y.A1); err != nil {
		return nil, err
	}
	return encoded, nil
}

func putBls12381FieldElementOnChain(dst []byte, element bls12381fp.Element) error {
	if len(dst) != 64 {
		return errors.Errorf("invalid bls12381 destination length: %d", len(dst))
	}

	var fieldBytes [48]byte
	bls12381fp.BigEndian.PutElement(&fieldBytes, element)
	copy(dst[16:64], fieldBytes[:])
	return nil
}

func (a RegistrationArtifact) MarshalJSON() ([]byte, error) {
	payload := struct {
		KeyTag       uint8  `json:"keyTag"`
		KeyHex       string `json:"keyHex"`
		SignatureHex string `json:"signatureHex"`
		ExtraDataHex string `json:"extraDataHex"`
	}{
		KeyTag:       uint8(a.KeyTag),
		KeyHex:       hexutil.Encode(a.Key),
		SignatureHex: hexutil.Encode(a.Signature),
		ExtraDataHex: hexutil.Encode(a.ExtraData),
	}
	if payload.ExtraDataHex == "" {
		payload.ExtraDataHex = "0x"
	}
	return json.Marshal(payload)
}
