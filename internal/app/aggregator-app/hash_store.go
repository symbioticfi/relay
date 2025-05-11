package signer_app

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/types"
)

type hashStore struct {
	m map[string]map[common.Address]hashWithValidator
	// todo ilya add mutex
}

type hashWithValidator struct {
	validator types.Validator
	hash      entity.SignatureHashMessage
}

func newHashStore() *hashStore {
	return &hashStore{
		m: make(map[string]map[common.Address]hashWithValidator),
	}
}

func (h *hashStore) PutHash(msg entity.SignatureHashMessage, val types.Validator) (*big.Int, error) {
	validators, ok := h.m[string(msg.MessageHash)]
	if !ok {
		validators = make(map[common.Address]hashWithValidator)
		h.m[string(msg.MessageHash)] = validators
	}
	if _, ok := validators[val.Operator]; ok {
		return nil, errors.Errorf("hash already exists for validator %s", val.Operator.Hex())
	}
	validators[val.Operator] = hashWithValidator{
		validator: val,
		hash:      msg,
	}

	totalVotingPower := new(big.Int)
	for _, validator := range validators {
		totalVotingPower = totalVotingPower.Add(totalVotingPower, validator.validator.VotingPower)
	}

	return totalVotingPower, nil
}
