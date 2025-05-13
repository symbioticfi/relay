package aggregator_app

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
)

type hashStore struct {
	mu sync.Mutex
	m  map[string]map[common.Address]hashWithValidator
}

type hashWithValidator struct {
	validator entity.Validator
	hash      entity.SignatureHashMessage
}

func newHashStore() *hashStore {
	return &hashStore{
		m: make(map[string]map[common.Address]hashWithValidator),
	}
}

type currentValues struct {
	votingPower    *big.Int
	aggSignature   *bls.G1
	aggPublicKeyG1 *bls.G1
	aggPublicKeyG2 *bls.G2
	validators     []entity.Validator
}

func (h *hashStore) PutHash(msg entity.SignatureHashMessage, val entity.Validator) (currentValues, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	validators, ok := h.m[string(msg.MessageHash)]
	if !ok {
		validators = make(map[common.Address]hashWithValidator)
		h.m[string(msg.MessageHash)] = validators
	}
	if _, ok = validators[val.Operator]; ok {
		return currentValues{}, errors.Errorf("hash already exists for validator %s", val.Operator.Hex())
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
			return currentValues{}, errors.Errorf("failed to deserialize signature: %w", err)
		}
		publicKeyG1, err := bls.DeserializeG1(validator.hash.PublicKeyG1)
		if err != nil {
			return currentValues{}, errors.Errorf("failed to deserialize public key G1: %w", err)
		}
		publicKeyG2, err := bls.DeserializeG2(validator.hash.PublicKeyG2)
		if err != nil {
			return currentValues{}, errors.Errorf("failed to deserialize public key G2: %w", err)
		}
		aggSignature = aggSignature.Add(signature)
		aggPublicKeyG1 = aggPublicKeyG1.Add(publicKeyG1)
		aggPublicKeyG2 = aggPublicKeyG2.Add(publicKeyG2)
	}

	return currentValues{
		votingPower:    totalVotingPower,
		aggSignature:   aggSignature,
		aggPublicKeyG1: aggPublicKeyG1,
		aggPublicKeyG2: aggPublicKeyG2,
		validators: lo.Map(lo.Values(validators), func(v hashWithValidator, _ int) entity.Validator {
			return v.validator
		}),
	}, nil
}
