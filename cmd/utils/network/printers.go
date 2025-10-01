package network

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/symbioticfi/relay/core/entity"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/samber/lo"
)

type settlementReplicaData struct {
	IsCommitted              bool
	HeaderHash               common.Hash
	MissedEpochs             uint64
	LastCommittedHeaderEpoch uint64
}

func printAddresses(driver entity.CrossChainAddress, networkConfig *entity.NetworkConfig) string {
	addressesTableData := pterm.TableData{
		{"Type", "Chain ID", "Address"},
		{"Driver", strconv.FormatUint(driver.ChainId, 10), driver.Address.String()},
		{"KeyRegistry", strconv.FormatUint(networkConfig.KeysProvider.ChainId, 10), networkConfig.KeysProvider.Address.String()},
	}
	for _, provider := range networkConfig.VotingPowerProviders {
		addressesTableData = append(addressesTableData, []string{
			"VotingPowerProvider",
			strconv.FormatUint(provider.ChainId, 10),
			provider.Address.String(),
		})
	}
	for _, settlement := range networkConfig.Settlements {
		addressesTableData = append(addressesTableData, []string{
			"Settlement",
			strconv.FormatUint(settlement.ChainId, 10),
			settlement.Address.String(),
		})
	}
	addressesText, _ := pterm.DefaultTable.WithHasHeader().WithData(addressesTableData).Srender()
	return addressesText
}

func printNetworkConfig(epochDuration uint64, networkConfig *entity.NetworkConfig) string {
	configText := fmt.Sprintf("Verification type: %s\n", networkConfig.VerificationType.String())
	configText += fmt.Sprintf("Max voting power: %0.4e\n", new(big.Float).SetInt(networkConfig.MaxVotingPower.Int))
	configText += fmt.Sprintf("Min inclusion voting power: %v\n", networkConfig.MinInclusionVotingPower)
	configText += fmt.Sprintf("Max validators count: %v\n", networkConfig.MaxValidatorsCount)
	configText += fmt.Sprintf("Epoch duration: %d sec\n", epochDuration)
	configText += fmt.Sprintf("Required key tags: %s\n", strings.Join(lo.Map(networkConfig.RequiredKeyTags, func(item entity.KeyTag, _ int) string {
		return strconv.FormatUint(uint64(item), 10)
	}), ", "))
	configText += fmt.Sprintf("Quorum thresholds (keyTag/%%): %s\n", strings.Join(lo.Map(networkConfig.QuorumThresholds, func(item entity.QuorumThreshold, _ int) string {
		return fmt.Sprintf("%d/%0.3f%%", uint8(item.KeyTag), cmdhelpers.GetPct(item.QuorumThreshold.Int, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	}), ", "))
	configText += fmt.Sprintf("Header key tag: %s\n", networkConfig.RequiredHeaderKeyTag.String())

	return configText
}

func printNetworkInfo(epoch entity.Epoch, epochStart entity.Timestamp, networkConfig *entity.NetworkConfig, valset *entity.ValidatorSet) string {
	infoText := fmt.Sprintf("Network epoch: %v\n", epoch)
	t := time.Unix(int64(epochStart), 0)
	tFormatted := t.Format("2006-01-02 15:04:05")
	infoText += fmt.Sprintf("Epoch start: %d (%s)\n", epochStart, tFormatted)
	infoText += fmt.Sprintf("Validators: %d\n", len(valset.Validators))
	infoText += fmt.Sprintf("Total voting power: %v\n", valset.GetTotalActiveVotingPower())
	infoText += fmt.Sprintf("Voting power providers: %d\n", len(networkConfig.VotingPowerProviders))
	infoText += fmt.Sprintf("Settlements: %d\n", len(networkConfig.Settlements))
	infoText += fmt.Sprintf("Header quorum threshold: %d (%0.3f%%)\n",
		valset.QuorumThreshold, cmdhelpers.GetPct(valset.QuorumThreshold.Int, valset.GetTotalActiveVotingPower().Int))
	return infoText
}

func printValidatorsTree(valset *entity.ValidatorSet) string {
	leveledList := pterm.LeveledList{}

	validators := valset.Validators

	for _, validator := range validators {
		leveledList = cmdhelpers.PrintTreeValidator(leveledList, validator, valset.GetTotalActiveVotingPower().Int)
	}

	// Render the tree structure using the default tree printer.
	text, _ := pterm.DefaultTree.WithRoot(putils.TreeFromLeveledList(leveledList)).Srender()
	return text
}

func printValidatorsTable(valset *entity.ValidatorSet) string {
	tableData := pterm.TableData{
		{"Address", "Status", "Voting Power", "Vaults", "Keys"},
	}

	validators := valset.Validators

	for _, validator := range validators {
		status := pterm.FgRed.Sprint("inactive")
		if validator.IsActive {
			status = pterm.FgGreen.Sprint("active")
		}
		pct := new(big.Float).SetInt(validator.VotingPower.Int)
		pct = pct.Mul(pct, big.NewFloat(100))
		pct = pct.Quo(pct, new(big.Float).SetInt(valset.GetTotalActiveVotingPower().Int))
		tableData = append(tableData, []string{
			validator.Operator.String(),
			status,
			fmt.Sprintf("%v (%0.3f)%%", validator.VotingPower, pct),
			strconv.Itoa(len(validator.Vaults)),
			strconv.Itoa(len(validator.Keys)),
		})
	}
	text, _ := pterm.DefaultTable.WithHasHeader().WithData(tableData).Srender()
	return text
}

func printHeaderTable(header entity.ValidatorSetHeader) string {
	headerTableData := pterm.TableData{
		{"Field", "Value"},
		{"Version", strconv.FormatUint(uint64(header.Version), 10)},
		{"Epoch", strconv.FormatUint(uint64(header.Epoch), 10)},
		{"CaptureTimestamp", fmt.Sprintf("%d (%s)",
			header.CaptureTimestamp,
			time.Unix(int64(header.CaptureTimestamp), 0).Format("2006-01-02 15:04:05"),
		)},
		{"RequiredKeyTag", header.RequiredKeyTag.String()},
		{"QuorumThreshold", fmt.Sprintf("%d", header.QuorumThreshold.Int)},
		{"TotalVotingPower", fmt.Sprintf("%d", header.TotalVotingPower.Int)},
		{"ValidatorsSszMRoot", fmt.Sprintf("0x%064x", header.ValidatorsSszMRoot)},
	}

	text, _ := pterm.DefaultTable.WithHasHeader().WithData(headerTableData).Srender()

	return text
}

func printExtraDataTable(extraData entity.ExtraDataList) string {
	extraDataTable := pterm.TableData{{"Key", "Value"}}

	for _, extraData := range extraData {
		extraDataTable = append(extraDataTable, []string{
			fmt.Sprintf("0x%064x", extraData.Key),
			fmt.Sprintf("0x%064x", extraData.Value),
		})
	}

	text, _ := pterm.DefaultTable.WithHasHeader().WithData(extraDataTable).Srender()
	return text
}

func printHeaderWithExtraDataToJSON(validatorSetHeader entity.ValidatorSetHeader, extraDataList entity.ExtraDataList) string {
	type jsonHeader struct {
		Version            uint8    `json:"version"`
		ValidatorsSszMRoot string   `json:"validatorsSszMRoot"` // hex string
		Epoch              uint64   `json:"epoch"`
		RequiredKeyTag     uint8    `json:"requiredKeyTag"`
		CaptureTimestamp   uint64   `json:"captureTimestamp"`
		QuorumThreshold    *big.Int `json:"quorumThreshold"`
		TotalVotingPower   *big.Int `json:"totalVotingPower"`
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
		Epoch:              uint64(validatorSetHeader.Epoch),
		RequiredKeyTag:     uint8(validatorSetHeader.RequiredKeyTag),
		CaptureTimestamp:   uint64(validatorSetHeader.CaptureTimestamp),
		QuorumThreshold:    validatorSetHeader.QuorumThreshold.Int,
		TotalVotingPower:   validatorSetHeader.TotalVotingPower.Int,
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

func printSettlementData(
	valsetHeader entity.ValidatorSetHeader,
	networkConfig entity.NetworkConfig,
	settlementData []settlementReplicaData,
) string {
	tableData := pterm.TableData{
		{"Address", "ChainID", "Status", "Integrity", "Latest Committed Epoch", "Missed Epochs", "Header hash"},
	}

	for i, settlement := range networkConfig.Settlements {
		hash := "N/A"
		status := "Missing"
		if settlementData[i].IsCommitted {
			status = "Committed"
			hash = settlementData[i].HeaderHash.String()
		}

		expectedHash, err := valsetHeader.Hash()
		if err != nil {
			panic(err)
		}

		integrity := "N/A"
		if settlementData[i].IsCommitted && settlementData[i].HeaderHash != expectedHash {
			integrity = "Failed"
		} else if settlementData[i].IsCommitted && settlementData[i].HeaderHash == expectedHash {
			integrity = "Ok"
		}

		tableData = append(tableData, []string{
			settlement.Address.String(),
			strconv.FormatUint(settlement.ChainId, 10),
			status,
			integrity,
			strconv.FormatUint(settlementData[i].LastCommittedHeaderEpoch, 10),
			strconv.FormatUint(settlementData[i].MissedEpochs, 10),
			hash,
		})
	}

	text, _ := pterm.DefaultTable.WithHasHeader().WithData(tableData).Srender()
	return text
}
