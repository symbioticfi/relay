package aggregator_app

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/core/entity"
)

type hashStore struct {
	mu sync.Mutex
	m  map[string]map[common.Address]hashWithValidator
}

type hashWithValidator struct {
	validator        entity.Validator
	signatureMessage entity.Signature
}

func newHashStore() *hashStore {
	return &hashStore{
		m: make(map[string]map[common.Address]hashWithValidator),
	}
}

type currentValues struct {
	votingPower *big.Int
	validators  []entity.Validator
}

func (h *hashStore) PutHash(msg entity.Signature, val entity.Validator) (currentValues, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	validators, ok := h.m[string(msg.MessageHash)]
	if !ok {
		validators = make(map[common.Address]hashWithValidator)
		h.m[string(msg.MessageHash)] = validators
	}
	if _, ok = validators[val.Operator]; ok {
		return currentValues{}, errors.Errorf("signature already exists for validator %s", val.Operator.Hex())
	}

	validators[val.Operator] = hashWithValidator{
		validator:        val,
		signatureMessage: msg,
	}

	totalVotingPower := new(big.Int)
	for _, validator := range validators {
		totalVotingPower = totalVotingPower.Add(totalVotingPower, validator.validator.VotingPower)
	}

	return currentValues{
		votingPower: totalVotingPower,
		validators: lo.Map(lo.Values(validators), func(v hashWithValidator, _ int) entity.Validator {
			return v.validator
		}),
	}, nil
}
