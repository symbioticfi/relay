package network

import (
	"encoding/json"
	"fmt"
	"math/big"

	"middleware-offchain/core/entity"
)

func printHeaderWithExtraDataToJSON(validatorSetHeader entity.ValidatorSetHeader, extraDataList entity.ExtraDataList) string {
	type jsonHeader struct {
		Version            uint8    `json:"version"`
		ValidatorsSszMRoot string   `json:"validatorsSszMRoot"` // hex string
		Epoch              uint64   `json:"epoch"`
		RequiredKeyTag     uint8    `json:"requiredKeyTag"`
		CaptureTimestamp   uint64   `json:"captureTimestamp"`
		QuorumThreshold    *big.Int `json:"quorumThreshold"`
		PreviousHeaderHash string   `json:"previousHeaderHash"` // hex string
	}

	type jsonExtraData struct {
		Key   string `json:"key"`   // hex string
		Value string `json:"value"` // hex string
	}

	type jsonValidatorSetHeaderWithExtraData struct {
		Header        jsonHeader      `json:"header"`
		ExtraDataList []jsonExtraData `json:"extraData"`
	}

	jsonHeaderData := jsonHeader{
		Version:            validatorSetHeader.Version,
		ValidatorsSszMRoot: fmt.Sprintf("0x%064x", validatorSetHeader.ValidatorsSszMRoot),
		Epoch:              validatorSetHeader.Epoch,
		RequiredKeyTag:     uint8(validatorSetHeader.RequiredKeyTag),
		CaptureTimestamp:   validatorSetHeader.CaptureTimestamp,
		QuorumThreshold:    validatorSetHeader.QuorumThreshold.Int,
		PreviousHeaderHash: fmt.Sprintf("0x%064x", validatorSetHeader.PreviousHeaderHash),
	}

	jsonExtraDataList := make([]jsonExtraData, len(extraDataList))
	for i, extraData := range extraDataList {
		jsonExtraDataList[i].Key = fmt.Sprintf("0x%064x", extraData.Key)
		jsonExtraDataList[i].Value = fmt.Sprintf("0x%064x", extraData.Value)
	}

	jsonValidatorSetHeaderWithExtraDataData := jsonValidatorSetHeaderWithExtraData{
		Header:        jsonHeaderData,
		ExtraDataList: jsonExtraDataList,
	}

	jsonData, err := json.MarshalIndent(jsonValidatorSetHeaderWithExtraDataData, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}
