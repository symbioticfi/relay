package aggregator

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
)

type Aggregator struct {
	zkProver *proof.ZkProver
}

func NewAggregator(prover *proof.ZkProver) *Aggregator {
	return &Aggregator{
		zkProver: prover,
	}
}

func (a *Aggregator) Aggregate(
	valset *entity.ValidatorSet,
	keyTag entity.KeyTag,
	verificationType entity.VerificationType,
	messageHash []byte,
	signatures []entity.Signature,
) (*entity.AggregationProof, error) {
	if !compareMessageHasher(signatures) {
		return nil, errors.New("message hashes mismatch")
	}

	switch verificationType {
	case entity.VerificationTypeZK:
		return a.zkAggregate(valset, keyTag, messageHash, signatures)
	case entity.VerificationTypeSimple:
		return a.simpleAggregate(valset, keyTag, messageHash, signatures)
	}
	return nil, errors.New("unknown verification type")
}

func compareMessageHasher(signatures []entity.Signature) bool {
	for i := range signatures {
		if i > 0 {
			if !bytes.Equal(signatures[i-1].MessageHash, signatures[i].MessageHash) {
				return false
			}
		}
	}
	return true
}

func (a *Aggregator) Verify(
	valset *entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof *entity.AggregationProof,
) (bool, error) {
	switch aggregationProof.VerificationType {
	case entity.VerificationTypeZK:
		return a.zkVerify(valset, keyTag, aggregationProof)
	case entity.VerificationTypeSimple:
		return a.simpleVerify(valset, keyTag, aggregationProof)
	}
	return false, errors.New("unknown verification type")
}

func (a *Aggregator) zkAggregate(
	valset *entity.ValidatorSet,
	keyTag entity.KeyTag,
	messageHash []byte,
	signatures []entity.Signature,
) (*entity.AggregationProof, error) {
	aggG1Sig := bls.ZeroG1()
	aggG2Key := bls.ZeroG2()
	signers := make(map[common.Address]bool)
	for _, sig := range signatures {
		g1, g2Key, err := bls.UnpackPublicG1G2(sig.PublicKey)
		if err != nil {
			return nil, err
		}
		val, ok := valset.FindValidatorByKey(keyTag, g1.Marshal())
		if !ok {
			return nil, errors.New("failed to find validator by key")
		}
		g1Sig, err := bls.DeserializeG1(sig.Signature)
		if err != nil {
			return nil, err
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
				return nil, errors.New("failed to find key by keyTag")
			}
			_, isSinger := signers[val.Operator]
			g1Key, err := bls.DeserializeG1(keyBytes)
			if err != nil {
				return nil, errors.Errorf("failed to deserialize G1 key: %w", err)
			}

			validatorsData = append(validatorsData, proof.ValidatorData{
				Key:         *g1Key.G1Affine,
				IsNonSigner: !isSinger,
				VotingPower: val.VotingPower,
			})
		}
	}

	proverInput := proof.ProveInput{
		ValidatorData:   proof.NormalizeValset(validatorsData),
		Message:         messageHash,
		Signature:       *aggG1Sig.G1Affine,
		SignersAggKeyG2: *aggG2Key.G2Affine,
	}
	proofData, err := a.zkProver.Prove(proverInput)
	if err != nil {
		return nil, err
	}
	return &entity.AggregationProof{
		VerificationType: entity.VerificationTypeZK,
		MessageHash:      messageHash,
		Proof:            proofData.Marshall(),
	}, nil
}

func (a *Aggregator) simpleAggregate(
	valset *entity.ValidatorSet,
	keyTag entity.KeyTag,
	messageHash []byte,
	signatures []entity.Signature,
) (*entity.AggregationProof, error) {
	return &entity.AggregationProof{}, nil
}

func (a *Aggregator) zkVerify(
	valset *entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof *entity.AggregationProof,
) (bool, error) {
	activeVals := 0
	for _, val := range valset.Validators {
		if val.IsActive {
			activeVals++
		}
	}

	mimcAccum, err := proof.ValidatorSetMimcAccumulator(valset.Validators, keyTag)
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

	ok, err := a.zkProver.Verify(activeVals, inpHash, aggregationProof.Proof)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (a *Aggregator) simpleVerify(
	valset *entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof *entity.AggregationProof,
) (bool, error) {
	return true, nil
}
