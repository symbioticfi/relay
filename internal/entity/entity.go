package entity

import (
	"math/big"
)

type StringError string

func (e StringError) Error() string {
	return string(e)
}

const (
	ErrPhaseFail = StringError("phase is fail")
)

const ValsetHeaderKeyTag uint8 = 15

type SignatureRequest struct {
	KeyTag        uint8
	RequiredEpoch *big.Int
	Message       []byte
}
