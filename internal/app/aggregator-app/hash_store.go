package aggregator_app

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"middleware-offchain/bls"
	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/types"
)

type hashStore struct {
	mu sync.Mutex
	m  map[string]map[common.Address]hashWithValidator
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

func (h *hashStore) PutHash(msg entity.SignatureHashMessage, val types.Validator) (*big.Int, *bls.G1, *bls.G1, *bls.G2, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	validators, ok := h.m[string(msg.MessageHash)]
	if !ok {
		validators = make(map[common.Address]hashWithValidator)
		h.m[string(msg.MessageHash)] = validators
	}
	if _, ok = validators[val.Operator]; ok {
		return nil, nil, nil, nil, errors.Errorf("hash already exists for validator %s", val.Operator.Hex())
	}
	validators[val.Operator] = hashWithValidator{
		validator: val,
		hash:      msg,
	}

	totalVotingPower := new(big.Int)
	aggSignature := bls.ZeroG1()
	aggPublicKeyG1 := bls.ZeroG1()
	aggPublicKeyG2 := bls.ZeroG2()
	for _, validator := range validators {
		totalVotingPower = totalVotingPower.Add(totalVotingPower, validator.validator.VotingPower)
		signature, err := bls.DeserializeG1(validator.hash.Signature)
		if err != nil {
			return nil, nil, nil, nil, errors.Errorf("failed to deserialize signature: %w", err)
		}
		publicKeyG1, err := bls.DeserializeG1(validator.hash.PublicKeyG1)
		if err != nil {
			return nil, nil, nil, nil, errors.Errorf("failed to deserialize public key G1: %w", err)
		}
		publicKeyG2, err := bls.DeserializeG2(validator.hash.PublicKeyG2)
		if err != nil {
			return nil, nil, nil, nil, errors.Errorf("failed to deserialize public key G2: %w", err)
		}
		aggSignature = aggSignature.Add(signature)
		aggPublicKeyG1 = aggPublicKeyG1.Add(publicKeyG1)
		aggPublicKeyG2 = aggPublicKeyG2.Add(publicKeyG2)
	}

	return totalVotingPower, aggSignature, aggPublicKeyG1, aggPublicKeyG2, nil
}
