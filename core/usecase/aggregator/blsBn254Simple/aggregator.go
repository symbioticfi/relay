package blsBn254Simple

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"reflect"
	"sort"

	"github.com/symbioticfi/relay/core/usecase/aggregator/helpers"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/bls"
)

type Aggregator struct{}

func NewAggregator() *Aggregator {
	return &Aggregator{}
}

func (a Aggregator) Aggregate(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	messageHash []byte,
	signatures []entity.SignatureExtended,
) (entity.AggregationProof, error) {
	if !helpers.CompareMessageHasher(signatures, messageHash) {
		return entity.AggregationProof{}, errors.New("message hashes mismatch")
	}

	type dtoG1Point struct {
		X *big.Int
		Y *big.Int
	}
	type dtoG2Point struct {
		X [2]*big.Int
		Y [2]*big.Int
	}
	type dtoValidatorData struct {
		KeySerialized common.Hash
		VotingPower   *big.Int
		isNonSigner   bool
	}
	var validatorsData []dtoValidatorData

	aggG1Sig := bls.ZeroG1()
	aggG2Key := bls.ZeroG2()
	signers := make(map[common.Address]bool)
	for _, sig := range signatures {
		g1, g2Key, err := bls.UnpackPublicG1G2(sig.PublicKey)
		if err != nil {
			return entity.AggregationProof{}, err
		}
		val, ok := valset.FindValidatorByKey(keyTag, g1.Marshal())
		if !ok {
			return entity.AggregationProof{}, errors.New("failed to find validator by key")
		}
		g1Sig, err := bls.DeserializeG1(sig.Signature)
		if err != nil {
			return entity.AggregationProof{}, err
		}
		aggG1Sig = aggG1Sig.Add(g1Sig)
		aggG2Key = aggG2Key.Add(&g2Key)
		signers[val.Operator] = true
	}

	for _, val := range valset.Validators {
		if val.IsActive {
			keyBytes, ok := val.FindKeyByKeyTag(keyTag)
			if !ok {
				return entity.AggregationProof{}, errors.New("failed to find key by keyTag")
			}
			_, isSinger := signers[val.Operator]
			g1Key, err := bls.DeserializeG1(keyBytes)
			if err != nil {
				return entity.AggregationProof{}, errors.Errorf("failed to deserialize G1 key: %w", err)
			}

			compressedKeyG1, err := bls.Compress(g1Key)
			if err != nil {
				return entity.AggregationProof{}, errors.Errorf("failed to compress G1 key: %w", err)
			}

			validatorsData = append(validatorsData, dtoValidatorData{
				KeySerialized: compressedKeyG1,
				VotingPower:   val.VotingPower.Int,
				isNonSigner:   !isSinger,
			})
		}
	}

	sort.Slice(validatorsData, func(i, j int) bool {
		// Compare keys (lower first)
		return validatorsData[i].KeySerialized.Cmp(validatorsData[j].KeySerialized) < 0
	})

	nonSigners := make([]int, 0, len(validatorsData))
	for i, val := range validatorsData {
		if val.isNonSigner {
			nonSigners = append(nonSigners, i)
		}
	}

	dtoG1AggSig := dtoG1Point{
		X: aggG1Sig.X.BigInt(new(big.Int)),
		Y: aggG1Sig.Y.BigInt(new(big.Int)),
	}

	dtoG2AggKey := dtoG2Point{}
	dtoG2AggKey.X[1] = aggG2Key.X.A0.BigInt(new(big.Int))
	dtoG2AggKey.X[0] = aggG2Key.X.A1.BigInt(new(big.Int))
	dtoG2AggKey.Y[1] = aggG2Key.Y.A0.BigInt(new(big.Int))
	dtoG2AggKey.Y[0] = aggG2Key.Y.A1.BigInt(new(big.Int))

	g2Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256[2]"},
		{Name: "Y", Type: "uint256[2]"},
	})
	if err != nil {
		return entity.AggregationProof{}, err
	}

	g1Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256"},
		{Name: "Y", Type: "uint256"},
	})
	if err != nil {
		return entity.AggregationProof{}, err
	}

	validatorsDataType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "keySerialized", Type: "bytes32"},
		{Name: "VotingPower", Type: "uint256"},
	})
	if err != nil {
		return entity.AggregationProof{}, err
	}

	g1PointAbiArgs := abi.Arguments{
		{
			Type: g1Type,
		},
	}

	g2PointAbiArgs := abi.Arguments{
		{
			Type: g2Type,
		},
	}

	validatorsDataAbiArgs := abi.Arguments{
		{
			Type: validatorsDataType,
		},
	}

	aggG1SigBytes, err := g1PointAbiArgs.Pack(dtoG1AggSig)
	if err != nil {
		return entity.AggregationProof{}, err
	}

	aggG2KeyBytes, err := g2PointAbiArgs.Pack(dtoG2AggKey)
	if err != nil {
		return entity.AggregationProof{}, err
	}

	nonSignersBytes := make([]byte, 0, len(nonSigners)*2)
	for _, nonSigner := range nonSigners {
		bidEndianBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(bidEndianBytes, uint16(nonSigner))
		nonSignersBytes = append(nonSignersBytes, bidEndianBytes...)
	}

	validatorsDataBytes, err := validatorsDataAbiArgs.Pack(validatorsData)
	if err != nil {
		return entity.AggregationProof{}, err
	}

	proofBytes := bytes.Clone(aggG1SigBytes)
	proofBytes = append(proofBytes, aggG2KeyBytes...)
	proofBytes = append(proofBytes, validatorsDataBytes[32:]...)
	proofBytes = append(proofBytes, nonSignersBytes...)

	return entity.AggregationProof{
		Proof:            proofBytes,
		MessageHash:      messageHash,
		VerificationType: entity.VerificationTypeBlsBn254Simple,
	}, nil
}

func (a Aggregator) Verify(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof entity.AggregationProof,
) (bool, error) {
	g2Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256[2]"},
		{Name: "Y", Type: "uint256[2]"},
	})
	if err != nil {
		return false, err
	}

	g1Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256"},
		{Name: "Y", Type: "uint256"},
	})
	if err != nil {
		return false, err
	}

	validatorsDataType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "keySerialized", Type: "bytes32"},
		{Name: "VotingPower", Type: "uint256"},
	})
	if err != nil {
		return false, err
	}

	g1PointAbiArgs := abi.Arguments{
		{
			Type: g1Type,
		},
	}

	g2PointAbiArgs := abi.Arguments{
		{
			Type: g2Type,
		},
	}

	validatorsDataAbiArgs := abi.Arguments{
		{
			Type: validatorsDataType,
		},
	}

	offset := 0
	length := 64
	aggG1SigTuple, err := g1PointAbiArgs.Unpack(aggregationProof.Proof[offset : offset+length])
	if err != nil {
		return false, err
	}
	offset += length

	length = 128
	aggG2KeyTuple, err := g2PointAbiArgs.Unpack(aggregationProof.Proof[offset : offset+length])
	if err != nil {
		return false, err
	}
	offset += length
	lengthBig := new(big.Int).SetBytes(aggregationProof.Proof[offset : offset+32])
	length = 32 + 64*int(lengthBig.Int64())
	validatorsDataBytes := make([]byte, 32, 32+length)
	validatorsDataBytes[31] = 32
	validatorsDataBytes = append(validatorsDataBytes, aggregationProof.Proof[offset:offset+length]...)
	validatorsDataRaw, err := validatorsDataAbiArgs.Unpack(validatorsDataBytes)
	if err != nil {
		return false, err
	}
	offset += length

	isNonSignersRaw := aggregationProof.Proof[offset:]

	validatorsData := validatorsDataRaw[0].([]struct {
		KeySerialized [32]byte `json:"keySerialized"`
		VotingPower   *big.Int `json:"VotingPower"`
	})

	nonSignersMap := make(map[uint16]bool)
	for i := 0; i < len(isNonSignersRaw); i += 2 {
		val := binary.BigEndian.Uint16(isNonSignersRaw[i : i+2])
		nonSignersMap[val] = true
	}

	aggG1SigData := aggG1SigTuple[0].(struct {
		X *big.Int `json:"X"`
		Y *big.Int `json:"Y"`
	})

	aggSig := new(bn254.G1Affine)
	aggSig.X.SetBigInt(aggG1SigData.X)
	aggSig.Y.SetBigInt(aggG1SigData.Y)

	aggG2KeyData := aggG2KeyTuple[0].(struct {
		X [2]*big.Int `json:"X"`
		Y [2]*big.Int `json:"Y"`
	})
	aggPubKeyG2 := new(bn254.G2Affine)
	aggPubKeyG2.X.A0.SetBigInt(aggG2KeyData.X[1])
	aggPubKeyG2.X.A1.SetBigInt(aggG2KeyData.X[0])
	aggPubKeyG2.Y.A0.SetBigInt(aggG2KeyData.Y[1])
	aggPubKeyG2.Y.A1.SetBigInt(aggG2KeyData.Y[0])

	valsetSorted := make([]entity.Validator, 0, len(valset.Validators))
	for _, val := range valset.Validators {
		if val.IsActive {
			valsetSorted = append(valsetSorted, val)
		}
	}
	if len(valsetSorted) != len(validatorsData) {
		return false, errors.Errorf("active validators length mismatch: got %d, expected %d", len(valsetSorted), len(validatorsData))
	}

	sort.Slice(valsetSorted, func(i, j int) bool {
		// Compare keys (lower first)
		keyBytes1, ok := valsetSorted[i].FindKeyByKeyTag(keyTag)
		if !ok {
			return false
		}
		g1Key1, err := bls.DeserializeG1(keyBytes1)
		if err != nil {
			return false
		}
		g1Compressed1, err := bls.Compress(g1Key1)
		if err != nil {
			return false
		}
		keyBytes2, ok := valsetSorted[j].FindKeyByKeyTag(keyTag)
		if !ok {
			return false
		}
		g1Key2, err := bls.DeserializeG1(keyBytes2)
		if err != nil {
			return false
		}
		g1Compressed2, err := bls.Compress(g1Key2)
		if err != nil {
			return false
		}
		return g1Compressed1.Cmp(g1Compressed2) < 0
	})

	aggPubKeyG1 := bls.ZeroG1()
	var signersVotingPower big.Int
	for i, val := range valsetSorted {
		keyBytes, ok := val.FindKeyByKeyTag(keyTag)
		if !ok {
			return false, errors.Errorf("keyTag not found for validator %s", val.Operator.Hex())
		}
		g1Key, err := bls.DeserializeG1(keyBytes)
		if err != nil {
			return false, errors.Errorf("failed to deserialize G1 key from valset: %w", err)
		}
		g1, err := bls.Decompress(validatorsData[i].KeySerialized)
		if err != nil {
			return false, errors.Errorf("failed to decompress G1 key from valset: %w", err)
		}
		if g1Key.X.BigInt(new(big.Int)).Cmp(g1.X.BigInt(new(big.Int))) != 0 ||
			g1Key.Y.BigInt(new(big.Int)).Cmp(g1.Y.BigInt(new(big.Int))) != 0 {
			return false, errors.Errorf("mismatch in validator G1 pubkey for val %s idx %d", val.Operator.Hex(), i)
		}
		if val.VotingPower.Cmp(validatorsData[i].VotingPower) != 0 {
			return false, errors.Errorf("voting power mismatch for val %s", val.Operator.Hex())
		}
		if !nonSignersMap[uint16(i)] {
			aggPubKeyG1 = aggPubKeyG1.Add(g1Key)
			signersVotingPower.Add(&signersVotingPower, val.VotingPower.Int)
		}
	}

	if signersVotingPower.Cmp(valset.QuorumThreshold.Int) < 0 {
		return false, errors.Errorf("signers do not meet threshold voting power (%s < %s)", signersVotingPower.String(), valset.QuorumThreshold.String())
	}

	if len(aggregationProof.MessageHash) != 32 {
		return false, errors.New("message hash must be 32 bytes")
	}

	messageHashG1, err := bls.HashToG1(aggregationProof.MessageHash)
	if err != nil {
		return false, errors.Errorf("failed to hash message to G1: %w", err)
	}

	aggPubKeyG1XBytes := make([]byte, 32)
	aggPubKeyG1YBytes := make([]byte, 32)
	aggPubKeyG1.X.BigInt(new(big.Int)).FillBytes(aggPubKeyG1XBytes)
	aggPubKeyG1.Y.BigInt(new(big.Int)).FillBytes(aggPubKeyG1YBytes)
	aggPubKeyG2X0Bytes := make([]byte, 32)
	aggPubKeyG2X1Bytes := make([]byte, 32)
	aggPubKeyG2Y0Bytes := make([]byte, 32)
	aggPubKeyG2Y1Bytes := make([]byte, 32)
	aggPubKeyG2.X.A0.BigInt(new(big.Int)).FillBytes(aggPubKeyG2X0Bytes)
	aggPubKeyG2.X.A1.BigInt(new(big.Int)).FillBytes(aggPubKeyG2X1Bytes)
	aggPubKeyG2.Y.A0.BigInt(new(big.Int)).FillBytes(aggPubKeyG2Y0Bytes)
	aggPubKeyG2.Y.A1.BigInt(new(big.Int)).FillBytes(aggPubKeyG2Y1Bytes)
	aggSigXBytes := make([]byte, 32)
	aggSigYBytes := make([]byte, 32)
	aggSig.X.BigInt(new(big.Int)).FillBytes(aggSigXBytes)
	aggSig.Y.BigInt(new(big.Int)).FillBytes(aggSigYBytes)

	alpha := new(big.Int).SetBytes(
		crypto.Keccak256(
			aggregationProof.MessageHash,
			aggPubKeyG1XBytes,
			aggPubKeyG1YBytes,
			aggPubKeyG2X0Bytes,
			aggPubKeyG2X1Bytes,
			aggPubKeyG2Y0Bytes,
			aggPubKeyG2Y1Bytes,
			aggSigXBytes,
			aggSigYBytes,
		),
	)
	alpha = new(big.Int).Mod(alpha, bls.FrModulus)
	_, _, g1, g2 := bn254.Generators()
	negG2 := new(bn254.G2Affine).Neg(&g2)

	p := [2]bn254.G1Affine{
		*new(bn254.G1Affine).Add(aggSig, new(bn254.G1Affine).ScalarMultiplication(aggPubKeyG1.G1Affine, alpha)),
		*new(bn254.G1Affine).Add(messageHashG1.G1Affine, new(bn254.G1Affine).ScalarMultiplication(&g1, alpha)),
	}
	q := [2]bn254.G2Affine{*negG2, *aggPubKeyG2}

	ok, err := bn254.PairingCheck(p[:], q[:])
	if err != nil {
		return false, errors.Errorf("pairing check failed: %w", err)
	}
	if !ok {
		return false, errors.New("pairing check failed")
	}

	return true, nil
}

func (a Aggregator) GenerateExtraData(valset entity.ValidatorSet, keyTags []entity.KeyTag) ([]entity.ExtraData, error) {
	extraData := make([]entity.ExtraData, 0)

	totalActiveVotingPowerKey, err := helpers.GetExtraDataKey(entity.VerificationTypeBlsBn254Simple, entity.SimpleVerificationTotalVotingPowerHash)
	if err != nil {
		return nil, errors.Errorf("failed to get extra data key: %w", err)
	}

	totalActiveVotingPower := valset.GetTotalActiveVotingPower()
	totalActiveVotingPowerBytes32 := common.Hash{}
	totalActiveVotingPower.FillBytes(totalActiveVotingPowerBytes32[:])
	extraData = append(extraData, entity.ExtraData{
		Key:   totalActiveVotingPowerKey,
		Value: totalActiveVotingPowerBytes32,
	})

	aggregatedPubKeys := helpers.GetAggregatedPubKeys(valset, keyTags)

	// pack keccak accumulators per keyTag
	for _, key := range aggregatedPubKeys {
		validatorSetHashKey, err := helpers.GetExtraDataKeyTagged(entity.VerificationTypeBlsBn254Simple, key.Tag, entity.SimpleVerificationValidatorSetHashKeccak256Hash)
		if err != nil {
			return nil, errors.Errorf("failed to get extra data key: %w", err)
		}

		keccakHashAccumulator, err := calcKeccakAccumulator(valset.Validators, key.Tag)
		if err != nil {
			return nil, errors.Errorf("failed to generate validator set MiMC accumulator: %w", err)
		}

		extraData = append(extraData, entity.ExtraData{
			Key:   validatorSetHashKey,
			Value: keccakHashAccumulator,
		})
	}

	// pack aggregated keys per keyTag
	for _, activeAggregatedKey := range aggregatedPubKeys {
		activeAggregatedKeyKey, err := helpers.GetExtraDataKeyTagged(entity.VerificationTypeBlsBn254Simple, activeAggregatedKey.Tag, entity.SimpleVerificationAggPublicKeyG1Hash)
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

	return extraData, nil
}

func calcKeccakAccumulator(validators []entity.Validator, requiredKeyTag entity.KeyTag) (common.Hash, error) {
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
					return common.Hash{}, errors.Errorf("failed to deserialize G1: %w", err)
				}

				compressedKeyG1, err := bls.Compress(validatorKeyG1)
				if err != nil {
					return [32]byte{}, errors.Errorf("failed to compress G1: %w", err)
				}

				votingPower := validatorVotingPower

				validatorsData = append(validatorsData, validatorDataTuple{
					KeySerialized: compressedKeyG1,
					VotingPower:   votingPower.Int,
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
		return common.Hash{}, errors.Errorf("failed to pack arguments: %w", err)
	}
	hash := crypto.Keccak256Hash(packed[32:])
	return hash, nil
}
