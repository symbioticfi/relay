package zk

import (
	"fmt"
	"math/big"
	"middleware-offchain/core/entity"
	types "middleware-offchain/core/usecase/aggregator/aggregator-types"
	"middleware-offchain/core/usecase/aggregator/helpers"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type Aggregator struct {
	prover types.Prover
}

func NewAggregator(prover types.Prover) *Aggregator {
	return &Aggregator{
		prover: prover,
	}
}

func (a Aggregator) Aggregate(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	messageHash []byte,
	signatures []entity.SignatureExtended,
) (entity.AggregationProof, error) {
	if !helpers.CompareMessageHasher(signatures, messageHash) {
		return entity.AggregationProof{}, errors.New("message hashes mismatch")
	}

	aggG1Sig := bls.ZeroG1()
	aggG2Key := bls.ZeroG2()
	signers := make(map[common.Address]bool)
	for _, sig := range signatures {
		g1, g2Key, err := bls.UnpackPublicG1G2(sig.PublicKey)
		if err != nil {
			return entity.AggregationProof{}, err
		}
		val, ok := valset.FindValidatorByKey(keyTag, g1.Marshal())
		if !ok {
			return entity.AggregationProof{}, errors.New("failed to find validator by key")
		}
		g1Sig, err := bls.DeserializeG1(sig.Signature)
		if err != nil {
			return entity.AggregationProof{}, err
		}
		aggG1Sig = aggG1Sig.Add(g1Sig)
		aggG2Key = aggG2Key.Add(&g2Key)
		signers[val.Operator] = true
	}

	var validatorsData []proof.ValidatorData
	for _, val := range valset.Validators {
		if val.IsActive {
			keyBytes, ok := val.FindKeyByKeyTag(keyTag)
			if !ok {
				return entity.AggregationProof{}, errors.New("failed to find key by keyTag")
			}
			_, isSinger := signers[val.Operator]
			g1Key, err := bls.DeserializeG1(keyBytes)
			if err != nil {
				return entity.AggregationProof{}, errors.Errorf("failed to deserialize G1 key: %w", err)
			}

			validatorsData = append(validatorsData, proof.ValidatorData{
				Key:         *g1Key.G1Affine,
				IsNonSigner: !isSinger,
				VotingPower: val.VotingPower.Int,
			})
		}
	}

	messageG1, err := bls.HashToG1(messageHash)
	if err != nil {
		return entity.AggregationProof{}, err
	}
	messageG1Bn254 := bn254.G1Affine{X: messageG1.X, Y: messageG1.Y}

	proverInput := proof.ProveInput{
		ValidatorData:   proof.NormalizeValset(validatorsData),
		MessageG1:       messageG1Bn254,
		Signature:       *aggG1Sig.G1Affine,
		SignersAggKeyG2: *aggG2Key.G2Affine,
	}

	proofData, err := a.prover.Prove(proverInput)
	if err != nil {
		return entity.AggregationProof{}, err
	}

	return entity.AggregationProof{
		VerificationType: entity.VerificationTypeZK,
		MessageHash:      messageHash,
		Proof:            proofData.Marshal(),
	}, nil
}

func (a Aggregator) Verify(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof entity.AggregationProof,
) (bool, error) {
	activeVals := 0
	for _, val := range valset.Validators {
		if val.IsActive {
			activeVals++
		}
	}

	mimcAccum, err := validatorSetMimcAccumulator(valset.Validators, keyTag)
	if err != nil {
		return false, err
	}
	// last 32 bytes is aggVotingPowerBytes
	aggVotingPowerBytes := aggregationProof.Proof[len(aggregationProof.Proof)-32:]

	messageG1, err := bls.HashToG1(aggregationProof.MessageHash)
	if err != nil {
		return false, errors.Errorf("failed to hash message to G1: %w", err)
	}
	messageG1Bytes := messageG1.RawBytes() // non compressed

	inpBytes := mimcAccum[:]
	inpBytes = append(inpBytes, aggVotingPowerBytes...)
	inpBytes = append(inpBytes, messageG1Bytes[:]...)
	inpHash := crypto.Keccak256Hash(inpBytes)

	ok, err := a.prover.Verify(activeVals, inpHash, aggregationProof.Proof)
	if err != nil {
		return false, err
	}

	aggVotingPower := new(big.Int).SetBytes(aggVotingPowerBytes)
	if aggVotingPower.Cmp(valset.QuorumThreshold.Int) < 0 {
		return false, fmt.Errorf("agg voting power %s is less than quorum threshold %s", aggVotingPower.String(), valset.QuorumThreshold.String())
	}

	return ok, nil
}

func (a Aggregator) GenerateExtraData(valset entity.ValidatorSet, keyTags []entity.KeyTag) ([]entity.ExtraData, error) {
	extraData := make([]entity.ExtraData, 0)

	totalActiveValidatorsKey, err := helpers.GetExtraDataKey(entity.VerificationTypeZK, entity.ZkVerificationTotalActiveValidatorsHash)
	if err != nil {
		return nil, errors.Errorf("failed to get extra data key: %w", err)
	}

	totalActiveValidators := big.NewInt(valset.GetTotalActiveValidators())
	totalActiveValidatorsBytes32 := common.Hash{}
	totalActiveValidators.FillBytes(totalActiveValidatorsBytes32[:])
	extraData = append(extraData, entity.ExtraData{
		Key:   totalActiveValidatorsKey,
		Value: totalActiveValidatorsBytes32,
	})

	aggregatedPubKeys := helpers.GetAggregatedPubKeys(valset, keyTags)

	for _, key := range aggregatedPubKeys {
		mimcAccumulator, err := validatorSetMimcAccumulator(valset.Validators, key.Tag)
		if err != nil {
			return nil, errors.Errorf("failed to generate validator set mimc accumulator: %w", err)
		}

		validatorSetHashKey, err := helpers.GetExtraDataKeyTagged(entity.VerificationTypeZK, key.Tag, entity.ZkVerificationValidatorSetHashMimcHash)
		if err != nil {
			return nil, errors.Errorf("failed to get extra data key: %w", err)
		}

		extraData = append(extraData, entity.ExtraData{
			Key:   validatorSetHashKey,
			Value: mimcAccumulator,
		})
	}

	return extraData, nil
}

func validatorSetMimcAccumulator(valset []entity.Validator, requiredKeyTag entity.KeyTag) (common.Hash, error) {
	validatorsData, err := toValidatorsData([]entity.Validator{}, valset, requiredKeyTag)
	if err != nil {
		return common.Hash{}, err
	}
	return common.Hash(proof.HashValset(validatorsData)), nil
}

func toValidatorsData(signerValidators []entity.Validator, allValidators []entity.Validator, requiredKeyTag entity.KeyTag) ([]proof.ValidatorData, error) {
	activeValidators := entity.GetActiveValidators(allValidators)
	valset := make([]proof.ValidatorData, 0)
	for i := range activeValidators {
		for _, key := range activeValidators[i].Keys {
			if key.Tag == requiredKeyTag {
				g1, err := bls.DeserializeG1(key.Payload)
				if err != nil {
					return nil, errors.Errorf("failed to deserialize G1: %w", err)
				}
				validatorData := proof.ValidatorData{Key: *g1.G1Affine, VotingPower: activeValidators[i].VotingPower.Int, IsNonSigner: true}

				for _, signer := range signerValidators {
					if signer.Operator.Cmp(activeValidators[i].Operator) == 0 {
						validatorData.IsNonSigner = false
						break
					}
				}

				valset = append(valset, validatorData)
				break
			}
		}
	}
	return proof.NormalizeValset(valset), nil
}
