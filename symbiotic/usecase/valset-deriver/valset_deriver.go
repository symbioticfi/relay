package valsetDeriver

import (
	"context"
	"log/slog"
	"maps"
	"math/big"
	"slices"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/ssz"
)

const (
	valsetVersion      = 1
	aggregatorRoleType = "AGGREGATOR"
	committerRoleType  = "COMMITTER"
)

//go:generate mockgen -source=valset_deriver.go -destination=mocks/deriver.go -package=mocks -mock_names=evmClient=MockEvmClient
type evmClient interface {
	GetConfig(ctx context.Context, timestamp symbiotic.Timestamp) (symbiotic.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.Timestamp, error)
	GetVotingPowers(ctx context.Context, address symbiotic.CrossChainAddress, timestamp symbiotic.Timestamp) ([]symbiotic.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address symbiotic.CrossChainAddress, timestamp symbiotic.Timestamp) ([]symbiotic.OperatorWithKeys, error)
	GetEip712Domain(ctx context.Context, addr symbiotic.CrossChainAddress) (symbiotic.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetSubnetwork(ctx context.Context) (common.Hash, error)
	GetNetworkAddress(ctx context.Context) (common.Address, error)
	GetHeaderHash(ctx context.Context, addr symbiotic.CrossChainAddress) (common.Hash, error)
	IsValsetHeaderCommittedAt(ctx context.Context, addr symbiotic.CrossChainAddress, epoch symbiotic.Epoch) (bool, error)
	GetHeaderHashAt(ctx context.Context, addr symbiotic.CrossChainAddress, epoch symbiotic.Epoch) (common.Hash, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr symbiotic.CrossChainAddress) (symbiotic.Epoch, error)
	GetOperators(ctx context.Context, address symbiotic.CrossChainAddress, timestamp symbiotic.Timestamp) ([]common.Address, error)
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

func (v *Deriver) GetNetworkData(ctx context.Context, addr symbiotic.CrossChainAddress) (symbiotic.NetworkData, error) {
	address, err := v.evmClient.GetNetworkAddress(ctx)
	if err != nil {
		return symbiotic.NetworkData{}, errors.Errorf("failed to get network address: %w", err)
	}

	subnetwork, err := v.evmClient.GetSubnetwork(ctx)
	if err != nil {
		return symbiotic.NetworkData{}, errors.Errorf("failed to get subnetwork: %w", err)
	}

	eip712Data, err := v.evmClient.GetEip712Domain(ctx, addr)
	if err != nil {
		return symbiotic.NetworkData{}, errors.Errorf("failed to get eip712 domain: %w", err)
	}

	return symbiotic.NetworkData{
		Address:    address,
		Subnetwork: subnetwork,
		Eip712Data: eip712Data,
	}, nil
}

type dtoOperatorVotingPower struct {
	chainId      uint64
	votingPowers []symbiotic.OperatorVotingPower
}

func (v *Deriver) GetValidatorSet(ctx context.Context, epoch symbiotic.Epoch, config symbiotic.NetworkConfig) (symbiotic.ValidatorSet, error) {
	timestamp, err := v.evmClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return symbiotic.ValidatorSet{}, errors.Errorf("failed to get epoch start timestamp: %w", err)
	}
	slog.DebugContext(ctx, "Got current valset timestamp", "timestamp", strconv.Itoa(int(timestamp)), "epoch", epoch)

	// Get voting powers from all voting power providers
	allVotingPowers := make([]dtoOperatorVotingPower, len(config.VotingPowerProviders))
	for i, provider := range config.VotingPowerProviders {
		votingPowers, err := v.evmClient.GetVotingPowers(ctx, provider, timestamp)
		if err != nil {
			return symbiotic.ValidatorSet{}, errors.Errorf("failed to get voting powers from provider %s: %w", provider.Address.Hex(), err)
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
		return symbiotic.ValidatorSet{}, errors.Errorf("failed to get keys: %w", err)
	}
	slog.DebugContext(ctx, "Got keys from provider", "provider", config.KeysProvider.Address.Hex(), "keys", keys)

	// form validators list from voting powers and keys using config
	validators := v.formValidators(config, allVotingPowers, keys)

	// calc new quorum threshold
	quorumThreshold, err := config.CalcQuorumThreshold(validators.GetTotalActiveVotingPower())
	if err != nil {
		return symbiotic.ValidatorSet{}, errors.Errorf("failed to calc quorum threshold: %w", err)
	}

	valset := symbiotic.ValidatorSet{
		Version:           valsetVersion,
		RequiredKeyTag:    config.RequiredHeaderKeyTag,
		Epoch:             epoch,
		CaptureTimestamp:  timestamp,
		QuorumThreshold:   quorumThreshold,
		Validators:        validators,
		Status:            symbiotic.HeaderDerived,
		AggregatorIndices: nil, // will be initialized later
		CommitterIndices:  nil, // will be initialized later
	}

	aggIndices, commIndices, err := GetSchedulerInfo(ctx, valset, config)
	if err != nil {
		return symbiotic.ValidatorSet{}, errors.Errorf("failed to get scheduler info: %w", err)
	}
	valset.AggregatorIndices = aggIndices
	valset.CommitterIndices = commIndices

	return valset, nil
}

func GetSchedulerInfo(_ context.Context, valset symbiotic.ValidatorSet, config symbiotic.NetworkConfig) (aggIndices []uint32, commIndices []uint32, err error) {
	// ensure validators sorted already, function expects sorted list
	if err := valset.Validators.CheckIsSortedByOperatorAddressAsc(); err != nil {
		return nil, nil, err
	}

	aggregatorIndices := map[uint32]struct{}{}
	committerIndices := map[uint32]struct{}{}

	header, err := valset.GetHeader()
	if err != nil {
		return nil, nil, errors.Errorf("failed to get valset header: %w", err)
	}

	headerHash, err := header.Hash()
	if err != nil {
		return nil, nil, errors.Errorf("failed to hash valset header: %w", err)
	}

	validatorCount := len(valset.Validators)
	if validatorCount == 0 {
		return []uint32{}, []uint32{}, nil
	}

	for i := 1; i <= int(config.NumAggregators); i++ {
		hash := new(big.Int).SetBytes(
			crypto.Keccak256Hash(
				[]byte(aggregatorRoleType),
				headerHash.Bytes(),
				new(big.Int).SetInt64(int64(i)).Bytes(),
			).Bytes())

		startIndex := new(big.Int).Mod(hash, big.NewInt(int64(validatorCount))).Uint64()
		foundIndex := findNextAvailableIndex(uint32(startIndex), validatorCount, aggregatorIndices)
		aggregatorIndices[foundIndex] = struct{}{}
	}

	for i := 1; i <= int(config.NumCommitters); i++ {
		hash := new(big.Int).SetBytes(
			crypto.Keccak256Hash(
				[]byte(committerRoleType),
				headerHash.Bytes(),
				new(big.Int).SetInt64(int64(i)).Bytes(),
			).Bytes())

		startIndex := new(big.Int).Mod(hash, big.NewInt(int64(validatorCount))).Uint64()
		foundIndex := findNextAvailableIndex(uint32(startIndex), validatorCount, committerIndices)
		committerIndices[foundIndex] = struct{}{}
	}

	return slices.Collect(maps.Keys(aggregatorIndices)), slices.Collect(maps.Keys(committerIndices)), nil
}

// Helper function for wrap-around search
func findNextAvailableIndex(startIndex uint32, validatorCount int, usedIndices map[uint32]struct{}) uint32 {
	for offset := 0; offset < validatorCount; offset++ {
		candidateIndex := (startIndex + uint32(offset)) % uint32(validatorCount)
		if _, exists := usedIndices[candidateIndex]; !exists {
			return candidateIndex
		}
	}
	// This should never happen if we don't request more roles than available validators
	panic("no available validator index found - this indicates a bug")
}

func (v *Deriver) formValidators(
	config symbiotic.NetworkConfig,
	votingPowers []dtoOperatorVotingPower,
	keys []symbiotic.OperatorWithKeys,
) symbiotic.Validators {
	validators := fillValidators(votingPowers, keys)

	markValidatorsActive(config, validators)

	validators.SortByOperatorAddressAsc()

	return validators
}

func markValidatorsActive(config symbiotic.NetworkConfig, validators symbiotic.Validators) {
	totalActive := 0

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
				validators[i].VotingPower = symbiotic.ToVotingPower(new(big.Int).Set(config.MaxVotingPower.Int))
			}
		}

		if config.MaxValidatorsCount.Int64() != 0 {
			if totalActive >= int(config.MaxValidatorsCount.Int64()) {
				break
			}
		}
	}
}

func fillValidators(votingPowers []dtoOperatorVotingPower, keys []symbiotic.OperatorWithKeys) symbiotic.Validators {
	// Create validators map to consolidate voting powers and keys
	validatorsMap := make(map[string]*symbiotic.Validator)

	// Process voting powers
	for _, chainVp := range votingPowers {
		for _, vp := range chainVp.votingPowers {
			operatorAddr := vp.Operator.Hex()
			if _, exists := validatorsMap[operatorAddr]; !exists {
				validatorsMap[operatorAddr] = &symbiotic.Validator{
					Operator:    vp.Operator,
					VotingPower: symbiotic.ToVotingPower(big.NewInt(0)),
					IsActive:    false, // Default to active, will filter later
					Keys:        []symbiotic.ValidatorKey{},
					Vaults:      []symbiotic.ValidatorVault{},
				}
			}

			// Add vaults and their voting powers
			for _, vault := range vp.Vaults {
				validatorsMap[operatorAddr].VotingPower = symbiotic.ToVotingPower(new(big.Int).Add(
					validatorsMap[operatorAddr].VotingPower.Int,
					vault.VotingPower.Int,
				))

				// Add vault to validator's vaults
				validatorsMap[operatorAddr].Vaults = append(validatorsMap[operatorAddr].Vaults, symbiotic.ValidatorVault{
					Vault:       vault.Vault,
					VotingPower: vault.VotingPower,
					ChainID:     chainVp.chainId,
				})
			}
		}
	}

	// filter by ssz max-vaults limit
	for val := range validatorsMap {
		validatorsMap[val].Vaults.SortVaultsByVotingPowerDescAndAddressAsc()
		if len(validatorsMap[val].Vaults) > ssz.VaultsListMaxElements {
			validatorsMap[val].Vaults = validatorsMap[val].Vaults[:ssz.VaultsListMaxElements]
		}

		validatorsMap[val].Vaults.SortByAddressAsc()

		totalVP := big.NewInt(0)
		for _, vault := range validatorsMap[val].Vaults {
			totalVP.Add(totalVP, vault.VotingPower.Int)
		}
		validatorsMap[val].VotingPower = symbiotic.ToVotingPower(totalVP)
	}

	// Process required keys
	for _, rk := range keys { // TODO: get required key tags from validator set config and fill with nils if needed
		operatorAddr := rk.Operator.Hex()
		if validator, exists := validatorsMap[operatorAddr]; exists {
			// Add all keys for this operator
			for _, key := range rk.Keys {
				validator.Keys = append(validator.Keys, symbiotic.ValidatorKey{
					Tag:     key.Tag,
					Payload: key.Payload,
				})
			}
		}
	}

	validators := symbiotic.Validators(lo.Map(lo.Values(validatorsMap), func(item *symbiotic.Validator, _ int) symbiotic.Validator {
		return *item
	}))

	// filter by ssz max-validators limit
	validators.SortByVotingPowerDescAndOperatorAddressAsc()
	if len(validators) > ssz.ValidatorsListMaxElements {
		validators = validators[:ssz.ValidatorsListMaxElements]
	}

	return validators
}
