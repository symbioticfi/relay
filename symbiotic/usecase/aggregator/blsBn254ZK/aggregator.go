package blsBn254ZK

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/symbioticfi/relay/pkg/proof"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	types "github.com/symbioticfi/relay/symbiotic/usecase/aggregator/aggregator-types"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator/helpers"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type Aggregator struct {
	prover types.Prover
}

func NewAggregator(prover types.Prover) (*Aggregator, error) {
	return &Aggregator{
		prover: prover,
	}, nil
}

func (a Aggregator) Aggregate(
	valset symbiotic.ValidatorSet,
	keyTag symbiotic.KeyTag,
	messageHash []byte,
	signatures []symbiotic.SignatureExtended,
) (symbiotic.AggregationProof, error) {
	if !helpers.CompareMessageHasher(signatures, messageHash) {
		return symbiotic.AggregationProof{}, errors.New("message hashes mismatch")
	}

	aggG1Sig := new(bn254.G1Affine)
	aggG2Key := new(bn254.G2Affine)
	signers := make(map[common.Address]bool)
	valKeysToIdx := helpers.GetValidatorsIndexesMapByKey(valset, keyTag)

	for _, sig := range signatures {
		pubKey, err := blsBn254.FromRaw(sig.PublicKey)
		if err != nil {
			return symbiotic.AggregationProof{}, err
		}
		idx, ok := valKeysToIdx[string(pubKey.OnChain())]
		if !ok {
			return symbiotic.AggregationProof{}, errors.New("failed to find validator by key")
		}
		val := valset.Validators[idx]
		if !val.IsActive {
			continue
		}
		g1Sig := new(bn254.G1Affine)
		_, err = g1Sig.SetBytes(sig.Signature)
		if err != nil {
			return symbiotic.AggregationProof{}, err
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
				return symbiotic.AggregationProof{}, errors.New("failed to find key by keyTag")
			}
			_, isSigner := signers[val.Operator]
			g1Key := new(bn254.G1Affine)
			_, err := g1Key.SetBytes(keyBytes)
			if err != nil {
				return symbiotic.AggregationProof{}, errors.Errorf("failed to deserialize G1 key: %w", err)
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
		return symbiotic.AggregationProof{}, err
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
		return symbiotic.AggregationProof{}, err
	}

	return symbiotic.AggregationProof{
		MessageHash: messageHash,
		KeyTag:      keyTag,
		Epoch:       valset.Epoch,
		Proof:       proofData.Marshal(),
	}, nil
}

func (a Aggregator) Verify(
	valset symbiotic.ValidatorSet,
	keyTag symbiotic.KeyTag,
	aggregationProof symbiotic.AggregationProof,
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

func (a Aggregator) GenerateExtraData(valset symbiotic.ValidatorSet, keyTags []symbiotic.KeyTag) ([]symbiotic.ExtraData, error) {
	extraData := make([]symbiotic.ExtraData, 0)

	totalActiveValidatorsKey, err := helpers.GetExtraDataKey(symbiotic.VerificationTypeBlsBn254ZK, symbiotic.ZkVerificationTotalActiveValidatorsHash)
	if err != nil {
		return nil, errors.Errorf("failed to get extra data key: %w", err)
	}

	totalActiveValidators := big.NewInt(valset.GetTotalActiveValidators())
	totalActiveValidatorsBytes32 := common.Hash{}
	totalActiveValidators.FillBytes(totalActiveValidatorsBytes32[:])
	extraData = append(extraData, symbiotic.ExtraData{
		Key:   totalActiveValidatorsKey,
		Value: totalActiveValidatorsBytes32,
	})

	aggregatedPubKeys := helpers.GetAggregatedPubKeys(valset, keyTags)

	for _, key := range aggregatedPubKeys {
		mimcAccumulator, err := validatorSetMimcAccumulator(valset.Validators, key.Tag)
		if err != nil {
			return nil, errors.Errorf("failed to generate validator set MiMC accumulator: %w", err)
		}

		validatorSetHashKey, err := helpers.GetExtraDataKeyTagged(symbiotic.VerificationTypeBlsBn254ZK, key.Tag, symbiotic.ZkVerificationValidatorSetHashMimcHash)
		if err != nil {
			return nil, errors.Errorf("failed to get extra data key: %w", err)
		}

		extraData = append(extraData, symbiotic.ExtraData{
			Key:   validatorSetHashKey,
			Value: mimcAccumulator,
		})
	}

	// sort extra data by key to ensure deterministic order
	sort.Slice(extraData, func(i, j int) bool {
		return bytes.Compare(extraData[i].Key[:], extraData[j].Key[:]) < 0
	})

	return extraData, nil
}

func validatorSetMimcAccumulator(valset []symbiotic.Validator, requiredKeyTag symbiotic.KeyTag) (common.Hash, error) {
	validatorsData, err := toValidatorsData([]symbiotic.Validator{}, valset, requiredKeyTag)
	if err != nil {
		return common.Hash{}, err
	}
	return common.Hash(proof.HashValset(validatorsData)), nil
}

func toValidatorsData(signerValidators []symbiotic.Validator, allValidators symbiotic.Validators, requiredKeyTag symbiotic.KeyTag) ([]proof.ValidatorData, error) {
	activeValidators := allValidators.GetActiveValidators()
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
