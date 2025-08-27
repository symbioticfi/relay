package blsBn254ZK

import (
	"math/big"

	"github.com/symbioticfi/relay/core/entity"
	types "github.com/symbioticfi/relay/core/usecase/aggregator/aggregator-types"
	"github.com/symbioticfi/relay/core/usecase/aggregator/helpers"
	"github.com/symbioticfi/relay/core/usecase/crypto/blsBn254"
	"github.com/symbioticfi/relay/pkg/proof"

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

	aggG1Sig := new(bn254.G1Affine)
	aggG2Key := new(bn254.G2Affine)
	signers := make(map[common.Address]bool)
	for _, sig := range signatures {
		pubKey, err := blsBn254.FromRaw(sig.PublicKey)
		if err != nil {
			return entity.AggregationProof{}, err
		}
		val, ok := valset.FindValidatorByKey(keyTag, pubKey.OnChain())
		if !ok {
			return entity.AggregationProof{}, errors.New("failed to find validator by key")
		}
		if !val.IsActive {
			// skip inactive validators
			continue
		}
		g1Sig := new(bn254.G1Affine)
		_, err = g1Sig.SetBytes(sig.Signature)
		if err != nil {
			return entity.AggregationProof{}, err
		}
		aggG1Sig = aggG1Sig.Add(aggG1Sig, g1Sig)
		aggG2Key = aggG2Key.Add(aggG2Key, pubKey.G2())
		signers[val.Operator] = true
	}

	var validatorsData []proof.ValidatorData
	for _, val := range valset.Validators {
		if val.IsActive {
			keyBytes, ok := val.FindKeyByKeyTag(keyTag)
			if !ok {
				return entity.AggregationProof{}, errors.New("failed to find key by keyTag")
			}
			_, isSigner := signers[val.Operator]
			g1Key := new(bn254.G1Affine)
			_, err := g1Key.SetBytes(keyBytes)
			if err != nil {
				return entity.AggregationProof{}, errors.Errorf("failed to deserialize G1 key: %w", err)
			}

			validatorsData = append(validatorsData, proof.ValidatorData{
				Key:         *g1Key,
				IsNonSigner: !isSigner,
				VotingPower: val.VotingPower.Int,
			})
		}
	}

	messageG1, err := blsBn254.HashToG1(messageHash)
	if err != nil {
		return entity.AggregationProof{}, err
	}
	messageG1Bn254 := bn254.G1Affine{X: messageG1.X, Y: messageG1.Y}

	proverInput := proof.ProveInput{
		ValidatorData:   proof.NormalizeValset(validatorsData),
		MessageG1:       messageG1Bn254,
		Signature:       *aggG1Sig,
		SignersAggKeyG2: *aggG2Key,
	}

	proofData, err := a.prover.Prove(proverInput)
	if err != nil {
		return entity.AggregationProof{}, err
	}

	return entity.AggregationProof{
		VerificationType: entity.VerificationTypeBlsBn254ZK,
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

	messageG1, err := blsBn254.HashToG1(aggregationProof.MessageHash)
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
		return false, errors.Errorf("agg voting power %s is less than quorum threshold %s", aggVotingPower.String(), valset.QuorumThreshold.String())
	}

	return ok, nil
}

func (a Aggregator) GenerateExtraData(valset entity.ValidatorSet, keyTags []entity.KeyTag) ([]entity.ExtraData, error) {
	extraData := make([]entity.ExtraData, 0)

	totalActiveValidatorsKey, err := helpers.GetExtraDataKey(entity.VerificationTypeBlsBn254ZK, entity.ZkVerificationTotalActiveValidatorsHash)
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
			return nil, errors.Errorf("failed to generate validator set MiMC accumulator: %w", err)
		}

		validatorSetHashKey, err := helpers.GetExtraDataKeyTagged(entity.VerificationTypeBlsBn254ZK, key.Tag, entity.ZkVerificationValidatorSetHashMimcHash)
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
				g1 := new(bn254.G1Affine)
				_, err := g1.SetBytes(key.Payload)
				if err != nil {
					return nil, errors.Errorf("failed to deserialize G1: %w", err)
				}
				validatorData := proof.ValidatorData{Key: *g1, VotingPower: activeValidators[i].VotingPower.Int, IsNonSigner: true}

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
