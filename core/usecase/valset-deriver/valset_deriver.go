package valsetDeriver

import (
	"bytes"
	"context"
	"log/slog"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/core/entity"
)

const valsetVersion = 1

//go:generate mockgen -source=valset_deriver.go -destination=mocks/deriver.go -package=mocks
type ethClient interface {
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error)
	GetEip712Domain(ctx context.Context, addr entity.CrossChainAddress) (entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetSubnetwork(ctx context.Context) (common.Hash, error)
	GetNetworkAddress(ctx context.Context) (*common.Address, error)
	GetHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (common.Hash, error)
	IsValsetHeaderCommittedAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (bool, error)
	GetPreviousHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
	GetHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error)
}

// Deriver coordinates the ETH services
type Deriver struct {
	ethClient ethClient
}

// NewDeriver creates a new valset deriver
func NewDeriver(ethClient ethClient) (*Deriver, error) {
	return &Deriver{
		ethClient: ethClient,
	}, nil
}

func (v *Deriver) GetNetworkData(ctx context.Context, addr entity.CrossChainAddress) (entity.NetworkData, error) {
	address, err := v.ethClient.GetNetworkAddress(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get network address: %w", err)
	}

	subnetwork, err := v.ethClient.GetSubnetwork(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get subnetwork: %w", err)
	}

	eip712Data, err := v.ethClient.GetEip712Domain(ctx, addr)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get eip712 domain: %w", err)
	}

	return entity.NetworkData{
		Address:    *address,
		Subnetwork: subnetwork,
		Eip712Data: eip712Data,
	}, nil
}

type dtoOperatorVotingPower struct {
	chainId      uint64
	votingPowers []entity.OperatorVotingPower
}

func (v *Deriver) GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error) {
	timestamp, err := v.ethClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to get epoch start timestamp: %w", err)
	}
	slog.DebugContext(ctx, "Got current valset timestamp", "timestamp", strconv.Itoa(int(timestamp)), "epoch", epoch)

	// Get voting powers from all voting power providers
	allVotingPowers := make([]dtoOperatorVotingPower, len(config.VotingPowerProviders))
	for i, provider := range config.VotingPowerProviders {
		votingPowers, err := v.ethClient.GetVotingPowers(ctx, provider, timestamp)
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
	keys, err := v.ethClient.GetKeys(ctx, config.KeysProvider, timestamp)
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

	committedAddr, isValsetCommitted, err := v.isValsetHeaderCommitted(ctx, config, epoch)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to check if validator committed at epoch %d: %w", epoch, err)
	}

	valset := entity.ValidatorSet{
		Version:          valsetVersion,
		RequiredKeyTag:   config.RequiredHeaderKeyTag,
		Epoch:            epoch,
		CaptureTimestamp: timestamp,
		QuorumThreshold:  quorumThreshold,
		Validators:       validators,
	}

	if isValsetCommitted {
		previousHeaderHash, err := v.ethClient.GetPreviousHeaderHashAt(ctx, committedAddr, epoch)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get previous header hash: %w", err)
		}
		valset.PreviousHeaderHash = previousHeaderHash

		// valset integrity check
		committedHash, err := v.ethClient.GetHeaderHashAt(ctx, committedAddr, epoch)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get header hash: %w", err)
		}
		valsetHeader, err := valset.GetHeader()
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get header hash: %w", err)
		}
		calculatedHash, err := valsetHeader.Hash()
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get header hash: %w", err)
		}

		if !bytes.Equal(committedHash[:], calculatedHash[:]) {
			slog.DebugContext(ctx, "Validator set integrity check failed", "committed hash", committedHash, "calculated hash", calculatedHash)
			return entity.ValidatorSet{}, errors.Errorf("validator set hash mistmach at epoch %d", epoch)
		}
		slog.DebugContext(ctx, "Validator set integrity check passed", "hash", committedHash)

		valset.Status = entity.HeaderCommitted
	} else {
		lastCommittedAddr, latestCommittedEpoch, err := v.getLastCommittedHeaderEpoch(ctx, config)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get current valset epoch: %w", err)
		}

		if epoch < latestCommittedEpoch {
			slog.DebugContext(ctx, "Header is not committed [missed header]", "epoch", epoch)
			valset.Status = entity.HeaderMissed
			// zero PreviousHeaderHash cos header is orphaned
		} else {
			slog.DebugContext(ctx, "Header is not committed [new header]", "epoch", epoch)
			previousHeaderHash, err := v.ethClient.GetHeaderHash(ctx, lastCommittedAddr)
			if err != nil {
				return entity.ValidatorSet{}, errors.Errorf("failed to get latest header hash: %w", err)
			}
			// trying to link to latest committed header
			valset.PreviousHeaderHash = previousHeaderHash
			valset.Status = entity.HeaderPending
		}
	}

	return valset, nil
}

func (v *Deriver) isValsetHeaderCommitted(ctx context.Context, config entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error) {
	for _, addr := range config.Replicas {
		isCommitted, err := v.ethClient.IsValsetHeaderCommittedAt(ctx, addr, epoch)
		if err != nil {
			return entity.CrossChainAddress{}, false, errors.Errorf("failed to check if valset header is committed at epoch %d: %w", epoch, err)
		}
		if isCommitted {
			return addr, true, nil
		}
	}
	return entity.CrossChainAddress{}, false, nil
}

func (v *Deriver) getLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (entity.CrossChainAddress, uint64, error) {
	maxEpoch := uint64(0)
	var maxEpochAddr entity.CrossChainAddress

	for _, addr := range config.Replicas {
		epoch, err := v.ethClient.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return entity.CrossChainAddress{}, 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch > maxEpoch {
			maxEpoch = epoch
			maxEpochAddr = addr
		}
	}

	return maxEpochAddr, maxEpoch, nil
}

func (v *Deriver) formValidators(
	config entity.NetworkConfig,
	votingPowers []dtoOperatorVotingPower,
	keys []entity.OperatorWithKeys,
) ([]entity.Validator, *big.Int) {
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
				// Compare voting powers (lower first)
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

	validators := lo.Map(lo.Values(validatorsMap), func(item *entity.Validator, _ int) entity.Validator {
		return *item
	})
	// Sort validators by voting power in descending order
	sort.Slice(validators, func(i, j int) bool {
		// Compare voting powers (higher first)
		return validators[i].VotingPower.Cmp(validators[j].VotingPower.Int) > 0
	})

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

	// Sort validators by address in ascending order
	sort.Slice(validators, func(i, j int) bool {
		// Compare voting powers (lower first)
		return validators[i].Operator.Cmp(validators[j].Operator) < 0
	})
	return validators, totalActiveVotingPower
}

func (v *Deriver) calcQuorumThreshold(config entity.NetworkConfig, totalVP *big.Int) (entity.VotingPower, error) {
	quorumThresholdPercent := big.NewInt(0)
	for _, quorumThreshold := range config.QuorumThresholds {
		if quorumThreshold.KeyTag == config.RequiredHeaderKeyTag {
			quorumThresholdPercent = quorumThreshold.QuorumThreshold.Int
		}
	}
	if quorumThresholdPercent.Cmp(big.NewInt(0)) == 0 {
		return entity.VotingPower{}, errors.Errorf("quorum threshold is zero")
	}
	maxThreshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	mul := new(big.Int).Mul(totalVP, quorumThresholdPercent)
	div := new(big.Int).Div(mul, maxThreshold)
	// add 1 to apply up rounding
	return entity.ToVotingPower(new(big.Int).Add(div, big.NewInt(1))), nil
}
