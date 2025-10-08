package symbiotic

import (
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

type EVMClient = evm.Client
type Aggregator = aggregator.Aggregator
type ValsetDeriver = valsetDeriver.Deriver
