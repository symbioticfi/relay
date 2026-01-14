package key_registerer

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/bls12381"
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
	currentOnchainEpoch, err := r.evmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get current epoch: %w", err)
	}

	captureTimestamp, err := r.evmClient.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get capture timestamp: %w", err)
	}

	networkConfig, err := r.evmClient.GetConfig(ctx, captureTimestamp, currentOnchainEpoch)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get config: %w", err)
	}

	eip712Domain, err := r.evmClient.GetEip712Domain(ctx, networkConfig.KeysProvider)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get eip712 domain: %w", err)
	}

	key := pk.PublicKey().OnChain()

	commitmentData, err := keyCommitmentData(eip712Domain, operatorAddress, key)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get commitment data: %w", err)
	}

	signature, _, err := pk.Sign(commitmentData)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to sign commitment data: %w", err)
	}

	// For ECDSA signatures, we need to adjust the recovery ID
	// Go's crypto.Sign() returns V as 0 or 1, but Ethereum expects 27 or 28
	if kt.Type() == symbiotic.KeyTypeEcdsaSecp256k1 && len(signature) == 65 {
		// Convert recovery ID from 0/1 to 27/28 for Ethereum
		signature[64] += 27
	}

	var extraData []byte
	switch kt.Type() {
	case symbiotic.KeyTypeBlsBn254:
		blsKey, err := blsBn254.FromRaw(pk.PublicKey().Raw())
		if err != nil {
			return symbiotic.TxResult{}, errors.Errorf("failed to parse BLS public key: %w", err)
		}
		rawByte := blsKey.G2().RawBytes()
		extraData = rawByte[:]
	case symbiotic.KeyTypeBls12381:
		blsKey, err := bls12381.FromRaw(pk.PublicKey().Raw())
		if err != nil {
			return symbiotic.TxResult{}, errors.Errorf("failed to parse BLS public key: %w", err)
		}
		rawByte := blsKey.G2().RawBytes()
		extraData = rawByte[:]
	case symbiotic.KeyTypeEcdsaSecp256k1:
		// no extra data needed for ECDSA keys
	case symbiotic.KeyTypeInvalid:
		return symbiotic.TxResult{}, errors.New("invalid key type")
	}

	// Use the adjusted signature for registration
	txResult, err := r.evmClient.RegisterKey(ctx, networkConfig.KeysProvider, kt, key, signature, extraData)
	if err != nil {
		return txResult, errors.Errorf("failed to register key: %w", err)
	}

	return txResult, nil
}

func keyCommitmentData(eip712Domain symbiotic.Eip712Domain, operator common.Address, keyBytes []byte) ([]byte, error) {
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
