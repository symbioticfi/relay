package valsetDeriver

import (
	"context"
	"log/slog"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
)

const valsetVersion = 1

//go:generate mockgen -source=valset_deriver.go -destination=mocks/deriver.go -package=mocks -mock_names=evmClient=MockEvmClient
type evmClient interface {
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error)
	GetEip712Domain(ctx context.Context, addr entity.CrossChainAddress) (entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetSubnetwork(ctx context.Context) (common.Hash, error)
	GetNetworkAddress(ctx context.Context) (common.Address, error)
	GetHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (common.Hash, error)
	IsValsetHeaderCommittedAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (bool, error)
	GetPreviousHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
	GetHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error)
	GetOperators(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]common.Address, error)
}

// Deriver coordinates the ETH services
type Deriver struct {
	evmClient evmClient
}

// NewDeriver creates a new valset deriver
func NewDeriver(evmClient evmClient) (*Deriver, error) {
	return &Deriver{
		evmClient: evmClient,
	}, nil
}

func (v *Deriver) GetNetworkData(ctx context.Context, addr entity.CrossChainAddress) (entity.NetworkData, error) {
	address, err := v.evmClient.GetNetworkAddress(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get network address: %w", err)
	}

	subnetwork, err := v.evmClient.GetSubnetwork(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get subnetwork: %w", err)
	}

	eip712Data, err := v.evmClient.GetEip712Domain(ctx, addr)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get eip712 domain: %w", err)
	}

	return entity.NetworkData{
		Address:    address,
		Subnetwork: subnetwork,
		Eip712Data: eip712Data,
	}, nil
}

type dtoOperatorVotingPower struct {
	chainId      uint64
	votingPowers []entity.OperatorVotingPower
}

func (v *Deriver) GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error) {
	timestamp, err := v.evmClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to get epoch start timestamp: %w", err)
	}
	slog.DebugContext(ctx, "Got current valset timestamp", "timestamp", strconv.Itoa(int(timestamp)), "epoch", epoch)

	// Get voting powers from all voting power providers
	allVotingPowers := make([]dtoOperatorVotingPower, len(config.VotingPowerProviders))
	for i, provider := range config.VotingPowerProviders {
		votingPowers, err := v.evmClient.GetVotingPowers(ctx, provider, timestamp)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get voting powers from provider %s: %w", provider.Address.Hex(), err)
		}

		slog.DebugContext(ctx, "Got voting powers from provider", "provider", provider.Address.Hex(), "votingPowers", votingPowers)

		allVotingPowers[i] = dtoOperatorVotingPower{
			chainId:      provider.ChainId,
			votingPowers: votingPowers,
		}
	}

	// Get keys from the keys provider
	keys, err := v.evmClient.GetKeys(ctx, config.KeysProvider, timestamp)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to get keys: %w", err)
	}
	slog.DebugContext(ctx, "Got keys from provider", "provider", config.KeysProvider.Address.Hex(), "keys", keys)

	// form validators list from voting powers and keys using config
	validators, totalVP := v.formValidators(config, allVotingPowers, keys)

	// calc new quorum threshold
	quorumThreshold, err := v.calcQuorumThreshold(config, totalVP)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to calc quorum threshold: %w", err)
	}

	valset := entity.ValidatorSet{
		Version:          valsetVersion,
		RequiredKeyTag:   config.RequiredHeaderKeyTag,
		Epoch:            epoch,
		CaptureTimestamp: timestamp,
		QuorumThreshold:  quorumThreshold,
		Validators:       validators,
	}

	return valset, nil
}

func (v *Deriver) formValidators(
	config entity.NetworkConfig,
	votingPowers []dtoOperatorVotingPower,
	keys []entity.OperatorWithKeys,
) ([]entity.Validator, entity.VotingPower) {
	validators := v.fillValidators(votingPowers, keys)

	totalActiveVotingPower := v.fillValidatorsActive(config, validators)

	validators.SortByOperatorAddressAsc()

	return validators, entity.ToVotingPower(totalActiveVotingPower)
}

func (v *Deriver) fillValidatorsActive(config entity.NetworkConfig, validators entity.Validators) *big.Int {
	validators.SortByVotingPowerDesc()

	totalActive := 0
	totalActiveVotingPower := big.NewInt(0)

	for i := range validators {
		// Check minimum voting power if configured
		if validators[i].VotingPower.Cmp(config.MinInclusionVotingPower.Int) < 0 {
			break
		}

		// Check if validator has at least one key
		if len(validators[i].Keys) == 0 {
			continue
		}

		totalActive++
		validators[i].IsActive = true

		if config.MaxVotingPower.Int64() != 0 {
			if validators[i].VotingPower.Cmp(config.MaxVotingPower.Int) > 0 {
				validators[i].VotingPower = entity.ToVotingPower(new(big.Int).Set(config.MaxVotingPower.Int))
			}
		}
		// Add to total active voting power if validator is active
		totalActiveVotingPower = new(big.Int).Add(totalActiveVotingPower, validators[i].VotingPower.Int)

		if config.MaxValidatorsCount.Int64() != 0 {
			if totalActive >= int(config.MaxValidatorsCount.Int64()) {
				break
			}
		}
	}

	return totalActiveVotingPower
}

func (v *Deriver) fillValidators(votingPowers []dtoOperatorVotingPower, keys []entity.OperatorWithKeys) entity.Validators {
	// Create validators map to consolidate voting powers and keys
	validatorsMap := make(map[string]*entity.Validator)

	// Process voting powers
	for _, chainVp := range votingPowers {
		for _, vp := range chainVp.votingPowers {
			operatorAddr := vp.Operator.Hex()
			if _, exists := validatorsMap[operatorAddr]; !exists {
				validatorsMap[operatorAddr] = &entity.Validator{
					Operator:    vp.Operator,
					VotingPower: entity.ToVotingPower(big.NewInt(0)),
					IsActive:    false, // Default to active, will filter later
					Keys:        []entity.ValidatorKey{},
					Vaults:      []entity.ValidatorVault{},
				}
			}

			// Add vaults and their voting powers
			for _, vault := range vp.Vaults {
				validatorsMap[operatorAddr].VotingPower = entity.ToVotingPower(new(big.Int).Add(
					validatorsMap[operatorAddr].VotingPower.Int,
					vault.VotingPower.Int,
				))

				// Add vault to validator's vaults
				validatorsMap[operatorAddr].Vaults = append(validatorsMap[operatorAddr].Vaults, entity.ValidatorVault{
					Vault:       vault.Vault,
					VotingPower: vault.VotingPower,
					ChainID:     chainVp.chainId,
				})
			}

			// Sort vaults by address in ascending order
			sort.Slice(validatorsMap[operatorAddr].Vaults, func(i, j int) bool {
				// Compare vault addresses (lower first)
				return validatorsMap[operatorAddr].Vaults[i].Vault.Cmp(validatorsMap[operatorAddr].Vaults[j].Vault) < 0
			})
		}
	}

	// Process required keys
	for _, rk := range keys { // TODO: get required key tags from validator set config and fill with nils if needed
		operatorAddr := rk.Operator.Hex()
		if validator, exists := validatorsMap[operatorAddr]; exists {
			// Add all keys for this operator
			for _, key := range rk.Keys {
				validator.Keys = append(validator.Keys, entity.ValidatorKey{
					Tag:     key.Tag,
					Payload: key.Payload,
				})
			}
		}
	}

	validators := entity.Validators(lo.Map(lo.Values(validatorsMap), func(item *entity.Validator, _ int) entity.Validator {
		return *item
	}))

	return validators
}

func maxThreshold() *big.Int {
	// 10^18 is the maximum threshold value
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
}

func (v *Deriver) calcQuorumThreshold(config entity.NetworkConfig, totalVP entity.VotingPower) (entity.VotingPower, error) {
	quorumThresholdPercent := big.NewInt(0)
	for _, quorumThreshold := range config.QuorumThresholds {
		if quorumThreshold.KeyTag == config.RequiredHeaderKeyTag {
			quorumThresholdPercent = quorumThreshold.QuorumThreshold.Int
		}
	}
	if quorumThresholdPercent.Cmp(big.NewInt(0)) == 0 {
		return entity.VotingPower{}, errors.Errorf("quorum threshold is zero")
	}

	mul := new(big.Int).Mul(totalVP.Int, quorumThresholdPercent)
	div := new(big.Int).Div(mul, maxThreshold())
	// add 1 to apply up rounding
	return entity.ToVotingPower(new(big.Int).Add(div, big.NewInt(1))), nil
}
