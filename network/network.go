package network

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"offchain-middleware/bls"
	"offchain-middleware/eth"
	"offchain-middleware/p2p"
	"offchain-middleware/proof"
	"offchain-middleware/storage"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/karalabe/ssz"
)

// NetworkService coordinates the P2P and ETH services
type NetworkService struct {
	storage    *storage.Storage
	p2pService *p2p.P2PService
	ethClient  *eth.EthClient
	keyPair    *bls.KeyPair
}

// NewNetworkService creates a new network service
func NewNetworkService(p2pService *p2p.P2PService, ethClient *eth.EthClient, storage *storage.Storage, keyPair *bls.KeyPair) (*NetworkService, error) {
	return &NetworkService{
		p2pService: p2pService,
		ethClient:  ethClient,
		storage:    storage,
		keyPair:    keyPair,
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
	epoch, err := n.ethClient.GetCurrentEpoch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current epoch: %w", err)
	}

	validatorSet, err := n.ethClient.GetValidatorSet(ctx, epoch)
	if err != nil {
		return fmt.Errorf("failed to get validator set: %w", err)
	}

	requiredKeyTag, err := n.ethClient.GetRequiredKeyTag(ctx)
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

	sszMroot := ssz.HashSequential(&validatorSet)

	valset := proof.ToValidatorsData(validatorSet.Validators, requiredKeyTag)
	extraData := proof.HashValset(&valset)

	validatorSetHeader := ValidatorSetHeader{
		ActiveAggregatedKeys:   []bls.G1{*aggPubkeyG1},
		TotalActiveVotingPower: validatorSet.TotalActiveVotingPower,
		ValidatorsSszMRoot:     sszMroot,
		ExtraData:              extraData,
	}

	validatorSetHeaderBytes, err := validatorSetHeader.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode validator set header: %w", err)
	}

	validatorSetHeaderHash := crypto.Keccak256(validatorSetHeaderBytes)

	signature, err := n.keyPair.Sign(validatorSetHeaderHash)
	if err != nil {
		return fmt.Errorf("failed to sign validator set header: %w", err)
	}

	validatorSetHeaderHashString := hex.EncodeToString(validatorSetHeaderHash)

	n.storage.AddSignature(epoch, validatorSetHeaderHashString, storage.Signature{
		Signature: signature.Marshal(),
		PublicKey: n.keyPair.PublicKeyG1.Marshal(),
	})

	return nil
}

func (n *NetworkService) checkIsValidator(validatorSet *eth.ValidatorSet) bool {
	for _, validator := range validatorSet.Validators {
		for _, key := range validator.Keys {
			if key.Tag == 0 && bytes.Equal(key.Payload, n.keyPair.PublicKeyG1.Marshal()) {
				return true
			}
		}
	}

	return false
}
