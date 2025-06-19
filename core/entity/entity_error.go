package entity

type StringError string

func (e StringError) Error() string {
	return string(e)
}

const (
	ErrEntityNotFound                 = StringError("entity not found")
	ErrEntityAlreadyExist             = StringError("entity already exists")
	ErrNotAnAggregator                = StringError("not an aggregator")
	ErrChainNotFound                  = StringError("chain not found")
	ErrValsetAlreadyCommittedForEpoch = StringError("valset already committed for epoch")
)
