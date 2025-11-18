package entity

type BlockNumber string

const (
	BlockNumberFinalized BlockNumber = "finalized"
	BlockNumberLatest    BlockNumber = "latest"
)

type EVMOptions struct {
	BlockNumber BlockNumber
}

func AppliedEVMOptions(opts ...EVMOption) *EVMOptions {
	options := &EVMOptions{
		BlockNumber: BlockNumberFinalized,
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
