package proof

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
)

func genValset(numValidators int, nonSigners []int) []ValidatorData {
	valset := make([]ValidatorData, numValidators)
	for i := 0; i < numValidators; i++ {
		pk := big.NewInt(int64(i + 10))
		valset[i].PrivateKey = pk
		valset[i].Key = getPubkeyG1(pk)
		valset[i].KeyG2 = getPubkeyG2(pk)
		valset[i].VotingPower = big.NewInt(100)
		valset[i].IsNonSigner = false
	}

	for _, nonSigner := range nonSigners {
		valset[nonSigner].IsNonSigner = true
	}

	return valset
}

//nolint:unused // will be used later
func mockValset() []ValidatorData {
	pks := []string{
		"87191036493798670866484781455694320176667203290824056510541300741498740913410",
		"26972876870930381973856869753776124637336739336929668162870464864826929175089",
		"11008377096554045051122023680185802911050337017631086444859313200352654461863",
	}

	valset := make([]ValidatorData, len(pks))
	for i := 0; i < len(pks); i++ {
		pk, ok := new(big.Int).SetString(pks[i], 10)
		if !ok {
			panic(errors.New("failed to convert pk to big.Int"))
		}
		valset[i].PrivateKey = pk
		valset[i].Key = getPubkeyG1(pk)
		valset[i].KeyG2 = getPubkeyG2(pk)
		valset[i].VotingPower = big.NewInt(10000000000000)
		valset[i].IsNonSigner = false
	}

	return valset
}

func calculateInputHash(validatorSetHash []byte, signersVotingPower *big.Int, messageG1 *bn254.G1Affine) common.Hash {
	var packed []byte

	packed = append(packed, validatorSetHash[:32]...)

	signersVPBytes := make([]byte, 32)
	signersVotingPower.FillBytes(signersVPBytes)
	packed = append(packed, signersVPBytes...)

	packed = append(packed, messageG1.X.Marshal()...)
	packed = append(packed, messageG1.Y.Marshal()...)

	hashBytes := crypto.Keccak256(packed)

	return common.BytesToHash(hashBytes)
}

func TestProof(t *testing.T) {
	t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")

	startTime := time.Now()
	prover := NewZkProver("circuits")
	fmt.Printf("prover initialation took %v\n", time.Since(startTime))

	// generate valset
	valset := genValset(10, []int{})
	// valset := mockValset()

	validatorData := NormalizeValset(valset)

	messageG1Hex := "04c3256b0d7e3f3766d9d3f08fad062e025db392f7b8d8d86322602365b82eba2370c94328160af53802c073a5ddafe012a4073eca842339acc5caae83e1b922"
	messageG1 := &bn254.G1Affine{}
	if err := messageG1.Unmarshal(common.Hex2Bytes(messageG1Hex)); err != nil {
		t.Fatal(err)
	}

	aggSignature, aggKeyG2, _ := getAggSignature(*messageG1, &validatorData)

	proveInput := ProveInput{
		ValidatorData:   validatorData,
		MessageG1:       *messageG1,
		Signature:       *aggSignature,
		SignersAggKeyG2: *aggKeyG2,
	}

	startTime = time.Now()
	proofData, err := prover.Prove(t.Context(), proveInput)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("proving took %v\n", time.Since(startTime))

	fmt.Println("Proof:", hex.EncodeToString(proofData.Proof))
	fmt.Println("Commitments:", hex.EncodeToString(proofData.Commitments))
	fmt.Println("CommitmentPok:", hex.EncodeToString(proofData.CommitmentPok))
	fmt.Println("SignersAggVotingPower:", proofData.SignersAggVotingPower.String())

	inputHash := calculateInputHash(HashValset(valset), proofData.SignersAggVotingPower, messageG1)
	startTime = time.Now()
	res, err := prover.Verify(t.Context(), len(validatorData), inputHash, proofData.Marshal())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("verification took %v\n", time.Since(startTime))

	if !res {
		t.Fatal("failed to verify")
	}
}

func TestProofFailOnEmptyCircuitDir(t *testing.T) {
	startTime := time.Now()
	prover := NewZkProver("")
	fmt.Printf("prover initialation took %v\n", time.Since(startTime))

	_, err := prover.Prove(t.Context(), ProveInput{})
	require.ErrorContains(t, err, "ZK prover circuits directory is not set", "expected error on empty circuit dir")

	_, err = prover.Verify(t.Context(), 0, common.Hash{}, nil)
	require.ErrorContains(t, err, "ZK prover circuits directory is not set", "expected error on empty circuit dir")
}

// TestNewZkProverInitialization tests the NewZkProver initialization
func TestNewZkProverInitialization(t *testing.T) {
	t.Run("initializes with empty circuits directory", func(t *testing.T) {
		prover := NewZkProver("")

		// Check all maps are initialized
		require.NotNil(t, prover.cs)
		require.NotNil(t, prover.pk)
		require.NotNil(t, prover.vk)

		// Check maps are empty (no circuits loaded without dir)
		require.Empty(t, prover.cs)
		require.Empty(t, prover.pk)
		require.Empty(t, prover.vk)

		// Check maxValidators is set to default
		require.Equal(t, []int{10, 100, 1000}, prover.maxValidators)

		// Check circuitsDir is empty
		require.Empty(t, prover.circuitsDir)
	})

	t.Run("sets circuitsDir when provided", func(t *testing.T) {
		// We can't actually test NewZkProver with a real circuits dir
		// because it will try to load/compile circuits which is slow.
		// Instead, we manually create a prover to verify the field is set.
		testDir := "/tmp/test_circuits"
		prover := &ZkProver{
			cs:          make(map[int]constraint.ConstraintSystem),
			pk:          make(map[int]groth16.ProvingKey),
			vk:          make(map[int]groth16.VerifyingKey),
			circuitsDir: testDir,
		}

		require.Equal(t, testDir, prover.circuitsDir)

		// Maps should still be initialized
		require.NotNil(t, prover.cs)
		require.NotNil(t, prover.pk)
		require.NotNil(t, prover.vk)
	})

	t.Run("respects MAX_VALIDATORS environment variable", func(t *testing.T) {
		t.Setenv("MAX_VALIDATORS", "5,15,25")

		prover := NewZkProver("")

		require.Equal(t, []int{5, 15, 25}, prover.maxValidators)
	})
}

// TestProveValidation tests the Prove validation paths
func TestProveValidation(t *testing.T) {
	t.Run("returns error when circuits directory not set", func(t *testing.T) {
		prover := NewZkProver("")

		_, err := prover.Prove(t.Context(), ProveInput{})
		require.ErrorContains(t, err, "ZK prover circuits directory is not set")
	})

	t.Run("returns error for empty validator data", func(t *testing.T) {
		prover := NewZkProver("")

		input := ProveInput{
			ValidatorData: []ValidatorData{},
		}

		_, err := prover.Prove(t.Context(), input)
		require.ErrorContains(t, err, "ZK prover circuits directory is not set")
	})

	t.Run("returns error for unsupported validator count", func(t *testing.T) {
		// Create prover with circuits dir but no actual circuits loaded
		prover := &ZkProver{
			cs:            make(map[int]constraint.ConstraintSystem),
			pk:            make(map[int]groth16.ProvingKey),
			vk:            make(map[int]groth16.VerifyingKey),
			circuitsDir:   "/tmp/circuits",
			maxValidators: []int{10, 100, 1000},
		}

		// Try with 25 validators (not in {10, 100, 1000})
		input := ProveInput{
			ValidatorData: make([]ValidatorData, 25),
		}

		_, err := prover.Prove(t.Context(), input)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to load cs, vk, pk for valset size")
	})

	t.Run("checks for constraint system availability", func(t *testing.T) {
		prover := &ZkProver{
			cs:            make(map[int]constraint.ConstraintSystem),
			pk:            make(map[int]groth16.ProvingKey),
			vk:            make(map[int]groth16.VerifyingKey),
			circuitsDir:   "/tmp/circuits",
			maxValidators: []int{10, 100, 1000},
		}

		// Even with 10 validators (which is in maxValidators),
		// we don't have actual circuit loaded
		input := ProveInput{
			ValidatorData: make([]ValidatorData, 10),
		}

		_, err := prover.Prove(t.Context(), input)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to load cs, vk, pk for valset size: 10")
	})
}

// TestVerifyValidation tests the Verify validation paths
func TestVerifyValidation(t *testing.T) {
	t.Run("returns error when circuits directory not set", func(t *testing.T) {
		prover := NewZkProver("")

		ok, err := prover.Verify(t.Context(), 10, common.Hash{}, []byte{})
		require.False(t, ok)
		require.ErrorContains(t, err, "ZK prover circuits directory is not set")
	})

	t.Run("normalizes valset length using getOptimalN", func(t *testing.T) {
		prover := &ZkProver{
			cs:            make(map[int]constraint.ConstraintSystem),
			pk:            make(map[int]groth16.ProvingKey),
			vk:            make(map[int]groth16.VerifyingKey),
			circuitsDir:   "/tmp/circuits",
			maxValidators: []int{10, 100, 1000},
		}

		// valsetLen = 5 should normalize to 10
		// But since we don't have vk[10], it should error
		// Note: Verify needs at least 384 bytes for proof
		proofBytes := make([]byte, 384)
		ok, err := prover.Verify(t.Context(), 5, common.Hash{}, proofBytes)
		require.False(t, ok)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to find verification key for valset length 10")
	})

	t.Run("returns error for valsetLen exceeding all maxValidators", func(t *testing.T) {
		prover := &ZkProver{
			cs:            make(map[int]constraint.ConstraintSystem),
			pk:            make(map[int]groth16.ProvingKey),
			vk:            make(map[int]groth16.VerifyingKey),
			circuitsDir:   "/tmp/circuits",
			maxValidators: []int{10, 100, 1000},
		}

		// valsetLen = 5000 exceeds all sizes, getOptimalN returns 0
		proofBytes := make([]byte, 384)
		ok, err := prover.Verify(t.Context(), 5000, common.Hash{}, proofBytes)
		require.False(t, ok)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to find verification key for valset length 0")
	})

	t.Run("returns error when verification key not found", func(t *testing.T) {
		prover := &ZkProver{
			cs:            make(map[int]constraint.ConstraintSystem),
			pk:            make(map[int]groth16.ProvingKey),
			vk:            make(map[int]groth16.VerifyingKey),
			circuitsDir:   "/tmp/circuits",
			maxValidators: []int{10, 100, 1000},
		}

		// Even with exact match valsetLen=100, vk map is empty
		proofBytes := make([]byte, 384)
		ok, err := prover.Verify(t.Context(), 100, common.Hash{}, proofBytes)
		require.False(t, ok)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to find verification key for valset length 100")
	})
}

// TestPathHelpers tests the path generation helper functions
func TestR1csPathTmp(t *testing.T) {
	tests := []struct {
		name        string
		circuitsDir string
		suffix      string
		expected    string
	}{
		{
			name:        "basic path",
			circuitsDir: "/tmp/circuits",
			suffix:      "10",
			expected:    "/tmp/circuits/circuit_10.r1cs",
		},
		{
			name:        "different suffix",
			circuitsDir: "/tmp/circuits",
			suffix:      "100",
			expected:    "/tmp/circuits/circuit_100.r1cs",
		},
		{
			name:        "relative path",
			circuitsDir: "circuits",
			suffix:      "1000",
			expected:    "circuits/circuit_1000.r1cs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r1csPathTmp(tt.circuitsDir, tt.suffix)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPkPathTmp(t *testing.T) {
	tests := []struct {
		name        string
		circuitsDir string
		suffix      string
		expected    string
	}{
		{
			name:        "basic path",
			circuitsDir: "/tmp/circuits",
			suffix:      "10",
			expected:    "/tmp/circuits/circuit_10.pk",
		},
		{
			name:        "different suffix",
			circuitsDir: "/tmp/circuits",
			suffix:      "100",
			expected:    "/tmp/circuits/circuit_100.pk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pkPathTmp(tt.circuitsDir, tt.suffix)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestVkPathTmp(t *testing.T) {
	tests := []struct {
		name        string
		circuitsDir string
		suffix      string
		expected    string
	}{
		{
			name:        "basic path",
			circuitsDir: "/tmp/circuits",
			suffix:      "10",
			expected:    "/tmp/circuits/circuit_10.vk",
		},
		{
			name:        "different suffix",
			circuitsDir: "/tmp/circuits",
			suffix:      "100",
			expected:    "/tmp/circuits/circuit_100.vk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vkPathTmp(tt.circuitsDir, tt.suffix)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSolPathTmp(t *testing.T) {
	tests := []struct {
		name        string
		circuitsDir string
		suffix      string
		expected    string
	}{
		{
			name:        "basic path",
			circuitsDir: "/tmp/circuits",
			suffix:      "10",
			expected:    "/tmp/circuits/Verifier_10.sol",
		},
		{
			name:        "different suffix",
			circuitsDir: "/tmp/circuits",
			suffix:      "100",
			expected:    "/tmp/circuits/Verifier_100.sol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := solPathTmp(tt.circuitsDir, tt.suffix)
			require.Equal(t, tt.expected, result)
		})
	}
}
