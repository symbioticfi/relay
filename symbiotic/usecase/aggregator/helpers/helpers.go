package helpers

import (
	"bytes"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func CompareMessageHasher(signatures []symbiotic.Signature, msgHash []byte) bool {
	for i := range signatures {
		if !bytes.Equal(msgHash, signatures[i].MessageHash) {
			return false
		}
	}
	return true
}

func GetExtraDataKey(verificationType symbiotic.VerificationType, nameHash common.Hash) (common.Hash, error) {
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

func GetExtraDataKeyTagged(verificationType symbiotic.VerificationType, keyTag symbiotic.KeyTag, nameHash common.Hash) (common.Hash, error) {
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
		symbiotic.ExtraDataKeyTagPrefixHash,
		keyTag,
		nameHash,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func GetAggregatedPubKeys(
	valset symbiotic.ValidatorSet,
	keyTags []symbiotic.KeyTag,
) []symbiotic.ValidatorKey {
	needToAggregateTags := map[symbiotic.KeyTag]interface{}{}
	for _, tag := range keyTags {
		// only bn254 bls for now
		if tag.Type() == symbiotic.KeyTypeBlsBn254 {
			needToAggregateTags[tag] = new(bn254.G1Affine)
		}
	}

	for _, validator := range valset.Validators {
		if !validator.IsActive {
			continue
		}

		for _, key := range validator.Keys {
			if keyValue, ok := needToAggregateTags[key.Tag]; ok {
				if key.Tag.Type() == symbiotic.KeyTypeBlsBn254 {
					aggG1Key := keyValue.(*bn254.G1Affine)
					valG1Key := new(bn254.G1Affine)
					valG1Key.SetBytes(key.Payload)
					// aggregate and save in map
					needToAggregateTags[key.Tag] = aggG1Key.Add(aggG1Key, valG1Key)
				}
			}
		}
	}

	var aggregatedPubKeys []symbiotic.ValidatorKey
	for tag, keyValue := range needToAggregateTags {
		if tag.Type() == symbiotic.KeyTypeBlsBn254 {
			aggG1Key := keyValue.(*bn254.G1Affine)
			// pack g1 point to bytes and add to list
			aggG1KeyBytes := aggG1Key.Bytes()
			aggregatedPubKeys = append(aggregatedPubKeys, symbiotic.ValidatorKey{
				Tag:     tag,
				Payload: aggG1KeyBytes[:],
			})
		}
	}

	return aggregatedPubKeys
}

// will be used later
func GetExtraDataKeyIndexed(
	verificationType symbiotic.VerificationType,
	keyTag symbiotic.KeyTag,
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

func GetValidatorsIndexesMapByKey(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag) map[string]int {
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
