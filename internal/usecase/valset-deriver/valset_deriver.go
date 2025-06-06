package valsetDeriver

import (
	"bytes"
	"context"
	"log/slog"
	"math/big"
	"reflect"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
)

const valsetVersion = 1

//go:generate mockgen -source=valset_deriver.go -destination=mocks/deriver.go -package=mocks
type ethClient interface {
	GetCaptureTimestamp(ctx context.Context) (uint64, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
	GetCurrentValsetTimestamp(ctx context.Context) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error)
	GetRequiredKeyTag(ctx context.Context, timestamp uint64) (entity.KeyTag, error)
	GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetSubnetwork(ctx context.Context) ([32]byte, error)
	GetNetworkAddress(ctx context.Context) (*common.Address, error)
	GetLatestHeaderHash(ctx context.Context) ([32]byte, error)
	IsValsetHeaderCommittedAt(ctx context.Context, epoch uint64) (bool, error)
	GetPreviousHeaderHashAt(ctx context.Context, epoch uint64) ([32]byte, error)
	GetHeaderHashAt(ctx context.Context, epoch uint64) ([32]byte, error)
	GetLastCommittedHeaderEpoch(ctx context.Context) (uint64, error)
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

func (v *Deriver) GetNetworkData(ctx context.Context) (entity.NetworkData, error) {
	address, err := v.ethClient.GetNetworkAddress(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get network address: %w", err)
	}

	subnetwork, err := v.ethClient.GetSubnetwork(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get subnetwork: %w", err)
	}

	eip712Data, err := v.ethClient.GetEip712Domain(ctx)
	if err != nil {
		return entity.NetworkData{}, errors.Errorf("failed to get eip712 domain: %w", err)
	}

	return entity.NetworkData{
		Address:    *address,
		Subnetwork: subnetwork,
		Eip712Data: eip712Data,
	}, nil
}

func (v *Deriver) GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error) {
	slog.DebugContext(ctx, "Trying to fetch current valset timestamp", "epoch", epoch)
	timestamp, err := v.ethClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to get epoch start timestamp: %w", err)
	}
	slog.DebugContext(ctx, "Got current valset timestamp", "timestamp", timestamp, "epoch", epoch)

	slog.DebugContext(ctx, "Got config", "timestamp", timestamp, "config", config)

	// Get voting powers from all voting power providers
	var allVotingPowers []entity.OperatorVotingPower
	for _, provider := range config.VotingPowerProviders {
		slog.DebugContext(ctx, "Trying to fetch voting powers from provider", "provider", provider.Address.Hex())
		votingPowers, err := v.ethClient.GetVotingPowers(ctx, provider, timestamp)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get voting powers from provider %s: %w", provider.Address.Hex(), err)
		}

		slog.DebugContext(ctx, "Got voting powers from provider", "provider", provider.Address.Hex(), "votingPowers", votingPowers)

		allVotingPowers = append(allVotingPowers, votingPowers...)
	}

	// Get keys from the keys provider
	slog.DebugContext(ctx, "Trying to fetch keys from provider", "provider", config.KeysProvider.Address.Hex())

	keys, err := v.ethClient.GetKeys(ctx, config.KeysProvider, timestamp)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to get keys: %w", err)
	}

	// form validators list from voting powers and keys using config
	validators, totalVP := v.formValidators(config, allVotingPowers, keys)

	// calc new quorum threshold
	quorumThreshold := v.calcQuorumThreshold(config, totalVP)

	requiredKeyTag, err := v.ethClient.GetRequiredKeyTag(ctx, timestamp)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to get required key tag: %w", err)
	}

	isValsetCommitted, err := v.ethClient.IsValsetHeaderCommittedAt(ctx, epoch)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to check if validator committed at epoch %d: %w", epoch, err)
	}

	valset := entity.ValidatorSet{
		Version:          valsetVersion,
		RequiredKeyTag:   requiredKeyTag,
		Epoch:            epoch,
		CaptureTimestamp: timestamp,
		QuorumThreshold:  quorumThreshold,
		Validators:       validators,
	}

	if isValsetCommitted {
		slog.DebugContext(ctx, "Validator set committed at epoch already, checking integrity", "epoch", epoch)
		previousHeaderHash, err := v.ethClient.GetPreviousHeaderHashAt(ctx, epoch)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get previous header hash: %w", err)
		}
		valset.PreviousHeaderHash = previousHeaderHash

		// valset integrity check
		committedHash, err := v.ethClient.GetHeaderHashAt(ctx, epoch)
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
			slog.DebugContext(ctx, "committed header hash", "hash", committedHash)
			slog.DebugContext(ctx, "calculated header hash", "hash", calculatedHash)
			return entity.ValidatorSet{}, errors.Errorf("validator set hash mistmach at epoch %d", epoch)
		}

		valset.Status = entity.HeaderCommitted
	} else {
		latestCommittedEpoch, err := v.ethClient.GetLastCommittedHeaderEpoch(ctx)
		if err != nil {
			return entity.ValidatorSet{}, errors.Errorf("failed to get current valset epoch: %w", err)
		}

		if epoch < latestCommittedEpoch {
			valset.Status = entity.HeaderMissed
			// zero PreviousHeaderHash cos header is orphaned
		} else {
			slog.DebugContext(ctx, "Validator set is not committed at epoch", "epoch", epoch)
			previousHeaderHash, err := v.ethClient.GetLatestHeaderHash(ctx)
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

func (v *Deriver) formValidators(
	config entity.NetworkConfig,
	votingPowers []entity.OperatorVotingPower,
	keys []entity.OperatorWithKeys,
) ([]entity.Validator, *big.Int) {
	// Create validators map to consolidate voting powers and keys
	validatorsMap := make(map[string]*entity.Validator)

	// Process voting powers
	for _, vp := range votingPowers {
		operatorAddr := vp.Operator.Hex()
		if _, exists := validatorsMap[operatorAddr]; !exists {
			validatorsMap[operatorAddr] = &entity.Validator{
				Operator:    vp.Operator,
				VotingPower: big.NewInt(0),
				IsActive:    false, // Default to active, will filter later
				Keys:        []entity.Key{},
				Vaults:      []entity.ValidatorVault{},
			}
		}

		// Add vaults and their voting powers
		for _, vault := range vp.Vaults {
			validatorsMap[operatorAddr].VotingPower = new(big.Int).Add(
				validatorsMap[operatorAddr].VotingPower,
				vault.VotingPower,
			)

			// Add vault to validator's vaults
			validatorsMap[operatorAddr].Vaults = append(validatorsMap[operatorAddr].Vaults, entity.ValidatorVault{
				Vault:       vault.Vault,
				VotingPower: vault.VotingPower,
			})
		}

		// Sort vaults by address in ascending order
		sort.Slice(validatorsMap[operatorAddr].Vaults, func(i, j int) bool {
			// Compare voting powers (higher first)
			return validatorsMap[operatorAddr].Vaults[i].Vault.Cmp(validatorsMap[operatorAddr].Vaults[j].Vault) < 0
		})
	}

	// Process required keys
	for _, rk := range keys { // TODO: get required key tags from validator set config and fill with nils if needed
		operatorAddr := rk.Operator.Hex()
		if validator, exists := validatorsMap[operatorAddr]; exists {
			// Add all keys for this operator
			for _, key := range rk.Keys {
				validator.Keys = append(validator.Keys, entity.Key{
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
		return validators[i].VotingPower.Cmp(validators[j].VotingPower) > 0
	})

	totalActive := 0

	totalActiveVotingPower := big.NewInt(0)
	for i := range validators {
		// Check minimum voting power if configured
		if validators[i].VotingPower.Cmp(config.MinInclusionVotingPower) < 0 {
			break
		}

		// Check if validator has at least one key
		if len(validators[i].Keys) == 0 {
			continue
		}

		totalActive++
		validators[i].IsActive = true

		if config.MaxVotingPower.Int64() != 0 {
			if validators[i].VotingPower.Cmp(config.MaxVotingPower) > 0 {
				validators[i].VotingPower = new(big.Int).Set(config.MaxVotingPower)
			}
		}
		// Add to total active voting power if validator is active
		totalActiveVotingPower = new(big.Int).Add(totalActiveVotingPower, validators[i].VotingPower)

		if config.MaxValidatorsCount.Int64() != 0 {
			if totalActive >= int(config.MaxValidatorsCount.Int64()) {
				break
			}
		}
	}

	// Sort validators by address in ascending order
	sort.Slice(validators, func(i, j int) bool {
		// Compare voting powers (higher first)
		return validators[i].Operator.Cmp(validators[j].Operator) < 0
	})
	return validators, totalActiveVotingPower
}

func (v *Deriver) calcQuorumThreshold(_ entity.NetworkConfig, totalVP *big.Int) *big.Int {
	// not using config now but later can
	mul := big.NewInt(1).Mul(totalVP, big.NewInt(2))
	div := big.NewInt(1).Div(mul, big.NewInt(3))
	return big.NewInt(0).Add(div, big.NewInt(1))
}

// TODO need to move to aggregator maybe
func (v *Deriver) GenerateExtraData(valset entity.ValidatorSet, config entity.NetworkConfig) ([]entity.ExtraData, error) {
	extraData := make([]entity.ExtraData, 0)

	switch config.VerificationType {
	case entity.VerificationTypeZK:
		{
			totalActiveValidatorsKey, err := v.getExtraDataKey(config.VerificationType, entity.ZkVerificationTotalActiveValidators)
			if err != nil {
				return nil, errors.Errorf("failed to get extra data key: %w", err)
			}

			totalActiveValidators := big.NewInt(valset.GetTotalActiveValidators())
			totalActiveValidatorsBytes32 := [32]byte{}
			totalActiveValidators.FillBytes(totalActiveValidatorsBytes32[:])
			extraData = append(extraData, entity.ExtraData{
				Key:   totalActiveValidatorsKey,
				Value: totalActiveValidatorsBytes32,
			})

			aggregatedPubKeys := v.getAggregatedPubKeys(valset, config)

			for _, key := range aggregatedPubKeys {
				mimcAccumulator, err := proof.ValidatorSetMimcAccumulator(valset.Validators, key.Tag)
				if err != nil {
					return nil, errors.Errorf("failed to generate validator set mimc accumulator: %w", err)
				}

				validatorSetHashKey, err := v.getExtraDataKeyTagged(config.VerificationType, key.Tag, entity.ZkVerificationValidatorSetHashMimc)
				if err != nil {
					return nil, errors.Errorf("failed to get extra data key: %w", err)
				}

				extraData = append(extraData, entity.ExtraData{
					Key:   validatorSetHashKey,
					Value: mimcAccumulator,
				})
			}
		}
	case entity.VerificationTypeSimple: // TODO: prettify/check
		totalActiveVotingPowerKey, err := v.getExtraDataKey(config.VerificationType, entity.SimpleVerificationTotalVotingPower)
		if err != nil {
			return nil, errors.Errorf("failed to get extra data key: %w", err)
		}

		totalActiveVotingPower := valset.GetTotalActiveVotingPower()
		totalActiveVotingPowerBytes32 := [32]byte{}
		totalActiveVotingPower.FillBytes(totalActiveVotingPowerBytes32[:])
		extraData = append(extraData, entity.ExtraData{
			Key:   totalActiveVotingPowerKey,
			Value: totalActiveVotingPowerBytes32,
		})

		aggregatedPubKeys := v.getAggregatedPubKeys(valset, config)

		// pack keccak accumulators per keyTag
		for _, key := range aggregatedPubKeys {
			validatorSetHashKey, err := v.getExtraDataKeyTagged(config.VerificationType, key.Tag, entity.SimpleVerificationValidatorSetHashKeccak256)
			if err != nil {
				return nil, errors.Errorf("failed to get extra data key: %w", err)
			}

			keccakHashAccumulator, err := CalcKeccakAccumulator(valset.Validators, key.Tag)
			if err != nil {
				return nil, errors.Errorf("failed to generate validator set mimc accumulator: %w", err)
			}

			extraData = append(extraData, entity.ExtraData{
				Key:   validatorSetHashKey,
				Value: keccakHashAccumulator,
			})
		}

		// pack aggregated keys per keyTag
		for _, activeAggregatedKey := range aggregatedPubKeys {
			activeAggregatedKeyKey, err := v.getExtraDataKeyTagged(config.VerificationType, activeAggregatedKey.Tag, entity.SimpleVerificationAggPublicKeyG1)
			if err != nil {
				return nil, errors.Errorf("failed to get extra data key: %w", err)
			}
			keyG1Raw, err := bls.DeserializeG1(activeAggregatedKey.Payload)
			if err != nil {
				return nil, errors.Errorf("failed to deserialize G1: %w", err)
			}

			x := keyG1Raw.X.BigInt(new(big.Int))
			y := keyG1Raw.Y.BigInt(new(big.Int))
			_, derivedY, err := bls.FindYFromX(x)
			if err != nil {
				return nil, errors.Errorf("failed to find Y from X: %w", err)
			}

			flag := y.Cmp(derivedY) != 0
			compressedKeyG1 := new(big.Int).Mul(x, big.NewInt(2))
			if flag {
				compressedKeyG1.Add(compressedKeyG1, big.NewInt(1))
			}

			compressedKeyG1Bytes := [32]byte{}
			compressedKeyG1.FillBytes(compressedKeyG1Bytes[:])
			extraData = append(extraData, entity.ExtraData{
				Key:   activeAggregatedKeyKey,
				Value: compressedKeyG1Bytes,
			})
		}
	}

	return extraData, nil
}

func (v *Deriver) getExtraDataKey(verificationType entity.VerificationType, name string) ([32]byte, error) {
	strTy, _ := abi.NewType("string", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)

	args := abi.Arguments{
		{Type: strTy},
		{Type: u32Ty},
		{Type: strTy},
	}

	packed, err := args.Pack(entity.ExtraDataGlobalKeyPrefix, uint32(verificationType), name)
	if err != nil {
		return [32]byte{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func (v *Deriver) getExtraDataKeyTagged(verificationType entity.VerificationType, keyTag entity.KeyTag, name string) ([32]byte, error) {
	strTy, _ := abi.NewType("string", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)
	u8Ty, _ := abi.NewType("uint8", "", nil)

	args := abi.Arguments{
		{Type: strTy},
		{Type: u32Ty},
		{Type: strTy},
		{Type: u8Ty},
		{Type: strTy},
	}

	packed, err := args.Pack(entity.ExtraDataGlobalKeyPrefix, uint32(verificationType), entity.ExtraDataKeyTagPrefix, keyTag, name)
	if err != nil {
		return [32]byte{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

//nolint:unused // will be used later
func (v *Deriver) getExtraDataKeyIndexed(
	verificationType entity.VerificationType,
	keyTag entity.KeyTag,
	name string,
	index *big.Int,
) ([32]byte, error) {
	baseHash, err := v.getExtraDataKeyTagged(verificationType, keyTag, name)
	if err != nil {
		return [32]byte{}, err
	}

	sum := new(big.Int).Add(new(big.Int).SetBytes(baseHash[:]), index)
	var out [32]byte
	sum.FillBytes(out[:])
	return out, nil
}

func (v *Deriver) getAggregatedPubKeys(
	valset entity.ValidatorSet,
	config entity.NetworkConfig,
) []entity.Key {
	needToAggregateTags := map[entity.KeyTag]interface{}{}
	for _, tag := range config.RequiredKeyTags {
		// only bn254 bls for now
		if tag.Type() == entity.KeyTypeBlsBn254 {
			needToAggregateTags[tag] = new(bn254.G1Affine)
		}
	}

	for _, validator := range valset.Validators {
		if !validator.IsActive {
			continue
		}

		for _, key := range validator.Keys {
			if keyValue, ok := needToAggregateTags[key.Tag]; ok {
				if key.Tag.Type() == entity.KeyTypeBlsBn254 {
					aggG1Key := keyValue.(*bn254.G1Affine)
					valG1Key := new(bn254.G1Affine)
					valG1Key.SetBytes(key.Payload)
					// aggregate and save in map
					needToAggregateTags[key.Tag] = aggG1Key.Add(aggG1Key, valG1Key)
				}
			}
		}
	}

	var aggregatedPubKeys []entity.Key
	for tag, keyValue := range needToAggregateTags {
		if tag.Type() == entity.KeyTypeBlsBn254 {
			aggG1Key := keyValue.(*bn254.G1Affine)
			// pack g1 point to bytes and add to list
			aggG1KeyBytes := aggG1Key.Bytes()
			aggregatedPubKeys = append(aggregatedPubKeys, entity.Key{
				Tag:     tag,
				Payload: aggG1KeyBytes[:],
			})
		}
	}

	return aggregatedPubKeys
}

func CalcKeccakAccumulator(validators []entity.Validator, requiredKeyTag entity.KeyTag) ([32]byte, error) {

	type validatorDataTuple struct {
		X, Y, VotingPower *big.Int
	}
	u256, _ := abi.NewType("uint256", "", nil)

	tupleType := abi.Type{
		T:             abi.TupleTy,
		TupleElems:    []*abi.Type{&u256, &u256, &u256},
		TupleRawNames: []string{"X", "Y", "votingPower"},
		TupleType:     reflect.TypeOf(validatorDataTuple{}),
	}

	arrayType := abi.Type{
		T:    abi.SliceTy,
		Elem: &tupleType,
	}

	args := abi.Arguments{{Type: arrayType}}
	validatorsData := make([]validatorDataTuple, 0, len(validators))
	for _, validator := range validators {
		validatorVotingPower := validator.VotingPower
		for _, validatorKey := range validator.Keys {
			if validatorKey.Tag == requiredKeyTag {
				validatorKeyG1, err := bls.DeserializeG1(validatorKey.Payload)
				if err != nil {
					return [32]byte{}, errors.Errorf("failed to deserialize G1: %w", err)
				}
				x := validatorKeyG1.X.BigInt(new(big.Int))
				y := validatorKeyG1.Y.BigInt(new(big.Int))

				votingPower := validatorVotingPower

				validatorsData = append(validatorsData, validatorDataTuple{
					X:           x,
					Y:           y,
					VotingPower: votingPower,
				})
			}
		}
	}

	sort.Slice(validatorsData, func(i, j int) bool {
		// Compare keys (lower first)
		return validatorsData[i].X.Cmp(validatorsData[j].X) > 0 || validatorsData[i].Y.Cmp(validatorsData[j].Y) > 0
	})

	packed, err := args.Pack(validatorsData)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to pack arguments: %w", err)
	}
	hash := crypto.Keccak256Hash(packed)
	return hash, nil
}
