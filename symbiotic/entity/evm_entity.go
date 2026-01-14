package entity

type BlockNumber string

const (
	BlockNumberFinalized BlockNumber = "finalized"
	BlockNumberLatest    BlockNumber = "latest"
)

type EVMOptions struct {
	BlockNumber        BlockNumber
	GasLimitMultiplier float64
}

func AppliedEVMOptions(opts ...EVMOption) *EVMOptions {
	options := &EVMOptions{
		BlockNumber:        BlockNumberFinalized,
		GasLimitMultiplier: .0,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

type EVMOption func(options *EVMOptions)

func WithEVMBlockNumber(blockNumber BlockNumber) EVMOption {
	return func(o *EVMOptions) {
		o.BlockNumber = blockNumber
	}
}

func WithGasLimitMultiplier(multiplier float64) EVMOption {
	return func(o *EVMOptions) {
		o.GasLimitMultiplier = multiplier
	}
}
