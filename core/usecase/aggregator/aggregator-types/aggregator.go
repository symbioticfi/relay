package aggregator_types

import (
	"middleware-offchain/core/entity"
	"middleware-offchain/pkg/proof"

	"github.com/ethereum/go-ethereum/common"
)

type Aggregator interface {
	Aggregate(valset entity.ValidatorSet, keyTag entity.KeyTag, messageHash []byte, signatures []entity.SignatureExtended) (entity.AggregationProof, error)
	Verify(valset entity.ValidatorSet, keyTag entity.KeyTag, aggregationProof entity.AggregationProof) (bool, error)
	GenerateExtraData(valset entity.ValidatorSet, keyTags []entity.KeyTag) ([]entity.ExtraData, error)
}

type Prover interface {
	Prove(proveInput proof.ProveInput) (proof.ProofData, error)
	Verify(valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error)
}
