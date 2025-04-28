package valset

import (
	"context"
	"fmt"
	"log"
	"offchain-middleware/bls"
	"offchain-middleware/eth"
	"offchain-middleware/proof"
	"offchain-middleware/valset/types"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/karalabe/ssz"
)

// ValsetGenerator handles the generation of validator set headers
type ValsetGenerator struct {
	deriver   *ValsetDeriver
	ethClient eth.IEthClient
}

// NewValsetGenerator creates a new validator set generator
func NewValsetGenerator(deriver *ValsetDeriver, ethClient eth.IEthClient) (*ValsetGenerator, error) {
	return &ValsetGenerator{
		deriver:   deriver,
		ethClient: ethClient,
	}, nil
}

// GenerateValidatorSetHeader generates a validator set header for the current epoch
func (v ValsetGenerator) GenerateValidatorSetHeader(ctx context.Context) (*types.ValidatorSetHeader, error) {
	log.Println("Generating validator set header")

	timestamp, err := v.ethClient.GetCaptureTimestamp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get capture timestamp: %w", err)
	}

	validatorSet, err := v.deriver.GetValidatorSet(ctx, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator set: %w", err)
	}

	requiredKeyTag, err := v.ethClient.GetRequiredKeyTag(ctx, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get required key tag: %w", err)
	}

	log.Println("Processing validator set")

	tags := []uint8{uint8(len(validatorSet.Validators[0].Keys))}
	for _, key := range validatorSet.Validators[0].Keys {
		if key.Tag == requiredKeyTag {
			tags = append(tags, key.Tag)
		}
	}

	// Create aggregated pubkeys for each required key tag
	aggPubkeysG1 := make([]*bls.G1, len(tags))
	for i := range tags {
		aggPubkeysG1[i] = &bls.G1{G1Affine: new(bn254.G1Affine)}
	}

	for _, validator := range validatorSet.Validators {
		if !validator.IsActive {
			continue
		}

		for _, key := range validator.Keys {
			for i, tag := range tags {
				if key.Tag == tag {
					aggPubkeysG1[i] = aggPubkeysG1[i].Add(bls.DeserializeG1(key.Payload))
				}
			}
		}
	}

	sszMroot := ssz.HashSequential(validatorSet)

	// Use the first key tag for proof generation
	valset := proof.ToValidatorsData(validatorSet.Validators, requiredKeyTag)
	extraData := proof.HashValset(&valset)

	// Format all aggregated keys for the header
	formattedKeys := make([]types.G1, len(aggPubkeysG1))
	for i, key := range aggPubkeysG1 {
		formattedKeys[i] = types.FormatG1(key)
	}

	validatorSetHeader := &types.ValidatorSetHeader{
		ActiveAggregatedKeys:   formattedKeys,
		TotalActiveVotingPower: validatorSet.TotalActiveVotingPower,
		ValidatorsSszMRoot:     sszMroot,
		ExtraData:              extraData,
	}

	log.Println("Generated validator set header")

	return validatorSetHeader, nil
}
