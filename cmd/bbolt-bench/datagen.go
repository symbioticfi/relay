package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type TestData struct {
	Keys               []crypto.PrivateKey
	Validators         symbiotic.Validators
	NetworkConfig      symbiotic.NetworkConfig
	TemplateSignatures []symbiotic.Signature
}

func GenerateTestData(numValidators int) (*TestData, error) {
	keys := make([]crypto.PrivateKey, numValidators)
	validators := make([]symbiotic.Validator, numValidators)

	for i := range numValidators {
		priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
		if err != nil {
			return nil, fmt.Errorf("generate key %d: %w", i, err)
		}
		keys[i] = priv

		opBytes := make([]byte, 20)
		big.NewInt(int64(i + 1)).FillBytes(opBytes)

		validators[i] = symbiotic.Validator{
			Operator:    common.BytesToAddress(opBytes),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{{
				Tag:     symbiotic.KeyTag(15),
				Payload: priv.PublicKey().OnChain(),
			}},
			Vaults: []symbiotic.ValidatorVault{{
				ChainID:     1,
				Vault:       common.BytesToAddress(opBytes),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			}},
		}
	}

	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Operator.Hex() < validators[j].Operator.Hex()
	})

	// Re-sort keys to match sorted validators
	keyMap := make(map[common.Address]crypto.PrivateKey)
	for _, v := range validators {
		for _, k := range keys {
			if common.Bytes2Hex(k.PublicKey().OnChain()) == common.Bytes2Hex(v.Keys[0].Payload) {
				keyMap[v.Operator] = k
				break
			}
		}
	}
	sortedKeys := make([]crypto.PrivateKey, numValidators)
	for i, v := range validators {
		sortedKeys[i] = keyMap[v.Operator]
	}

	// Sign one template message to get valid signature bytes
	templateMessage := randomBytes32()
	templateSigs := make([]symbiotic.Signature, numValidators)
	for i, priv := range sortedKeys {
		rawSig, messageHash, err := priv.Sign(templateMessage)
		if err != nil {
			return nil, fmt.Errorf("sign template with key %d: %w", i, err)
		}
		templateSigs[i] = symbiotic.Signature{
			KeyTag:      symbiotic.KeyTag(15),
			Epoch:       0,
			MessageHash: messageHash,
			Signature:   rawSig,
			PublicKey:   priv.PublicKey(),
		}
	}

	networkCfg := symbiotic.NetworkConfig{
		VotingPowerProviders:    []symbiotic.CrossChainAddress{{ChainId: 1, Address: common.HexToAddress("0x1")}},
		KeysProvider:            symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x2")},
		Settlements:             []symbiotic.CrossChainAddress{{ChainId: 1, Address: common.HexToAddress("0x3")}},
		VerificationType:        1,
		MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(100000)),
		MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(1)),
		MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(1000)),
		RequiredKeyTags:         []symbiotic.KeyTag{15},
		RequiredHeaderKeyTag:    15,
		QuorumThresholds:        []symbiotic.QuorumThreshold{{KeyTag: 15, QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(50))}},
		EpochDuration:           3600,
		NumAggregators:          1,
		NumCommitters:           1,
		CommitterSlotDuration:   60,
	}

	return &TestData{
		Keys:               sortedKeys,
		Validators:         validators,
		NetworkConfig:      networkCfg,
		TemplateSignatures: templateSigs,
	}, nil
}

// MakeSignatures copies template signatures with substituted epoch and messageHash.
// The raw signature bytes are reused (crypto-invalid for new messageHash, but
// the repository does not verify signatures on save).
func (td *TestData) MakeSignatures(epoch symbiotic.Epoch, messageHash []byte) []symbiotic.Signature {
	sigs := make([]symbiotic.Signature, len(td.TemplateSignatures))
	for i, tmpl := range td.TemplateSignatures {
		sigs[i] = symbiotic.Signature{
			KeyTag:      tmpl.KeyTag,
			Epoch:       epoch,
			MessageHash: messageHash,
			Signature:   tmpl.Signature,
			PublicKey:   tmpl.PublicKey,
		}
	}
	return sigs
}

func (td *TestData) MakeValidatorSet(epoch symbiotic.Epoch) symbiotic.ValidatorSet {
	return symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1000000,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(500)),
		Validators:       td.Validators,
		Status:           symbiotic.HeaderDerived,
	}
}

func (td *TestData) MakeNextValsetData(epoch symbiotic.Epoch) entity.NextValsetData {
	prevEpoch := epoch - 1
	if epoch == 0 {
		prevEpoch = 0
	}

	requestID := common.BytesToHash(randomBytes32())

	return entity.NextValsetData{
		PrevValidatorSet:  td.MakeValidatorSet(prevEpoch),
		PrevNetworkConfig: td.NetworkConfig,
		NextValidatorSet:  td.MakeValidatorSet(epoch),
		NextNetworkConfig: td.NetworkConfig,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{
			RequestID:      requestID,
			Epoch:          epoch,
			CommitmentData: randomBytes32(),
		},
	}
}

func randomBytes32() []byte {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	return buf
}
