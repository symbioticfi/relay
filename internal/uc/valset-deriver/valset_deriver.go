package valsetDeriver

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
	"middleware-offchain/pkg/ssz"
)

const valsetVersion = 1

//go:generate mockgen -source=deriver.go -destination=mocks/deriver.go -package=mocks
type ethClient interface {
	GetCaptureTimestamp(ctx context.Context) (*big.Int, error)
	GetEpochStart(ctx context.Context, epoch *big.Int) (*big.Int, error)
	GetCurrentValsetTimestamp(ctx context.Context) (*big.Int, error)
	GetConfig(ctx context.Context, timestamp *big.Int) (entity.Config, error)
	GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp *big.Int) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp *big.Int) ([]entity.OperatorWithKeys, error)
	GetRequiredKeyTag(ctx context.Context, timestamp *big.Int) (uint8, error)
	GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (*big.Int, error)
	GetSubnetwork(ctx context.Context) ([]byte, error)
}

// Deriver coordinates the ETH services
type Deriver struct {
	ethClient ethClient
}

// NewDeriver creates a new valset deriver
func NewDeriver(ethClient ethClient) (*Deriver, error) {
	return &Deriver{
		ethClient: ethClient,
	}, nil
}

func (v *Deriver) GetValidatorSetExtraForEpoch(ctx context.Context, epoch *big.Int) (entity.ValidatorSetExtra, error) {
	slog.DebugContext(ctx, "Trying to fetch current valset timestamp", "epoch", epoch.String())
	timestamp, err := v.ethClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get epoch start timestamp: %w", err)
	}
	slog.DebugContext(ctx, "Got current valset timestamp", "timestamp", timestamp.String(), "epoch", epoch.String())

	slog.DebugContext(ctx, "Trying to fetch config", "timestamp", timestamp.String())
	config, err := v.ethClient.GetConfig(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get config: %w", err)
	}
	slog.DebugContext(ctx, "Got config", "timestamp", timestamp.String(), "config", config)

	// Get voting powers from all voting power providers
	var allVotingPowers []entity.OperatorVotingPower
	for _, provider := range config.VotingPowerProviders {
		slog.DebugContext(ctx, "Trying to fetch voting powers from provider", "provider", provider.Address.Hex())
		votingPowers, err := v.ethClient.GetVotingPowers(ctx, provider, timestamp)
		if err != nil {
			return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get voting powers from provider %s: %w", provider.Address.Hex(), err)
		}

		slog.DebugContext(ctx, "Got voting powers from provider", "provider", provider.Address.Hex(), "votingPowers", votingPowers)

		allVotingPowers = append(allVotingPowers, votingPowers...)
	}

	// Get keys from the keys provider
	slog.DebugContext(ctx, "Trying to fetch keys from provider", "provider", config.KeysProvider.Address.Hex())

	keys, err := v.ethClient.GetKeys(ctx, config.KeysProvider, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get keys: %w", err)
	}

	requiredKeyTag, err := v.ethClient.GetRequiredKeyTag(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get required key tag: %w", err)
	}

	captureTimestamp, err := v.ethClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get capture timestamp: %w", err)
	}

	domainEip712, err := v.ethClient.GetEip712Domain(ctx)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get eip712 domain: %w", err)
	}

	subnetwork, err := v.ethClient.GetSubnetwork(ctx)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get subnetwork: %w", err)
	}

	return entity.ValidatorSetExtra{
		Version:              valsetVersion,
		RequiredKeyTag:       requiredKeyTag,
		Config:               config,
		CaptureTimestamp:     captureTimestamp,
		DomainEip712:         domainEip712,
		Keys:                 keys,
		Subnetwork:           subnetwork,
		OperatorVotingPowers: allVotingPowers,
		Epoch:                epoch,
	}, nil
}

func (v *Deriver) MakeValsetHeader(ctx context.Context, extra entity.ValidatorSetExtra) (entity.ValidatorSetHeader, error) {
	validatorSet := extra.MakeValidatorSet()

	tags := []uint8{uint8(len(validatorSet.Validators[0].Keys))}
	for _, key := range validatorSet.Validators[0].Keys {
		if key.Tag == extra.RequiredKeyTag { // TODO: major - get required key tags from validator set config
			tags = append(tags, key.Tag)
		}
	}

	// Create aggregated pubkeys for each required key tag
	aggPubkeysG1 := make([]*bls.G1, len(tags)) // TODO: minor - potentially not only BLS
	for i := range tags {
		if tags[i]>>4 == 0 {
			aggPubkeysG1[i] = &bls.G1{G1Affine: new(bn254.G1Affine)}
		}
	}

	for _, validator := range validatorSet.Validators {
		if !validator.IsActive {
			continue
		}

		for _, key := range validator.Keys {
			for i, tag := range tags {
				if key.Tag == tag {
					if tag>>4 == 0 {
						g1, err := bls.DeserializeG1(key.Payload)
						if err != nil {
							return entity.ValidatorSetHeader{}, fmt.Errorf("failed to deserialize G1: %w", err)
						}
						aggPubkeysG1[i] = aggPubkeysG1[i].Add(g1)
					}
				}
			}
		}
	}

	sszMroot, err := ssz.HashTreeRoot(validatorSet)
	if err != nil {
		return entity.ValidatorSetHeader{}, fmt.Errorf("failed to get hash tree root: %w", err)
	}

	// Format all aggregated keys for the header
	formattedKeys := make([]entity.Key, 0, len(aggPubkeysG1))
	for i, key := range aggPubkeysG1 {
		if key != nil && !key.IsInfinity() {
			formattedKeys = append(formattedKeys, entity.Key{
				Tag:     tags[i],
				Payload: bls.SerializeG1(key),
			})
		}
	}

	formattedHashesMimc := make([]entity.ValidatorSetHash, 0, len(formattedKeys))
	for _, key := range formattedKeys {
		validatorsData, err := proof.ToValidatorsData(validatorSet.Validators, validatorSet.Validators, key.Tag)
		if err != nil {
			return entity.ValidatorSetHeader{}, fmt.Errorf("failed to convert validators to data: %w", err)
		}
		hash := proof.HashValset(validatorsData)
		formattedHashesMimc = append(formattedHashesMimc, entity.ValidatorSetHash{
			KeyTag: key.Tag,
			Hash:   [32]byte(hash),
		})
	}

	formattedHashesKeccak256 := make([]entity.ValidatorSetHash, 0, len(formattedKeys)) // TODO: prettify/check
	type validatorDataTuple struct {
		X, Y, VotingPower *big.Int
	}
	u256, _ := abi.NewType("uint256", "", nil)

	tupleType := abi.Type{
		T:             abi.TupleTy,
		TupleElems:    []*abi.Type{&u256, &u256, &u256},
		TupleRawNames: []string{"X", "Y", "votingPower"},
		TupleType:     reflect.TypeOf(validatorDataTuple{}),
	}

	arrayType := abi.Type{
		T:    abi.SliceTy,
		Elem: &tupleType,
	}

	args := abi.Arguments{{Type: arrayType}}
	for _, key := range formattedKeys {
		validatorsData := make([]validatorDataTuple, 0, len(validatorSet.Validators))
		for _, validator := range validatorSet.Validators {
			validatorVotingPower := validator.VotingPower
			for _, validatorKey := range validator.Keys {
				if validatorKey.Tag == key.Tag {
					validatorKeyG1, err := bls.DeserializeG1(validatorKey.Payload)
					if err != nil {
						return entity.ValidatorSetHeader{}, fmt.Errorf("failed to deserialize G1: %w", err)
					}
					x := validatorKeyG1.X.BigInt(new(big.Int))
					y := validatorKeyG1.Y.BigInt(new(big.Int))

					votingPower := validatorVotingPower

					validatorsData = append(validatorsData, validatorDataTuple{
						X:           x,
						Y:           y,
						VotingPower: votingPower,
					})
				}
			}
		}

		packed, err := args.Pack(validatorsData)
		if err != nil {
			return entity.ValidatorSetHeader{}, fmt.Errorf("failed to pack arguments: %w", err)
		}
		hash := crypto.Keccak256Hash(packed)
		formattedHashesKeccak256 = append(formattedHashesKeccak256, entity.ValidatorSetHash{
			KeyTag: key.Tag,
			Hash:   hash,
		})
	}

	quorumThreshold := new(big.Int).Mul(validatorSet.TotalActiveVotingPower, entity.QuorumThresholdPercentage)
	quorumThreshold.Add(quorumThreshold, entity.QuorumThresholdBase)
	quorumThreshold.Sub(quorumThreshold, big.NewInt(1))
	quorumThreshold.Div(quorumThreshold, entity.QuorumThresholdBase)

	previousHeaderHash := [32]byte{} // TODO: get previous header hash from the previous header
	big.NewInt(0).FillBytes(previousHeaderHash[:])

	slog.DebugContext(ctx, "Generated validator set header", "formattedKeys", formattedKeys)

	return entity.ValidatorSetHeader{
		Version:                     validatorSet.Version,
		TotalActiveValidators:       new(big.Int).SetInt64(int64(len(proof.GetActiveValidators(validatorSet.Validators)))),
		ActiveAggregatedKeys:        formattedKeys,
		TotalActiveVotingPower:      validatorSet.TotalActiveVotingPower,
		ValidatorsSszMRoot:          sszMroot,
		Epoch:                       extra.Epoch,
		RequiredKeyTag:              extra.RequiredKeyTag,
		CaptureTimestamp:            extra.CaptureTimestamp,
		VerificationType:            extra.Config.VerificationType,
		ValidatorSetHashesMimc:      formattedHashesMimc,
		ValidatorSetHashesKeccak256: formattedHashesKeccak256,
		QuorumThreshold:             quorumThreshold,
		PreviousHeaderHash:          previousHeaderHash,
	}, nil
}

func (v *Deriver) MakeValidatorSetHeaderHash(ctx context.Context, valsetExtra entity.ValidatorSetExtra, extraData []entity.ExtraData) ([]byte, error) {
	header, err := v.MakeValsetHeader(ctx, valsetExtra)
	if err != nil {
		return nil, fmt.Errorf("failed to make valset header: %w", err)
	}

	headerHash, err := hashHeader(header, valsetExtra)
	if err != nil {
		return nil, fmt.Errorf("failed to hash valset header: %w", err)
	}

	extraDataHash, err := entity.ExtraDataList(extraData).Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash extra data: %w", err)
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
			},
			"ValSetHeaderCommit": []apitypes.Type{
				{Name: "subnetwork", Type: "bytes32"},
				{Name: "epoch", Type: "uint48"},
				{Name: "headerHash", Type: "bytes32"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:    valsetExtra.DomainEip712.Name,
			Version: valsetExtra.DomainEip712.Version,
		},
		PrimaryType: "ValSetHeaderCommit",
		Message: map[string]interface{}{
			"subnetwork": valsetExtra.Subnetwork,
			"epoch":      valsetExtra.Epoch,
			"headerHash": headerHash,
			"extraData":  extraDataHash,
		},
	}

	hashBytes, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to get typed data hash: %w", err)
	}

	return hashBytes, nil
}

func hashHeader(v entity.ValidatorSetHeader, extra entity.ValidatorSetExtra) ([]byte, error) {
	bytes, err := encodeHeader(v, extra)
	if err != nil {
		return nil, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return crypto.Keccak256(bytes), nil
}

func encodeHeader(v entity.ValidatorSetHeader, extra entity.ValidatorSetExtra) ([]byte, error) {
	arguments := abi.Arguments{
		{
			Name: "version",
			Type: abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Name: "requiredKeyTag",
			Type: abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Name: "epoch",
			Type: abi.Type{T: abi.UintTy, Size: 48},
		},
		{
			Name: "captureTimestamp",
			Type: abi.Type{T: abi.UintTy, Size: 48},
		},
		{
			Name: "verificationType",
			Type: abi.Type{T: abi.UintTy, Size: 32},
		},
		{
			Name: "quorumThreshold",
			Type: abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Name: "validatorsSszMRoot",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
		{
			Name: "previousHeaderHash",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
	}

	// Prepend the initial 32-byte offset (value 32 = 0x20)
	initialOffset := make([]byte, 32)
	offsetValue := big.NewInt(32)
	// FillBytes puts the big.Int's value into the byte slice, padded left with zeros
	offsetBytes := offsetValue.FillBytes(make([]byte, 32))
	copy(initialOffset, offsetBytes) // Copy the padded value into our prefix slice

	pack, err := arguments.Pack(v.Version, extra.RequiredKeyTag, extra.Epoch, extra.CaptureTimestamp, v.QuorumThreshold, v.ValidatorsSszMRoot, v.PreviousHeaderHash)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return append(initialOffset, pack...), err //nolint:makezero // intentionally appending to the initial offset
}

func (v *Deriver) GetExtraDataKey(verificationType uint32, name string) ([32]byte, error) {
	strTy, _ := abi.NewType("string", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)

	args := abi.Arguments{
		{Type: strTy},
		{Type: u32Ty},
		{Type: strTy},
	}

	packed, err := args.Pack(entity.ExtraDataGlobalKeyPrefix, verificationType, name)
	if err != nil {
		return [32]byte{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func (v *Deriver) GetExtraDataKeyTagged(verificationType uint32, keyTag uint8, name string) ([32]byte, error) {
	strTy, _ := abi.NewType("string", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)
	u8Ty, _ := abi.NewType("uint8", "", nil)

	args := abi.Arguments{
		{Type: strTy},
		{Type: u32Ty},
		{Type: strTy},
		{Type: u8Ty},
		{Type: strTy},
	}

	packed, err := args.Pack(entity.ExtraDataGlobalKeyPrefix, verificationType, entity.ExtraDataKeyTagPrefix, keyTag, name)
	if err != nil {
		return [32]byte{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func (v *Deriver) GetExtraDataKeyIndexed(
	verificationType uint32,
	keyTag uint8,
	name string,
	index *big.Int,
) ([32]byte, error) {
	baseHash, err := v.GetExtraDataKeyTagged(verificationType, keyTag, name)
	if err != nil {
		return [32]byte{}, err
	}

	sum := new(big.Int).Add(new(big.Int).SetBytes(baseHash[:]), index)
	var out [32]byte
	sum.FillBytes(out[:])
	return out, nil
}

func (v *Deriver) GenerateExtraData(ctx context.Context, valsetHeader entity.ValidatorSetHeader, verificationType uint32) ([]entity.ExtraData, error) {
	extraData := make([]entity.ExtraData, 0)

	switch verificationType {
	case entity.ZkVerificationType:
		{
			totalActiveValidatorsKey, err := v.GetExtraDataKey(verificationType, entity.ZkVerificationTotalActiveValidators)
			if err != nil {
				return nil, fmt.Errorf("failed to get extra data key: %w", err)
			}
			totalActiveValidatorsBytes32 := [32]byte{}
			valsetHeader.TotalActiveValidators.FillBytes(totalActiveValidatorsBytes32[:])
			extraData = append(extraData, entity.ExtraData{
				Key:   totalActiveValidatorsKey,
				Value: totalActiveValidatorsBytes32,
			})

			for _, validatorSetHash := range valsetHeader.ValidatorSetHashesMimc {
				validatorSetHashKey, err := v.GetExtraDataKeyTagged(verificationType, validatorSetHash.KeyTag, entity.ZkVerificationValidatorSetHashMimc)
				if err != nil {
					return nil, fmt.Errorf("failed to get extra data key: %w", err)
				}

				extraData = append(extraData, entity.ExtraData{
					Key:   validatorSetHashKey,
					Value: validatorSetHash.Hash,
				})
			}
		}
	case entity.SimpleVerificationType: // TODO: prettify/check
		totalActiveValidatorsKey, err := v.GetExtraDataKey(verificationType, entity.SimpleVerificationTotalVotingPower)
		if err != nil {
			return nil, fmt.Errorf("failed to get extra data key: %w", err)
		}
		totalActiveValidatorsBytes32 := [32]byte{}
		valsetHeader.TotalActiveValidators.FillBytes(totalActiveValidatorsBytes32[:])
		extraData = append(extraData, entity.ExtraData{
			Key:   totalActiveValidatorsKey,
			Value: totalActiveValidatorsBytes32,
		})

		for _, validatorSetHash := range valsetHeader.ValidatorSetHashesKeccak256 {
			validatorSetHashKey, err := v.GetExtraDataKeyTagged(verificationType, validatorSetHash.KeyTag, entity.SimpleVerificationValidatorSetHashKeccak256)
			if err != nil {
				return nil, fmt.Errorf("failed to get extra data key: %w", err)
			}

			extraData = append(extraData, entity.ExtraData{
				Key:   validatorSetHashKey,
				Value: validatorSetHash.Hash,
			})
		}

		for _, activeAggregatedKey := range valsetHeader.ActiveAggregatedKeys {
			activeAggregatedKeyKey, err := v.GetExtraDataKeyTagged(verificationType, activeAggregatedKey.Tag, entity.SimpleVerificationAggPublicKeyG1)
			if err != nil {
				return nil, fmt.Errorf("failed to get extra data key: %w", err)
			}
			keyG1Raw, err := bls.DeserializeG1(activeAggregatedKey.Payload)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize G1: %w", err)
			}

			x := keyG1Raw.X.BigInt(new(big.Int))
			y := keyG1Raw.Y.BigInt(new(big.Int))
			_, derivedY, err := bls.FindYFromX(x)
			if err != nil {
				return nil, fmt.Errorf("failed to find Y from X: %w", err)
			}

			flag := y.Cmp(derivedY) != 0
			compressedKeyG1 := new(big.Int).Mul(x, big.NewInt(2))
			if flag {
				compressedKeyG1.Add(compressedKeyG1, big.NewInt(1))
			}

			compressedKeyG1Bytes := [32]byte{}
			compressedKeyG1.FillBytes(compressedKeyG1Bytes[:])
			extraData = append(extraData, entity.ExtraData{
				Key:   activeAggregatedKeyKey,
				Value: compressedKeyG1Bytes,
			})
		}
	}

	return extraData, nil
}
