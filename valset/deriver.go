package valset

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math/big"
	"slices"
	"sort"

	"github.com/ethereum/go-ethereum/common"

	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/types"
)

const VALSET_VERSION = 1

type ethClient interface {
	GetCaptureTimestamp(ctx context.Context) (*big.Int, error)
	GetMasterConfig(ctx context.Context, timestamp *big.Int) (entity.MasterConfig, error)
	GetValSetConfig(ctx context.Context, timestamp *big.Int) (entity.ValSetConfig, error)
	GetVotingPowers(ctx context.Context, address common.Address, timestamp *big.Int) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address common.Address, timestamp *big.Int) ([]entity.OperatorWithKeys, error)
	GetRequiredKeyTag(ctx context.Context, timestamp *big.Int) (uint8, error)
	GetEip712Domain(ctx context.Context) (*entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (*big.Int, error)
	GetSubnetwork(ctx context.Context) ([]byte, error)
}

// ValsetDeriver coordinates the ETH services
type ValsetDeriver struct {
	ethClient ethClient
}

// NewValsetDeriver creates a new valset deriver
func NewValsetDeriver(ethClient ethClient) (*ValsetDeriver, error) {
	return &ValsetDeriver{
		ethClient: ethClient,
	}, nil
}

func (v ValsetDeriver) GetValidatorSet(ctx context.Context, timestamp *big.Int) (*types.ValidatorSet, error) {
	slog.DebugContext(ctx, "Trying to fetch master config", "timestamp", timestamp.String())
	masterConfig, err := v.ethClient.GetMasterConfig(ctx, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get master config: %w", err)
	}
	slog.DebugContext(ctx, "Got master config", "timestamp", timestamp.String(), "config", masterConfig)

	slog.DebugContext(ctx, "Trying to getch val set config", "timestamp", timestamp.String())
	valSetConfig, err := v.ethClient.GetValSetConfig(ctx, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get val set config: %w", err)
	}
	slog.DebugContext(ctx, "Got val set config", "timestamp", timestamp.String(), "config", valSetConfig)

	// Get voting powers from all voting power providers
	var allVotingPowers []entity.OperatorVotingPower
	for _, provider := range masterConfig.VotingPowerProviders {
		slog.DebugContext(ctx, "Trying to fetch voting powers from provider", "provider", provider.Address.Hex())
		votingPowers, err := v.ethClient.GetVotingPowers(ctx, provider.Address, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to get voting powers from provider %s: %w", provider.Address.Hex(), err)
		}
		allVotingPowers = append(allVotingPowers, votingPowers...)
		slog.DebugContext(ctx, "Got voting powers from provider", "provider", provider.Address.Hex(), "votingPowers", votingPowers)
	}

	// Get keys from the keys provider
	slog.DebugContext(ctx, "Trying to fetch keys from provider", "provider", masterConfig.KeysProvider.Address.Hex())
	keys, err := v.ethClient.GetKeys(ctx, masterConfig.KeysProvider.Address, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}
	slog.DebugContext(ctx, "Got keys from provider", "provider", masterConfig.KeysProvider.Address.Hex(), "keys", keys)

	// Create validators map to consolidate voting powers and keys
	validatorsMap := make(map[string]*types.Validator)

	// Process voting powers
	for _, vp := range allVotingPowers {
		operatorAddr := vp.Operator.Hex()
		if _, exists := validatorsMap[operatorAddr]; !exists {
			validatorsMap[operatorAddr] = &types.Validator{
				Operator:    vp.Operator,
				VotingPower: big.NewInt(0),
				IsActive:    false, // Default to active, will filter later
				Keys:        []*types.Key{},
				Vaults:      []*types.Vault{},
			}
		}

		// Add vaults and their voting powers
		for _, vault := range vp.Vaults {
			validatorsMap[operatorAddr].VotingPower = new(big.Int).Add(
				validatorsMap[operatorAddr].VotingPower,
				vault.VotingPower,
			)

			// Add vault to validator's vaults
			validatorsMap[operatorAddr].Vaults = append(validatorsMap[operatorAddr].Vaults, &types.Vault{
				Vault:       vault.Vault,
				VotingPower: vault.VotingPower,
			})
		}
	}

	// Process required keys
	for _, rk := range keys { // TODO: get required key tags from validator set config and fill with nills if needed
		operatorAddr := rk.Operator.Hex()
		if validator, exists := validatorsMap[operatorAddr]; exists {
			// Add all keys for this operator
			for _, key := range rk.Keys {
				validator.Keys = append(validator.Keys, &types.Key{
					Tag:     key.Tag,
					Payload: key.Payload,
				})
			}
		}
	}

	validators := slices.Collect(maps.Values(validatorsMap))
	// Sort validators by voting power in descending order
	sort.Slice(validators, func(i, j int) bool {
		// Compare voting powers (higher first)
		return validators[i].VotingPower.Cmp(validators[j].VotingPower) > 0
	})

	// Apply filters from valSetConfig
	totalActiveVotingPower := big.NewInt(0)
	totalActive := 0

	for i := range validators {
		// Check minimum voting power if configured
		if validators[i].VotingPower.Cmp(valSetConfig.MinInclusionVotingPower) < 0 {
			break
		}

		// Check if validator has at least one key
		if len(validators[i].Keys) == 0 {
			continue
		}

		totalActive++
		validators[i].IsActive = true

		if valSetConfig.MaxVotingPower.Int64() != 0 {
			if validators[i].VotingPower.Cmp(valSetConfig.MaxVotingPower) > 0 {
				validators[i].VotingPower = new(big.Int).Set(valSetConfig.MaxVotingPower)
			}
		}
		// Add to total active voting power if validator is active
		totalActiveVotingPower = new(big.Int).Add(totalActiveVotingPower, validators[i].VotingPower)

		if valSetConfig.MaxValidatorsCount.Int64() != 0 {
			if totalActive >= int(valSetConfig.MaxValidatorsCount.Int64()) {
				break
			}
		}
	}

	// Create the validator set
	validatorSet := types.ValidatorSet{
		Version:                VALSET_VERSION,
		TotalActiveVotingPower: totalActiveVotingPower,
		Validators:             validators,
	}

	return &validatorSet, nil
}
