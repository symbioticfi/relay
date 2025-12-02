package blsBn254Simple

import (
	"bytes"
	"context"
	"encoding/binary"
	"math/big"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator/helpers"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
)

const maxValidators = 65_536

type abiTypes struct {
	g1Type             abi.Type
	g2Type             abi.Type
	validatorsDataType abi.Type
	g1Args             abi.Arguments
	g2Args             abi.Arguments
	validatorsArgs     abi.Arguments
}

type Aggregator struct {
	abiTypes abiTypes
}

type ValidatorData struct {
	KeySerialized common.Hash
	VotingPower   *big.Int
}

func NewAggregator() (*Aggregator, error) {
	abis, err := createABITypes()
	if err != nil {
		return nil, err
	}
	return &Aggregator{
		abiTypes: abis,
	}, nil
}

// createABITypes creates and returns all ABI types
func createABITypes() (abiTypes, error) {
	g1Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256"},
		{Name: "Y", Type: "uint256"},
	})
	if err != nil {
		return abiTypes{}, err
	}

	g2Type, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256[2]"},
		{Name: "Y", Type: "uint256[2]"},
	})
	if err != nil {
		return abiTypes{}, err
	}

	validatorsDataType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "keySerialized", Type: "bytes32"},
		{Name: "VotingPower", Type: "uint256"},
	})
	if err != nil {
		return abiTypes{}, err
	}

	return abiTypes{
		g1Type:             g1Type,
		g2Type:             g2Type,
		validatorsDataType: validatorsDataType,
		g1Args:             abi.Arguments{{Type: g1Type}},
		g2Args:             abi.Arguments{{Type: g2Type}},
		validatorsArgs:     abi.Arguments{{Type: validatorsDataType}},
	}, nil
}

func (a Aggregator) Aggregate(ctx context.Context, valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, messageHash []byte, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error) {
	_, span := tracing.StartSpan(ctx, "aggregator.Aggregate",
		tracing.AttrEpoch.Int64(int64(valset.Epoch)),
		tracing.AttrValidatorCount.Int(len(valset.Validators)),
		tracing.AttrSignatureCount.Int(len(signatures)),
		tracing.AttrKeyTag.String(keyTag.String()),
		tracing.AttrProofType.String("bls-bn254-simple"),
	)
	defer span.End()

	tracing.AddEvent(span, "validating_inputs")
	if !helpers.CompareMessageHasher(signatures, messageHash) {
		err := errors.New("message hashes mismatch")
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, err
	}
	if err := valset.Validators.CheckIsSortedByOperatorAddressAsc(); err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("valset is not sorted by operator address asc: %w", err)
	}

	tracing.AddEvent(span, "processing_validators")
	validatorsData, err := processValidators(valset.Validators, keyTag)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, err
	}

	nonSigners := make([]int, 0)
	signersMap := make(map[common.Hash]struct{})

	aggG1Sig := new(bn254.G1Affine)
	aggG2Key := new(bn254.G2Affine)

	valKeysToIdx := helpers.GetValidatorsIndexesMapByKey(valset, keyTag)

	tracing.AddEvent(span, "aggregating_signatures")
	for _, sig := range signatures {
		pubKey, err := blsBn254.FromRaw(sig.PublicKey.Raw())
		if err != nil {
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, err
		}

		idx, ok := valKeysToIdx[string(pubKey.OnChain())]
		if !ok {
			err := errors.New("failed to find validator by key")
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, err
		}

		val := valset.Validators[idx]
		if !val.IsActive {
			continue
		}

		g1Key := new(bn254.G1Affine)
		_, err = g1Key.SetBytes(pubKey.OnChain())
		if err != nil {
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, err
		}

		compressedKeyG1, err := compress(g1Key)
		if err != nil {
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, errors.Errorf("failed to compress G1 key: %w", err)
		}

		if _, exists := signersMap[compressedKeyG1]; exists {
			err := errors.Errorf("duplicate signature from validator")
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, err
		}
		signersMap[compressedKeyG1] = struct{}{}

		g1Sig := new(bn254.G1Affine)
		_, err = g1Sig.SetBytes(sig.Signature)
		if err != nil {
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, err
		}

		aggG1Sig = aggG1Sig.Add(aggG1Sig, g1Sig)
		aggG2Key = aggG2Key.Add(aggG2Key, pubKey.G2())
	}

	tracing.AddEvent(span, "identifying_non_signers")
	for i, val := range validatorsData {
		if _, isSigner := signersMap[val.KeySerialized]; !isSigner {
			nonSigners = append(nonSigners, i)
		}
	}

	tracing.SetAttributes(span, tracing.AttrValidatorCount.Int(len(nonSigners)))
	tracing.AddEvent(span, "packing_proof")

	aggG1SigBytes, err := a.abiTypes.g1Args.Pack(struct {
		X *big.Int
		Y *big.Int
	}{
		X: aggG1Sig.X.BigInt(new(big.Int)),
		Y: aggG1Sig.Y.BigInt(new(big.Int)),
	})
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, err
	}

	aggG2KeyBytes, err := a.abiTypes.g2Args.Pack(struct {
		X [2]*big.Int
		Y [2]*big.Int
	}{
		X: [2]*big.Int{
			aggG2Key.X.A1.BigInt(new(big.Int)), // index 0
			aggG2Key.X.A0.BigInt(new(big.Int)), // index 1
		},
		Y: [2]*big.Int{
			aggG2Key.Y.A1.BigInt(new(big.Int)), // index 0
			aggG2Key.Y.A0.BigInt(new(big.Int)), // index 1
		},
	})
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, err
	}

	// Pack validators data with anonymous structs
	validatorsDataBytes, err := a.packValidatorsData(validatorsData)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, err
	}

	// Encode non-signers indices
	nonSignersBytes := make([]byte, 0, len(nonSigners)*2)
	for _, nonSigner := range nonSigners {
		bidEndianBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(bidEndianBytes, uint16(nonSigner))
		nonSignersBytes = append(nonSignersBytes, bidEndianBytes...)
	}

	tracing.AddEvent(span, "assembling_proof")
	// Assemble proof
	proofBytes := bytes.Clone(aggG1SigBytes)
	proofBytes = append(proofBytes, aggG2KeyBytes...)
	proofBytes = append(proofBytes, validatorsDataBytes[32:]...)
	proofBytes = append(proofBytes, nonSignersBytes...)

	tracing.SetAttributes(span, tracing.AttrProofSize.Int(len(proofBytes)))
	tracing.AddEvent(span, "aggregation_completed")

	return symbiotic.AggregationProof{
		MessageHash: messageHash,
		KeyTag:      keyTag,
		Epoch:       valset.Epoch,
		Proof:       proofBytes,
	}, nil
}

func (a Aggregator) Verify(
	ctx context.Context,
	valset symbiotic.ValidatorSet,
	keyTag symbiotic.KeyTag,
	aggregationProof symbiotic.AggregationProof,
) (bool, error) {
	_, span := tracing.StartSpan(ctx, "aggregator.Verify",
		tracing.AttrEpoch.Int64(int64(valset.Epoch)),
		tracing.AttrValidatorCount.Int(len(valset.Validators)),
		tracing.AttrKeyTag.String(keyTag.String()),
		tracing.AttrProofType.String("bls-bn254-simple"),
		tracing.AttrProofSize.Int(len(aggregationProof.Proof)),
	)
	defer span.End()

	tracing.AddEvent(span, "validating_inputs")
	// Check key tag type
	if keyTag.Type() != symbiotic.KeyTypeBlsBn254 {
		err := errors.New("unsupported key tag")
		tracing.RecordError(span, err)
		return false, err
	}

	if len(aggregationProof.MessageHash) != 32 {
		err := errors.New("aggregation proof message hash has invalid length")
		tracing.RecordError(span, err)
		return false, err
	}

	if len(aggregationProof.Proof) < 224 {
		err := errors.New("aggregation proof is too short")
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "parsing_proof_components")
	// Parse proof components
	offset := 0
	length := 64

	rawAggSig, err := a.abiTypes.g1Args.Unpack(aggregationProof.Proof[offset : offset+length])
	if err != nil {
		tracing.RecordError(span, err)
		return false, err
	}

	aggG1Data := rawAggSig[0].(struct {
		X *big.Int `json:"X"`
		Y *big.Int `json:"Y"`
	})

	aggSig := new(bn254.G1Affine)
	aggSig.X.SetBigInt(aggG1Data.X)
	aggSig.Y.SetBigInt(aggG1Data.Y)

	offset += length
	length = 128

	rawAggPubKeyG2, err := a.abiTypes.g2Args.Unpack(aggregationProof.Proof[offset : offset+length])
	if err != nil {
		return false, err
	}

	aggG2Data := rawAggPubKeyG2[0].(struct {
		X [2]*big.Int `json:"X"`
		Y [2]*big.Int `json:"Y"`
	})

	aggPubKeyG2 := new(bn254.G2Affine)
	aggPubKeyG2.X.A0.SetBigInt(aggG2Data.X[1])
	aggPubKeyG2.X.A1.SetBigInt(aggG2Data.X[0])
	aggPubKeyG2.Y.A0.SetBigInt(aggG2Data.Y[1])
	aggPubKeyG2.Y.A1.SetBigInt(aggG2Data.Y[0])

	offset += length

	// Parse validators data length
	lengthBig := new(big.Int).SetBytes(aggregationProof.Proof[offset : offset+32])
	if lengthBig.Uint64() == 0 {
		return false, nil
	}

	validatorsDataLength := int(lengthBig.Int64())
	if validatorsDataLength > maxValidators {
		return false, errors.New("too many validators")
	}

	// Calculate non-signers offset
	nonSignersOffset := 224 + validatorsDataLength*64
	if len(aggregationProof.Proof) < nonSignersOffset {
		return false, errors.New("proof too short for validators data")
	}

	// Verify validators data hash matches
	length = 32 + 64*int(lengthBig.Int64())
	validatorsDataBytes := make([]byte, 32, 32+length)
	validatorsDataBytes[31] = 32
	validatorsDataBytes = append(validatorsDataBytes, aggregationProof.Proof[offset:offset+length]...)

	expectedValidatorsData, err := processValidators(valset.Validators, keyTag)
	if err != nil {
		return false, err
	}

	expectedKeccak, err := a.calculateValidatorsKeccak(expectedValidatorsData)
	if err != nil {
		return false, err
	}

	localKeccak := crypto.Keccak256Hash(validatorsDataBytes[32:])
	if expectedKeccak.Cmp(localKeccak) != 0 {
		return false, nil
	}

	// Parse non-signers with proper validation
	nonSignersRaw := aggregationProof.Proof[nonSignersOffset:]
	nonSignersLength := len(nonSignersRaw) / 2

	// Validate proof length matches expected non-signers data
	if len(aggregationProof.Proof) != nonSignersOffset+nonSignersLength*2 {
		return false, errors.New("invalid proof length")
	}

	rawValidatorsData, err := a.abiTypes.validatorsArgs.Unpack(validatorsDataBytes)
	if err != nil {
		return false, err
	}

	validatorsRaw := rawValidatorsData[0].([]struct {
		KeySerialized [32]byte `json:"keySerialized"`
		VotingPower   *big.Int `json:"VotingPower"`
	})

	validatorsData := make([]ValidatorData, len(validatorsRaw))
	for i, v := range validatorsRaw {
		validatorsData[i] = ValidatorData{
			KeySerialized: v.KeySerialized,
			VotingPower:   v.VotingPower,
		}
	}

	// Parse and validate non-signers with ordering check
	nonSignersMap := make(map[uint16]bool)
	var nonSignersVotingPower big.Int
	var nonSignersPublicKeyG1 *bn254.G1Affine
	nonSignersPublicKeyG1 = new(bn254.G1Affine) // Initialize to zero point

	var prevNonSignerIndex uint16
	for i := 0; i < nonSignersLength; i++ {
		currentNonSignerIndex := binary.BigEndian.Uint16(nonSignersRaw[i*2 : (i+1)*2])

		// Validate non-signer index
		if currentNonSignerIndex >= uint16(validatorsDataLength) {
			return false, errors.New("invalid non-signer index")
		}

		// Check ordering (must be ascending)
		if i > 0 && prevNonSignerIndex >= currentNonSignerIndex {
			return false, errors.New("invalid non-signers order")
		}

		nonSignersMap[currentNonSignerIndex] = true

		// Add non-signer's voting power
		nonSignersVotingPower.Add(&nonSignersVotingPower, validatorsData[currentNonSignerIndex].VotingPower)

		// Add non-signer's public key
		g1Key, err := decompress(validatorsData[currentNonSignerIndex].KeySerialized)
		if err != nil {
			return false, errors.Errorf("failed to decompress non-signer G1 key: %w", err)
		}
		nonSignersPublicKeyG1 = nonSignersPublicKeyG1.Add(nonSignersPublicKeyG1, g1Key)

		prevNonSignerIndex = currentNonSignerIndex
	}

	// Verify validators match expected
	if len(expectedValidatorsData) != len(validatorsData) {
		return false, errors.Errorf("active validators length mismatch: got %d, expected %d", len(expectedValidatorsData), len(validatorsData))
	}

	for i, expectedVal := range expectedValidatorsData {
		// Verify validator data matches
		if expectedVal.KeySerialized != validatorsData[i].KeySerialized {
			return false, errors.Errorf("mismatch in validator key at index %d", i)
		}
		if expectedVal.VotingPower.Cmp(validatorsData[i].VotingPower) != 0 {
			return false, errors.Errorf("voting power mismatch at index %d", i)
		}
	}

	tracing.AddEvent(span, "checking_quorum")
	// Check quorum using the same logic as Solidity
	totalActiveVotingPower := valset.GetTotalActiveVotingPower()
	signersVotingPower := new(big.Int).Sub(totalActiveVotingPower.Int, &nonSignersVotingPower)

	if valset.QuorumThreshold.Cmp(signersVotingPower) > 0 {
		err := errors.Errorf("signers do not meet threshold voting power (%s < %s)", signersVotingPower.String(), valset.QuorumThreshold.String())
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "verifying_bls_signature")
	// Get aggregated public key from valset (equivalent to extra data in Solidity)
	aggregatedPubKeys := helpers.GetAggregatedPubKeys(valset, []symbiotic.KeyTag{keyTag})
	if len(aggregatedPubKeys) == 0 {
		return false, errors.New("no aggregated public key found")
	}

	aggPubKeyG1Raw := new(bn254.G1Affine)
	_, err = aggPubKeyG1Raw.SetBytes(aggregatedPubKeys[0].Payload)
	if err != nil {
		return false, errors.Errorf("failed to deserialize aggregated G1 key: %w", err)
	}

	// Calculate effective public key: aggPubKeyG1 - nonSignersPublicKeyG1
	// This matches the Solidity logic: aggPubKeyG1.plus(nonSignersPublicKeyG1.negate())
	negNonSignersKey := new(bn254.G1Affine).Neg(nonSignersPublicKeyG1)
	effectivePubKeyG1 := new(bn254.G1Affine).Add(aggPubKeyG1Raw, negNonSignersKey)

	// Verify signature using BLS verification
	if len(aggregationProof.MessageHash) != 32 {
		return false, errors.New("message hash must be 32 bytes")
	}

	messageHashG1, err := blsBn254.HashToG1(aggregationProof.MessageHash)
	if err != nil {
		return false, errors.Errorf("failed to hash message to G1: %w", err)
	}

	// Prepare bytes for alpha calculation
	alpha := calcAlpha(effectivePubKeyG1, aggPubKeyG2, aggSig, aggregationProof.MessageHash)

	_, _, g1, g2 := bn254.Generators()
	negG2 := new(bn254.G2Affine).Neg(&g2)

	p := [2]bn254.G1Affine{
		*new(bn254.G1Affine).Add(aggSig, new(bn254.G1Affine).ScalarMultiplication(effectivePubKeyG1, alpha)),
		*new(bn254.G1Affine).Add(messageHashG1, new(bn254.G1Affine).ScalarMultiplication(&g1, alpha)),
	}
	q := [2]bn254.G2Affine{*negG2, *aggPubKeyG2}

	ok, err := bn254.PairingCheck(p[:], q[:])
	if err != nil {
		tracing.RecordError(span, err)
		return false, errors.Errorf("pairing check failed: %w", err)
	}
	if !ok {
		err := errors.New("pairing check failed")
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "verification_successful")
	return true, nil
}

func calcAlpha(aggPubKeyG1 *bn254.G1Affine, aggPubKeyG2 *bn254.G2Affine, aggSig *bn254.G1Affine, messageHash []byte) *big.Int {
	alphaBytes := make([][]byte, 0, 9)

	alphaBytes = append(alphaBytes, messageHash)

	// G1 public key bytes
	aggPubKeyG1XBytes := make([]byte, 32)
	aggPubKeyG1YBytes := make([]byte, 32)
	aggPubKeyG1.X.BigInt(new(big.Int)).FillBytes(aggPubKeyG1XBytes)
	aggPubKeyG1.Y.BigInt(new(big.Int)).FillBytes(aggPubKeyG1YBytes)
	alphaBytes = append(alphaBytes, aggPubKeyG1XBytes, aggPubKeyG1YBytes)

	// G2 public key bytes
	aggPubKeyG2X0Bytes := make([]byte, 32)
	aggPubKeyG2X1Bytes := make([]byte, 32)
	aggPubKeyG2Y0Bytes := make([]byte, 32)
	aggPubKeyG2Y1Bytes := make([]byte, 32)
	aggPubKeyG2.X.A0.BigInt(new(big.Int)).FillBytes(aggPubKeyG2X0Bytes)
	aggPubKeyG2.X.A1.BigInt(new(big.Int)).FillBytes(aggPubKeyG2X1Bytes)
	aggPubKeyG2.Y.A0.BigInt(new(big.Int)).FillBytes(aggPubKeyG2Y0Bytes)
	aggPubKeyG2.Y.A1.BigInt(new(big.Int)).FillBytes(aggPubKeyG2Y1Bytes)
	alphaBytes = append(alphaBytes, aggPubKeyG2X0Bytes, aggPubKeyG2X1Bytes, aggPubKeyG2Y0Bytes, aggPubKeyG2Y1Bytes)

	// Signature bytes
	aggSigXBytes := make([]byte, 32)
	aggSigYBytes := make([]byte, 32)
	aggSig.X.BigInt(new(big.Int)).FillBytes(aggSigXBytes)
	aggSig.Y.BigInt(new(big.Int)).FillBytes(aggSigYBytes)
	alphaBytes = append(alphaBytes, aggSigXBytes, aggSigYBytes)

	alpha := new(big.Int).SetBytes(crypto.Keccak256(alphaBytes...))
	alpha = new(big.Int).Mod(alpha, fr.Modulus())

	return alpha
}

func (a Aggregator) GenerateExtraData(ctx context.Context, valset symbiotic.ValidatorSet, keyTags []symbiotic.KeyTag) ([]symbiotic.ExtraData, error) {
	_, span := tracing.StartSpan(ctx, "GenerateExtraData",
		tracing.AttrEpoch.Int64(int64(valset.Epoch)),
		tracing.AttrValidatorCount.Int(len(valset.Validators)),
	)
	defer span.End()

	extraData := make([]symbiotic.ExtraData, 0)

	aggregatedPubKeys := helpers.GetAggregatedPubKeys(valset, keyTags)
	tracing.SetAttributes(span, tracing.AttrKeyTag.Int(len(aggregatedPubKeys)))

	for _, key := range aggregatedPubKeys {
		tracing.AddEvent(span, "processing_key_tag", tracing.AttrKeyTag.String(key.Tag.String()))

		validatorsData, err := processValidators(valset.Validators, key.Tag)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, errors.Errorf("failed to encode validators: %w", err)
		}

		validatorSetHashKey, err := helpers.GetExtraDataKeyTagged(symbiotic.VerificationTypeBlsBn254Simple, key.Tag, symbiotic.SimpleVerificationValidatorSetHashKeccak256Hash)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, errors.Errorf("failed to get extra data key: %w", err)
		}

		tracing.AddEvent(span, "calculating_validators_keccak")
		keccakHashAccumulator, err := a.calculateValidatorsKeccak(validatorsData)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, errors.Errorf("failed to generate validator set keccak accumulator: %w", err)
		}

		extraData = append(extraData, symbiotic.ExtraData{
			Key:   validatorSetHashKey,
			Value: keccakHashAccumulator,
		})

		// Pack aggregated keys
		activeAggregatedKeyKey, err := helpers.GetExtraDataKeyTagged(symbiotic.VerificationTypeBlsBn254Simple, key.Tag, symbiotic.SimpleVerificationAggPublicKeyG1Hash)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, errors.Errorf("failed to get extra data key: %w", err)
		}

		keyG1Raw := new(bn254.G1Affine)
		_, err = keyG1Raw.SetBytes(key.Payload)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, errors.Errorf("failed to deserialize G1: %w", err)
		}

		tracing.AddEvent(span, "compressing_g1_key")
		compressedG1, err := compress(keyG1Raw)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, errors.Errorf("failed to compress G1: %w", err)
		}

		extraData = append(extraData, symbiotic.ExtraData{
			Key:   activeAggregatedKeyKey,
			Value: compressedG1,
		})

		tracing.AddEvent(span, "key_tag_processed")
	}

	tracing.AddEvent(span, "sorting_extra_data")
	// sort extra data by key to ensure deterministic order
	sort.Slice(extraData, func(i, j int) bool {
		return bytes.Compare(extraData[i].Key[:], extraData[j].Key[:]) < 0
	})

	tracing.SetAttributes(span, tracing.AttrKeyTag.Int(len(extraData)))
	return extraData, nil
}

func (a Aggregator) packValidatorsData(validatorsData []ValidatorData) ([]byte, error) {
	abiData := make([]struct {
		KeySerialized [32]byte
		VotingPower   *big.Int
	}, len(validatorsData))

	for i, v := range validatorsData {
		abiData[i].KeySerialized = v.KeySerialized
		abiData[i].VotingPower = v.VotingPower
	}

	return a.abiTypes.validatorsArgs.Pack(abiData)
}

func processValidators(validators []symbiotic.Validator, keyTag symbiotic.KeyTag) ([]ValidatorData, error) {
	validatorsData := make([]ValidatorData, 0, len(validators))

	for _, val := range validators {
		if !val.IsActive {
			continue
		}

		keyBytes, ok := val.FindKeyByKeyTag(keyTag)
		if !ok {
			return nil, errors.Errorf("failed to find key by keyTag for validator %s", val.Operator.Hex())
		}

		g1Key := new(bn254.G1Affine)
		_, err := g1Key.SetBytes(keyBytes)
		if err != nil {
			return nil, errors.Errorf("failed to deserialize G1 key: %w", err)
		}

		compressedKeyG1, err := compress(g1Key)
		if err != nil {
			return nil, errors.Errorf("failed to compress G1 key: %w", err)
		}

		validatorsData = append(validatorsData, ValidatorData{
			KeySerialized: compressedKeyG1,
			VotingPower:   val.VotingPower.Int,
		})
	}

	sort.Slice(validatorsData, func(i, j int) bool {
		return validatorsData[i].KeySerialized.Cmp(validatorsData[j].KeySerialized) < 0
	})

	return validatorsData, nil
}

func (a Aggregator) calculateValidatorsKeccak(validatorsData []ValidatorData) (common.Hash, error) {
	packed, err := a.packValidatorsData(validatorsData)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(packed[32:]), nil
}

func compress(g1 *bn254.G1Affine) (common.Hash, error) {
	x := g1.X.BigInt(new(big.Int))
	y := g1.Y.BigInt(new(big.Int))
	derivedY, err := findYFromX(x)
	if err != nil {
		return common.Hash{}, errors.New("failed to find Y from X")
	}

	flag := y.Cmp(derivedY) != 0
	compressedKeyG1 := new(big.Int).Mul(x, big.NewInt(2))
	if flag {
		compressedKeyG1.Add(compressedKeyG1, big.NewInt(1))
	}

	compressedKeyG1Bytes := [32]byte{}
	compressedKeyG1.FillBytes(compressedKeyG1Bytes[:])

	return compressedKeyG1Bytes, nil
}

func decompress(compressed [32]byte) (*bn254.G1Affine, error) {
	x, flag := new(big.Int).DivMod(new(big.Int).SetBytes(compressed[:32]), big.NewInt(2), big.NewInt(2))
	y, err := findYFromX(x)
	if err != nil {
		return nil, err
	}
	g1 := new(bn254.G1Affine)
	g1.X.SetBigInt(x)
	g1.Y.SetBigInt(y)
	if flag.Cmp(big.NewInt(1)) == 0 {
		g1.Neg(g1)
	}

	return g1, nil
}

func findYFromX(x *big.Int) (y *big.Int, err error) {
	fpModulus := fp.Modulus()

	// Calculate beta = x^3 + 3 mod p
	beta := new(big.Int).Exp(x, big.NewInt(3), fpModulus) // x^3
	beta.Add(beta, big.NewInt(3))                         // x^3 + 3
	beta.Mod(beta, fpModulus)                             // (x^3 + 3) mod p

	// Calculate y = beta^((p+1)/4) mod p
	exponent, success := new(big.Int).SetString("c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52", 16)
	if !success {
		return nil, errors.New("blsBn254: failed to set exponent")
	}

	y = new(big.Int).Exp(beta, exponent, fpModulus)

	return y, nil
}
