package entity

import (
	"math/big"
)

type StringError string

func (e StringError) Error() string {
	return string(e)
}

const (
	ErrEntityNotFound         = StringError("entity not found")
	ErrPhaseNotCommit         = StringError("phase is not commit")
	ErrSignatureRequestExists = StringError("signature request already exists")
)

const ValsetHeaderKeyTag uint8 = 15
const MaxSavedEpochs int64 = 10

type SignatureRequest struct {
	KeyTag        uint8
	RequiredEpoch *big.Int
	Message       []byte
}

// AggregationProof aggregator.proof(signatures []Signature) -> AggregationProof
type AggregationProof struct {
	VerificationType uint32 // proof verification type
	MessageHash      []byte // scheme depends on KeyTag
	Proof            []byte // parse based on KeyTag & VerificationType
}

const (
	ZkVerificationType     = 0
	SimpleVerificationType = 1
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

var (
	QuorumThresholdBase       = big.NewInt(1e18)
	QuorumThresholdPercentage = big.NewInt(666666666666666667)
)
