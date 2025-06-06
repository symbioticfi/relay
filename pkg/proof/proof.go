//nolint:forbidigo // ignore this linter for now todo ilya
package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/consensys/gnark/std/math/bits"
	"github.com/consensys/gnark/std/math/uints"
	"math/big"
	"os"
	"sort"
	"strconv"

	"github.com/go-errors/errors"

	"log/slog"
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
	gnarkSha3 "github.com/consensys/gnark/std/hash/sha3"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	//MaxValidators = []int{10, 100, 1000}
	MaxValidators = []int{10}
)

func InitCircuitsDir(newCircuitsDir string) {
	circuitsDir = newCircuitsDir
}

var (
	circuitsDir = "circuits"
)

func r1csPathTmp(suffix string) string {
	return fmt.Sprintf(circuitsDir+"/circuit_%s.r1cs", suffix)
}

func pkPathTmp(suffix string) string {
	return fmt.Sprintf(circuitsDir+"/circuit_%s.pk", suffix)
}

func vkPathTmp(suffix string) string {
	return fmt.Sprintf(circuitsDir+"/circuit_%s.vk", suffix)
}

func solPathTmp(suffix string) string {
	return fmt.Sprintf(circuitsDir+"/Verifier_%s.sol", suffix)
}

type ProofData struct {
	Proof                 []byte
	Commitments           []byte
	CommitmentPok         []byte
	SignersAggVotingPower *big.Int
}

func (p ProofData) Marshall() []byte {
	var result bytes.Buffer

	result.Write(p.Proof)
	result.Write(p.Commitments)
	result.Write(p.CommitmentPok)
	signersAggVotingPowerBuffer := make([]byte, 32)
	p.SignersAggVotingPower.FillBytes(signersAggVotingPowerBuffer)
	result.Write(signersAggVotingPowerBuffer)

	return result.Bytes()
}

type RawProveInput struct {
	SignerValidators []entity.Validator
	AllValidators    []entity.Validator
	RequiredKeyTag   entity.KeyTag
	Message          []byte
	Signature        bls.G1
	SignersAggKeyG2  bls.G2
}

type ProveInput struct {
	ValidatorData   []ValidatorData
	Message         []byte
	Signature       bn254.G1Affine
	SignersAggKeyG2 bn254.G2Affine
}

type ValidatorData struct {
	PrivateKey  *big.Int
	Key         bn254.G1Affine
	KeyG2       bn254.G2Affine
	VotingPower *big.Int
	IsNonSigner bool
}

type ValidatorDataCircuit struct {
	Key         sw_bn254.G1Affine
	VotingPower frontend.Variable
	IsNonSigner frontend.Variable
}

// Circuit defines a pre-image knowledge proof
type Circuit struct {
	InputHash             frontend.Variable      `gnark:",public"`  // 254 bits
	SignersAggVotingPower frontend.Variable      `gnark:",private"` // 254 bits, virtually public
	Message               sw_bn254.G1Affine      `gnark:",private"` // virtually public
	Signature             sw_bn254.G1Affine      `gnark:",private"`
	SignersAggKeyG2       sw_bn254.G2Affine      `gnark:",private"`
	ValidatorData         []ValidatorDataCircuit `gnark:",private"`
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

func variableToBytes(api frontend.API, u64api *uints.BinaryField[uints.U64], variable frontend.Variable) []uints.U8 {
	res := make([]uints.U8, 32)
	hexVar := bits.ToBinary(api, variable, bits.WithNbDigits(256))
	for i := range 32 {
		res[i] = u64api.ByteValueOf(api.Add(
			api.Mul(1<<7, hexVar[8*(32-i)-1]),
			api.Mul(1<<6, hexVar[8*(32-i)-2]),
			api.Mul(1<<5, hexVar[8*(32-i)-3]),
			api.Mul(1<<4, hexVar[8*(32-i)-4]),
			api.Mul(1<<3, hexVar[8*(32-i)-5]),
			api.Mul(1<<2, hexVar[8*(32-i)-6]),
			api.Mul(1<<1, hexVar[8*(32-i)-7]),
			api.Mul(1<<0, hexVar[8*(32-i)-8]),
		))
	}

	return res
}

func keyToBytes(u64api *uints.BinaryField[uints.U64], key *sw_bn254.G1Affine) []uints.U8 {
	xLimbs := key.X.Limbs
	yLimbs := key.Y.Limbs

	result := limbsToBytes(u64api, xLimbs)
	return append(result, limbsToBytes(u64api, yLimbs)...)
}

func limbsToBytes(u64api *uints.BinaryField[uints.U64], limbs []frontend.Variable) []uints.U8 {
	result := make([]uints.U8, 0, len(limbs)*8)
	for i := range limbs {
		u64 := u64api.ValueOf(limbs[len(limbs)-1-i])
		result = append(result, u64api.UnpackMSB(u64)...)
	}
	return result
}

// Define declares the circuit's constraints
func (circuit *Circuit) Define(api frontend.API) error {
	// --------------------------------------- Prove ValSet consistency ---------------------------------------
	curveApi, err := sw_emulated.New[emulated.BN254Fp, emulated.BN254Fr](api, sw_emulated.GetBN254Params())
	if err != nil {
		return err
	}

	fieldFpApi, err := emulated.NewField[emulated.BN254Fp](api)
	if err != nil {
		return err
	}

	fieldFrApi, err := emulated.NewField[emulated.BN254Fr](api)
	if err != nil {
		return err
	}

	mimcApi, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	keccak256Api, err := gnarkSha3.NewLegacyKeccak256(api)
	if err != nil {
		return err
	}

	u64Api, err := uints.New[uints.U64](api)
	if err != nil {
		return err
	}

	pairingApi, err := sw_bn254.NewPairing(api)
	if err != nil {
		return err
	}

	valsetHash := frontend.Variable(0)
	signersAggKey := &sw_bn254.G1Affine{
		X: emulated.ValueOf[emulated.BN254Fp](0),
		Y: emulated.ValueOf[emulated.BN254Fp](0),
	}
	signersAggVotingPower := frontend.Variable(0)

	// calc valset hash, agg key and agg voting power
	for i := range circuit.ValidatorData {
		hashAffineG1(&mimcApi, &circuit.ValidatorData[i].Key)
		mimcApi.Write(circuit.ValidatorData[i].VotingPower)
		valsetHashTemp := mimcApi.Sum()

		valsetHash = api.Select(
			api.And(fieldFpApi.IsZero(&circuit.ValidatorData[i].Key.X), fieldFpApi.IsZero(&circuit.ValidatorData[i].Key.Y)),
			valsetHash,
			valsetHashTemp,
		)

		// get power if NON-SIGNER otherwise 0
		pow := api.Select(circuit.ValidatorData[i].IsNonSigner, frontend.Variable(0), circuit.ValidatorData[i].VotingPower)
		signersAggVotingPower = api.Add(signersAggVotingPower, pow)

		// get key if SIGNER otherwise zero point
		point := curveApi.Select(api.IsZero(circuit.ValidatorData[i].IsNonSigner), &circuit.ValidatorData[i].Key, &sw_bn254.G1Affine{
			X: emulated.ValueOf[emulated.BN254Fp](0),
			Y: emulated.ValueOf[emulated.BN254Fp](0),
		})
		signersAggKey = curveApi.AddUnified(signersAggKey, point)
	}

	// compare with public inputs
	api.AssertIsEqual(signersAggVotingPower, circuit.SignersAggVotingPower)

	// --------------------------------------- Prove Input consistency ---------------------------------------

	// valset consistency checked against InputHash which is Hash{valset-hash|non-signers-vp|message}
	hashBytes := variableToBytes(api, u64Api, valsetHash)

	api.Println("HashBytes:", hashBytes)
	keccak256Api.Write(hashBytes)
	aggVotingPowerBytes := variableToBytes(api, u64Api, circuit.SignersAggVotingPower)

	api.Println("aggVotingPowerBytes:", aggVotingPowerBytes)
	keccak256Api.Write(aggVotingPowerBytes)
	messageBytes := keyToBytes(u64Api, &circuit.Message)

	api.Println("MessageBytes:", messageBytes)
	keccak256Api.Write(messageBytes)
	inputDataHash := keccak256Api.Sum()
	api.Println("InputDataHash:", inputDataHash)
	inputHashBytes := variableToBytes(api, u64Api, circuit.InputHash)

	inputDataHash[0] = u64Api.ByteValueOf(u64Api.ToValue(u64Api.And(u64Api.ValueOf(inputDataHash[0].Val), uints.NewU64(0x1f)))) // zero two first bits
	for i := range inputHashBytes {
		u64Api.ByteAssertEq(inputDataHash[i], inputHashBytes[i])
	}

	// --------------------------------------- Verify Signature ---------------------------------------

	// calc alpha
	mimcApi.Reset()
	hashAffineG1(&mimcApi, &circuit.Signature)
	hashAffineG1(&mimcApi, signersAggKey)
	hashAffineG2(&mimcApi, &circuit.SignersAggKeyG2)
	hashAffineG1(&mimcApi, &circuit.Message)
	//TODO optimize
	alpha := fieldFrApi.FromBits(bits.ToBinary(api, mimcApi.Sum())...)

	// pairing check
	_, _, g1Gen, g2Gen := bn254.Generators()
	g1GenAffine := sw_bn254.NewG1Affine(g1Gen)
	negG2GenAffine := sw_bn254.NewG2Affine(*g2Gen.Neg(&g2Gen))
	err = pairingApi.PairingCheck(
		[]*sw_bn254.G1Affine{
			curveApi.AddUnified(&circuit.Signature, curveApi.ScalarMul(signersAggKey, alpha)),
			curveApi.AddUnified(&circuit.Message, curveApi.ScalarMul(&g1GenAffine, alpha)),
		},
		[]*sw_bn254.G2Affine{
			&negG2GenAffine,
			&circuit.SignersAggKeyG2,
		},
	)

	return err
}

// helper functions
func HashValset(valset []ValidatorData) []byte {
	h := mimc_native.NewMiMC()
	zeroPoint := new(bn254.G1Affine)
	zeroPoint.SetInfinity()
	for i := range valset {
		if valset[i].Key.X.Cmp(&zeroPoint.X) == 0 && valset[i].Key.Y.Cmp(&zeroPoint.Y) == 0 {
			break
		}

		xBytes := valset[i].Key.X.Bytes()
		yBytes := valset[i].Key.Y.Bytes()

		// hash by limbs as it's done inside circuit
		h.Write(xBytes[24:32])
		h.Write(xBytes[16:24])
		h.Write(xBytes[8:16])
		h.Write(xBytes[0:8])

		h.Write(yBytes[24:32])
		h.Write(yBytes[16:24])
		h.Write(yBytes[8:16])
		h.Write(yBytes[0:8])

		votingPowerBuf := make([]byte, 32)
		valset[i].VotingPower.FillBytes(votingPowerBuf)
		h.Write(votingPowerBuf)

		//	outerHash.Write(innerHash.Sum(nil))
	}
	return h.Sum(nil)
}

func ValidatorSetMimcAccumulator(valset []entity.Validator, requiredKeyTag entity.KeyTag) ([32]byte, error) {
	validatorsData, err := ToValidatorsData([]entity.Validator{}, valset, requiredKeyTag)
	if err != nil {
		return [32]byte{}, err
	}
	return [32]byte(HashValset(validatorsData)), nil
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

func getNonSignersData(valset []ValidatorData) (aggKey *bn254.G1Affine, aggVotingPower *big.Int, totalVotingPower *big.Int) { //nolint:unparam // maybe needed later
	aggVotingPower = big.NewInt(0)
	totalVotingPower = big.NewInt(0)
	aggKey = new(bn254.G1Affine)
	aggKey.SetInfinity()
	for i := range valset {
		if valset[i].IsNonSigner {
			aggKey = aggKey.Add(aggKey, &(valset)[i].Key)
			aggVotingPower = aggVotingPower.Add(aggVotingPower, valset[i].VotingPower)
		}
		totalVotingPower = totalVotingPower.Add(totalVotingPower, valset[i].VotingPower)
	}
	return aggKey, aggVotingPower, totalVotingPower
}

func getAggSignature(message bn254.G1Affine, valset *[]ValidatorData) (signature *bn254.G1Affine, aggKeyG2 *bn254.G2Affine, aggKeyG1 *bn254.G1Affine) {
	aggKeyG2 = new(bn254.G2Affine)
	aggKeyG2.SetInfinity()

	aggSignature := new(bn254.G1Affine)
	aggSignature.SetInfinity()

	aggKeyG1 = new(bn254.G1Affine)
	aggKeyG1.SetInfinity()

	for i := range *valset {
		if !(*valset)[i].IsNonSigner {
			aggKeyG2 = aggKeyG2.Add(aggKeyG2, &(*valset)[i].KeyG2)
			aggKeyG1 = aggKeyG1.Add(aggKeyG1, &(*valset)[i].Key)
			msg := bn254.G1Affine{X: message.X, Y: message.Y} // have to copy msg since ScalarMultiplication rewrite it
			sig := msg.ScalarMultiplication(&msg, (*valset)[i].PrivateKey)
			aggSignature = aggSignature.Add(aggSignature, sig)
		}
	}

	return aggSignature, aggKeyG2, aggKeyG1
}

func setCircuitData(circuit *Circuit, proveInput ProveInput) error {
	circuit.ValidatorData = make([]ValidatorDataCircuit, len(proveInput.ValidatorData))
	for i := range proveInput.ValidatorData {
		circuit.ValidatorData[i].Key = sw_bn254.NewG1Affine(proveInput.ValidatorData[i].Key)
		circuit.ValidatorData[i].VotingPower = proveInput.ValidatorData[i].VotingPower
		circuit.ValidatorData[i].IsNonSigner = *big.NewInt(0)

		if proveInput.ValidatorData[i].IsNonSigner {
			circuit.ValidatorData[i].IsNonSigner = *big.NewInt(1)
		}
	}

	messageG1, err := bls.HashToG1(proveInput.Message)
	if err != nil {
		return errors.Errorf("failed to hash message to G1: %w", err)
	}
	messageG1Bn254 := bn254.G1Affine{X: messageG1.X, Y: messageG1.Y}

	_, nonSignersAggVotingPower, totalVotingPower := getNonSignersData(proveInput.ValidatorData)
	signersAggVotingPower := new(big.Int).Sub(totalVotingPower, nonSignersAggVotingPower)
	valsetHash := HashValset(proveInput.ValidatorData)

	circuit.SignersAggVotingPower = *signersAggVotingPower

	//fmt.Println("proveInput.ValidatorData:", proveInput.ValidatorData)
	//fmt.Println("proveInput.Signature:", proveInput.Signature)
	//fmt.Println("messageG1Bn254.X:", messageG1Bn254)
	//fmt.Println("proveInput.SignersAggKeyG2:", proveInput.SignersAggKeyG2)

	circuit.Signature = sw_bn254.NewG1Affine(proveInput.Signature)
	circuit.Message = sw_bn254.NewG1Affine(messageG1Bn254)
	circuit.SignersAggKeyG2 = sw_bn254.NewG2Affine(proveInput.SignersAggKeyG2)

	messageBytes := messageG1Bn254.RawBytes()
	aggVotingPowerBuffer := make([]byte, 32)
	signersAggVotingPower.FillBytes(aggVotingPowerBuffer)

	//fmt.Println("signersAggVotingPower:", signersAggVotingPower)
	inputHashBytes := valsetHash
	inputHashBytes = append(inputHashBytes, aggVotingPowerBuffer...)
	inputHashBytes = append(inputHashBytes, messageBytes[:]...)
	inputHash := crypto.Keccak256(inputHashBytes)

	slog.Debug("signersAggVotingPower", "vp", signersAggVotingPower.String())
	slog.Debug("signed message", "message", messageG1Bn254.String())
	slog.Debug("signed message", "message.X", messageG1Bn254.X.String())
	slog.Debug("signed message", "message.Y", messageG1Bn254.Y.String())
	slog.Debug("mimc hash", "hash", hex.EncodeToString(valsetHash))

	//fmt.Println("InputHashBytes:", hex.EncodeToString(inputHashBytes))
	//fmt.Println(hex.EncodeToString(inputHashBytes))
	//fmt.Println("inputHash:", hex.EncodeToString(inputHash))
	inputHashInt := new(big.Int).SetBytes(inputHash)
	mask, _ := big.NewInt(0).SetString("1FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
	inputHashInt.And(inputHashInt, mask)

	//fmt.Println("inputHashHex:", inputHashInt.Text(16))
	circuit.InputHash = inputHashInt

	slog.Debug("[Prove] input hash", "hash", hex.EncodeToString(inputHashInt.Bytes()))

	return nil
}

type ZkProver struct {
	cs map[int]constraint.ConstraintSystem
	pk map[int]groth16.ProvingKey
	vk map[int]groth16.VerifyingKey
}

func NewZkProver() *ZkProver {
	p := ZkProver{
		cs: make(map[int]constraint.ConstraintSystem),
		pk: make(map[int]groth16.ProvingKey),
		vk: make(map[int]groth16.VerifyingKey),
	}
	p.init()
	return &p
}

func (p *ZkProver) init() {
	slog.Warn("ZK prover initialization started (might take a few seconds)")
	for _, size := range MaxValidators {
		cs, pk, vk, err := loadOrInit(size)
		if err != nil {
			panic(err)
		}
		p.cs[size] = cs
		p.pk[size] = pk
		p.vk[size] = vk
	}
	slog.Info("ZK prover initialization is done")
}

func (p *ZkProver) DoProve(rawProveInput RawProveInput) (ProofData, error) {
	data, err := ToValidatorsData(rawProveInput.SignerValidators, rawProveInput.AllValidators, rawProveInput.RequiredKeyTag)
	if err != nil {
		return ProofData{}, errors.Errorf("failed to convert validators to data: %w", err)
	}

	proofData, err := p.Prove(ProveInput{
		ValidatorData:   data,
		Message:         rawProveInput.Message,
		Signature:       bn254.G1Affine{X: rawProveInput.Signature.X, Y: rawProveInput.Signature.Y},
		SignersAggKeyG2: bn254.G2Affine{X: rawProveInput.SignersAggKeyG2.X, Y: rawProveInput.SignersAggKeyG2.Y},
	})
	if err != nil {
		return ProofData{}, errors.Errorf("failed to prove: %w", err)
	}

	return proofData, nil
}

func (p *ZkProver) Verify(valsetLen int, publicInputHash [32]byte, proofBytes []byte) (bool, error) {
	valsetLen = getOptimalN(valsetLen)
	assignment := Circuit{}
	publicInputHashInt := new(big.Int).SetBytes(publicInputHash[:])
	mask, _ := big.NewInt(0).SetString("1FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
	publicInputHashInt.And(publicInputHashInt, mask)
	assignment.InputHash = publicInputHashInt

	slog.Debug("[Verify] input hash", "hash", hex.EncodeToString(publicInputHashInt.Bytes()))

	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	publicWitness, _ := witness.Public()

	rawProofBytes := bytes.Clone(proofBytes[:256])
	rawProofBytes = append(rawProofBytes, []byte{0, 0, 0, 1}...) //dirty hack
	rawProofBytes = append(rawProofBytes, proofBytes[256:384]...)
	reader := bytes.NewReader(rawProofBytes)
	proof := groth16.NewProof(ecc.BN254)
	_, err := proof.ReadFrom(reader)
	if err != nil {
		return false, fmt.Errorf("failed to read proof: %w", err)
	}

	vk, ok := p.vk[valsetLen]
	if !ok {
		return false, fmt.Errorf("failed to find verification key for valset length %d", valsetLen)
	}

	err = groth16.Verify(proof, vk, publicWitness, backend.WithVerifierHashToFieldFunction(sha256.New()))
	if err != nil {
		return false, fmt.Errorf("failed to verify: %w", err)
	}
	return true, nil
}

func (p *ZkProver) Prove(proveInput ProveInput) (ProofData, error) {
	pk := p.pk[len(proveInput.ValidatorData)]
	vk := p.vk[len(proveInput.ValidatorData)]
	r1cs, ok := p.cs[len(proveInput.ValidatorData)]
	if !ok {
		return ProofData{}, errors.Errorf("failed to load cs, vk, pk for valset size: %d", len(proveInput.ValidatorData))
	}

	// witness definition
	assignment := Circuit{}
	err := setCircuitData(&assignment, proveInput)
	if err != nil {
		return ProofData{}, errors.Errorf("failed to set circuit data: %w", err)
	}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()

	// groth16: Prove & Verify
	proof, err := groth16.Prove(r1cs, pk, witness, backend.WithProverHashToFieldFunction(sha256.New()))
	if err != nil {
		return ProofData{}, errors.Errorf("failed to prove: %w", err)
	}

	publicInputs := publicWitness.Vector().(fr.Vector)
	// Format for the specific Solidity interface
	formattedInputs := make([]*big.Int, 0, len(publicInputs))

	// Format the vector of public inputs as hex strings
	for _, input := range publicInputs {
		formattedInputs = append(formattedInputs, new(big.Int).SetBytes(input.Marshal()))
	}

	// If more than 10 inputs (unlikely), you'll need to adapt the interface
	if len(formattedInputs) > 10 {
		return ProofData{}, errors.Errorf("more than 10 public inputs")
	}

	_, ok = proof.(interface{ MarshalSolidity() []byte })
	if !ok {
		panic("proof does not implement MarshalSolidity()")
	}

	// verify proof
	err = groth16.Verify(proof, vk, publicWitness, backend.WithVerifierHashToFieldFunction(sha256.New()))
	if err != nil {
		return ProofData{}, err
	}

	// Serialize the proof
	var proofBuffer bytes.Buffer
	_, err = proof.WriteRawTo(&proofBuffer)
	if err != nil {
		return ProofData{}, errors.Errorf("failed to write proof: %w", err)
	}
	proofBytes := proofBuffer.Bytes()
	//fmt.Println("proofBytes:", proofBytes) //nolint:staticcheck // will fix later
	//fmt.Println("hex:", common.Bytes2Hex(proofBytes))

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

	//fmt.Println("proof: ", standardProof)
	//fmt.Println("commitments: ", commitments)
	//fmt.Println("commitmentPok: ", commitmentPok)
	//fmt.Println("inputs", formattedInputs)
	//// Extract public inputs
	//for i := 0; i < publicWitness.Vector(); i++ {
	//	val, _ := publicWitness.GetValue(i)
	//	publicInputs[i] = new(big.Int).SetBytes(val.Bytes())
	//}

	_, nonSignersAggVotingPower, totalVotingPower := getNonSignersData(proveInput.ValidatorData)
	return ProofData{
		Proof:                 proofBytes[:256],
		Commitments:           proofBytes[260:324],
		CommitmentPok:         proofBytes[324:388],
		SignersAggVotingPower: new(big.Int).Sub(totalVotingPower, nonSignersAggVotingPower),
	}, nil
}

func GetActiveValidators(allValidators []entity.Validator) []entity.Validator {
	activeValidators := make([]entity.Validator, 0)
	for _, validator := range allValidators {
		if validator.IsActive {
			activeValidators = append(activeValidators, validator)
		}
	}
	return activeValidators
}

// todo ilya
func ToValidatorsData(signerValidators []entity.Validator, allValidators []entity.Validator, requiredKeyTag entity.KeyTag) ([]ValidatorData, error) {
	activeValidators := GetActiveValidators(allValidators)
	valset := make([]ValidatorData, 0)
	for i := range activeValidators {
		for _, key := range activeValidators[i].Keys {
			if key.Tag == requiredKeyTag {
				g1, err := bls.DeserializeG1(key.Payload)
				if err != nil {
					return nil, fmt.Errorf("failed to deserialize G1: %w", err)
				}
				validatorData := ValidatorData{Key: *g1.G1Affine, VotingPower: activeValidators[i].VotingPower, IsNonSigner: true}

				for _, signer := range signerValidators {
					if signer.Operator.Cmp(activeValidators[i].Operator) == 0 {
						validatorData.IsNonSigner = false
						break
					}
				}

				valset = append(valset, validatorData)
				break
			}
		}
	}
	return NormalizeValset(valset), nil
}

func loadOrInit(valsetLen int) (constraint.ConstraintSystem, groth16.ProvingKey, groth16.VerifyingKey, error) {
	suffix := strconv.Itoa(valsetLen)
	r1csP := r1csPathTmp(suffix)
	pkP := pkPathTmp(suffix)
	vkP := vkPathTmp(suffix)
	solP := solPathTmp(suffix)

	if exists(r1csP) && exists(pkP) && exists(vkP) && exists(solP) {
		r1csCS := groth16.NewCS(bn254.ID)
		data, err := os.Open(r1csP)
		if err != nil {
			return nil, nil, nil, errors.Errorf("failed to open r1cs: %w", err)
		}
		defer data.Close()
		if _, err := r1csCS.ReadFrom(data); err != nil {
			return nil, nil, nil, errors.Errorf("failed to read r1cs: %w", err)
		}

		pk := groth16.NewProvingKey(bn254.ID)
		data, err = os.Open(pkP)
		if err != nil {
			return nil, nil, nil, errors.Errorf("failed to open pk: %w", err)
		}
		defer data.Close()
		if _, err := pk.UnsafeReadFrom(data); err != nil {
			return nil, nil, nil, errors.Errorf("failed to read pk: %w", err)
		}

		vk := groth16.NewVerifyingKey(bn254.ID)
		data, err = os.Open(vkP)
		if err != nil {
			return nil, nil, nil, errors.Errorf("failed to open vk: %w", err)
		}
		defer data.Close()
		if _, err := vk.UnsafeReadFrom(data); err != nil {
			return nil, nil, nil, errors.Errorf("failed to read vk: %w", err)
		}

		return r1csCS, pk, vk, nil
	}

	if err := os.MkdirAll(circuitsDir, 0o755); err != nil {
		return nil, nil, nil, err
	}

	for _, m := range MaxValidators {
		suf := strconv.Itoa(m)
		r1csFile := r1csPathTmp(suf)
		pkFile := pkPathTmp(suf)
		vkFile := vkPathTmp(suf)
		solFile := solPathTmp(suf)

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

	return loadOrInit(valsetLen)
}

func NormalizeValset(valset []ValidatorData) []ValidatorData {
	// Sort validators by key in ascending order
	sort.Slice(valset, func(i, j int) bool {
		// Compare keys (lower first)
		return valset[i].Key.X.Cmp(&valset[j].Key.X) > 0 || valset[i].Key.Y.Cmp(&valset[j].Key.Y) > 0
	})
	n := getOptimalN(len(valset))
	normalizedValset := make([]ValidatorData, n)
	for i := range n {
		if i < len(valset) {
			normalizedValset[i] = valset[i]
		} else {
			zeroPoint := new(bn254.G1Affine)
			zeroPoint.SetInfinity()
			zeroPointG2 := new(bn254.G2Affine)
			zeroPointG2.SetInfinity()
			normalizedValset[i] = ValidatorData{PrivateKey: big.NewInt(0), Key: *zeroPoint, KeyG2: *zeroPointG2, VotingPower: big.NewInt(0), IsNonSigner: false}
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
