package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/big"
	"offchain-middleware/bls"
	"time"
)

// NetworkService coordinates the P2P and ETH services
type NetworkService struct {
	storage    *Storage
	p2pService *P2PService
	ethClient  *EthClient
	keyPair    *bls.KeyPair
}

// ValidatorSetHeaderInput represents the input for validator set header
type ValidatorSetHeaderInput struct {
	ActiveAggregatedKeys   []Key
	TotalActiveVotingPower *big.Int
	ValidatorsSszMRoot     [32]byte
	ExtraData              []byte
}

// NewNetworkService creates a new network service
func NewNetworkService(p2pService *P2PService, ethClient *EthClient, storage *Storage) (*NetworkService, error) {
	return &NetworkService{
		p2pService: p2pService,
		ethClient:  ethClient,
		storage:    storage,
	}, nil
}

// Start begins all service operations
func (n *NetworkService) Start(interval time.Duration) error {
	go func() {
		for {
			time.Sleep(interval)
			if err := n.signValidatorSet(context.Background()); err != nil {
				log.Printf("failed to sign validator set: %v", err)
			}
		}
	}()

	return nil
}

func (n *NetworkService) signValidatorSet(ctx context.Context) error {
	epoch, err := n.ethClient.getCurrentEpoch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current epoch: %w", err)
	}

	validatorSet, err := n.ethClient.getValidatorSet(ctx, epoch)
	if err != nil {
		return fmt.Errorf("failed to get validator set: %w", err)
	}

	requiredKeyTag, err := n.ethClient.getRequiredKeyTag(ctx)
	if err != nil {
		return fmt.Errorf("failed to get required key tag: %w", err)
	}

	aggPubkeyG1 := new(bls.G1)

	for _, validator := range validatorSet.Validators {
		if !validator.IsActive {
			continue
		}

		for _, key := range validator.Keys {
			if key.Tag == requiredKeyTag {
				aggPubkeyG1 = aggPubkeyG1.Add(bls.DeserializeG1(key.Payload))
			}
		}
	}

	return nil
}

func (n *NetworkService) checkIsValidator(validatorSet *ValidatorSet) bool {
	for _, validator := range validatorSet.Validators {
		for _, key := range validator.Keys {
			if key.Tag == 0 && bytes.Equal(key.Payload, n.keyPair.PublicKeyG1.Marshal()) {
				return true
			}
		}
	}

	return false
}
