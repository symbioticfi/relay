package aggregator

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
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

type prover interface {
	Prove(proveInput proof.ProveInput) (proof.ProofData, error)
	Verify(valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error)
}

type Aggregator struct {
	zkProver prover
}

func NewAggregator(prover prover) *Aggregator {
	return &Aggregator{
		zkProver: prover,
	}
}

func (a *Aggregator) Aggregate(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	verificationType entity.VerificationType,
	messageHash []byte,
	signatures []entity.SignatureExtended,
) (entity.AggregationProof, error) {
	if !compareMessageHasher(signatures, messageHash) {
		return entity.AggregationProof{}, errors.New("message hashes mismatch")
	}

	switch verificationType {
	case entity.VerificationTypeZK:
		return a.zkAggregate(valset, keyTag, messageHash, signatures)
	case entity.VerificationTypeSimple:
		return a.simpleAggregate(valset, keyTag, messageHash, signatures)
	}
	return entity.AggregationProof{}, errors.New("unknown verification type")
}

func compareMessageHasher(signatures []entity.SignatureExtended, msgHash []byte) bool {
	for i := range signatures {
		if !bytes.Equal(msgHash, signatures[i].MessageHash) {
			return false
		}
	}
	return true
}

func (a *Aggregator) Verify(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof entity.AggregationProof,
) (bool, error) {
	switch aggregationProof.VerificationType {
	case entity.VerificationTypeZK:
		return a.zkVerify(valset, keyTag, aggregationProof)
	case entity.VerificationTypeSimple:
		return a.simpleVerify(valset, keyTag, aggregationProof)
	}
	return false, errors.New("unknown verification type")
}

func (a *Aggregator) zkAggregate(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	messageHash []byte,
	signatures []entity.SignatureExtended,
) (entity.AggregationProof, error) {
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

	var validatorsData []proof.ValidatorData
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

			validatorsData = append(validatorsData, proof.ValidatorData{
				Key:         *g1Key.G1Affine,
				IsNonSigner: !isSinger,
				VotingPower: val.VotingPower.Int,
			})
		}
	}

	messageG1, err := bls.HashToG1(messageHash)
	if err != nil {
		return entity.AggregationProof{}, err
	}
	messageG1Bn254 := bn254.G1Affine{X: messageG1.X, Y: messageG1.Y}

	proverInput := proof.ProveInput{
		ValidatorData:   proof.NormalizeValset(validatorsData),
		MessageG1:       messageG1Bn254,
		Signature:       *aggG1Sig.G1Affine,
		SignersAggKeyG2: *aggG2Key.G2Affine,
	}
	proofData, err := a.zkProver.Prove(proverInput)
	if err != nil {
		return entity.AggregationProof{}, err
	}

	return entity.AggregationProof{
		VerificationType: entity.VerificationTypeZK,
		MessageHash:      messageHash,
		Proof:            proofData.Marshal(),
	}, nil
}

func (a *Aggregator) simpleAggregate(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	messageHash []byte,
	signatures []entity.SignatureExtended,
) (entity.AggregationProof, error) {
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
				return entity.AggregationProof{}, fmt.Errorf("failed to deserialize G1 key: %w", err)
			}

			compressedKeyG1, err := bls.Compress(g1Key)
			if err != nil {
				return entity.AggregationProof{}, fmt.Errorf("failed to compress G1 key: %w", err)
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
		littleEndianBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(littleEndianBytes, uint16(nonSigner))
		nonSignersBytes = append(nonSignersBytes, littleEndianBytes...)
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
		VerificationType: entity.VerificationTypeSimple,
	}, nil
}

func (a *Aggregator) zkVerify(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof entity.AggregationProof,
) (bool, error) {
	activeVals := 0
	for _, val := range valset.Validators {
		if val.IsActive {
			activeVals++
		}
	}

	mimcAccum, err := validatorSetMimcAccumulator(valset.Validators, keyTag)
	if err != nil {
		return false, err
	}
	// last 32 bytes is aggVotingPowerBytes
	aggVotingPowerBytes := aggregationProof.Proof[len(aggregationProof.Proof)-32:]

	messageG1, err := bls.HashToG1(aggregationProof.MessageHash)
	if err != nil {
		return false, errors.Errorf("failed to hash message to G1: %w", err)
	}
	messageG1Bytes := messageG1.RawBytes() // non compressed

	inpBytes := mimcAccum[:]
	inpBytes = append(inpBytes, aggVotingPowerBytes...)
	inpBytes = append(inpBytes, messageG1Bytes[:]...)
	inpHash := crypto.Keccak256Hash(inpBytes)

	ok, err := a.zkProver.Verify(activeVals, inpHash, aggregationProof.Proof)
	if err != nil {
		return false, err
	}

	aggVotingPower := new(big.Int).SetBytes(aggVotingPowerBytes)
	if aggVotingPower.Cmp(valset.QuorumThreshold.Int) < 0 {
		return false, fmt.Errorf("agg voting power %s is less than quorum threshold %s", aggVotingPower.String(), valset.QuorumThreshold.String())
	}

	return ok, nil
}

func (a *Aggregator) simpleVerify(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof entity.AggregationProof,
) (bool, error) {
	// TODO fix and prettify
	return true, nil
	//g2Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
	//	{Name: "X", Type: "uint256[2]"},
	//	{Name: "Y", Type: "uint256[2]"},
	//})
	//if err != nil {
	//	return false, err
	//}
	//
	//g1Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
	//	{Name: "X", Type: "uint256"},
	//	{Name: "Y", Type: "uint256"},
	//})
	//if err != nil {
	//	return false, err
	//}
	//
	//validatorsDataType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
	//	{Name: "KeyCompressed", Type: "bytes32"},
	//	{Name: "VotingPower", Type: "uint256"},
	//})
	//if err != nil {
	//	return false, err
	//}
	//
	//isNonSignersType, err := abi.NewType("bool[]", "", []abi.ArgumentMarshaling{})
	//if err != nil {
	//	return false, err
	//}
	//
	//g1PointAbiArgs := abi.Arguments{
	//	{
	//		Type: g1Type,
	//	},
	//}
	//
	//g2PointAbiArgs := abi.Arguments{
	//	{
	//		Type: g2Type,
	//	},
	//}
	//
	//validatorsDataAbiArgs := abi.Arguments{
	//	{
	//		Type: validatorsDataType,
	//	},
	//}
	//
	//isNonSignersAbiArgs := abi.Arguments{
	//	{
	//		Type: isNonSignersType,
	//	},
	//}
	//
	//offset := 0
	//length := 64
	//aggG1SigTuple, err := g1PointAbiArgs.Unpack(aggregationProof.Proof[offset : offset+length])
	//if err != nil {
	//	return false, err
	//}
	//offset += length
	//
	//length = 128
	//aggG2KeyTuple, err := g2PointAbiArgs.Unpack(aggregationProof.Proof[offset : offset+length])
	//if err != nil {
	//	return false, err
	//}
	//offset += length
	//
	//lengthBig := new(big.Int).SetBytes(aggregationProof.Proof[offset+32 : offset+64])
	//length = 64 + 96*int(lengthBig.Int64())
	//validatorsDataRaw, err := validatorsDataAbiArgs.Unpack(aggregationProof.Proof[offset : offset+length])
	//if err != nil {
	//	return false, err
	//}
	//offset += length
	//
	//length = 64 + 32*int(lengthBig.Int64())
	//isNonSignersRaw, err := isNonSignersAbiArgs.Unpack(aggregationProof.Proof[offset : offset+length])
	//if err != nil {
	//	return false, err
	//}
	//offset += length
	//
	//if offset != len(aggregationProof.Proof) {
	//	return false, errors.Errorf("length mismatch in aggregation proof unpacking: expected %d, got %d", len(aggregationProof.Proof), offset)
	//}
	//
	//validatorsData := validatorsDataRaw[0].([]struct {
	//	G1PubKey struct {
	//		X *big.Int `json:"X"`
	//		Y *big.Int `json:"Y"`
	//	} `json:"G1PubKey"`
	//	VotingPower *big.Int `json:"VotingPower"`
	//})
	//
	//isNonSigners := isNonSignersRaw[0].([]bool)
	//
	//aggG1SigData := aggG1SigTuple[0].(struct {
	//	X *big.Int `json:"X"`
	//	Y *big.Int `json:"Y"`
	//})
	//
	//aggSig := new(bn254.G1Affine)
	//aggSig.X.SetBigInt(aggG1SigData.X)
	//aggSig.Y.SetBigInt(aggG1SigData.Y)
	//
	//aggG2KeyData := aggG2KeyTuple[0].(struct {
	//	X [2]*big.Int `json:"X"`
	//	Y [2]*big.Int `json:"Y"`
	//})
	//aggPubKeyG2 := new(bn254.G2Affine)
	//aggPubKeyG2.X.A0.SetBigInt(aggG2KeyData.X[1])
	//aggPubKeyG2.X.A1.SetBigInt(aggG2KeyData.X[0])
	//aggPubKeyG2.Y.A0.SetBigInt(aggG2KeyData.Y[1])
	//aggPubKeyG2.Y.A1.SetBigInt(aggG2KeyData.Y[0])
	//
	//valsetSorted := make([]entity.Validator, 0, len(valset.Validators))
	//for _, val := range valset.Validators {
	//	if val.IsActive {
	//		valsetSorted = append(valsetSorted, val)
	//	}
	//}
	//if len(valsetSorted) != len(validatorsData) {
	//	return false, errors.Errorf("active validators length mismatch: got %d, expected %d", len(valsetSorted), len(validatorsData))
	//}
	//
	//sort.Slice(valsetSorted, func(i, j int) bool {
	//	// Compare keys (lower first)
	//	keyBytes1, ok := valsetSorted[i].FindKeyByKeyTag(keyTag)
	//	if !ok {
	//		return false
	//	}
	//	g1Key1, err := bls.DeserializeG1(keyBytes1)
	//	if err != nil {
	//		return false
	//	}
	//	keyBytes2, ok := valsetSorted[j].FindKeyByKeyTag(keyTag)
	//	if !ok {
	//		return false
	//	}
	//	g1Key2, err := bls.DeserializeG1(keyBytes2)
	//	if err != nil {
	//		return false
	//	}
	//	return g1Key1.X.BigInt(new(big.Int)).Cmp(g1Key2.X.BigInt(new(big.Int))) < 0 || g1Key1.Y.BigInt(new(big.Int)).Cmp(g1Key2.Y.BigInt(new(big.Int))) < 0
	//})
	//
	//aggPubKeyG1 := new(bn254.G1Affine)
	//var signersVotingPower big.Int
	//for i, val := range valsetSorted {
	//	keyBytes, ok := val.FindKeyByKeyTag(keyTag)
	//	if !ok {
	//		return false, fmt.Errorf("keyTag not found for validator %s", val.Operator.Hex())
	//	}
	//	g1Key, err := bls.DeserializeG1(keyBytes)
	//	if err != nil {
	//		return false, fmt.Errorf("failed to deserialize G1 key from valset: %w", err)
	//	}
	//	if g1Key.X.BigInt(new(big.Int)).Cmp(validatorsData[i].G1PubKey.X) != 0 ||
	//		g1Key.Y.BigInt(new(big.Int)).Cmp(validatorsData[i].G1PubKey.Y) != 0 {
	//		return false, fmt.Errorf("mismatch in validator G1 pubkey for val %s", val.Operator.Hex())
	//	}
	//	if val.VotingPower.Cmp(validatorsData[i].VotingPower) != 0 {
	//		return false, fmt.Errorf("voting power mismatch for val %s", val.Operator.Hex())
	//	}
	//	if !isNonSigners[i] {
	//		aggPubKeyG1 = new(bn254.G1Affine).Add(aggPubKeyG1, g1Key.G1Affine)
	//		signersVotingPower.Add(&signersVotingPower, val.VotingPower)
	//	}
	//}
	//
	//if signersVotingPower.Cmp(valset.QuorumThreshold) < 0 {
	//	return false, errors.Errorf("signers do not meet threshold voting power (%s < %s)", signersVotingPower.String(), valset.QuorumThreshold.String())
	//}
	//
	//if len(aggregationProof.MessageHash) != 32 {
	//	return false, errors.New("message hash must be 32 bytes")
	//}
	//
	//messageHashG1, err := bls.HashToG1(aggregationProof.MessageHash)
	//if err != nil {
	//	return false, errors.Errorf("failed to hash message to G1: %w", err)
	//}
	//
	//aggPubKeyG1XBytes := make([]byte, 32)
	//aggPubKeyG1YBytes := make([]byte, 32)
	//aggPubKeyG1.X.BigInt(new(big.Int)).FillBytes(aggPubKeyG1XBytes)
	//aggPubKeyG1.Y.BigInt(new(big.Int)).FillBytes(aggPubKeyG1YBytes)
	//aggPubKeyG2X0Bytes := make([]byte, 32)
	//aggPubKeyG2X1Bytes := make([]byte, 32)
	//aggPubKeyG2Y0Bytes := make([]byte, 32)
	//aggPubKeyG2Y1Bytes := make([]byte, 32)
	//aggPubKeyG2.X.A0.BigInt(new(big.Int)).FillBytes(aggPubKeyG2X0Bytes)
	//aggPubKeyG2.X.A1.BigInt(new(big.Int)).FillBytes(aggPubKeyG2X1Bytes)
	//aggPubKeyG2.Y.A0.BigInt(new(big.Int)).FillBytes(aggPubKeyG2Y0Bytes)
	//aggPubKeyG2.Y.A1.BigInt(new(big.Int)).FillBytes(aggPubKeyG2Y1Bytes)
	//aggSigXBytes := make([]byte, 32)
	//aggSigYBytes := make([]byte, 32)
	//aggSig.X.BigInt(new(big.Int)).FillBytes(aggSigXBytes)
	//aggSig.Y.BigInt(new(big.Int)).FillBytes(aggSigYBytes)
	//
	//alpha := new(big.Int).SetBytes(
	//	crypto.Keccak256(
	//		aggregationProof.MessageHash,
	//		aggPubKeyG1XBytes,
	//		aggPubKeyG1YBytes,
	//		aggPubKeyG2X0Bytes,
	//		aggPubKeyG2X1Bytes,
	//		aggPubKeyG2Y0Bytes,
	//		aggPubKeyG2Y1Bytes,
	//		aggSigXBytes,
	//		aggSigYBytes,
	//	),
	//)
	//alpha = new(big.Int).Mod(alpha, bls.FrModulus)
	//
	//_, _, g1, g2 := bn254.Generators()
	//negG2 := new(bn254.G2Affine).Neg(&g2)
	//
	//p := [2]bn254.G1Affine{
	//	*new(bn254.G1Affine).Add(aggSig, new(bn254.G1Affine).ScalarMultiplication(aggPubKeyG1, alpha)),
	//	*new(bn254.G1Affine).Add(messageHashG1.G1Affine, new(bn254.G1Affine).ScalarMultiplication(&g1, alpha)),
	//}
	//q := [2]bn254.G2Affine{*negG2, *aggPubKeyG2}
	//
	//ok, err := bn254.PairingCheck(p[:], q[:])
	//if err != nil {
	//	return false, errors.Errorf("pairing check failed: %w", err)
	//}
	//if !ok {
	//	return false, errors.New("pairing check failed")
	//}
	//
	//return true, nil
}
