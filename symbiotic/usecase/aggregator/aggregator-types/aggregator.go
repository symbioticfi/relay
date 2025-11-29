package aggregator_types

import (
	"context"

	"github.com/symbioticfi/relay/pkg/proof"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/ethereum/go-ethereum/common"
)

type Aggregator interface {
	Aggregate(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, messageHash []byte, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error)
	Verify(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, aggregationProof symbiotic.AggregationProof) (bool, error)
	GenerateExtraData(ctx context.Context, valset symbiotic.ValidatorSet, keyTags []symbiotic.KeyTag) ([]symbiotic.ExtraData, error)
}

type Prover interface {
	Prove(proveInput proof.ProveInput) (proof.ProofData, error)
	Verify(valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error)
}
