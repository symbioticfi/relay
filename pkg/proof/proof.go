//nolint:forbidigo // ignore this linter for now todo ilya
package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	mimc_native "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/solidity"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bn254"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/ethereum/go-ethereum/common"
)

var (
	MaxValidators = []int{10, 50}
)

const (
	circuitsDir = "circuits"
)

const (
	r1csPathTmp = circuitsDir + "/circuit_%s.r1cs"
	pkPathTmp   = circuitsDir + "/circuit_%s.pk"
	vkPathTmp   = circuitsDir + "/circuit_%s.vk"
	solPathTmp  = circuitsDir + "/Verifier_%s.sol"
)

type ValidatorDataCircuit struct {
	Key         sw_emulated.AffinePoint[emulated.BN254Fp]
	VotingPower frontend.Variable
	IsNonSigner frontend.Variable
}

type ValidatorData struct {
	Key         bn254.G1Affine
	VotingPower *big.Int
	IsNonSigner bool
}

// Circuit defines a pre-image knowledge proof
type Circuit struct {
	NonSignersAggKey         sw_emulated.AffinePoint[emulated.BN254Fp] `gnark:",public"`
	Hash                     frontend.Variable                         `gnark:",public"`
	NonSignersAggVotingPower frontend.Variable                         `gnark:",public"`
	ValidatorData            []ValidatorDataCircuit                    `gnark:",private"`
	ZeroPoint                sw_emulated.AffinePoint[emulated.BN254Fp] `gnark:",private"`
}

// Define declares the circuit's constraints
func (circuit *Circuit) Define(api frontend.API) error {
	curve, err := sw_emulated.New[emulated.BN254Fp, emulated.BN254Fr](api, sw_emulated.GetBN254Params())
	if err != nil {
		return err
	}

	field, err := emulated.NewField[emulated.BN254Fp](api)
	if err != nil {
		return err
	}

	// check if zero point is zero
	api.AssertIsEqual(field.IsZero(&circuit.ZeroPoint.X), 1)
	api.AssertIsEqual(field.IsZero(&circuit.ZeroPoint.Y), 1)

	aggKey := &circuit.ZeroPoint
	aggVotingPower := frontend.Variable(0)

	mimcOuter, _ := mimc.NewMiMC(api)
	mimcInner, _ := mimc.NewMiMC(api)

	for i := 0; i < len(circuit.ValidatorData); i++ {
		mimcInner.Reset()
		xVar := field.ToBits(&circuit.ValidatorData[i].Key.X)
		yVar := field.ToBits(&circuit.ValidatorData[i].Key.Y)
		mimcInner.Write(api.FromBinary(xVar...))
		mimcInner.Write(api.FromBinary(yVar...))
		mimcInner.Write(circuit.ValidatorData[i].VotingPower)
		mimcOuter.Write(mimcInner.Sum())

		// get power if non-signer otherwise 0
		pow := api.Select(circuit.ValidatorData[i].IsNonSigner, circuit.ValidatorData[i].VotingPower, frontend.Variable(0))
		aggVotingPower = api.Add(aggVotingPower, pow)

		// get key if non-signer otherwise zero point
		point := curve.Select(circuit.ValidatorData[i].IsNonSigner, &circuit.ValidatorData[i].Key, &circuit.ZeroPoint)
		aggKey = curve.AddUnified(aggKey, point)
	}

	curve.AssertIsEqual(aggKey, &circuit.NonSignersAggKey)
	api.AssertIsEqual(aggVotingPower, circuit.NonSignersAggVotingPower)
	api.AssertIsEqual(circuit.Hash, mimcOuter.Sum())

	return nil
}

// helper functions
func getPubkey(pk *big.Int) bn254.G1Affine {
	_, _, g1Aff, _ := bn254.Generators()
	var p bn254.G1Affine
	p.ScalarMultiplication(&g1Aff, pk)
	return p
}

func HashValset(valset []ValidatorData) []byte {
	outerHash := mimc_native.NewMiMC()
	for i := 0; i < len(valset); i++ {
		innerHash := mimc_native.NewMiMC()
		xBytes := valset[i].Key.X.Bytes()
		yBytes := valset[i].Key.Y.Bytes()

		innerHash.Write(xBytes[:])
		innerHash.Write(yBytes[:])
		votingPowerBuf := make([]byte, 32)
		valset[i].VotingPower.FillBytes(votingPowerBuf)
		innerHash.Write(votingPowerBuf)

		outerHash.Write(innerHash.Sum(nil))
	}
	return outerHash.Sum(nil)
}

func getNonSignersData(valset []ValidatorData) (aggKey *bn254.G1Affine, aggVotingPower *big.Int) {
	aggVotingPower = big.NewInt(0)
	aggKey = new(bn254.G1Affine)
	aggKey.SetInfinity()
	for i := 0; i < len(valset); i++ {
		if valset[i].IsNonSigner {
			aggKey = aggKey.Add(aggKey, &(valset)[i].Key)
			aggVotingPower = aggVotingPower.Add(aggVotingPower, valset[i].VotingPower)
		}
	}
	return aggKey, aggVotingPower
}

func setCircuitData(circuit *Circuit, valset []ValidatorData) {
	circuit.ValidatorData = make([]ValidatorDataCircuit, len(valset))
	for i := 0; i < len(valset); i++ {
		circuit.ValidatorData[i].Key = sw_bn254.NewG1Affine(valset[i].Key)
		circuit.ValidatorData[i].VotingPower = valset[i].VotingPower
		circuit.ValidatorData[i].IsNonSigner = *big.NewInt(0)

		if valset[i].IsNonSigner {
			circuit.ValidatorData[i].IsNonSigner = *big.NewInt(1)
		}
	}

	aggKey, aggVotingPower := getNonSignersData(valset)
	circuit.NonSignersAggKey = sw_bn254.NewG1Affine(*aggKey)
	circuit.NonSignersAggVotingPower = *aggVotingPower
	circuit.Hash = HashValset(valset)
	zeroPoint := new(bn254.G1Affine)
	zeroPoint.SetInfinity()
	circuit.ZeroPoint = sw_bn254.NewG1Affine(*zeroPoint)
}

func DoProve(validators []entity.Validator, requiredKeyTag uint8) ([]byte, error) {
	data, err := ToValidatorsData(validators, requiredKeyTag)
	if err != nil {
		return nil, errors.Errorf("failed to convert validators to data: %w", err)
	}

	prove, err := Prove(data)
	if err != nil {
		return nil, errors.Errorf("failed to prove: %w", err)
	}

	return prove, nil
}

func ToValidatorsData(validators []entity.Validator, requiredKeyTag uint8) ([]ValidatorData, error) {
	valset := make([]ValidatorData, 0)
	for i := 0; i < len(validators); i++ {
		if !validators[i].IsActive {
			continue
		}
		for _, key := range validators[i].Keys {
			if key.Tag == requiredKeyTag {
				g1, err := bls.DeserializeG1(key.Payload)
				if err != nil {
					return nil, fmt.Errorf("failed to deserialize G1: %w", err)
				}
				valset = append(valset, ValidatorData{Key: *g1.G1Affine, VotingPower: validators[i].VotingPower})
			}
		}
	}
	return normalizeValset(valset), nil
}

// todo ilya
//func ToValidatorsData(signerValidators []entity.Validator, allValidators []entity.Validator, requiredKeyTag uint8) ([]ValidatorData, error) {
//	activeValidators := make([]entity.Validator, 0)
//	for _, validator := range allValidators {
//		if validator.IsActive {
//			activeValidators = append(activeValidators, validator)
//		}
//	}
//	valset := make([]ValidatorData, 0)
//	for i := 0; i < len(activeValidators); i++ {
//		for _, key := range activeValidators[i].Keys {
//			if key.Tag == requiredKeyTag {
//				g1, err := bls.DeserializeG1(key.Payload)
//				if err != nil {
//					return nil, fmt.Errorf("failed to deserialize G1: %w", err)
//				}
//				validatorData := ValidatorData{Key: *g1.G1Affine, VotingPower: activeValidators[i].VotingPower, IsNonSigner: true}
//
//				for _, signer := range signerValidators {
//					if signer.Operator.Cmp(activeValidators[i].Operator) == 0 {
//						validatorData.IsNonSigner = false
//					}
//				}
//
//				valset = append(valset, validatorData)
//			}
//		}
//	}
//	return normalizeValset(valset), nil
//}

func Prove(valset []ValidatorData) ([]byte, error) {
	r1cs, pk, vk, err := loadOrInit(valset)
	if err != nil {
		return nil, errors.Errorf("failed to load or init: %w", err)
	}

	// witness definition
	assignment := Circuit{}
	setCircuitData(&assignment, valset)
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()

	// groth16: Prove & Verify
	proof, err := groth16.Prove(r1cs, pk, witness, backend.WithProverHashToFieldFunction(sha256.New()))
	if err != nil {
		return nil, errors.Errorf("failed to prove: %w", err)
	}
	fmt.Println(proof.CurveID())

	publicInputs := publicWitness.Vector().(fr.Vector)
	// Format for the specific Solidity interface
	formattedInputs := make([]*big.Int, 0, len(publicInputs))

	// Format the vector of public inputs as hex strings
	for _, input := range publicInputs {
		formattedInputs = append(formattedInputs, new(big.Int).SetBytes(input.Marshal()))
	}

	// If more than 10 inputs (unlikely), you'll need to adapt the interface
	if len(formattedInputs) > 10 {
		fmt.Println("Warning: More public inputs than the interface supports")
	}

	_proof, ok := proof.(interface{ MarshalSolidity() []byte })
	if !ok {
		panic("proof does not implement MarshalSolidity()")
	}

	proofBytes := _proof.MarshalSolidity()
	fmt.Println(len(proofBytes))
	fmt.Println("Proof:", hex.EncodeToString(proofBytes))
	// verify proof
	err = groth16.Verify(proof, vk, publicWitness, backend.WithVerifierHashToFieldFunction(sha256.New()))
	if err != nil {
		return nil, err
	}

	// Serialize the proof
	var proofBuffer bytes.Buffer
	_, err = proof.WriteRawTo(&proofBuffer)
	if err != nil {
		return nil, errors.Errorf("failed to write proof: %w", err)
	}
	proofBytes = proofBuffer.Bytes()
	fmt.Println("proofBytes:", proofBytes) //nolint:staticcheck // will fix later
	fmt.Println("hex:", common.Bytes2Hex(proofBytes))

	// Assuming fpSize is 32 bytes for BN254
	const fpSize = 32

	standardProof := [8]*big.Int{}
	standardProof[0] = new(big.Int).SetBytes(proofBytes[fpSize*0 : fpSize*1]) // Ar.x
	standardProof[1] = new(big.Int).SetBytes(proofBytes[fpSize*1 : fpSize*2]) // Ar.y
	standardProof[2] = new(big.Int).SetBytes(proofBytes[fpSize*2 : fpSize*3]) // Bs.x[0]
	standardProof[3] = new(big.Int).SetBytes(proofBytes[fpSize*3 : fpSize*4]) // Bs.x[1]
	standardProof[4] = new(big.Int).SetBytes(proofBytes[fpSize*4 : fpSize*5]) // Bs.y[0]
	standardProof[5] = new(big.Int).SetBytes(proofBytes[fpSize*5 : fpSize*6]) // Bs.y[1]
	standardProof[6] = new(big.Int).SetBytes(proofBytes[fpSize*6 : fpSize*7]) // Krs.x
	standardProof[7] = new(big.Int).SetBytes(proofBytes[fpSize*7 : fpSize*8]) // Krs.y

	commitments := [2]*big.Int{}
	commitments[0] = new(big.Int).SetBytes(proofBytes[4+fpSize*8 : 4+fpSize*9])  // Commitment.x
	commitments[1] = new(big.Int).SetBytes(proofBytes[4+fpSize*9 : 4+fpSize*10]) // Commitment.y

	commitmentPok := [2]*big.Int{}
	commitmentPok[0] = new(big.Int).SetBytes(proofBytes[4+fpSize*10 : 4+fpSize*11]) // CommitmentPok.x
	commitmentPok[1] = new(big.Int).SetBytes(proofBytes[4+fpSize*11 : 4+fpSize*12]) // CommitmentPok.y

	fmt.Println("proof: ", standardProof)
	fmt.Println("commitments: ", commitments)
	fmt.Println("commitmentPok: ", commitmentPok)
	fmt.Println("inputs", formattedInputs)
	//// Extract public inputs
	//for i := 0; i < publicWitness.Vector(); i++ {
	//	val, _ := publicWitness.GetValue(i)
	//	publicInputs[i] = new(big.Int).SetBytes(val.Bytes())
	//}

	return proofBytes, nil
}

func loadOrInit(valset []ValidatorData) (constraint.ConstraintSystem, groth16.ProvingKey, groth16.VerifyingKey, error) {
	suffix := strconv.Itoa(len(valset))
	r1csP := fmt.Sprintf(r1csPathTmp, suffix)
	pkP := fmt.Sprintf(pkPathTmp, suffix)
	vkP := fmt.Sprintf(vkPathTmp, suffix)
	solP := fmt.Sprintf(solPathTmp, suffix)

	if exists(r1csP) && exists(pkP) && exists(vkP) && exists(solP) {
		r1csCS := groth16.NewCS(bn254.ID)
		data, _ := os.ReadFile(r1csP)
		r1csCS.ReadFrom(bytes.NewReader(data))
		pk := groth16.NewProvingKey(bn254.ID)
		data, _ = os.ReadFile(pkP)
		pk.UnsafeReadFrom(bytes.NewReader(data))
		vk := groth16.NewVerifyingKey(bn254.ID)
		data, _ = os.ReadFile(vkP)
		vk.UnsafeReadFrom(bytes.NewReader(data))

		return r1csCS, pk, vk, nil
	}

	if err := os.MkdirAll(circuitsDir, 0o755); err != nil {
		return nil, nil, nil, err
	}

	for _, m := range MaxValidators {
		suf := strconv.Itoa(m)
		r1csFile := fmt.Sprintf(r1csPathTmp, suf)
		pkFile := fmt.Sprintf(pkPathTmp, suf)
		vkFile := fmt.Sprintf(vkPathTmp, suf)
		solFile := fmt.Sprintf(solPathTmp, suf)

		if exists(r1csFile) && exists(pkFile) && exists(vkFile) && exists(solFile) {
			continue
		}

		circ := Circuit{
			ValidatorData: make([]ValidatorDataCircuit, m),
		}

		cs_i, err := frontend.Compile(bn254.ID.ScalarField(), r1cs.NewBuilder, &circ)
		if err != nil {
			return nil, nil, nil, err
		}
		pk_i, vk_i, err := groth16.Setup(cs_i)
		if err != nil {
			return nil, nil, nil, err
		}

		{
			var buf bytes.Buffer
			cs_i.WriteTo(&buf)
			os.WriteFile(r1csFile, buf.Bytes(), 0600)
		}
		{
			f, _ := os.Create(pkFile)
			pk_i.WriteRawTo(f)
			f.Close()
			f, _ = os.Create(vkFile)
			vk_i.WriteRawTo(f)
			f.Close()
		}
		{
			f, _ := os.Create(solFile)
			vk_i.ExportSolidity(f, solidity.WithHashToFieldFunction(sha256.New()))
			f.Close()
		}
	}

	return loadOrInit(valset)
}

func normalizeValset(valset []ValidatorData) []ValidatorData {
	// Sort validators by key in ascending order
	sort.Slice(valset, func(i, j int) bool {
		// Compare keys (lower first)
		return valset[i].Key.X.Cmp(&valset[j].Key.X) > 0 || valset[i].Key.Y.Cmp(&valset[j].Key.Y) > 0
	})
	n := getOptimalN(len(valset))
	normalizedValset := make([]ValidatorData, n)
	for i := 0; i < n; i++ {
		if i < len(valset) {
			normalizedValset[i] = valset[i]
		} else {
			zeroPoint := new(bn254.G1Affine)
			zeroPoint.SetInfinity()
			normalizedValset[i] = ValidatorData{Key: *zeroPoint, VotingPower: big.NewInt(0), IsNonSigner: false}
		}
	}
	return normalizedValset
}

func getOptimalN(valsetLength int) int {
	var capSize int
	for _, m := range MaxValidators {
		if m >= valsetLength {
			capSize = m
			break
		}
	}
	if capSize == 0 {
		return 0
	}
	return capSize
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

//0000000a000000000000000a
//248c8c7d61427e104037798d87f2f6744bd4c718 c1a38841625fb1c780
//13af800000000000000000000000000000000000000000
//00000000b649840c052bf8920000000000000000000000000000000000000000
//00000000d0da6e916f5f61710000000000000000000000000000000000000000
//00000000c357f5c82f2c87ab0000000000000000000000000000000000000000
//000000001bf3ebe16a0321c00000000000000000000000000000000000000000
//000000002a382506552049b10000000000000000000000000000000000000000
//000000000385b691c3bc64430000000000000000000000000000000000000000
//000000000472e0def08271b50000000000000000000000000000000000000000
//000000002cc236a9e084af730000000000000000000000000000000000000000
//00000000000000000000012c

//2d25f5c066d29834d177791291f349683e867d08995de3f5776d33c784bd001d
//0ce82dca365ebb9ea753db43eea1ff1f687c33c03844acb75fff0b0e3508579b
//0fcd9204c7be444da8c47fb0d0e60727d05cc52f0e99af07f1051f02984653cc
//0f0ffad297c637840b697f90e3fa2b0b35bf92981aab35a94f305e5c7f609ffc
//0862c1d904db6087742382c42c74664cf2f905427a1c057587b18cdbc156da4b
//14313b1c8e2a7d3e950ebf2a119e9328949a6c9ff1bce6e2a5f756ef973aed22
//1d858e4ea671962aa7e55da5bc42bd6acc78adf910534bdd416165340730c390
//204cf9f09f2c0187be1b086b1be3c57b3e0bee0d7f65e24516acf26896e836a0
//00000001
//0ee469efe3c0db390334c9ce5d35a9d5bd7da9a329067e9b77cb109b8050a49e
//024bf94511a482636b7a3c73fc4f331e6fb4707827c10e708f487497d4d8ba8a
//1ab00b01d82e838ea8d5440f5b5685915966e99ace64341e3d7133e2236d2484
//305c8ddbaf5c8ad20a1e0fb0943cf50a9a10296326fe59b9641f89ecb646f24a
