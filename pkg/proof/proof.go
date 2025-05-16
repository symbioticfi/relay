//nolint:forbidigo // ignore this linter for now todo ilya
package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/bits"
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
	MaxValidators = []int{10}
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
	Key         sw_bn254.G1Affine
	VotingPower frontend.Variable
	IsNonSigner frontend.Variable
}

type ValidatorData struct {
	PrivateKey  big.Int
	Key         bn254.G1Affine
	KeyG2       bn254.G2Affine
	VotingPower big.Int
	IsNonSigner bool
}

// Circuit defines a pre-image knowledge proof
type Circuit struct {
	Hash                     frontend.Variable      `gnark:",public"` // 254 bits
	NonSignersAggVotingPower frontend.Variable      `gnark:",public"` // 254 bits
	Message                  sw_bn254.G1Affine      `gnark:",public"`
	Signature                sw_bn254.G1Affine      `gnark:",private"`
	SignersAggKeyG2          sw_bn254.G2Affine      `gnark:",private"`
	ValidatorData            []ValidatorDataCircuit `gnark:",private"`
}

func hashAffineG1(h *mimc.MiMC, g1 *sw_bn254.G1Affine) {
	h.Write(g1.X.Limbs...)
	h.Write(g1.Y.Limbs...)
}

func hashAffineG2(h *mimc.MiMC, g2 *sw_bn254.G2Affine) {
	h.Write(g2.P.X.A0.Limbs...)
	h.Write(g2.P.X.A1.Limbs...)
	h.Write(g2.P.Y.A0.Limbs...)
	h.Write(g2.P.Y.A1.Limbs...)
}

// Define declares the circuit's constraints
func (circuit *Circuit) Define(api frontend.API) error {
	curve, err := sw_emulated.New[emulated.BN254Fp, emulated.BN254Fr](api, sw_emulated.GetBN254Params())
	if err != nil {
		return err
	}

	fieldFr, err := emulated.NewField[emulated.BN254Fr](api)
	if err != nil {
		return err
	}

	mimcOuter, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	mimcInner, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	signersAggKey := &sw_bn254.G1Affine{
		X: emulated.ValueOf[emulated.BN254Fp](0),
		Y: emulated.ValueOf[emulated.BN254Fp](0),
	}
	nonSignersAggVotingPower := frontend.Variable(0)

	// calc valset hash, agg key and agg voting power
	for i := range circuit.ValidatorData {
		mimcInner.Reset()
		hashAffineG1(&mimcInner, &circuit.ValidatorData[i].Key)
		mimcInner.Write(circuit.ValidatorData[i].VotingPower)
		mimcOuter.Write(mimcInner.Sum())

		// get power if NON-SIGNER otherwise 0
		pow := api.Select(circuit.ValidatorData[i].IsNonSigner, circuit.ValidatorData[i].VotingPower, frontend.Variable(0))
		nonSignersAggVotingPower = api.Add(nonSignersAggVotingPower, pow)

		// get key if SIGNER otherwise zero point
		point := curve.Select(api.IsZero(circuit.ValidatorData[i].IsNonSigner), &circuit.ValidatorData[i].Key, &sw_bn254.G1Affine{
			X: emulated.ValueOf[emulated.BN254Fp](0),
			Y: emulated.ValueOf[emulated.BN254Fp](0),
		})
		signersAggKey = curve.AddUnified(signersAggKey, point)
	}
	valsetHash := mimcOuter.Sum()

	// compare with public inputs
	api.AssertIsEqual(nonSignersAggVotingPower, circuit.NonSignersAggVotingPower)
	api.AssertIsEqual(valsetHash, circuit.Hash)

	// calc alpha
	mimcOuter.Reset()
	hashAffineG1(&mimcOuter, &circuit.Signature)
	hashAffineG1(&mimcOuter, signersAggKey)
	hashAffineG2(&mimcOuter, &circuit.SignersAggKeyG2)
	hashAffineG1(&mimcOuter, &circuit.Message)
	// TODO optimize
	alpha := fieldFr.FromBits(bits.ToBinary(api, mimcOuter.Sum())...)

	// pairing check
	_, _, g1Gen, g2Gen := bn254.Generators()
	g1GenAffine := sw_bn254.NewG1Affine(g1Gen)
	negG2GenAffine := sw_bn254.NewG2Affine(*g2Gen.Neg(&g2Gen))
	evmprecompiles.ECPair(api,
		[]*sw_bn254.G1Affine{
			curve.AddUnified(&circuit.Signature, curve.ScalarMul(signersAggKey, alpha)),
			curve.AddUnified(&circuit.Message, curve.ScalarMul(&g1GenAffine, alpha)),
		},
		[]*sw_bn254.G2Affine{
			&negG2GenAffine,
			&circuit.SignersAggKeyG2,
		},
	)

	return nil
}

// helper functions
func HashValset(valset *[]ValidatorData) []byte {
	outerHash := mimc_native.NewMiMC()
	for i := 0; i < len(*valset); i++ {
		innerHash := mimc_native.NewMiMC()
		xBytes := (*valset)[i].Key.X.Bytes()
		yBytes := (*valset)[i].Key.Y.Bytes()

		// hash by limbs as it's done inside circuit
		innerHash.Write(xBytes[24:32])
		innerHash.Write(xBytes[16:24])
		innerHash.Write(xBytes[8:16])
		innerHash.Write(xBytes[0:8])

		innerHash.Write(yBytes[24:32])
		innerHash.Write(yBytes[16:24])
		innerHash.Write(yBytes[8:16])
		innerHash.Write(yBytes[0:8])

		votingPowerBuf := make([]byte, 32)
		(*valset)[i].VotingPower.FillBytes(votingPowerBuf)
		innerHash.Write(votingPowerBuf)

		outerHash.Write(innerHash.Sum(nil))
	}
	return outerHash.Sum(nil)
}

func getPubkeyG1(pk *big.Int) bn254.G1Affine {
	_, _, g1Aff, _ := bn254.Generators()
	var p bn254.G1Affine
	p.ScalarMultiplication(&g1Aff, pk)
	return p
}

func getPubkeyG2(pk *big.Int) bn254.G2Affine {
	_, _, _, g2Aff := bn254.Generators()
	var p bn254.G2Affine
	p.ScalarMultiplication(&g2Aff, pk)
	return p
}

func getNonSignersData(valset *[]ValidatorData) (aggKey *bn254.G1Affine, aggVotingPower *big.Int) {
	aggVotingPower = big.NewInt(0)
	aggKey = new(bn254.G1Affine)
	aggKey.SetInfinity()
	for i := 0; i < len(*valset); i++ {
		if (*valset)[i].IsNonSigner {
			aggKey = aggKey.Add(aggKey, &(*valset)[i].Key)
			aggVotingPower = aggVotingPower.Add(aggVotingPower, &(*valset)[i].VotingPower)
		}
	}
	return aggKey, aggVotingPower
}

func getAggSignature(message bn254.G1Affine, valset *[]ValidatorData) (signature *bn254.G1Affine, aggKeyG2 *bn254.G2Affine, aggKeyG1 *bn254.G1Affine) {
	aggKeyG2 = new(bn254.G2Affine)
	aggKeyG2.SetInfinity()

	aggSignature := new(bn254.G1Affine)
	aggSignature.SetInfinity()

	aggKeyG1 = new(bn254.G1Affine)
	aggKeyG1.SetInfinity()

	for i := 0; i < len(*valset); i++ {
		if !(*valset)[i].IsNonSigner {
			aggKeyG2 = aggKeyG2.Add(aggKeyG2, &(*valset)[i].KeyG2)
			aggKeyG1 = aggKeyG1.Add(aggKeyG1, &(*valset)[i].Key)
			msg := bn254.G1Affine{X: message.X, Y: message.Y} // have to copy msg since ScalarMultiplication rewrite it
			sig := msg.ScalarMultiplication(&msg, &(*valset)[i].PrivateKey)
			aggSignature = aggSignature.Add(aggSignature, sig)
		}
	}

	return aggSignature, aggKeyG2, aggKeyG1
}

func setCircuitData(circuit *Circuit, valset *[]ValidatorData) {
	circuit.ValidatorData = make([]ValidatorDataCircuit, len(*valset))
	for i := 0; i < len(*valset); i++ {
		circuit.ValidatorData[i].Key = sw_bn254.NewG1Affine((*valset)[i].Key)
		circuit.ValidatorData[i].VotingPower = (*valset)[i].VotingPower
		circuit.ValidatorData[i].IsNonSigner = *big.NewInt(0)

		if (*valset)[i].IsNonSigner {
			circuit.ValidatorData[i].IsNonSigner = *big.NewInt(1)
		}
	}

	_, _, message, _ := bn254.Generators()
	message = *message.ScalarMultiplication(&message, big.NewInt(101))

	_, aggVotingPower := getNonSignersData(valset)
	aggSignature, aggKeyG2, _ := getAggSignature(message, valset)
	valsetHash := HashValset(valset)

	circuit.NonSignersAggVotingPower = *aggVotingPower
	circuit.Hash = valsetHash
	circuit.Signature = sw_bn254.NewG1Affine(*aggSignature)
	circuit.Message = sw_bn254.NewG1Affine(message)
	circuit.SignersAggKeyG2 = sw_bn254.NewG2Affine(*aggKeyG2)
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
				valset = append(valset, ValidatorData{Key: *g1.G1Affine, VotingPower: *validators[i].VotingPower})
			}
		}
	}
	return normalizeValset(valset), nil
}

func Prove(valset []ValidatorData) ([]byte, error) {
	r1cs, pk, vk, err := loadOrInit(valset)
	if err != nil {
		return nil, errors.Errorf("failed to load or init: %w", err)
	}

	// witness definition
	assignment := Circuit{}
	setCircuitData(&assignment, &valset)
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()

	fmt.Println()

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
			zeroPointG2 := new(bn254.G2Affine)
			zeroPointG2.SetInfinity()
			normalizedValset[i] = ValidatorData{Key: *zeroPoint, KeyG2: *zeroPointG2, VotingPower: *big.NewInt(0), IsNonSigner: false}
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
