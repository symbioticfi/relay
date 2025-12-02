package proof

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/solidity"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	"github.com/symbioticfi/relay/pkg/tracing"
)

var (
	defaultMaxValidators = []int{10, 100, 1000}
)

func GetMaxValidators() []int {
	if os.Getenv("MAX_VALIDATORS") != "" {
		countList := strings.Split(os.Getenv("MAX_VALIDATORS"), ",")
		var newMaxValidators []int
		for _, countStr := range countList {
			count, err := strconv.Atoi(strings.TrimSpace(countStr))
			if err != nil {
				slog.Error("Invalid MAX_VALIDATORS value, must be comma-separated integers", "value", os.Getenv("MAX_VALIDATORS"))
				continue
			}
			if count <= 0 {
				slog.Error("Invalid MAX_VALIDATORS value, must be positive integers", "value", count)
				continue
			}
			newMaxValidators = append(newMaxValidators, count)
		}
		if len(newMaxValidators) > 0 {
			slog.Info("Using custom MAX_VALIDATORS", "list", newMaxValidators)
			return newMaxValidators
		}
	}

	return defaultMaxValidators
}

func r1csPathTmp(circuitsDir, suffix string) string {
	return fmt.Sprintf(circuitsDir+"/circuit_%s.r1cs", suffix)
}

func pkPathTmp(circuitsDir, suffix string) string {
	return fmt.Sprintf(circuitsDir+"/circuit_%s.pk", suffix)
}

func vkPathTmp(circuitsDir, suffix string) string {
	return fmt.Sprintf(circuitsDir+"/circuit_%s.vk", suffix)
}

func solPathTmp(circuitsDir, suffix string) string {
	return fmt.Sprintf(circuitsDir+"/Verifier_%s.sol", suffix)
}

type ProofData struct {
	Proof                 []byte
	Commitments           []byte
	CommitmentPok         []byte
	SignersAggVotingPower *big.Int
}

type ValidatorData struct {
	PrivateKey  *big.Int
	Key         bn254.G1Affine
	KeyG2       bn254.G2Affine
	VotingPower *big.Int
	IsNonSigner bool
}

type ZkProver struct {
	cs            map[int]constraint.ConstraintSystem
	pk            map[int]groth16.ProvingKey
	vk            map[int]groth16.VerifyingKey
	circuitsDir   string
	maxValidators []int
}

func NewZkProver(circuitsDir string) *ZkProver {
	p := ZkProver{
		cs:            make(map[int]constraint.ConstraintSystem),
		pk:            make(map[int]groth16.ProvingKey),
		vk:            make(map[int]groth16.VerifyingKey),
		circuitsDir:   circuitsDir,
		maxValidators: GetMaxValidators(),
	}
	if circuitsDir != "" {
		p.init()
	} else {
		slog.Warn("ZK prover circuits directory is not set, cannot run zk verify/proofs")
	}
	return &p
}

func (p *ZkProver) init() {
	slog.Warn("ZK prover initialization started (might take a few seconds)")
	for _, size := range p.maxValidators {
		cs, pk, vk, err := p.loadOrInit(size)
		if err != nil {
			panic(err)
		}
		p.cs[size] = cs
		p.pk[size] = pk
		p.vk[size] = vk
	}
	slog.Info("ZK prover initialization is done")
}

func (p *ZkProver) Verify(ctx context.Context, valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error) {
	_, span := tracing.StartSpan(ctx, "zkprover.Verify",
		tracing.AttrValidatorCount.Int(valsetLen),
		tracing.AttrProofSize.Int(len(proofBytes)),
	)
	defer span.End()

	if p.circuitsDir == "" {
		err := errors.New("ZK prover circuits directory is not set, cannot run zk verify")
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "preparing_inputs")
	valsetLen = getOptimalN(valsetLen)
	assignment := Circuit{}
	publicInputHashInt := new(big.Int).SetBytes(publicInputHash[:])
	mask, _ := big.NewInt(0).SetString("1FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
	publicInputHashInt.And(publicInputHashInt, mask)
	assignment.InputHash = publicInputHashInt

	slog.DebugContext(ctx, "[Verify] input hash", "hash", hex.EncodeToString(publicInputHashInt.Bytes()))

	tracing.AddEvent(span, "creating_witness")
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	publicWitness, _ := witness.Public()

	tracing.AddEvent(span, "deserializing_proof")
	//nolint:gosec // G602: proofBytes length is validated by caller, slicing is safe
	rawProofBytes := bytes.Clone(proofBytes[:256])
	rawProofBytes = append(rawProofBytes, []byte{0, 0, 0, 1}...) //dirty hack
	//nolint:gosec // G602: proofBytes length is validated by caller, slicing is safe
	rawProofBytes = append(rawProofBytes, proofBytes[256:384]...)
	reader := bytes.NewReader(rawProofBytes)
	proof := groth16.NewProof(ecc.BN254)
	_, err := proof.ReadFrom(reader)
	if err != nil {
		err = errors.Errorf("failed to read proof: %w", err)
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "loading_verification_key")
	vk, ok := p.vk[valsetLen]
	if !ok {
		err := errors.Errorf("failed to find verification key for valset length %d", valsetLen)
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "verifying_groth16_proof")
	err = groth16.Verify(proof, vk, publicWitness, backend.WithVerifierHashToFieldFunction(sha256.New()))
	if err != nil {
		err = errors.Errorf("failed to verify: %w", err)
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.AddEvent(span, "verification_successful")
	return true, nil
}

func (p *ZkProver) Prove(ctx context.Context, proveInput ProveInput) (ProofData, error) {
	_, span := tracing.StartSpan(ctx, "zkprover.Prove",
		tracing.AttrValidatorCount.Int(len(proveInput.ValidatorData)),
	)
	defer span.End()

	if p.circuitsDir == "" {
		err := errors.New("ZK prover circuits directory is not set, cannot run zk proofs")
		tracing.RecordError(span, err)
		return ProofData{}, err
	}

	tracing.AddEvent(span, "loading_circuit_parameters")
	pk := p.pk[len(proveInput.ValidatorData)]
	vk := p.vk[len(proveInput.ValidatorData)]
	r1cs, ok := p.cs[len(proveInput.ValidatorData)]
	if !ok {
		err := errors.Errorf("failed to load cs, vk, pk for valset size: %d", len(proveInput.ValidatorData))
		tracing.RecordError(span, err)
		return ProofData{}, err
	}

	tracing.AddEvent(span, "creating_witness")
	// witness definition
	assignment := Circuit{}
	setCircuitData(&assignment, proveInput)

	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		tracing.RecordError(span, err)
		return ProofData{}, errors.Errorf("failed to create witness: %w", err)
	}
	publicWitness, err := witness.Public()
	if err != nil {
		tracing.RecordError(span, err)
		return ProofData{}, errors.Errorf("failed to create public witness: %w", err)
	}

	tracing.AddEvent(span, "generating_groth16_proof")
	// groth16: Prove & Verify
	proof, err := groth16.Prove(r1cs, pk, witness, backend.WithProverHashToFieldFunction(sha256.New()))
	if err != nil {
		tracing.RecordError(span, err)
		return ProofData{}, errors.Errorf("failed to prove: %w", err)
	}

	tracing.AddEvent(span, "formatting_public_inputs")
	publicInputs := publicWitness.Vector().(fr.Vector)
	// Format for the specific Solidity interface
	formattedInputs := make([]*big.Int, 0, len(publicInputs))

	// Format the vector of public inputs as hex strings
	for _, input := range publicInputs {
		formattedInputs = append(formattedInputs, new(big.Int).SetBytes(input.Marshal()))
	}

	// If more than 10 inputs (unlikely), you'll need to adapt the interface
	if len(formattedInputs) > 10 {
		err := errors.Errorf("more than 10 public inputs")
		tracing.RecordError(span, err)
		return ProofData{}, err
	}

	_, ok = proof.(interface{ MarshalSolidity() []byte })
	if !ok {
		panic("proof does not implement MarshalSolidity()")
	}

	tracing.AddEvent(span, "verifying_proof")
	// verify proof
	err = groth16.Verify(proof, vk, publicWitness, backend.WithVerifierHashToFieldFunction(sha256.New()))
	if err != nil {
		tracing.RecordError(span, err)
		return ProofData{}, err
	}

	tracing.AddEvent(span, "serializing_proof")
	// Serialize the proof
	var proofBuffer bytes.Buffer
	_, err = proof.WriteRawTo(&proofBuffer)
	if err != nil {
		tracing.RecordError(span, err)
		return ProofData{}, errors.Errorf("failed to write proof: %w", err)
	}
	proofBytes := proofBuffer.Bytes()

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

	_, nonSignersAggVotingPower, totalVotingPower := getNonSignersData(proveInput.ValidatorData)

	tracing.AddEvent(span, "proof_generation_completed")
	return ProofData{
		Proof:                 proofBytes[:256],
		Commitments:           proofBytes[260:324],
		CommitmentPok:         proofBytes[324:388],
		SignersAggVotingPower: new(big.Int).Sub(totalVotingPower, nonSignersAggVotingPower),
	}, nil
}

//nolint:revive // function-result-limit: This function needs to return multiple complex types for cryptographic operations
func (p *ZkProver) loadOrInit(valsetLen int) (constraint.ConstraintSystem, groth16.ProvingKey, groth16.VerifyingKey, error) {
	slog.Info("Loading or initializing zk circuit files", "valsetLen", valsetLen, "dir", p.circuitsDir)
	suffix := strconv.Itoa(valsetLen)
	r1csP := r1csPathTmp(p.circuitsDir, suffix)
	pkP := pkPathTmp(p.circuitsDir, suffix)
	vkP := vkPathTmp(p.circuitsDir, suffix)

	if exists(r1csP) && exists(pkP) && exists(vkP) {
		slog.Warn("Using existing zk circuit files", "r1cs", r1csP, "pk", pkP, "vk", vkP)
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

	if err := os.MkdirAll(p.circuitsDir, 0o755); err != nil {
		return nil, nil, nil, err
	}

	for _, m := range p.maxValidators {
		suf := strconv.Itoa(m)
		r1csFile := r1csPathTmp(p.circuitsDir, suf)
		pkFile := pkPathTmp(p.circuitsDir, suf)
		vkFile := vkPathTmp(p.circuitsDir, suf)
		solFile := solPathTmp(p.circuitsDir, suf)

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

	return p.loadOrInit(valsetLen)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
