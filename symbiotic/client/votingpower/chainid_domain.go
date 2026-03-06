package votingpower

const (
	ExternalVotingPowerChainIDMin uint64 = 4_000_000_000
	ExternalVotingPowerChainIDMax uint64 = 4_100_000_000
)

func IsExternalVotingPowerChainID(chainID uint64) bool {
	return chainID >= ExternalVotingPowerChainIDMin && chainID <= ExternalVotingPowerChainIDMax
}
