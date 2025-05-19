package valset

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
	"middleware-offchain/pkg/ssz"
)

//go:generate mockgen -source=generator.go -destination=mocks/generator.go -package=mocks
type deriver interface {
	GetValidatorSet(ctx context.Context, timestamp *big.Int) (entity.ValidatorSet, error)
}

// Generator handles the generation of validator set headers
type Generator struct {
	deriver   deriver
	ethClient ethClient
}

// NewGenerator creates a new validator set generator
func NewGenerator(deriver deriver, ethClient ethClient) (*Generator, error) {
	return &Generator{
		deriver:   deriver,
		ethClient: ethClient,
	}, nil
}

// GenerateValidatorSetHeader generates a validator set header for the current epoch
func (v *Generator) GenerateValidatorSetHeader(ctx context.Context) (entity.ValidatorSetHeader, error) {
	slog.DebugContext(ctx, "Generating validator set header")

	slog.DebugContext(ctx, "Trying to get capture timestamp")
	timestamp, err := v.ethClient.GetCaptureTimestamp(ctx)
	if err != nil {
		return entity.ValidatorSetHeader{}, fmt.Errorf("failed to get capture timestamp: %w", err)
	}
	slog.DebugContext(ctx, "Got capture timestamp", "timestamp", timestamp.String())

	validatorSet, err := v.deriver.GetValidatorSet(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetHeader{}, fmt.Errorf("failed to get validator set: %w", err)
	}

	requiredKeyTag, err := v.ethClient.GetRequiredKeyTag(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSetHeader{}, fmt.Errorf("failed to get required key tag: %w", err)
	}

	slog.DebugContext(ctx, "Got validator set", "validatorSet", validatorSet)

	tags := []uint8{uint8(len(validatorSet.Validators[0].Keys))}
	for _, key := range validatorSet.Validators[0].Keys {
		if key.Tag == requiredKeyTag { // TODO: major - get required key tags from validator set config
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
	valset, err := proof.ToValidatorsData(validatorSet.Validators, requiredKeyTag)
	if err != nil {
		return entity.ValidatorSetHeader{}, fmt.Errorf("failed to convert validators to data: %w", err)
	}
	extraData := proof.HashValset(&valset)

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

func (v *Generator) GenerateValidatorSetHeaderHash(ctx context.Context, validatorSetHeader entity.ValidatorSetHeader) ([]byte, error) {
	hash, err := validatorSetHeader.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash validator set header: %w", err)
	}

	domainEip712, err := v.ethClient.GetEip712Domain(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get eip712 domain: %w", err)
	}

	currentEpoch, err := v.ethClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current epoch: %w", err)
	}

	subnetwork, err := v.ethClient.GetSubnetwork(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnetwork: %w", err)
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
			Name:    domainEip712.Name,
			Version: domainEip712.Version,
		},
		PrimaryType: "ValSetHeaderCommit",
		Message: map[string]interface{}{
			"subnetwork": subnetwork,
			"epoch":      currentEpoch,
			"headerHash": hash,
		},
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("error hashing domain: %w", err)
	}

	typeHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("error hashing message: %w", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typeHash)))

	return crypto.Keccak256(rawData), nil
}
