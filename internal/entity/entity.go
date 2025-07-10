package entity

import (
	"math/big"

	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

type AggregationStatus struct {
	VotingPower *big.Int
	Validators  []entity.Validator
}

type CMDCrossChainAddress struct {
	ChainID uint64 `mapstructure:"chain-id" validate:"required"`
	Address string `mapstructure:"address" validate:"required"`
}

func NewChainsRpcURl(chainsId []uint64, chainsUrl []string) ([]entity.ChainURL, error) {
	if len(chainsId) != len(chainsUrl) {
		return nil, errors.New("chains id and chains rpc url length do not match")
	}

	chains := make([]entity.ChainURL, len(chainsId))
	for i := range chains {
		chains[i] = entity.ChainURL{
			ChainID: chainsId[i],
			RPCURL:  chainsUrl[i],
		}
	}

	return chains, nil
}
