package helpers

import (
	"bytes"
	"math/big"

	"github.com/symbioticfi/relay/core/entity"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"
)

func CompareMessageHasher(signatures []entity.SignatureExtended, msgHash []byte) bool {
	for i := range signatures {
		if !bytes.Equal(msgHash, signatures[i].MessageHash) {
			return false
		}
	}
	return true
}

func GetExtraDataKey(verificationType entity.VerificationType, nameHash common.Hash) (common.Hash, error) {
	bytes32Ty, _ := abi.NewType("bytes32", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)

	args := abi.Arguments{
		{Type: u32Ty},
		{Type: bytes32Ty},
	}

	packed, err := args.Pack(
		uint32(verificationType),
		nameHash,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func GetExtraDataKeyTagged(verificationType entity.VerificationType, keyTag entity.KeyTag, nameHash common.Hash) (common.Hash, error) {
	bytes32Ty, _ := abi.NewType("bytes32", "", nil)
	u32Ty, _ := abi.NewType("uint32", "", nil)
	u8Ty, _ := abi.NewType("uint8", "", nil)

	args := abi.Arguments{
		{Type: u32Ty},
		{Type: bytes32Ty},
		{Type: u8Ty},
		{Type: bytes32Ty},
	}

	packed, err := args.Pack(
		uint32(verificationType),
		entity.ExtraDataKeyTagPrefixHash,
		keyTag,
		nameHash,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func GetAggregatedPubKeys(
	valset entity.ValidatorSet,
	keyTags []entity.KeyTag,
) []entity.ValidatorKey {
	needToAggregateTags := map[entity.KeyTag]interface{}{}
	for _, tag := range keyTags {
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

	var aggregatedPubKeys []entity.ValidatorKey
	for tag, keyValue := range needToAggregateTags {
		if tag.Type() == entity.KeyTypeBlsBn254 {
			aggG1Key := keyValue.(*bn254.G1Affine)
			// pack g1 point to bytes and add to list
			aggG1KeyBytes := aggG1Key.Bytes()
			aggregatedPubKeys = append(aggregatedPubKeys, entity.ValidatorKey{
				Tag:     tag,
				Payload: aggG1KeyBytes[:],
			})
		}
	}

	return aggregatedPubKeys
}

// will be used later
func GetExtraDataKeyIndexed(
	verificationType entity.VerificationType,
	keyTag entity.KeyTag,
	nameHash common.Hash,
	index *big.Int,
) (common.Hash, error) {
	baseHash, err := GetExtraDataKeyTagged(verificationType, keyTag, nameHash)
	if err != nil {
		return common.Hash{}, err
	}

	sum := new(big.Int).Add(new(big.Int).SetBytes(baseHash[:]), index)
	var out common.Hash
	sum.FillBytes(out[:])
	return out, nil
}

func GetValidatorsIndexesMapByKey(valset entity.ValidatorSet, keyTag entity.KeyTag) map[string]int {
	keysMap := make(map[string]int)

	for i, validator := range valset.Validators {
		if !validator.IsActive {
			continue
		}

		for _, key := range validator.Keys {
			if key.Tag == keyTag {
				keysMap[string(key.Payload)] = i
			}
		}
	}

	return keysMap
}
