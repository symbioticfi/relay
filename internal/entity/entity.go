package entity

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type StringError string

func (e StringError) Error() string {
	return string(e)
}

const (
	ErrEntityNotFound     = StringError("entity not found")
	ErrEntityAlreadyExist = StringError("entity already exists")
	ErrPhaseNotCommit     = StringError("phase is not commit")
)

const ValsetHeaderKeyTag = KeyTag(15)
const MaxSavedEpochs int64 = 10

// SignatureRequest signature request message
// RequestHash = sha256(SignatureRequest) (use as identifier later)
type SignatureRequest struct {
	KeyTag        KeyTag
	RequiredEpoch uint64
	Message       []byte
}

func (r SignatureRequest) Hash() common.Hash {
	return crypto.Keccak256Hash([]byte{uint8(r.KeyTag)}, new(big.Int).SetInt64(int64(r.RequiredEpoch)).Bytes(), r.Message)
}

type SignatureMessage struct {
	RequestHash common.Hash
	KeyTag      KeyTag
	Epoch       uint64
	Signature   Signature // parse based on KeyTag
}

type AggregationState struct {
	SignaturesCnt       uint32
	CurrentVotingPower  *big.Int
	RequiredVotingPower *big.Int
}

// AggregationProof aggregator.proof(signatures []Signature) -> AggregationProof
type AggregationProof struct {
	VerificationType VerificationType // proof verification type
	MessageHash      []byte           // scheme depends on KeyTag
	Proof            []byte           // parse based on KeyTag & VerificationType
}

type AggregatedSignatureMessage struct {
	RequestHash      common.Hash
	KeyTag           KeyTag
	Epoch            uint64
	AggregationProof AggregationProof
}

type VerificationType uint32

const (
	VerificationTypeZK     VerificationType = 0
	VerificationTypeSimple VerificationType = 1
)

const (
	ExtraDataGlobalKeyPrefix = "symbiotic.Settlement.extraData."
	ExtraDataKeyTagPrefix    = "keyTag."
)

const (
	ZkVerificationTotalActiveValidators = "totalActiveValidators"
	ZkVerificationValidatorSetHashMimc  = "validatorSetHashMimc"
)

const (
	SimpleVerificationValidatorSetHashKeccak256 = "validatorSetHashKeccak256"
	SimpleVerificationTotalVotingPower          = "totalVotingPower"
	SimpleVerificationAggPublicKeyG1            = "aggPublicKeyG1"
)
