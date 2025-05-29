package entity

import (
	"math/big"
)

type StringError string

func (e StringError) Error() string {
	return string(e)
}

const (
	ErrPhaseFail              = StringError("phase is fail")
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
