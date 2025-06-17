package aggregator

import (
	"math/big"
	"reflect"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
)

func (a *Aggregator) GenerateExtraData(valset entity.ValidatorSet, config entity.NetworkConfig) ([]entity.ExtraData, error) {
	extraData := make([]entity.ExtraData, 0)

	switch config.VerificationType {
	case entity.VerificationTypeZK:
		{
			totalActiveValidatorsKey, err := a.getExtraDataKey(config.VerificationType, entity.ZkVerificationTotalActiveValidatorsHash)
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

			aggregatedPubKeys := a.getAggregatedPubKeys(valset, config)

			for _, key := range aggregatedPubKeys {
				mimcAccumulator, err := proof.ValidatorSetMimcAccumulator(valset.Validators, key.Tag)
				if err != nil {
					return nil, errors.Errorf("failed to generate validator set mimc accumulator: %w", err)
				}

				validatorSetHashKey, err := a.getExtraDataKeyTagged(config.VerificationType, key.Tag, entity.ZkVerificationValidatorSetHashMimcHash)
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
		totalActiveVotingPowerKey, err := a.getExtraDataKey(config.VerificationType, entity.SimpleVerificationTotalVotingPowerHash)
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

		aggregatedPubKeys := a.getAggregatedPubKeys(valset, config)

		// pack keccak accumulators per keyTag
		for _, key := range aggregatedPubKeys {
			validatorSetHashKey, err := a.getExtraDataKeyTagged(config.VerificationType, key.Tag, entity.SimpleVerificationValidatorSetHashKeccak256Hash)
			if err != nil {
				return nil, errors.Errorf("failed to get extra data key: %w", err)
			}

			keccakHashAccumulator, err := calcKeccakAccumulator(valset.Validators, key.Tag)
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
			activeAggregatedKeyKey, err := a.getExtraDataKeyTagged(config.VerificationType, activeAggregatedKey.Tag, entity.SimpleVerificationAggPublicKeyG1Hash)
			if err != nil {
				return nil, errors.Errorf("failed to get extra data key: %w", err)
			}
			keyG1Raw, err := bls.DeserializeG1(activeAggregatedKey.Payload)
			if err != nil {
				return nil, errors.Errorf("failed to deserialize G1: %w", err)
			}

			compressedKeyG1, err := bls.Compress(keyG1Raw)
			if err != nil {
				return nil, errors.Errorf("failed to compress G1: %w", err)
			}

			extraData = append(extraData, entity.ExtraData{
				Key:   activeAggregatedKeyKey,
				Value: compressedKeyG1,
			})
		}
	}

	return extraData, nil
}

func (a *Aggregator) getExtraDataKey(verificationType entity.VerificationType, nameHash common.Hash) ([32]byte, error) {
	bytes32Ty, _ := abi.NewType("bytes32", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)

	args := abi.Arguments{
		{Type: bytes32Ty},
		{Type: u32Ty},
		{Type: bytes32Ty},
	}

	packed, err := args.Pack(
		entity.ExtraDataGlobalKeyPrefixHash,
		uint32(verificationType),
		nameHash,
	)
	if err != nil {
		return [32]byte{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func (a *Aggregator) getExtraDataKeyTagged(verificationType entity.VerificationType, keyTag entity.KeyTag, nameHash common.Hash) ([32]byte, error) {
	bytes32Ty, _ := abi.NewType("bytes32", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)
	u8Ty, _ := abi.NewType("uint8", "", nil)

	args := abi.Arguments{
		{Type: bytes32Ty},
		{Type: u32Ty},
		{Type: bytes32Ty},
		{Type: u8Ty},
		{Type: bytes32Ty},
	}

	packed, err := args.Pack(
		entity.ExtraDataGlobalKeyPrefixHash,
		uint32(verificationType),
		entity.ExtraDataKeyTagPrefixHash,
		keyTag,
		nameHash,
	)
	if err != nil {
		return [32]byte{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func (a *Aggregator) getAggregatedPubKeys(
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

//nolint:unused // will be used later
func (a *Aggregator) getExtraDataKeyIndexed(
	verificationType entity.VerificationType,
	keyTag entity.KeyTag,
	nameHash common.Hash,
	index *big.Int,
) ([32]byte, error) {
	baseHash, err := a.getExtraDataKeyTagged(verificationType, keyTag, nameHash)
	if err != nil {
		return [32]byte{}, err
	}

	sum := new(big.Int).Add(new(big.Int).SetBytes(baseHash[:]), index)
	var out [32]byte
	sum.FillBytes(out[:])
	return out, nil
}

func calcKeccakAccumulator(validators []entity.Validator, requiredKeyTag entity.KeyTag) ([32]byte, error) {
	type validatorDataTuple struct {
		KeySerialized common.Hash
		VotingPower   *big.Int
	}
	u256, _ := abi.NewType("uint256", "", nil)
	b32, _ := abi.NewType("bytes32", "", nil)

	tupleType := abi.Type{
		T:             abi.TupleTy,
		TupleElems:    []*abi.Type{&b32, &u256},
		TupleRawNames: []string{"keySerialized", "votingPower"},
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

				compressedKeyG1, err := bls.Compress(validatorKeyG1)
				if err != nil {
					return [32]byte{}, errors.Errorf("failed to compress G1: %w", err)
				}

				votingPower := validatorVotingPower

				validatorsData = append(validatorsData, validatorDataTuple{
					KeySerialized: compressedKeyG1,
					VotingPower:   votingPower,
				})
			}
		}
	}

	sort.Slice(validatorsData, func(i, j int) bool {
		// Compare keys (lower first)
		return validatorsData[i].KeySerialized.Cmp(validatorsData[j].KeySerialized) < 0
	})

	packed, err := args.Pack(validatorsData)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to pack arguments: %w", err)
	}
	hash := crypto.Keccak256Hash(packed[32:])
	return hash, nil
}
