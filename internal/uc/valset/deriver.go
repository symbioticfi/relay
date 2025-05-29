package valset

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
	GetMasterConfig(ctx context.Context, timestamp *big.Int) (entity.MasterConfig, error)
	GetValSetConfig(ctx context.Context, timestamp *big.Int) (entity.ValSetConfig, error)
	GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp *big.Int) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp *big.Int) ([]entity.OperatorWithKeys, error)
	GetRequiredKeyTag(ctx context.Context, timestamp *big.Int) (uint8, error)
	GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (*big.Int, error)
	GetSubnetwork(ctx context.Context) ([]byte, error)
}

type repo interface {
	GetLatestValsetExtra(ctx context.Context) (entity.ValidatorSetExtra, error)
}

// Deriver coordinates the ETH services
type Deriver struct {
	ethClient ethClient
	repo      repo
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

	slog.DebugContext(ctx, "Trying to fetch master config", "timestamp", timestamp.String())
	masterConfig, err := v.ethClient.GetMasterConfig(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get master config: %w", err)
	}
	slog.DebugContext(ctx, "Got master config", "timestamp", timestamp.String(), "config", masterConfig)

	slog.DebugContext(ctx, "Trying to getch val set config", "timestamp", timestamp.String())
	valSetConfig, err := v.ethClient.GetValSetConfig(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get val set config: %w", err)
	}
	slog.DebugContext(ctx, "Got val set config", "timestamp", timestamp.String(), "config", valSetConfig)

	// Get voting powers from all voting power providers
	var allVotingPowers []entity.OperatorVotingPower
	for _, provider := range masterConfig.VotingPowerProviders {
		slog.DebugContext(ctx, "Trying to fetch voting powers from provider", "provider", provider.Address.Hex())
		votingPowers, err := v.ethClient.GetVotingPowers(ctx, provider, timestamp)
		if err != nil {
			return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get voting powers from provider %s: %w", provider.Address.Hex(), err)
		}

		slog.DebugContext(ctx, "Got voting powers from provider", "provider", provider.Address.Hex(), "votingPowers", votingPowers)

		allVotingPowers = append(allVotingPowers, votingPowers...)
	}

	// Get keys from the keys provider
	slog.DebugContext(ctx, "Trying to fetch keys from provider", "provider", masterConfig.KeysProvider.Address.Hex())

	keys, err := v.ethClient.GetKeys(ctx, masterConfig.KeysProvider, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get keys: %w", err)
	}

	requiredKeyTag, err := v.ethClient.GetRequiredKeyTag(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetExtra{}, fmt.Errorf("failed to get required key tag: %w", err)
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
		MasterConfig:         masterConfig,
		ValSetConfig:         valSetConfig,
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

	// Use the first key tag for proof generation
	valsetExtra, err := proof.ToValidatorsData(validatorSet.Validators, validatorSet.Validators, extra.RequiredKeyTag)
	if err != nil {
		return entity.ValidatorSetHeader{}, fmt.Errorf("failed to convert validators to data: %w", err)
	}
	extraData := proof.HashValset(valsetExtra)

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

	slog.DebugContext(ctx, "Generated validator set header", "formattedKeys", formattedKeys)

	return entity.ValidatorSetHeader{
		Version:                validatorSet.Version,
		ActiveAggregatedKeys:   formattedKeys,
		TotalActiveVotingPower: validatorSet.TotalActiveVotingPower,
		ValidatorsSszMRoot:     sszMroot,
		ExtraData:              extraData,
	}, nil
}

func (v *Deriver) MakeValidatorSetHeaderHash(ctx context.Context, extra entity.ValidatorSetExtra) ([]byte, error) {
	header, err := v.MakeValsetHeader(ctx, extra)
	if err != nil {
		return nil, fmt.Errorf("failed to make valset header: %w", err)
	}

	hash, err := hashHeader(header)
	if err != nil {
		return nil, fmt.Errorf("failed to hash valset header: %w", err)
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
			Name:    extra.DomainEip712.Name,
			Version: extra.DomainEip712.Version,
		},
		PrimaryType: "ValSetHeaderCommit",
		Message: map[string]interface{}{
			"subnetwork": extra.Subnetwork,
			"epoch":      extra.Epoch,
			"headerHash": hash,
		},
	}

	hashBytes, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to get typed data hash: %w", err)
	}

	return hashBytes, nil
}

func hashHeader(v entity.ValidatorSetHeader) ([]byte, error) {
	bytes, err := encodeHeader(v)
	if err != nil {
		return nil, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return crypto.Keccak256(bytes), nil
}

func encodeHeader(v entity.ValidatorSetHeader) ([]byte, error) {
	arguments := abi.Arguments{
		{
			Name: "version",
			Type: abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Name: "activeAggregatedKeys",
			Type: abi.Type{
				T: abi.SliceTy,
				Elem: &abi.Type{
					T: abi.TupleTy,
					TupleElems: []*abi.Type{
						{T: abi.UintTy, Size: 8},
						{T: abi.BytesTy},
					},
					TupleRawNames: []string{"tag", "payload"},
					TupleType:     reflect.TypeOf(entity.Key{}),
				},
			},
		},
		{
			Name: "totalActiveVotingPower",
			Type: abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Name: "validatorsSszMRoot",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
		{
			Name: "extraData",
			Type: abi.Type{T: abi.BytesTy},
		},
	}

	// Prepend the initial 32-byte offset (value 32 = 0x20)
	initialOffset := make([]byte, 32)
	offsetValue := big.NewInt(32)
	// FillBytes puts the big.Int's value into the byte slice, padded left with zeros
	offsetBytes := offsetValue.FillBytes(make([]byte, 32))
	copy(initialOffset, offsetBytes) // Copy the padded value into our prefix slice

	pack, err := arguments.Pack(v.Version, v.ActiveAggregatedKeys, v.TotalActiveVotingPower, v.ValidatorsSszMRoot, v.ExtraData)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return append(initialOffset, pack...), err //nolint:makezero // intentionally appending to the initial offset
}
