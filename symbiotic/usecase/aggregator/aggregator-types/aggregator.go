package aggregator_types

import (
	"context"

	"github.com/symbioticfi/relay/pkg/proof"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/ethereum/go-ethereum/common"
)

type Aggregator interface {
	Aggregate(ctx context.Context, valset symbiotic.ValidatorSet, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error)
	Verify(ctx context.Context, valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, aggregationProof symbiotic.AggregationProof) (bool, error)
	GenerateExtraData(ctx context.Context, valset symbiotic.ValidatorSet, keyTags []symbiotic.KeyTag) ([]symbiotic.ExtraData, error)
}

type Prover interface {
	Prove(ctx context.Context, proveInput proof.ProveInput) (proof.ProofData, error)
	Verify(ctx context.Context, valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error)
}
