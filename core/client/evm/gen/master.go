// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gen

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IConfigProviderConfig is an auto generated low-level Go binding around an user-defined struct.
type IConfigProviderConfig struct {
	VotingPowerProviders    []IConfigProviderCrossChainAddress
	KeysProvider            IConfigProviderCrossChainAddress
	Replicas                []IConfigProviderCrossChainAddress
	VerificationType        uint32
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
}

// IConfigProviderConfigProviderInitParams is an auto generated low-level Go binding around an user-defined struct.
type IConfigProviderConfigProviderInitParams struct {
	VotingPowerProviders    []IConfigProviderCrossChainAddress
	KeysProvider            IConfigProviderCrossChainAddress
	Replicas                []IConfigProviderCrossChainAddress
	VerificationType        uint32
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
}

// IConfigProviderCrossChainAddress is an auto generated low-level Go binding around an user-defined struct.
type IConfigProviderCrossChainAddress struct {
	Addr    common.Address
	ChainId uint64
}

// IEpochManagerEpochManagerInitParams is an auto generated low-level Go binding around an user-defined struct.
type IEpochManagerEpochManagerInitParams struct {
	EpochDuration          *big.Int
	EpochDurationTimestamp *big.Int
}

// INetworkManagerNetworkManagerInitParams is an auto generated low-level Go binding around an user-defined struct.
type INetworkManagerNetworkManagerInitParams struct {
	Network      common.Address
	SubnetworkID *big.Int
}

// IOzEIP712OzEIP712InitParams is an auto generated low-level Go binding around an user-defined struct.
type IOzEIP712OzEIP712InitParams struct {
	Name    string
	Version string
}

// ISettlementExtraData is an auto generated low-level Go binding around an user-defined struct.
type ISettlementExtraData struct {
	Key   [32]byte
	Value [32]byte
}

// ISettlementSettlementInitParams is an auto generated low-level Go binding around an user-defined struct.
type ISettlementSettlementInitParams struct {
	NetworkManagerInitParams INetworkManagerNetworkManagerInitParams
	EpochManagerInitParams   IEpochManagerEpochManagerInitParams
	OzEip712InitParams       IOzEIP712OzEIP712InitParams
	CommitDuration           *big.Int
	ProlongDuration          *big.Int
	RequiredKeyTag           uint8
	SigVerifier              common.Address
}

// ISettlementValSetHeader is an auto generated low-level Go binding around an user-defined struct.
type ISettlementValSetHeader struct {
	Version            uint8
	RequiredKeyTag     uint8
	Epoch              *big.Int
	CaptureTimestamp   *big.Int
	QuorumThreshold    *big.Int
	ValidatorsSszMRoot [32]byte
	PreviousHeaderHash [32]byte
}

// MasterMetaData contains all meta data concerning the Master contract.
var MasterMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"ConfigProvider_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"EpochManager_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"NETWORK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"NetworkManager_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"OzAccessControl_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OzEIP712_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PermissionManager_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SUBNETWORK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SUBNETWORK_IDENTIFIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint96\",\"internalType\":\"uint96\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"Settlement_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"VALIDATOR_SET_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"addReplica\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addVotingPowerProvider\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"commitValSetHeader\",\"inputs\":[{\"name\":\"header\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"captureTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousHeaderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"extraData\",\"type\":\"tuple[]\",\"internalType\":\"structISettlement.ExtraData[]\",\"components\":[{\"name\":\"key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"value\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"proof\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCaptureTimestamp\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCaptureTimestampFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCaptureTimestampFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCommitDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCommitDurationAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getConfig\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.Config\",\"components\":[{\"name\":\"votingPowerProviders\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"replicas\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"verificationType\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxValidatorsCount\",\"type\":\"uint208\",\"internalType\":\"uint208\"},{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getConfigAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.Config\",\"components\":[{\"name\":\"votingPowerProviders\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"replicas\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"verificationType\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxValidatorsCount\",\"type\":\"uint208\",\"internalType\":\"uint208\"},{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpochDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpochStart\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentPhase\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumISettlement.ValSetPhase\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentValSetEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentValSetTimestamp\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochDuration\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochIndex\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochStart\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExtraData\",\"inputs\":[{\"name\":\"key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExtraDataAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getKeysProvider\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getKeysProviderAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLastCommittedHeaderEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxValidatorsCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint208\",\"internalType\":\"uint208\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxValidatorsCountAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint208\",\"internalType\":\"uint208\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxVotingPower\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxVotingPowerAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMinInclusionVotingPower\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMinInclusionVotingPowerAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNextEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNextEpochDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNextEpochStart\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPreviousHeaderHashFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPreviousHeaderHashFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getProlongDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getProlongDurationAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getQuorumThresholdFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getQuorumThresholdFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReplicas\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReplicasAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTag\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTagAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTagFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTagFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTags\",\"inputs\":[],\"outputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTagsAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRole\",\"inputs\":[{\"name\":\"selector\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSigVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSigVerifierAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"header\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"captureTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousHeaderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"captureTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousHeaderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetHeaderHash\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetHeaderHashAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValidatorsSszMRootFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValidatorsSszMRootFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVerificationType\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVerificationTypeAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVersionFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVersionFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVotingPowerProviders\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVotingPowerProvidersAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4CrossChain\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"settlementInitParams\",\"type\":\"tuple\",\"internalType\":\"structISettlement.SettlementInitParams\",\"components\":[{\"name\":\"networkManagerInitParams\",\"type\":\"tuple\",\"internalType\":\"structINetworkManager.NetworkManagerInitParams\",\"components\":[{\"name\":\"network\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"subnetworkID\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]},{\"name\":\"epochManagerInitParams\",\"type\":\"tuple\",\"internalType\":\"structIEpochManager.EpochManagerInitParams\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"epochDurationTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]},{\"name\":\"ozEip712InitParams\",\"type\":\"tuple\",\"internalType\":\"structIOzEIP712.OzEIP712InitParams\",\"components\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"name\":\"commitDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"prolongDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"sigVerifier\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"name\":\"configProviderInitParams\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.ConfigProviderInitParams\",\"components\":[{\"name\":\"votingPowerProviders\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"replicas\",\"type\":\"tuple[]\",\"internalType\":\"structIConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"verificationType\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxValidatorsCount\",\"type\":\"uint208\",\"internalType\":\"uint208\"},{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}]},{\"name\":\"defaultAdmin\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isReplicaRegistered\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isReplicaRegisteredAt\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isValSetHeaderCommitted\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isValSetHeaderCommittedAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isVotingPowerProviderRegistered\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isVotingPowerProviderRegisteredAt\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"multicall\",\"inputs\":[{\"name\":\"data\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"results\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeReplica\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeVotingPowerProvider\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setCommitDuration\",\"inputs\":[{\"name\":\"commitDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setEpochDuration\",\"inputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGenesis\",\"inputs\":[{\"name\":\"valSetHeader\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"captureTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousHeaderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"extraData\",\"type\":\"tuple[]\",\"internalType\":\"structISettlement.ExtraData[]\",\"components\":[{\"name\":\"key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"value\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setKeysProvider\",\"inputs\":[{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxValidatorsCount\",\"inputs\":[{\"name\":\"maxValidatorsCount\",\"type\":\"uint208\",\"internalType\":\"uint208\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxVotingPower\",\"inputs\":[{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinInclusionVotingPower\",\"inputs\":[{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setProlongDuration\",\"inputs\":[{\"name\":\"prolongDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRequiredKeyTag\",\"inputs\":[{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRequiredKeyTags\",\"inputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSigVerifier\",\"inputs\":[{\"name\":\"sigVerifier\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setVerificationType\",\"inputs\":[{\"name\":\"verificationType\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"staticDelegateCall\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifyQuorumSig\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"proof\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"AddReplica\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AddVotingPowerProvider\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CommitValSetHeader\",\"inputs\":[{\"name\":\"valSetHeader\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"captureTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousHeaderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"extraData\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structISettlement.ExtraData[]\",\"components\":[{\"name\":\"key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"value\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitCommitDuration\",\"inputs\":[{\"name\":\"commitDuration\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitEIP712\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitEpochDuration\",\"inputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"},{\"name\":\"epochDurationTimestamp\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitProlongDuration\",\"inputs\":[{\"name\":\"prolongDuration\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitRequiredKeyTag\",\"inputs\":[{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitSigVerifier\",\"inputs\":[{\"name\":\"sigVerifier\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitSubnetwork\",\"inputs\":[{\"name\":\"network\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"subnetworkID\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RemoveReplica\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RemoveVotingPowerProvider\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetCommitDuration\",\"inputs\":[{\"name\":\"commitDuration\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetEpochDuration\",\"inputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetGenesis\",\"inputs\":[{\"name\":\"valSetHeader\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"captureTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"quorumThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousHeaderHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"extraData\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structISettlement.ExtraData[]\",\"components\":[{\"name\":\"key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"value\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetKeysProvider\",\"inputs\":[{\"name\":\"keysProvider\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetMaxValidatorsCount\",\"inputs\":[{\"name\":\"maxValidatorsCount\",\"type\":\"uint208\",\"indexed\":false,\"internalType\":\"uint208\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetMaxVotingPower\",\"inputs\":[{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetMinInclusionVotingPower\",\"inputs\":[{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetProlongDuration\",\"inputs\":[{\"name\":\"prolongDuration\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetRequiredKeyTag\",\"inputs\":[{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetRequiredKeyTags\",\"inputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"indexed\":false,\"internalType\":\"uint8[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetSelectorRole\",\"inputs\":[{\"name\":\"selector\",\"type\":\"bytes4\",\"indexed\":true,\"internalType\":\"bytes4\"},{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetSigVerifier\",\"inputs\":[{\"name\":\"sigVerifier\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetVerificationType\",\"inputs\":[{\"name\":\"verificationType\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ConfigProvider_AlreadyAdded\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ConfigProvider_NotAdded\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_InvalidEpochDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_InvalidEpochDurationIndex\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_InvalidEpochDurationTimestamp\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_NoCheckpoint\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_CommitDurationTooLong\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_CommitDurationTooShort\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_Duplicate\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_EpochDurationTooShort\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidCaptureTimestamp\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidEpoch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidPhase\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidVersion\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_ValSetHeaderAlreadySubmitted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_VerificationFailed\",\"inputs\":[]}]",
}

// MasterABI is the input ABI used to generate the binding from.
// Deprecated: Use MasterMetaData.ABI instead.
var MasterABI = MasterMetaData.ABI

// Master is an auto generated Go binding around an Ethereum contract.
type Master struct {
	MasterCaller     // Read-only binding to the contract
	MasterTransactor // Write-only binding to the contract
	MasterFilterer   // Log filterer for contract events
}

// MasterCaller is an auto generated read-only Go binding around an Ethereum contract.
type MasterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MasterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MasterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MasterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MasterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MasterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MasterSession struct {
	Contract     *Master           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MasterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MasterCallerSession struct {
	Contract *MasterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// MasterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MasterTransactorSession struct {
	Contract     *MasterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MasterRaw is an auto generated low-level Go binding around an Ethereum contract.
type MasterRaw struct {
	Contract *Master // Generic contract binding to access the raw methods on
}

// MasterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MasterCallerRaw struct {
	Contract *MasterCaller // Generic read-only contract binding to access the raw methods on
}

// MasterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MasterTransactorRaw struct {
	Contract *MasterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMaster creates a new instance of Master, bound to a specific deployed contract.
func NewMaster(address common.Address, backend bind.ContractBackend) (*Master, error) {
	contract, err := bindMaster(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Master{MasterCaller: MasterCaller{contract: contract}, MasterTransactor: MasterTransactor{contract: contract}, MasterFilterer: MasterFilterer{contract: contract}}, nil
}

// NewMasterCaller creates a new read-only instance of Master, bound to a specific deployed contract.
func NewMasterCaller(address common.Address, caller bind.ContractCaller) (*MasterCaller, error) {
	contract, err := bindMaster(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MasterCaller{contract: contract}, nil
}

// NewMasterTransactor creates a new write-only instance of Master, bound to a specific deployed contract.
func NewMasterTransactor(address common.Address, transactor bind.ContractTransactor) (*MasterTransactor, error) {
	contract, err := bindMaster(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MasterTransactor{contract: contract}, nil
}

// NewMasterFilterer creates a new log filterer instance of Master, bound to a specific deployed contract.
func NewMasterFilterer(address common.Address, filterer bind.ContractFilterer) (*MasterFilterer, error) {
	contract, err := bindMaster(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MasterFilterer{contract: contract}, nil
}

// bindMaster binds a generic wrapper to an already deployed contract.
func bindMaster(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MasterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Master *MasterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Master.Contract.MasterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Master *MasterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Master.Contract.MasterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Master *MasterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Master.Contract.MasterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Master *MasterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Master.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Master *MasterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Master.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Master *MasterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Master.Contract.contract.Transact(opts, method, params...)
}

// ConfigProviderVERSION is a free data retrieval call binding the contract method 0x00ff780d.
//
// Solidity: function ConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterCaller) ConfigProviderVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "ConfigProvider_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// ConfigProviderVERSION is a free data retrieval call binding the contract method 0x00ff780d.
//
// Solidity: function ConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterSession) ConfigProviderVERSION() (uint64, error) {
	return _Master.Contract.ConfigProviderVERSION(&_Master.CallOpts)
}

// ConfigProviderVERSION is a free data retrieval call binding the contract method 0x00ff780d.
//
// Solidity: function ConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterCallerSession) ConfigProviderVERSION() (uint64, error) {
	return _Master.Contract.ConfigProviderVERSION(&_Master.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Master *MasterCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Master *MasterSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Master.Contract.DEFAULTADMINROLE(&_Master.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Master *MasterCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Master.Contract.DEFAULTADMINROLE(&_Master.CallOpts)
}

// EpochManagerVERSION is a free data retrieval call binding the contract method 0xe7e77e3f.
//
// Solidity: function EpochManager_VERSION() pure returns(uint64)
func (_Master *MasterCaller) EpochManagerVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "EpochManager_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// EpochManagerVERSION is a free data retrieval call binding the contract method 0xe7e77e3f.
//
// Solidity: function EpochManager_VERSION() pure returns(uint64)
func (_Master *MasterSession) EpochManagerVERSION() (uint64, error) {
	return _Master.Contract.EpochManagerVERSION(&_Master.CallOpts)
}

// EpochManagerVERSION is a free data retrieval call binding the contract method 0xe7e77e3f.
//
// Solidity: function EpochManager_VERSION() pure returns(uint64)
func (_Master *MasterCallerSession) EpochManagerVERSION() (uint64, error) {
	return _Master.Contract.EpochManagerVERSION(&_Master.CallOpts)
}

// NETWORK is a free data retrieval call binding the contract method 0x8759e6d1.
//
// Solidity: function NETWORK() view returns(address)
func (_Master *MasterCaller) NETWORK(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "NETWORK")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NETWORK is a free data retrieval call binding the contract method 0x8759e6d1.
//
// Solidity: function NETWORK() view returns(address)
func (_Master *MasterSession) NETWORK() (common.Address, error) {
	return _Master.Contract.NETWORK(&_Master.CallOpts)
}

// NETWORK is a free data retrieval call binding the contract method 0x8759e6d1.
//
// Solidity: function NETWORK() view returns(address)
func (_Master *MasterCallerSession) NETWORK() (common.Address, error) {
	return _Master.Contract.NETWORK(&_Master.CallOpts)
}

// NetworkManagerVERSION is a free data retrieval call binding the contract method 0x50e5963b.
//
// Solidity: function NetworkManager_VERSION() pure returns(uint64)
func (_Master *MasterCaller) NetworkManagerVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "NetworkManager_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// NetworkManagerVERSION is a free data retrieval call binding the contract method 0x50e5963b.
//
// Solidity: function NetworkManager_VERSION() pure returns(uint64)
func (_Master *MasterSession) NetworkManagerVERSION() (uint64, error) {
	return _Master.Contract.NetworkManagerVERSION(&_Master.CallOpts)
}

// NetworkManagerVERSION is a free data retrieval call binding the contract method 0x50e5963b.
//
// Solidity: function NetworkManager_VERSION() pure returns(uint64)
func (_Master *MasterCallerSession) NetworkManagerVERSION() (uint64, error) {
	return _Master.Contract.NetworkManagerVERSION(&_Master.CallOpts)
}

// OzAccessControlVERSION is a free data retrieval call binding the contract method 0xc52a6697.
//
// Solidity: function OzAccessControl_VERSION() view returns(uint64)
func (_Master *MasterCaller) OzAccessControlVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "OzAccessControl_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// OzAccessControlVERSION is a free data retrieval call binding the contract method 0xc52a6697.
//
// Solidity: function OzAccessControl_VERSION() view returns(uint64)
func (_Master *MasterSession) OzAccessControlVERSION() (uint64, error) {
	return _Master.Contract.OzAccessControlVERSION(&_Master.CallOpts)
}

// OzAccessControlVERSION is a free data retrieval call binding the contract method 0xc52a6697.
//
// Solidity: function OzAccessControl_VERSION() view returns(uint64)
func (_Master *MasterCallerSession) OzAccessControlVERSION() (uint64, error) {
	return _Master.Contract.OzAccessControlVERSION(&_Master.CallOpts)
}

// OzEIP712VERSION is a free data retrieval call binding the contract method 0x12691577.
//
// Solidity: function OzEIP712_VERSION() view returns(uint64)
func (_Master *MasterCaller) OzEIP712VERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "OzEIP712_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// OzEIP712VERSION is a free data retrieval call binding the contract method 0x12691577.
//
// Solidity: function OzEIP712_VERSION() view returns(uint64)
func (_Master *MasterSession) OzEIP712VERSION() (uint64, error) {
	return _Master.Contract.OzEIP712VERSION(&_Master.CallOpts)
}

// OzEIP712VERSION is a free data retrieval call binding the contract method 0x12691577.
//
// Solidity: function OzEIP712_VERSION() view returns(uint64)
func (_Master *MasterCallerSession) OzEIP712VERSION() (uint64, error) {
	return _Master.Contract.OzEIP712VERSION(&_Master.CallOpts)
}

// PermissionManagerVERSION is a free data retrieval call binding the contract method 0x997f3b38.
//
// Solidity: function PermissionManager_VERSION() view returns(uint64)
func (_Master *MasterCaller) PermissionManagerVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "PermissionManager_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// PermissionManagerVERSION is a free data retrieval call binding the contract method 0x997f3b38.
//
// Solidity: function PermissionManager_VERSION() view returns(uint64)
func (_Master *MasterSession) PermissionManagerVERSION() (uint64, error) {
	return _Master.Contract.PermissionManagerVERSION(&_Master.CallOpts)
}

// PermissionManagerVERSION is a free data retrieval call binding the contract method 0x997f3b38.
//
// Solidity: function PermissionManager_VERSION() view returns(uint64)
func (_Master *MasterCallerSession) PermissionManagerVERSION() (uint64, error) {
	return _Master.Contract.PermissionManagerVERSION(&_Master.CallOpts)
}

// SUBNETWORK is a free data retrieval call binding the contract method 0x773e6b54.
//
// Solidity: function SUBNETWORK() view returns(bytes32)
func (_Master *MasterCaller) SUBNETWORK(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "SUBNETWORK")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SUBNETWORK is a free data retrieval call binding the contract method 0x773e6b54.
//
// Solidity: function SUBNETWORK() view returns(bytes32)
func (_Master *MasterSession) SUBNETWORK() ([32]byte, error) {
	return _Master.Contract.SUBNETWORK(&_Master.CallOpts)
}

// SUBNETWORK is a free data retrieval call binding the contract method 0x773e6b54.
//
// Solidity: function SUBNETWORK() view returns(bytes32)
func (_Master *MasterCallerSession) SUBNETWORK() ([32]byte, error) {
	return _Master.Contract.SUBNETWORK(&_Master.CallOpts)
}

// SUBNETWORKIDENTIFIER is a free data retrieval call binding the contract method 0xabacb807.
//
// Solidity: function SUBNETWORK_IDENTIFIER() view returns(uint96)
func (_Master *MasterCaller) SUBNETWORKIDENTIFIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "SUBNETWORK_IDENTIFIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SUBNETWORKIDENTIFIER is a free data retrieval call binding the contract method 0xabacb807.
//
// Solidity: function SUBNETWORK_IDENTIFIER() view returns(uint96)
func (_Master *MasterSession) SUBNETWORKIDENTIFIER() (*big.Int, error) {
	return _Master.Contract.SUBNETWORKIDENTIFIER(&_Master.CallOpts)
}

// SUBNETWORKIDENTIFIER is a free data retrieval call binding the contract method 0xabacb807.
//
// Solidity: function SUBNETWORK_IDENTIFIER() view returns(uint96)
func (_Master *MasterCallerSession) SUBNETWORKIDENTIFIER() (*big.Int, error) {
	return _Master.Contract.SUBNETWORKIDENTIFIER(&_Master.CallOpts)
}

// SettlementVERSION is a free data retrieval call binding the contract method 0xce934e06.
//
// Solidity: function Settlement_VERSION() pure returns(uint64)
func (_Master *MasterCaller) SettlementVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "Settlement_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// SettlementVERSION is a free data retrieval call binding the contract method 0xce934e06.
//
// Solidity: function Settlement_VERSION() pure returns(uint64)
func (_Master *MasterSession) SettlementVERSION() (uint64, error) {
	return _Master.Contract.SettlementVERSION(&_Master.CallOpts)
}

// SettlementVERSION is a free data retrieval call binding the contract method 0xce934e06.
//
// Solidity: function Settlement_VERSION() pure returns(uint64)
func (_Master *MasterCallerSession) SettlementVERSION() (uint64, error) {
	return _Master.Contract.SettlementVERSION(&_Master.CallOpts)
}

// VALIDATORSETVERSION is a free data retrieval call binding the contract method 0x321d7b8d.
//
// Solidity: function VALIDATOR_SET_VERSION() pure returns(uint8)
func (_Master *MasterCaller) VALIDATORSETVERSION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "VALIDATOR_SET_VERSION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// VALIDATORSETVERSION is a free data retrieval call binding the contract method 0x321d7b8d.
//
// Solidity: function VALIDATOR_SET_VERSION() pure returns(uint8)
func (_Master *MasterSession) VALIDATORSETVERSION() (uint8, error) {
	return _Master.Contract.VALIDATORSETVERSION(&_Master.CallOpts)
}

// VALIDATORSETVERSION is a free data retrieval call binding the contract method 0x321d7b8d.
//
// Solidity: function VALIDATOR_SET_VERSION() pure returns(uint8)
func (_Master *MasterCallerSession) VALIDATORSETVERSION() (uint8, error) {
	return _Master.Contract.VALIDATORSETVERSION(&_Master.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Master *MasterCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Master *MasterSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _Master.Contract.Eip712Domain(&_Master.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Master *MasterCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _Master.Contract.Eip712Domain(&_Master.CallOpts)
}

// GetCaptureTimestamp is a free data retrieval call binding the contract method 0xdb3adf12.
//
// Solidity: function getCaptureTimestamp() view returns(uint48)
func (_Master *MasterCaller) GetCaptureTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCaptureTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCaptureTimestamp is a free data retrieval call binding the contract method 0xdb3adf12.
//
// Solidity: function getCaptureTimestamp() view returns(uint48)
func (_Master *MasterSession) GetCaptureTimestamp() (*big.Int, error) {
	return _Master.Contract.GetCaptureTimestamp(&_Master.CallOpts)
}

// GetCaptureTimestamp is a free data retrieval call binding the contract method 0xdb3adf12.
//
// Solidity: function getCaptureTimestamp() view returns(uint48)
func (_Master *MasterCallerSession) GetCaptureTimestamp() (*big.Int, error) {
	return _Master.Contract.GetCaptureTimestamp(&_Master.CallOpts)
}

// GetCaptureTimestampFromValSetHeader is a free data retrieval call binding the contract method 0xf4935d39.
//
// Solidity: function getCaptureTimestampFromValSetHeader() view returns(uint48)
func (_Master *MasterCaller) GetCaptureTimestampFromValSetHeader(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCaptureTimestampFromValSetHeader")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCaptureTimestampFromValSetHeader is a free data retrieval call binding the contract method 0xf4935d39.
//
// Solidity: function getCaptureTimestampFromValSetHeader() view returns(uint48)
func (_Master *MasterSession) GetCaptureTimestampFromValSetHeader() (*big.Int, error) {
	return _Master.Contract.GetCaptureTimestampFromValSetHeader(&_Master.CallOpts)
}

// GetCaptureTimestampFromValSetHeader is a free data retrieval call binding the contract method 0xf4935d39.
//
// Solidity: function getCaptureTimestampFromValSetHeader() view returns(uint48)
func (_Master *MasterCallerSession) GetCaptureTimestampFromValSetHeader() (*big.Int, error) {
	return _Master.Contract.GetCaptureTimestampFromValSetHeader(&_Master.CallOpts)
}

// GetCaptureTimestampFromValSetHeaderAt is a free data retrieval call binding the contract method 0x5485b549.
//
// Solidity: function getCaptureTimestampFromValSetHeaderAt(uint48 epoch) view returns(uint48)
func (_Master *MasterCaller) GetCaptureTimestampFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCaptureTimestampFromValSetHeaderAt", epoch)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCaptureTimestampFromValSetHeaderAt is a free data retrieval call binding the contract method 0x5485b549.
//
// Solidity: function getCaptureTimestampFromValSetHeaderAt(uint48 epoch) view returns(uint48)
func (_Master *MasterSession) GetCaptureTimestampFromValSetHeaderAt(epoch *big.Int) (*big.Int, error) {
	return _Master.Contract.GetCaptureTimestampFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetCaptureTimestampFromValSetHeaderAt is a free data retrieval call binding the contract method 0x5485b549.
//
// Solidity: function getCaptureTimestampFromValSetHeaderAt(uint48 epoch) view returns(uint48)
func (_Master *MasterCallerSession) GetCaptureTimestampFromValSetHeaderAt(epoch *big.Int) (*big.Int, error) {
	return _Master.Contract.GetCaptureTimestampFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetCommitDuration is a free data retrieval call binding the contract method 0x62a0befd.
//
// Solidity: function getCommitDuration() view returns(uint48)
func (_Master *MasterCaller) GetCommitDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCommitDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommitDuration is a free data retrieval call binding the contract method 0x62a0befd.
//
// Solidity: function getCommitDuration() view returns(uint48)
func (_Master *MasterSession) GetCommitDuration() (*big.Int, error) {
	return _Master.Contract.GetCommitDuration(&_Master.CallOpts)
}

// GetCommitDuration is a free data retrieval call binding the contract method 0x62a0befd.
//
// Solidity: function getCommitDuration() view returns(uint48)
func (_Master *MasterCallerSession) GetCommitDuration() (*big.Int, error) {
	return _Master.Contract.GetCommitDuration(&_Master.CallOpts)
}

// GetCommitDurationAt is a free data retrieval call binding the contract method 0x8f7a22c6.
//
// Solidity: function getCommitDurationAt(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterCaller) GetCommitDurationAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCommitDurationAt", timestamp, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommitDurationAt is a free data retrieval call binding the contract method 0x8f7a22c6.
//
// Solidity: function getCommitDurationAt(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterSession) GetCommitDurationAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetCommitDurationAt(&_Master.CallOpts, timestamp, hint)
}

// GetCommitDurationAt is a free data retrieval call binding the contract method 0x8f7a22c6.
//
// Solidity: function getCommitDurationAt(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterCallerSession) GetCommitDurationAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetCommitDurationAt(&_Master.CallOpts, timestamp, hint)
}

// GetConfig is a free data retrieval call binding the contract method 0xc3f909d4.
//
// Solidity: function getConfig() view returns(((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]))
func (_Master *MasterCaller) GetConfig(opts *bind.CallOpts) (IConfigProviderConfig, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getConfig")

	if err != nil {
		return *new(IConfigProviderConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IConfigProviderConfig)).(*IConfigProviderConfig)

	return out0, err

}

// GetConfig is a free data retrieval call binding the contract method 0xc3f909d4.
//
// Solidity: function getConfig() view returns(((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]))
func (_Master *MasterSession) GetConfig() (IConfigProviderConfig, error) {
	return _Master.Contract.GetConfig(&_Master.CallOpts)
}

// GetConfig is a free data retrieval call binding the contract method 0xc3f909d4.
//
// Solidity: function getConfig() view returns(((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]))
func (_Master *MasterCallerSession) GetConfig() (IConfigProviderConfig, error) {
	return _Master.Contract.GetConfig(&_Master.CallOpts)
}

// GetConfigAt is a free data retrieval call binding the contract method 0xf633dfc6.
//
// Solidity: function getConfigAt(uint48 timestamp, bytes hints) view returns(((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]))
func (_Master *MasterCaller) GetConfigAt(opts *bind.CallOpts, timestamp *big.Int, hints []byte) (IConfigProviderConfig, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getConfigAt", timestamp, hints)

	if err != nil {
		return *new(IConfigProviderConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IConfigProviderConfig)).(*IConfigProviderConfig)

	return out0, err

}

// GetConfigAt is a free data retrieval call binding the contract method 0xf633dfc6.
//
// Solidity: function getConfigAt(uint48 timestamp, bytes hints) view returns(((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]))
func (_Master *MasterSession) GetConfigAt(timestamp *big.Int, hints []byte) (IConfigProviderConfig, error) {
	return _Master.Contract.GetConfigAt(&_Master.CallOpts, timestamp, hints)
}

// GetConfigAt is a free data retrieval call binding the contract method 0xf633dfc6.
//
// Solidity: function getConfigAt(uint48 timestamp, bytes hints) view returns(((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]))
func (_Master *MasterCallerSession) GetConfigAt(timestamp *big.Int, hints []byte) (IConfigProviderConfig, error) {
	return _Master.Contract.GetConfigAt(&_Master.CallOpts, timestamp, hints)
}

// GetCurrentEpoch is a free data retrieval call binding the contract method 0xb97dd9e2.
//
// Solidity: function getCurrentEpoch() view returns(uint48)
func (_Master *MasterCaller) GetCurrentEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCurrentEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentEpoch is a free data retrieval call binding the contract method 0xb97dd9e2.
//
// Solidity: function getCurrentEpoch() view returns(uint48)
func (_Master *MasterSession) GetCurrentEpoch() (*big.Int, error) {
	return _Master.Contract.GetCurrentEpoch(&_Master.CallOpts)
}

// GetCurrentEpoch is a free data retrieval call binding the contract method 0xb97dd9e2.
//
// Solidity: function getCurrentEpoch() view returns(uint48)
func (_Master *MasterCallerSession) GetCurrentEpoch() (*big.Int, error) {
	return _Master.Contract.GetCurrentEpoch(&_Master.CallOpts)
}

// GetCurrentEpochDuration is a free data retrieval call binding the contract method 0x558e2eb6.
//
// Solidity: function getCurrentEpochDuration() view returns(uint48 epochDuration)
func (_Master *MasterCaller) GetCurrentEpochDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCurrentEpochDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentEpochDuration is a free data retrieval call binding the contract method 0x558e2eb6.
//
// Solidity: function getCurrentEpochDuration() view returns(uint48 epochDuration)
func (_Master *MasterSession) GetCurrentEpochDuration() (*big.Int, error) {
	return _Master.Contract.GetCurrentEpochDuration(&_Master.CallOpts)
}

// GetCurrentEpochDuration is a free data retrieval call binding the contract method 0x558e2eb6.
//
// Solidity: function getCurrentEpochDuration() view returns(uint48 epochDuration)
func (_Master *MasterCallerSession) GetCurrentEpochDuration() (*big.Int, error) {
	return _Master.Contract.GetCurrentEpochDuration(&_Master.CallOpts)
}

// GetCurrentEpochStart is a free data retrieval call binding the contract method 0xa6e16c4d.
//
// Solidity: function getCurrentEpochStart() view returns(uint48)
func (_Master *MasterCaller) GetCurrentEpochStart(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCurrentEpochStart")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentEpochStart is a free data retrieval call binding the contract method 0xa6e16c4d.
//
// Solidity: function getCurrentEpochStart() view returns(uint48)
func (_Master *MasterSession) GetCurrentEpochStart() (*big.Int, error) {
	return _Master.Contract.GetCurrentEpochStart(&_Master.CallOpts)
}

// GetCurrentEpochStart is a free data retrieval call binding the contract method 0xa6e16c4d.
//
// Solidity: function getCurrentEpochStart() view returns(uint48)
func (_Master *MasterCallerSession) GetCurrentEpochStart() (*big.Int, error) {
	return _Master.Contract.GetCurrentEpochStart(&_Master.CallOpts)
}

// GetCurrentPhase is a free data retrieval call binding the contract method 0xa3a40ea5.
//
// Solidity: function getCurrentPhase() view returns(uint8)
func (_Master *MasterCaller) GetCurrentPhase(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCurrentPhase")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetCurrentPhase is a free data retrieval call binding the contract method 0xa3a40ea5.
//
// Solidity: function getCurrentPhase() view returns(uint8)
func (_Master *MasterSession) GetCurrentPhase() (uint8, error) {
	return _Master.Contract.GetCurrentPhase(&_Master.CallOpts)
}

// GetCurrentPhase is a free data retrieval call binding the contract method 0xa3a40ea5.
//
// Solidity: function getCurrentPhase() view returns(uint8)
func (_Master *MasterCallerSession) GetCurrentPhase() (uint8, error) {
	return _Master.Contract.GetCurrentPhase(&_Master.CallOpts)
}

// GetCurrentValSetEpoch is a free data retrieval call binding the contract method 0xbb72f2e8.
//
// Solidity: function getCurrentValSetEpoch() view returns(uint48)
func (_Master *MasterCaller) GetCurrentValSetEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCurrentValSetEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentValSetEpoch is a free data retrieval call binding the contract method 0xbb72f2e8.
//
// Solidity: function getCurrentValSetEpoch() view returns(uint48)
func (_Master *MasterSession) GetCurrentValSetEpoch() (*big.Int, error) {
	return _Master.Contract.GetCurrentValSetEpoch(&_Master.CallOpts)
}

// GetCurrentValSetEpoch is a free data retrieval call binding the contract method 0xbb72f2e8.
//
// Solidity: function getCurrentValSetEpoch() view returns(uint48)
func (_Master *MasterCallerSession) GetCurrentValSetEpoch() (*big.Int, error) {
	return _Master.Contract.GetCurrentValSetEpoch(&_Master.CallOpts)
}

// GetCurrentValSetTimestamp is a free data retrieval call binding the contract method 0x1bd046bb.
//
// Solidity: function getCurrentValSetTimestamp() view returns(uint48)
func (_Master *MasterCaller) GetCurrentValSetTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCurrentValSetTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentValSetTimestamp is a free data retrieval call binding the contract method 0x1bd046bb.
//
// Solidity: function getCurrentValSetTimestamp() view returns(uint48)
func (_Master *MasterSession) GetCurrentValSetTimestamp() (*big.Int, error) {
	return _Master.Contract.GetCurrentValSetTimestamp(&_Master.CallOpts)
}

// GetCurrentValSetTimestamp is a free data retrieval call binding the contract method 0x1bd046bb.
//
// Solidity: function getCurrentValSetTimestamp() view returns(uint48)
func (_Master *MasterCallerSession) GetCurrentValSetTimestamp() (*big.Int, error) {
	return _Master.Contract.GetCurrentValSetTimestamp(&_Master.CallOpts)
}

// GetEpochDuration is a free data retrieval call binding the contract method 0x3a5f8abd.
//
// Solidity: function getEpochDuration(uint48 epoch, bytes hint) view returns(uint48 epochDuration)
func (_Master *MasterCaller) GetEpochDuration(opts *bind.CallOpts, epoch *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getEpochDuration", epoch, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochDuration is a free data retrieval call binding the contract method 0x3a5f8abd.
//
// Solidity: function getEpochDuration(uint48 epoch, bytes hint) view returns(uint48 epochDuration)
func (_Master *MasterSession) GetEpochDuration(epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetEpochDuration(&_Master.CallOpts, epoch, hint)
}

// GetEpochDuration is a free data retrieval call binding the contract method 0x3a5f8abd.
//
// Solidity: function getEpochDuration(uint48 epoch, bytes hint) view returns(uint48 epochDuration)
func (_Master *MasterCallerSession) GetEpochDuration(epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetEpochDuration(&_Master.CallOpts, epoch, hint)
}

// GetEpochIndex is a free data retrieval call binding the contract method 0x36913ca9.
//
// Solidity: function getEpochIndex(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterCaller) GetEpochIndex(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getEpochIndex", timestamp, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochIndex is a free data retrieval call binding the contract method 0x36913ca9.
//
// Solidity: function getEpochIndex(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterSession) GetEpochIndex(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetEpochIndex(&_Master.CallOpts, timestamp, hint)
}

// GetEpochIndex is a free data retrieval call binding the contract method 0x36913ca9.
//
// Solidity: function getEpochIndex(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterCallerSession) GetEpochIndex(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetEpochIndex(&_Master.CallOpts, timestamp, hint)
}

// GetEpochStart is a free data retrieval call binding the contract method 0x40a29a88.
//
// Solidity: function getEpochStart(uint48 epoch, bytes hint) view returns(uint48)
func (_Master *MasterCaller) GetEpochStart(opts *bind.CallOpts, epoch *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getEpochStart", epoch, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochStart is a free data retrieval call binding the contract method 0x40a29a88.
//
// Solidity: function getEpochStart(uint48 epoch, bytes hint) view returns(uint48)
func (_Master *MasterSession) GetEpochStart(epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetEpochStart(&_Master.CallOpts, epoch, hint)
}

// GetEpochStart is a free data retrieval call binding the contract method 0x40a29a88.
//
// Solidity: function getEpochStart(uint48 epoch, bytes hint) view returns(uint48)
func (_Master *MasterCallerSession) GetEpochStart(epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetEpochStart(&_Master.CallOpts, epoch, hint)
}

// GetExtraData is a free data retrieval call binding the contract method 0xecae6344.
//
// Solidity: function getExtraData(bytes32 key) view returns(bytes32)
func (_Master *MasterCaller) GetExtraData(opts *bind.CallOpts, key [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getExtraData", key)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetExtraData is a free data retrieval call binding the contract method 0xecae6344.
//
// Solidity: function getExtraData(bytes32 key) view returns(bytes32)
func (_Master *MasterSession) GetExtraData(key [32]byte) ([32]byte, error) {
	return _Master.Contract.GetExtraData(&_Master.CallOpts, key)
}

// GetExtraData is a free data retrieval call binding the contract method 0xecae6344.
//
// Solidity: function getExtraData(bytes32 key) view returns(bytes32)
func (_Master *MasterCallerSession) GetExtraData(key [32]byte) ([32]byte, error) {
	return _Master.Contract.GetExtraData(&_Master.CallOpts, key)
}

// GetExtraDataAt is a free data retrieval call binding the contract method 0x52bb038a.
//
// Solidity: function getExtraDataAt(uint48 epoch, bytes32 key) view returns(bytes32)
func (_Master *MasterCaller) GetExtraDataAt(opts *bind.CallOpts, epoch *big.Int, key [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getExtraDataAt", epoch, key)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetExtraDataAt is a free data retrieval call binding the contract method 0x52bb038a.
//
// Solidity: function getExtraDataAt(uint48 epoch, bytes32 key) view returns(bytes32)
func (_Master *MasterSession) GetExtraDataAt(epoch *big.Int, key [32]byte) ([32]byte, error) {
	return _Master.Contract.GetExtraDataAt(&_Master.CallOpts, epoch, key)
}

// GetExtraDataAt is a free data retrieval call binding the contract method 0x52bb038a.
//
// Solidity: function getExtraDataAt(uint48 epoch, bytes32 key) view returns(bytes32)
func (_Master *MasterCallerSession) GetExtraDataAt(epoch *big.Int, key [32]byte) ([32]byte, error) {
	return _Master.Contract.GetExtraDataAt(&_Master.CallOpts, epoch, key)
}

// GetKeysProvider is a free data retrieval call binding the contract method 0x297d29b8.
//
// Solidity: function getKeysProvider() view returns((address,uint64))
func (_Master *MasterCaller) GetKeysProvider(opts *bind.CallOpts) (IConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getKeysProvider")

	if err != nil {
		return *new(IConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new(IConfigProviderCrossChainAddress)).(*IConfigProviderCrossChainAddress)

	return out0, err

}

// GetKeysProvider is a free data retrieval call binding the contract method 0x297d29b8.
//
// Solidity: function getKeysProvider() view returns((address,uint64))
func (_Master *MasterSession) GetKeysProvider() (IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProvider(&_Master.CallOpts)
}

// GetKeysProvider is a free data retrieval call binding the contract method 0x297d29b8.
//
// Solidity: function getKeysProvider() view returns((address,uint64))
func (_Master *MasterCallerSession) GetKeysProvider() (IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProvider(&_Master.CallOpts)
}

// GetKeysProviderAt is a free data retrieval call binding the contract method 0xb405b818.
//
// Solidity: function getKeysProviderAt(uint48 timestamp, bytes hint) view returns((address,uint64))
func (_Master *MasterCaller) GetKeysProviderAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (IConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getKeysProviderAt", timestamp, hint)

	if err != nil {
		return *new(IConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new(IConfigProviderCrossChainAddress)).(*IConfigProviderCrossChainAddress)

	return out0, err

}

// GetKeysProviderAt is a free data retrieval call binding the contract method 0xb405b818.
//
// Solidity: function getKeysProviderAt(uint48 timestamp, bytes hint) view returns((address,uint64))
func (_Master *MasterSession) GetKeysProviderAt(timestamp *big.Int, hint []byte) (IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProviderAt(&_Master.CallOpts, timestamp, hint)
}

// GetKeysProviderAt is a free data retrieval call binding the contract method 0xb405b818.
//
// Solidity: function getKeysProviderAt(uint48 timestamp, bytes hint) view returns((address,uint64))
func (_Master *MasterCallerSession) GetKeysProviderAt(timestamp *big.Int, hint []byte) (IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProviderAt(&_Master.CallOpts, timestamp, hint)
}

// GetLastCommittedHeaderEpoch is a free data retrieval call binding the contract method 0x65b0849b.
//
// Solidity: function getLastCommittedHeaderEpoch() view returns(uint48)
func (_Master *MasterCaller) GetLastCommittedHeaderEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getLastCommittedHeaderEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLastCommittedHeaderEpoch is a free data retrieval call binding the contract method 0x65b0849b.
//
// Solidity: function getLastCommittedHeaderEpoch() view returns(uint48)
func (_Master *MasterSession) GetLastCommittedHeaderEpoch() (*big.Int, error) {
	return _Master.Contract.GetLastCommittedHeaderEpoch(&_Master.CallOpts)
}

// GetLastCommittedHeaderEpoch is a free data retrieval call binding the contract method 0x65b0849b.
//
// Solidity: function getLastCommittedHeaderEpoch() view returns(uint48)
func (_Master *MasterCallerSession) GetLastCommittedHeaderEpoch() (*big.Int, error) {
	return _Master.Contract.GetLastCommittedHeaderEpoch(&_Master.CallOpts)
}

// GetMaxValidatorsCount is a free data retrieval call binding the contract method 0x06ce894d.
//
// Solidity: function getMaxValidatorsCount() view returns(uint208)
func (_Master *MasterCaller) GetMaxValidatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMaxValidatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxValidatorsCount is a free data retrieval call binding the contract method 0x06ce894d.
//
// Solidity: function getMaxValidatorsCount() view returns(uint208)
func (_Master *MasterSession) GetMaxValidatorsCount() (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCount(&_Master.CallOpts)
}

// GetMaxValidatorsCount is a free data retrieval call binding the contract method 0x06ce894d.
//
// Solidity: function getMaxValidatorsCount() view returns(uint208)
func (_Master *MasterCallerSession) GetMaxValidatorsCount() (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCount(&_Master.CallOpts)
}

// GetMaxValidatorsCountAt is a free data retrieval call binding the contract method 0x2e1f3b08.
//
// Solidity: function getMaxValidatorsCountAt(uint48 timestamp, bytes hint) view returns(uint208)
func (_Master *MasterCaller) GetMaxValidatorsCountAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMaxValidatorsCountAt", timestamp, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxValidatorsCountAt is a free data retrieval call binding the contract method 0x2e1f3b08.
//
// Solidity: function getMaxValidatorsCountAt(uint48 timestamp, bytes hint) view returns(uint208)
func (_Master *MasterSession) GetMaxValidatorsCountAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCountAt(&_Master.CallOpts, timestamp, hint)
}

// GetMaxValidatorsCountAt is a free data retrieval call binding the contract method 0x2e1f3b08.
//
// Solidity: function getMaxValidatorsCountAt(uint48 timestamp, bytes hint) view returns(uint208)
func (_Master *MasterCallerSession) GetMaxValidatorsCountAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCountAt(&_Master.CallOpts, timestamp, hint)
}

// GetMaxVotingPower is a free data retrieval call binding the contract method 0x9f9c3080.
//
// Solidity: function getMaxVotingPower() view returns(uint256)
func (_Master *MasterCaller) GetMaxVotingPower(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMaxVotingPower")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxVotingPower is a free data retrieval call binding the contract method 0x9f9c3080.
//
// Solidity: function getMaxVotingPower() view returns(uint256)
func (_Master *MasterSession) GetMaxVotingPower() (*big.Int, error) {
	return _Master.Contract.GetMaxVotingPower(&_Master.CallOpts)
}

// GetMaxVotingPower is a free data retrieval call binding the contract method 0x9f9c3080.
//
// Solidity: function getMaxVotingPower() view returns(uint256)
func (_Master *MasterCallerSession) GetMaxVotingPower() (*big.Int, error) {
	return _Master.Contract.GetMaxVotingPower(&_Master.CallOpts)
}

// GetMaxVotingPowerAt is a free data retrieval call binding the contract method 0xab94b4e6.
//
// Solidity: function getMaxVotingPowerAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterCaller) GetMaxVotingPowerAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMaxVotingPowerAt", timestamp, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxVotingPowerAt is a free data retrieval call binding the contract method 0xab94b4e6.
//
// Solidity: function getMaxVotingPowerAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterSession) GetMaxVotingPowerAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMaxVotingPowerAt(&_Master.CallOpts, timestamp, hint)
}

// GetMaxVotingPowerAt is a free data retrieval call binding the contract method 0xab94b4e6.
//
// Solidity: function getMaxVotingPowerAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterCallerSession) GetMaxVotingPowerAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMaxVotingPowerAt(&_Master.CallOpts, timestamp, hint)
}

// GetMinInclusionVotingPower is a free data retrieval call binding the contract method 0xb6a94695.
//
// Solidity: function getMinInclusionVotingPower() view returns(uint256)
func (_Master *MasterCaller) GetMinInclusionVotingPower(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMinInclusionVotingPower")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinInclusionVotingPower is a free data retrieval call binding the contract method 0xb6a94695.
//
// Solidity: function getMinInclusionVotingPower() view returns(uint256)
func (_Master *MasterSession) GetMinInclusionVotingPower() (*big.Int, error) {
	return _Master.Contract.GetMinInclusionVotingPower(&_Master.CallOpts)
}

// GetMinInclusionVotingPower is a free data retrieval call binding the contract method 0xb6a94695.
//
// Solidity: function getMinInclusionVotingPower() view returns(uint256)
func (_Master *MasterCallerSession) GetMinInclusionVotingPower() (*big.Int, error) {
	return _Master.Contract.GetMinInclusionVotingPower(&_Master.CallOpts)
}

// GetMinInclusionVotingPowerAt is a free data retrieval call binding the contract method 0x02c2b11f.
//
// Solidity: function getMinInclusionVotingPowerAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterCaller) GetMinInclusionVotingPowerAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMinInclusionVotingPowerAt", timestamp, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinInclusionVotingPowerAt is a free data retrieval call binding the contract method 0x02c2b11f.
//
// Solidity: function getMinInclusionVotingPowerAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterSession) GetMinInclusionVotingPowerAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMinInclusionVotingPowerAt(&_Master.CallOpts, timestamp, hint)
}

// GetMinInclusionVotingPowerAt is a free data retrieval call binding the contract method 0x02c2b11f.
//
// Solidity: function getMinInclusionVotingPowerAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterCallerSession) GetMinInclusionVotingPowerAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMinInclusionVotingPowerAt(&_Master.CallOpts, timestamp, hint)
}

// GetNextEpoch is a free data retrieval call binding the contract method 0xefe97d05.
//
// Solidity: function getNextEpoch() view returns(uint48)
func (_Master *MasterCaller) GetNextEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getNextEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNextEpoch is a free data retrieval call binding the contract method 0xefe97d05.
//
// Solidity: function getNextEpoch() view returns(uint48)
func (_Master *MasterSession) GetNextEpoch() (*big.Int, error) {
	return _Master.Contract.GetNextEpoch(&_Master.CallOpts)
}

// GetNextEpoch is a free data retrieval call binding the contract method 0xefe97d05.
//
// Solidity: function getNextEpoch() view returns(uint48)
func (_Master *MasterCallerSession) GetNextEpoch() (*big.Int, error) {
	return _Master.Contract.GetNextEpoch(&_Master.CallOpts)
}

// GetNextEpochDuration is a free data retrieval call binding the contract method 0x038cf1c0.
//
// Solidity: function getNextEpochDuration() view returns(uint48)
func (_Master *MasterCaller) GetNextEpochDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getNextEpochDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNextEpochDuration is a free data retrieval call binding the contract method 0x038cf1c0.
//
// Solidity: function getNextEpochDuration() view returns(uint48)
func (_Master *MasterSession) GetNextEpochDuration() (*big.Int, error) {
	return _Master.Contract.GetNextEpochDuration(&_Master.CallOpts)
}

// GetNextEpochDuration is a free data retrieval call binding the contract method 0x038cf1c0.
//
// Solidity: function getNextEpochDuration() view returns(uint48)
func (_Master *MasterCallerSession) GetNextEpochDuration() (*big.Int, error) {
	return _Master.Contract.GetNextEpochDuration(&_Master.CallOpts)
}

// GetNextEpochStart is a free data retrieval call binding the contract method 0x65c5f94a.
//
// Solidity: function getNextEpochStart() view returns(uint48)
func (_Master *MasterCaller) GetNextEpochStart(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getNextEpochStart")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNextEpochStart is a free data retrieval call binding the contract method 0x65c5f94a.
//
// Solidity: function getNextEpochStart() view returns(uint48)
func (_Master *MasterSession) GetNextEpochStart() (*big.Int, error) {
	return _Master.Contract.GetNextEpochStart(&_Master.CallOpts)
}

// GetNextEpochStart is a free data retrieval call binding the contract method 0x65c5f94a.
//
// Solidity: function getNextEpochStart() view returns(uint48)
func (_Master *MasterCallerSession) GetNextEpochStart() (*big.Int, error) {
	return _Master.Contract.GetNextEpochStart(&_Master.CallOpts)
}

// GetPreviousHeaderHashFromValSetHeader is a free data retrieval call binding the contract method 0xe82a99e1.
//
// Solidity: function getPreviousHeaderHashFromValSetHeader() view returns(bytes32)
func (_Master *MasterCaller) GetPreviousHeaderHashFromValSetHeader(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getPreviousHeaderHashFromValSetHeader")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetPreviousHeaderHashFromValSetHeader is a free data retrieval call binding the contract method 0xe82a99e1.
//
// Solidity: function getPreviousHeaderHashFromValSetHeader() view returns(bytes32)
func (_Master *MasterSession) GetPreviousHeaderHashFromValSetHeader() ([32]byte, error) {
	return _Master.Contract.GetPreviousHeaderHashFromValSetHeader(&_Master.CallOpts)
}

// GetPreviousHeaderHashFromValSetHeader is a free data retrieval call binding the contract method 0xe82a99e1.
//
// Solidity: function getPreviousHeaderHashFromValSetHeader() view returns(bytes32)
func (_Master *MasterCallerSession) GetPreviousHeaderHashFromValSetHeader() ([32]byte, error) {
	return _Master.Contract.GetPreviousHeaderHashFromValSetHeader(&_Master.CallOpts)
}

// GetPreviousHeaderHashFromValSetHeaderAt is a free data retrieval call binding the contract method 0x60053c6b.
//
// Solidity: function getPreviousHeaderHashFromValSetHeaderAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterCaller) GetPreviousHeaderHashFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getPreviousHeaderHashFromValSetHeaderAt", epoch)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetPreviousHeaderHashFromValSetHeaderAt is a free data retrieval call binding the contract method 0x60053c6b.
//
// Solidity: function getPreviousHeaderHashFromValSetHeaderAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterSession) GetPreviousHeaderHashFromValSetHeaderAt(epoch *big.Int) ([32]byte, error) {
	return _Master.Contract.GetPreviousHeaderHashFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetPreviousHeaderHashFromValSetHeaderAt is a free data retrieval call binding the contract method 0x60053c6b.
//
// Solidity: function getPreviousHeaderHashFromValSetHeaderAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterCallerSession) GetPreviousHeaderHashFromValSetHeaderAt(epoch *big.Int) ([32]byte, error) {
	return _Master.Contract.GetPreviousHeaderHashFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetProlongDuration is a free data retrieval call binding the contract method 0x4fa1481b.
//
// Solidity: function getProlongDuration() view returns(uint48)
func (_Master *MasterCaller) GetProlongDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getProlongDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProlongDuration is a free data retrieval call binding the contract method 0x4fa1481b.
//
// Solidity: function getProlongDuration() view returns(uint48)
func (_Master *MasterSession) GetProlongDuration() (*big.Int, error) {
	return _Master.Contract.GetProlongDuration(&_Master.CallOpts)
}

// GetProlongDuration is a free data retrieval call binding the contract method 0x4fa1481b.
//
// Solidity: function getProlongDuration() view returns(uint48)
func (_Master *MasterCallerSession) GetProlongDuration() (*big.Int, error) {
	return _Master.Contract.GetProlongDuration(&_Master.CallOpts)
}

// GetProlongDurationAt is a free data retrieval call binding the contract method 0x3586b9ea.
//
// Solidity: function getProlongDurationAt(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterCaller) GetProlongDurationAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getProlongDurationAt", timestamp, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProlongDurationAt is a free data retrieval call binding the contract method 0x3586b9ea.
//
// Solidity: function getProlongDurationAt(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterSession) GetProlongDurationAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetProlongDurationAt(&_Master.CallOpts, timestamp, hint)
}

// GetProlongDurationAt is a free data retrieval call binding the contract method 0x3586b9ea.
//
// Solidity: function getProlongDurationAt(uint48 timestamp, bytes hint) view returns(uint48)
func (_Master *MasterCallerSession) GetProlongDurationAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetProlongDurationAt(&_Master.CallOpts, timestamp, hint)
}

// GetQuorumThresholdFromValSetHeader is a free data retrieval call binding the contract method 0xe586b38e.
//
// Solidity: function getQuorumThresholdFromValSetHeader() view returns(uint256)
func (_Master *MasterCaller) GetQuorumThresholdFromValSetHeader(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getQuorumThresholdFromValSetHeader")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetQuorumThresholdFromValSetHeader is a free data retrieval call binding the contract method 0xe586b38e.
//
// Solidity: function getQuorumThresholdFromValSetHeader() view returns(uint256)
func (_Master *MasterSession) GetQuorumThresholdFromValSetHeader() (*big.Int, error) {
	return _Master.Contract.GetQuorumThresholdFromValSetHeader(&_Master.CallOpts)
}

// GetQuorumThresholdFromValSetHeader is a free data retrieval call binding the contract method 0xe586b38e.
//
// Solidity: function getQuorumThresholdFromValSetHeader() view returns(uint256)
func (_Master *MasterCallerSession) GetQuorumThresholdFromValSetHeader() (*big.Int, error) {
	return _Master.Contract.GetQuorumThresholdFromValSetHeader(&_Master.CallOpts)
}

// GetQuorumThresholdFromValSetHeaderAt is a free data retrieval call binding the contract method 0x1d86bd88.
//
// Solidity: function getQuorumThresholdFromValSetHeaderAt(uint48 epoch) view returns(uint256)
func (_Master *MasterCaller) GetQuorumThresholdFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getQuorumThresholdFromValSetHeaderAt", epoch)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetQuorumThresholdFromValSetHeaderAt is a free data retrieval call binding the contract method 0x1d86bd88.
//
// Solidity: function getQuorumThresholdFromValSetHeaderAt(uint48 epoch) view returns(uint256)
func (_Master *MasterSession) GetQuorumThresholdFromValSetHeaderAt(epoch *big.Int) (*big.Int, error) {
	return _Master.Contract.GetQuorumThresholdFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetQuorumThresholdFromValSetHeaderAt is a free data retrieval call binding the contract method 0x1d86bd88.
//
// Solidity: function getQuorumThresholdFromValSetHeaderAt(uint48 epoch) view returns(uint256)
func (_Master *MasterCallerSession) GetQuorumThresholdFromValSetHeaderAt(epoch *big.Int) (*big.Int, error) {
	return _Master.Contract.GetQuorumThresholdFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetReplicas is a free data retrieval call binding the contract method 0x4df9ffb4.
//
// Solidity: function getReplicas() view returns((address,uint64)[])
func (_Master *MasterCaller) GetReplicas(opts *bind.CallOpts) ([]IConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getReplicas")

	if err != nil {
		return *new([]IConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IConfigProviderCrossChainAddress)).(*[]IConfigProviderCrossChainAddress)

	return out0, err

}

// GetReplicas is a free data retrieval call binding the contract method 0x4df9ffb4.
//
// Solidity: function getReplicas() view returns((address,uint64)[])
func (_Master *MasterSession) GetReplicas() ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetReplicas(&_Master.CallOpts)
}

// GetReplicas is a free data retrieval call binding the contract method 0x4df9ffb4.
//
// Solidity: function getReplicas() view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetReplicas() ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetReplicas(&_Master.CallOpts)
}

// GetReplicasAt is a free data retrieval call binding the contract method 0xf9d85e6d.
//
// Solidity: function getReplicasAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCaller) GetReplicasAt(opts *bind.CallOpts, timestamp *big.Int, hints [][]byte) ([]IConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getReplicasAt", timestamp, hints)

	if err != nil {
		return *new([]IConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IConfigProviderCrossChainAddress)).(*[]IConfigProviderCrossChainAddress)

	return out0, err

}

// GetReplicasAt is a free data retrieval call binding the contract method 0xf9d85e6d.
//
// Solidity: function getReplicasAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterSession) GetReplicasAt(timestamp *big.Int, hints [][]byte) ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetReplicasAt(&_Master.CallOpts, timestamp, hints)
}

// GetReplicasAt is a free data retrieval call binding the contract method 0xf9d85e6d.
//
// Solidity: function getReplicasAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetReplicasAt(timestamp *big.Int, hints [][]byte) ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetReplicasAt(&_Master.CallOpts, timestamp, hints)
}

// GetRequiredKeyTag is a free data retrieval call binding the contract method 0x2e897f93.
//
// Solidity: function getRequiredKeyTag() view returns(uint8)
func (_Master *MasterCaller) GetRequiredKeyTag(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTag")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRequiredKeyTag is a free data retrieval call binding the contract method 0x2e897f93.
//
// Solidity: function getRequiredKeyTag() view returns(uint8)
func (_Master *MasterSession) GetRequiredKeyTag() (uint8, error) {
	return _Master.Contract.GetRequiredKeyTag(&_Master.CallOpts)
}

// GetRequiredKeyTag is a free data retrieval call binding the contract method 0x2e897f93.
//
// Solidity: function getRequiredKeyTag() view returns(uint8)
func (_Master *MasterCallerSession) GetRequiredKeyTag() (uint8, error) {
	return _Master.Contract.GetRequiredKeyTag(&_Master.CallOpts)
}

// GetRequiredKeyTagAt is a free data retrieval call binding the contract method 0x5b1bcecc.
//
// Solidity: function getRequiredKeyTagAt(uint48 timestamp, bytes hint) view returns(uint8)
func (_Master *MasterCaller) GetRequiredKeyTagAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTagAt", timestamp, hint)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRequiredKeyTagAt is a free data retrieval call binding the contract method 0x5b1bcecc.
//
// Solidity: function getRequiredKeyTagAt(uint48 timestamp, bytes hint) view returns(uint8)
func (_Master *MasterSession) GetRequiredKeyTagAt(timestamp *big.Int, hint []byte) (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagAt(&_Master.CallOpts, timestamp, hint)
}

// GetRequiredKeyTagAt is a free data retrieval call binding the contract method 0x5b1bcecc.
//
// Solidity: function getRequiredKeyTagAt(uint48 timestamp, bytes hint) view returns(uint8)
func (_Master *MasterCallerSession) GetRequiredKeyTagAt(timestamp *big.Int, hint []byte) (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagAt(&_Master.CallOpts, timestamp, hint)
}

// GetRequiredKeyTagFromValSetHeader is a free data retrieval call binding the contract method 0xb91a434a.
//
// Solidity: function getRequiredKeyTagFromValSetHeader() view returns(uint8)
func (_Master *MasterCaller) GetRequiredKeyTagFromValSetHeader(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTagFromValSetHeader")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRequiredKeyTagFromValSetHeader is a free data retrieval call binding the contract method 0xb91a434a.
//
// Solidity: function getRequiredKeyTagFromValSetHeader() view returns(uint8)
func (_Master *MasterSession) GetRequiredKeyTagFromValSetHeader() (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagFromValSetHeader(&_Master.CallOpts)
}

// GetRequiredKeyTagFromValSetHeader is a free data retrieval call binding the contract method 0xb91a434a.
//
// Solidity: function getRequiredKeyTagFromValSetHeader() view returns(uint8)
func (_Master *MasterCallerSession) GetRequiredKeyTagFromValSetHeader() (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagFromValSetHeader(&_Master.CallOpts)
}

// GetRequiredKeyTagFromValSetHeaderAt is a free data retrieval call binding the contract method 0xe4378ed2.
//
// Solidity: function getRequiredKeyTagFromValSetHeaderAt(uint48 epoch) view returns(uint8)
func (_Master *MasterCaller) GetRequiredKeyTagFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTagFromValSetHeaderAt", epoch)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRequiredKeyTagFromValSetHeaderAt is a free data retrieval call binding the contract method 0xe4378ed2.
//
// Solidity: function getRequiredKeyTagFromValSetHeaderAt(uint48 epoch) view returns(uint8)
func (_Master *MasterSession) GetRequiredKeyTagFromValSetHeaderAt(epoch *big.Int) (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetRequiredKeyTagFromValSetHeaderAt is a free data retrieval call binding the contract method 0xe4378ed2.
//
// Solidity: function getRequiredKeyTagFromValSetHeaderAt(uint48 epoch) view returns(uint8)
func (_Master *MasterCallerSession) GetRequiredKeyTagFromValSetHeaderAt(epoch *big.Int) (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetRequiredKeyTags is a free data retrieval call binding the contract method 0xf9bfa78a.
//
// Solidity: function getRequiredKeyTags() view returns(uint8[] requiredKeyTags)
func (_Master *MasterCaller) GetRequiredKeyTags(opts *bind.CallOpts) ([]uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTags")

	if err != nil {
		return *new([]uint8), err
	}

	out0 := *abi.ConvertType(out[0], new([]uint8)).(*[]uint8)

	return out0, err

}

// GetRequiredKeyTags is a free data retrieval call binding the contract method 0xf9bfa78a.
//
// Solidity: function getRequiredKeyTags() view returns(uint8[] requiredKeyTags)
func (_Master *MasterSession) GetRequiredKeyTags() ([]uint8, error) {
	return _Master.Contract.GetRequiredKeyTags(&_Master.CallOpts)
}

// GetRequiredKeyTags is a free data retrieval call binding the contract method 0xf9bfa78a.
//
// Solidity: function getRequiredKeyTags() view returns(uint8[] requiredKeyTags)
func (_Master *MasterCallerSession) GetRequiredKeyTags() ([]uint8, error) {
	return _Master.Contract.GetRequiredKeyTags(&_Master.CallOpts)
}

// GetRequiredKeyTagsAt is a free data retrieval call binding the contract method 0x6803a63e.
//
// Solidity: function getRequiredKeyTagsAt(uint48 timestamp, bytes hint) view returns(uint8[] requiredKeyTags)
func (_Master *MasterCaller) GetRequiredKeyTagsAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) ([]uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTagsAt", timestamp, hint)

	if err != nil {
		return *new([]uint8), err
	}

	out0 := *abi.ConvertType(out[0], new([]uint8)).(*[]uint8)

	return out0, err

}

// GetRequiredKeyTagsAt is a free data retrieval call binding the contract method 0x6803a63e.
//
// Solidity: function getRequiredKeyTagsAt(uint48 timestamp, bytes hint) view returns(uint8[] requiredKeyTags)
func (_Master *MasterSession) GetRequiredKeyTagsAt(timestamp *big.Int, hint []byte) ([]uint8, error) {
	return _Master.Contract.GetRequiredKeyTagsAt(&_Master.CallOpts, timestamp, hint)
}

// GetRequiredKeyTagsAt is a free data retrieval call binding the contract method 0x6803a63e.
//
// Solidity: function getRequiredKeyTagsAt(uint48 timestamp, bytes hint) view returns(uint8[] requiredKeyTags)
func (_Master *MasterCallerSession) GetRequiredKeyTagsAt(timestamp *big.Int, hint []byte) ([]uint8, error) {
	return _Master.Contract.GetRequiredKeyTagsAt(&_Master.CallOpts, timestamp, hint)
}

// GetRole is a free data retrieval call binding the contract method 0xa846156d.
//
// Solidity: function getRole(bytes4 selector) view returns(bytes32)
func (_Master *MasterCaller) GetRole(opts *bind.CallOpts, selector [4]byte) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRole", selector)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRole is a free data retrieval call binding the contract method 0xa846156d.
//
// Solidity: function getRole(bytes4 selector) view returns(bytes32)
func (_Master *MasterSession) GetRole(selector [4]byte) ([32]byte, error) {
	return _Master.Contract.GetRole(&_Master.CallOpts, selector)
}

// GetRole is a free data retrieval call binding the contract method 0xa846156d.
//
// Solidity: function getRole(bytes4 selector) view returns(bytes32)
func (_Master *MasterCallerSession) GetRole(selector [4]byte) ([32]byte, error) {
	return _Master.Contract.GetRole(&_Master.CallOpts, selector)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Master *MasterCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Master *MasterSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Master.Contract.GetRoleAdmin(&_Master.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Master *MasterCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Master.Contract.GetRoleAdmin(&_Master.CallOpts, role)
}

// GetSigVerifier is a free data retrieval call binding the contract method 0x5b28556d.
//
// Solidity: function getSigVerifier() view returns(address)
func (_Master *MasterCaller) GetSigVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getSigVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetSigVerifier is a free data retrieval call binding the contract method 0x5b28556d.
//
// Solidity: function getSigVerifier() view returns(address)
func (_Master *MasterSession) GetSigVerifier() (common.Address, error) {
	return _Master.Contract.GetSigVerifier(&_Master.CallOpts)
}

// GetSigVerifier is a free data retrieval call binding the contract method 0x5b28556d.
//
// Solidity: function getSigVerifier() view returns(address)
func (_Master *MasterCallerSession) GetSigVerifier() (common.Address, error) {
	return _Master.Contract.GetSigVerifier(&_Master.CallOpts)
}

// GetSigVerifierAt is a free data retrieval call binding the contract method 0xa54ce263.
//
// Solidity: function getSigVerifierAt(uint48 timestamp, bytes hint) view returns(address)
func (_Master *MasterCaller) GetSigVerifierAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (common.Address, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getSigVerifierAt", timestamp, hint)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetSigVerifierAt is a free data retrieval call binding the contract method 0xa54ce263.
//
// Solidity: function getSigVerifierAt(uint48 timestamp, bytes hint) view returns(address)
func (_Master *MasterSession) GetSigVerifierAt(timestamp *big.Int, hint []byte) (common.Address, error) {
	return _Master.Contract.GetSigVerifierAt(&_Master.CallOpts, timestamp, hint)
}

// GetSigVerifierAt is a free data retrieval call binding the contract method 0xa54ce263.
//
// Solidity: function getSigVerifierAt(uint48 timestamp, bytes hint) view returns(address)
func (_Master *MasterCallerSession) GetSigVerifierAt(timestamp *big.Int, hint []byte) (common.Address, error) {
	return _Master.Contract.GetSigVerifierAt(&_Master.CallOpts, timestamp, hint)
}

// GetValSetHeader is a free data retrieval call binding the contract method 0xadc91fc8.
//
// Solidity: function getValSetHeader() view returns((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) header)
func (_Master *MasterCaller) GetValSetHeader(opts *bind.CallOpts) (ISettlementValSetHeader, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValSetHeader")

	if err != nil {
		return *new(ISettlementValSetHeader), err
	}

	out0 := *abi.ConvertType(out[0], new(ISettlementValSetHeader)).(*ISettlementValSetHeader)

	return out0, err

}

// GetValSetHeader is a free data retrieval call binding the contract method 0xadc91fc8.
//
// Solidity: function getValSetHeader() view returns((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) header)
func (_Master *MasterSession) GetValSetHeader() (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeader(&_Master.CallOpts)
}

// GetValSetHeader is a free data retrieval call binding the contract method 0xadc91fc8.
//
// Solidity: function getValSetHeader() view returns((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) header)
func (_Master *MasterCallerSession) GetValSetHeader() (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeader(&_Master.CallOpts)
}

// GetValSetHeaderAt is a free data retrieval call binding the contract method 0x4addaee7.
//
// Solidity: function getValSetHeaderAt(uint48 epoch) view returns((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32))
func (_Master *MasterCaller) GetValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (ISettlementValSetHeader, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValSetHeaderAt", epoch)

	if err != nil {
		return *new(ISettlementValSetHeader), err
	}

	out0 := *abi.ConvertType(out[0], new(ISettlementValSetHeader)).(*ISettlementValSetHeader)

	return out0, err

}

// GetValSetHeaderAt is a free data retrieval call binding the contract method 0x4addaee7.
//
// Solidity: function getValSetHeaderAt(uint48 epoch) view returns((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32))
func (_Master *MasterSession) GetValSetHeaderAt(epoch *big.Int) (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetValSetHeaderAt is a free data retrieval call binding the contract method 0x4addaee7.
//
// Solidity: function getValSetHeaderAt(uint48 epoch) view returns((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32))
func (_Master *MasterCallerSession) GetValSetHeaderAt(epoch *big.Int) (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetValSetHeaderHash is a free data retrieval call binding the contract method 0x32624bf3.
//
// Solidity: function getValSetHeaderHash() view returns(bytes32)
func (_Master *MasterCaller) GetValSetHeaderHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValSetHeaderHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetValSetHeaderHash is a free data retrieval call binding the contract method 0x32624bf3.
//
// Solidity: function getValSetHeaderHash() view returns(bytes32)
func (_Master *MasterSession) GetValSetHeaderHash() ([32]byte, error) {
	return _Master.Contract.GetValSetHeaderHash(&_Master.CallOpts)
}

// GetValSetHeaderHash is a free data retrieval call binding the contract method 0x32624bf3.
//
// Solidity: function getValSetHeaderHash() view returns(bytes32)
func (_Master *MasterCallerSession) GetValSetHeaderHash() ([32]byte, error) {
	return _Master.Contract.GetValSetHeaderHash(&_Master.CallOpts)
}

// GetValSetHeaderHashAt is a free data retrieval call binding the contract method 0xf35d12a3.
//
// Solidity: function getValSetHeaderHashAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterCaller) GetValSetHeaderHashAt(opts *bind.CallOpts, epoch *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValSetHeaderHashAt", epoch)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetValSetHeaderHashAt is a free data retrieval call binding the contract method 0xf35d12a3.
//
// Solidity: function getValSetHeaderHashAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterSession) GetValSetHeaderHashAt(epoch *big.Int) ([32]byte, error) {
	return _Master.Contract.GetValSetHeaderHashAt(&_Master.CallOpts, epoch)
}

// GetValSetHeaderHashAt is a free data retrieval call binding the contract method 0xf35d12a3.
//
// Solidity: function getValSetHeaderHashAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterCallerSession) GetValSetHeaderHashAt(epoch *big.Int) ([32]byte, error) {
	return _Master.Contract.GetValSetHeaderHashAt(&_Master.CallOpts, epoch)
}

// GetValidatorsSszMRootFromValSetHeader is a free data retrieval call binding the contract method 0x0167166e.
//
// Solidity: function getValidatorsSszMRootFromValSetHeader() view returns(bytes32)
func (_Master *MasterCaller) GetValidatorsSszMRootFromValSetHeader(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValidatorsSszMRootFromValSetHeader")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetValidatorsSszMRootFromValSetHeader is a free data retrieval call binding the contract method 0x0167166e.
//
// Solidity: function getValidatorsSszMRootFromValSetHeader() view returns(bytes32)
func (_Master *MasterSession) GetValidatorsSszMRootFromValSetHeader() ([32]byte, error) {
	return _Master.Contract.GetValidatorsSszMRootFromValSetHeader(&_Master.CallOpts)
}

// GetValidatorsSszMRootFromValSetHeader is a free data retrieval call binding the contract method 0x0167166e.
//
// Solidity: function getValidatorsSszMRootFromValSetHeader() view returns(bytes32)
func (_Master *MasterCallerSession) GetValidatorsSszMRootFromValSetHeader() ([32]byte, error) {
	return _Master.Contract.GetValidatorsSszMRootFromValSetHeader(&_Master.CallOpts)
}

// GetValidatorsSszMRootFromValSetHeaderAt is a free data retrieval call binding the contract method 0x230ae408.
//
// Solidity: function getValidatorsSszMRootFromValSetHeaderAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterCaller) GetValidatorsSszMRootFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValidatorsSszMRootFromValSetHeaderAt", epoch)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetValidatorsSszMRootFromValSetHeaderAt is a free data retrieval call binding the contract method 0x230ae408.
//
// Solidity: function getValidatorsSszMRootFromValSetHeaderAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterSession) GetValidatorsSszMRootFromValSetHeaderAt(epoch *big.Int) ([32]byte, error) {
	return _Master.Contract.GetValidatorsSszMRootFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetValidatorsSszMRootFromValSetHeaderAt is a free data retrieval call binding the contract method 0x230ae408.
//
// Solidity: function getValidatorsSszMRootFromValSetHeaderAt(uint48 epoch) view returns(bytes32)
func (_Master *MasterCallerSession) GetValidatorsSszMRootFromValSetHeaderAt(epoch *big.Int) ([32]byte, error) {
	return _Master.Contract.GetValidatorsSszMRootFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetVerificationType is a free data retrieval call binding the contract method 0x24acc119.
//
// Solidity: function getVerificationType() view returns(uint32)
func (_Master *MasterCaller) GetVerificationType(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getVerificationType")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetVerificationType is a free data retrieval call binding the contract method 0x24acc119.
//
// Solidity: function getVerificationType() view returns(uint32)
func (_Master *MasterSession) GetVerificationType() (uint32, error) {
	return _Master.Contract.GetVerificationType(&_Master.CallOpts)
}

// GetVerificationType is a free data retrieval call binding the contract method 0x24acc119.
//
// Solidity: function getVerificationType() view returns(uint32)
func (_Master *MasterCallerSession) GetVerificationType() (uint32, error) {
	return _Master.Contract.GetVerificationType(&_Master.CallOpts)
}

// GetVerificationTypeAt is a free data retrieval call binding the contract method 0x01e5050a.
//
// Solidity: function getVerificationTypeAt(uint48 timestamp, bytes hint) view returns(uint32)
func (_Master *MasterCaller) GetVerificationTypeAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (uint32, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getVerificationTypeAt", timestamp, hint)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetVerificationTypeAt is a free data retrieval call binding the contract method 0x01e5050a.
//
// Solidity: function getVerificationTypeAt(uint48 timestamp, bytes hint) view returns(uint32)
func (_Master *MasterSession) GetVerificationTypeAt(timestamp *big.Int, hint []byte) (uint32, error) {
	return _Master.Contract.GetVerificationTypeAt(&_Master.CallOpts, timestamp, hint)
}

// GetVerificationTypeAt is a free data retrieval call binding the contract method 0x01e5050a.
//
// Solidity: function getVerificationTypeAt(uint48 timestamp, bytes hint) view returns(uint32)
func (_Master *MasterCallerSession) GetVerificationTypeAt(timestamp *big.Int, hint []byte) (uint32, error) {
	return _Master.Contract.GetVerificationTypeAt(&_Master.CallOpts, timestamp, hint)
}

// GetVersionFromValSetHeader is a free data retrieval call binding the contract method 0xd2df9fb6.
//
// Solidity: function getVersionFromValSetHeader() view returns(uint8)
func (_Master *MasterCaller) GetVersionFromValSetHeader(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getVersionFromValSetHeader")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetVersionFromValSetHeader is a free data retrieval call binding the contract method 0xd2df9fb6.
//
// Solidity: function getVersionFromValSetHeader() view returns(uint8)
func (_Master *MasterSession) GetVersionFromValSetHeader() (uint8, error) {
	return _Master.Contract.GetVersionFromValSetHeader(&_Master.CallOpts)
}

// GetVersionFromValSetHeader is a free data retrieval call binding the contract method 0xd2df9fb6.
//
// Solidity: function getVersionFromValSetHeader() view returns(uint8)
func (_Master *MasterCallerSession) GetVersionFromValSetHeader() (uint8, error) {
	return _Master.Contract.GetVersionFromValSetHeader(&_Master.CallOpts)
}

// GetVersionFromValSetHeaderAt is a free data retrieval call binding the contract method 0x548202ad.
//
// Solidity: function getVersionFromValSetHeaderAt(uint48 epoch) view returns(uint8)
func (_Master *MasterCaller) GetVersionFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getVersionFromValSetHeaderAt", epoch)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetVersionFromValSetHeaderAt is a free data retrieval call binding the contract method 0x548202ad.
//
// Solidity: function getVersionFromValSetHeaderAt(uint48 epoch) view returns(uint8)
func (_Master *MasterSession) GetVersionFromValSetHeaderAt(epoch *big.Int) (uint8, error) {
	return _Master.Contract.GetVersionFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetVersionFromValSetHeaderAt is a free data retrieval call binding the contract method 0x548202ad.
//
// Solidity: function getVersionFromValSetHeaderAt(uint48 epoch) view returns(uint8)
func (_Master *MasterCallerSession) GetVersionFromValSetHeaderAt(epoch *big.Int) (uint8, error) {
	return _Master.Contract.GetVersionFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetVotingPowerProviders is a free data retrieval call binding the contract method 0x3e39b8db.
//
// Solidity: function getVotingPowerProviders() view returns((address,uint64)[])
func (_Master *MasterCaller) GetVotingPowerProviders(opts *bind.CallOpts) ([]IConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getVotingPowerProviders")

	if err != nil {
		return *new([]IConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IConfigProviderCrossChainAddress)).(*[]IConfigProviderCrossChainAddress)

	return out0, err

}

// GetVotingPowerProviders is a free data retrieval call binding the contract method 0x3e39b8db.
//
// Solidity: function getVotingPowerProviders() view returns((address,uint64)[])
func (_Master *MasterSession) GetVotingPowerProviders() ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetVotingPowerProviders(&_Master.CallOpts)
}

// GetVotingPowerProviders is a free data retrieval call binding the contract method 0x3e39b8db.
//
// Solidity: function getVotingPowerProviders() view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetVotingPowerProviders() ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetVotingPowerProviders(&_Master.CallOpts)
}

// GetVotingPowerProvidersAt is a free data retrieval call binding the contract method 0x5bf4eef7.
//
// Solidity: function getVotingPowerProvidersAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCaller) GetVotingPowerProvidersAt(opts *bind.CallOpts, timestamp *big.Int, hints [][]byte) ([]IConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getVotingPowerProvidersAt", timestamp, hints)

	if err != nil {
		return *new([]IConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IConfigProviderCrossChainAddress)).(*[]IConfigProviderCrossChainAddress)

	return out0, err

}

// GetVotingPowerProvidersAt is a free data retrieval call binding the contract method 0x5bf4eef7.
//
// Solidity: function getVotingPowerProvidersAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterSession) GetVotingPowerProvidersAt(timestamp *big.Int, hints [][]byte) ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetVotingPowerProvidersAt(&_Master.CallOpts, timestamp, hints)
}

// GetVotingPowerProvidersAt is a free data retrieval call binding the contract method 0x5bf4eef7.
//
// Solidity: function getVotingPowerProvidersAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetVotingPowerProvidersAt(timestamp *big.Int, hints [][]byte) ([]IConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetVotingPowerProvidersAt(&_Master.CallOpts, timestamp, hints)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Master *MasterCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Master *MasterSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Master.Contract.HasRole(&_Master.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Master *MasterCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Master.Contract.HasRole(&_Master.CallOpts, role, account)
}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_Master *MasterCaller) HashTypedDataV4(opts *bind.CallOpts, structHash [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "hashTypedDataV4", structHash)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_Master *MasterSession) HashTypedDataV4(structHash [32]byte) ([32]byte, error) {
	return _Master.Contract.HashTypedDataV4(&_Master.CallOpts, structHash)
}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_Master *MasterCallerSession) HashTypedDataV4(structHash [32]byte) ([32]byte, error) {
	return _Master.Contract.HashTypedDataV4(&_Master.CallOpts, structHash)
}

// HashTypedDataV4CrossChain is a free data retrieval call binding the contract method 0x518dcf3b.
//
// Solidity: function hashTypedDataV4CrossChain(bytes32 structHash) view returns(bytes32)
func (_Master *MasterCaller) HashTypedDataV4CrossChain(opts *bind.CallOpts, structHash [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "hashTypedDataV4CrossChain", structHash)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashTypedDataV4CrossChain is a free data retrieval call binding the contract method 0x518dcf3b.
//
// Solidity: function hashTypedDataV4CrossChain(bytes32 structHash) view returns(bytes32)
func (_Master *MasterSession) HashTypedDataV4CrossChain(structHash [32]byte) ([32]byte, error) {
	return _Master.Contract.HashTypedDataV4CrossChain(&_Master.CallOpts, structHash)
}

// HashTypedDataV4CrossChain is a free data retrieval call binding the contract method 0x518dcf3b.
//
// Solidity: function hashTypedDataV4CrossChain(bytes32 structHash) view returns(bytes32)
func (_Master *MasterCallerSession) HashTypedDataV4CrossChain(structHash [32]byte) ([32]byte, error) {
	return _Master.Contract.HashTypedDataV4CrossChain(&_Master.CallOpts, structHash)
}

// IsReplicaRegistered is a free data retrieval call binding the contract method 0x77958baa.
//
// Solidity: function isReplicaRegistered((address,uint64) replica) view returns(bool)
func (_Master *MasterCaller) IsReplicaRegistered(opts *bind.CallOpts, replica IConfigProviderCrossChainAddress) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isReplicaRegistered", replica)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReplicaRegistered is a free data retrieval call binding the contract method 0x77958baa.
//
// Solidity: function isReplicaRegistered((address,uint64) replica) view returns(bool)
func (_Master *MasterSession) IsReplicaRegistered(replica IConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsReplicaRegistered(&_Master.CallOpts, replica)
}

// IsReplicaRegistered is a free data retrieval call binding the contract method 0x77958baa.
//
// Solidity: function isReplicaRegistered((address,uint64) replica) view returns(bool)
func (_Master *MasterCallerSession) IsReplicaRegistered(replica IConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsReplicaRegistered(&_Master.CallOpts, replica)
}

// IsReplicaRegisteredAt is a free data retrieval call binding the contract method 0x5456ee40.
//
// Solidity: function isReplicaRegisteredAt((address,uint64) replica, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCaller) IsReplicaRegisteredAt(opts *bind.CallOpts, replica IConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isReplicaRegisteredAt", replica, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReplicaRegisteredAt is a free data retrieval call binding the contract method 0x5456ee40.
//
// Solidity: function isReplicaRegisteredAt((address,uint64) replica, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterSession) IsReplicaRegisteredAt(replica IConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsReplicaRegisteredAt(&_Master.CallOpts, replica, timestamp, hint)
}

// IsReplicaRegisteredAt is a free data retrieval call binding the contract method 0x5456ee40.
//
// Solidity: function isReplicaRegisteredAt((address,uint64) replica, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCallerSession) IsReplicaRegisteredAt(replica IConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsReplicaRegisteredAt(&_Master.CallOpts, replica, timestamp, hint)
}

// IsValSetHeaderCommitted is a free data retrieval call binding the contract method 0x39e28545.
//
// Solidity: function isValSetHeaderCommitted() view returns(bool)
func (_Master *MasterCaller) IsValSetHeaderCommitted(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isValSetHeaderCommitted")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValSetHeaderCommitted is a free data retrieval call binding the contract method 0x39e28545.
//
// Solidity: function isValSetHeaderCommitted() view returns(bool)
func (_Master *MasterSession) IsValSetHeaderCommitted() (bool, error) {
	return _Master.Contract.IsValSetHeaderCommitted(&_Master.CallOpts)
}

// IsValSetHeaderCommitted is a free data retrieval call binding the contract method 0x39e28545.
//
// Solidity: function isValSetHeaderCommitted() view returns(bool)
func (_Master *MasterCallerSession) IsValSetHeaderCommitted() (bool, error) {
	return _Master.Contract.IsValSetHeaderCommitted(&_Master.CallOpts)
}

// IsValSetHeaderCommittedAt is a free data retrieval call binding the contract method 0x5fa4bbd2.
//
// Solidity: function isValSetHeaderCommittedAt(uint48 epoch) view returns(bool)
func (_Master *MasterCaller) IsValSetHeaderCommittedAt(opts *bind.CallOpts, epoch *big.Int) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isValSetHeaderCommittedAt", epoch)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValSetHeaderCommittedAt is a free data retrieval call binding the contract method 0x5fa4bbd2.
//
// Solidity: function isValSetHeaderCommittedAt(uint48 epoch) view returns(bool)
func (_Master *MasterSession) IsValSetHeaderCommittedAt(epoch *big.Int) (bool, error) {
	return _Master.Contract.IsValSetHeaderCommittedAt(&_Master.CallOpts, epoch)
}

// IsValSetHeaderCommittedAt is a free data retrieval call binding the contract method 0x5fa4bbd2.
//
// Solidity: function isValSetHeaderCommittedAt(uint48 epoch) view returns(bool)
func (_Master *MasterCallerSession) IsValSetHeaderCommittedAt(epoch *big.Int) (bool, error) {
	return _Master.Contract.IsValSetHeaderCommittedAt(&_Master.CallOpts, epoch)
}

// IsVotingPowerProviderRegistered is a free data retrieval call binding the contract method 0xd89751f1.
//
// Solidity: function isVotingPowerProviderRegistered((address,uint64) votingPowerProvider) view returns(bool)
func (_Master *MasterCaller) IsVotingPowerProviderRegistered(opts *bind.CallOpts, votingPowerProvider IConfigProviderCrossChainAddress) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isVotingPowerProviderRegistered", votingPowerProvider)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsVotingPowerProviderRegistered is a free data retrieval call binding the contract method 0xd89751f1.
//
// Solidity: function isVotingPowerProviderRegistered((address,uint64) votingPowerProvider) view returns(bool)
func (_Master *MasterSession) IsVotingPowerProviderRegistered(votingPowerProvider IConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderRegistered(&_Master.CallOpts, votingPowerProvider)
}

// IsVotingPowerProviderRegistered is a free data retrieval call binding the contract method 0xd89751f1.
//
// Solidity: function isVotingPowerProviderRegistered((address,uint64) votingPowerProvider) view returns(bool)
func (_Master *MasterCallerSession) IsVotingPowerProviderRegistered(votingPowerProvider IConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderRegistered(&_Master.CallOpts, votingPowerProvider)
}

// IsVotingPowerProviderRegisteredAt is a free data retrieval call binding the contract method 0x9156e8b4.
//
// Solidity: function isVotingPowerProviderRegisteredAt((address,uint64) votingPowerProvider, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCaller) IsVotingPowerProviderRegisteredAt(opts *bind.CallOpts, votingPowerProvider IConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isVotingPowerProviderRegisteredAt", votingPowerProvider, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsVotingPowerProviderRegisteredAt is a free data retrieval call binding the contract method 0x9156e8b4.
//
// Solidity: function isVotingPowerProviderRegisteredAt((address,uint64) votingPowerProvider, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterSession) IsVotingPowerProviderRegisteredAt(votingPowerProvider IConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderRegisteredAt(&_Master.CallOpts, votingPowerProvider, timestamp, hint)
}

// IsVotingPowerProviderRegisteredAt is a free data retrieval call binding the contract method 0x9156e8b4.
//
// Solidity: function isVotingPowerProviderRegisteredAt((address,uint64) votingPowerProvider, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCallerSession) IsVotingPowerProviderRegisteredAt(votingPowerProvider IConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderRegisteredAt(&_Master.CallOpts, votingPowerProvider, timestamp, hint)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Master *MasterCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Master *MasterSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Master.Contract.SupportsInterface(&_Master.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Master *MasterCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Master.Contract.SupportsInterface(&_Master.CallOpts, interfaceId)
}

// VerifyQuorumSig is a free data retrieval call binding the contract method 0x4f785398.
//
// Solidity: function verifyQuorumSig(uint48 epoch, bytes message, uint8 keyTag, uint256 quorumThreshold, bytes proof, bytes hint) view returns(bool)
func (_Master *MasterCaller) VerifyQuorumSig(opts *bind.CallOpts, epoch *big.Int, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte, hint []byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "verifyQuorumSig", epoch, message, keyTag, quorumThreshold, proof, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyQuorumSig is a free data retrieval call binding the contract method 0x4f785398.
//
// Solidity: function verifyQuorumSig(uint48 epoch, bytes message, uint8 keyTag, uint256 quorumThreshold, bytes proof, bytes hint) view returns(bool)
func (_Master *MasterSession) VerifyQuorumSig(epoch *big.Int, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte, hint []byte) (bool, error) {
	return _Master.Contract.VerifyQuorumSig(&_Master.CallOpts, epoch, message, keyTag, quorumThreshold, proof, hint)
}

// VerifyQuorumSig is a free data retrieval call binding the contract method 0x4f785398.
//
// Solidity: function verifyQuorumSig(uint48 epoch, bytes message, uint8 keyTag, uint256 quorumThreshold, bytes proof, bytes hint) view returns(bool)
func (_Master *MasterCallerSession) VerifyQuorumSig(epoch *big.Int, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte, hint []byte) (bool, error) {
	return _Master.Contract.VerifyQuorumSig(&_Master.CallOpts, epoch, message, keyTag, quorumThreshold, proof, hint)
}

// AddReplica is a paid mutator transaction binding the contract method 0x5c0bc730.
//
// Solidity: function addReplica((address,uint64) replica) returns()
func (_Master *MasterTransactor) AddReplica(opts *bind.TransactOpts, replica IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "addReplica", replica)
}

// AddReplica is a paid mutator transaction binding the contract method 0x5c0bc730.
//
// Solidity: function addReplica((address,uint64) replica) returns()
func (_Master *MasterSession) AddReplica(replica IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddReplica(&_Master.TransactOpts, replica)
}

// AddReplica is a paid mutator transaction binding the contract method 0x5c0bc730.
//
// Solidity: function addReplica((address,uint64) replica) returns()
func (_Master *MasterTransactorSession) AddReplica(replica IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddReplica(&_Master.TransactOpts, replica)
}

// AddVotingPowerProvider is a paid mutator transaction binding the contract method 0x88514824.
//
// Solidity: function addVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactor) AddVotingPowerProvider(opts *bind.TransactOpts, votingPowerProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "addVotingPowerProvider", votingPowerProvider)
}

// AddVotingPowerProvider is a paid mutator transaction binding the contract method 0x88514824.
//
// Solidity: function addVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterSession) AddVotingPowerProvider(votingPowerProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// AddVotingPowerProvider is a paid mutator transaction binding the contract method 0x88514824.
//
// Solidity: function addVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactorSession) AddVotingPowerProvider(votingPowerProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// CommitValSetHeader is a paid mutator transaction binding the contract method 0x5a810418.
//
// Solidity: function commitValSetHeader((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) header, (bytes32,bytes32)[] extraData, bytes proof, bytes hint) returns()
func (_Master *MasterTransactor) CommitValSetHeader(opts *bind.TransactOpts, header ISettlementValSetHeader, extraData []ISettlementExtraData, proof []byte, hint []byte) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "commitValSetHeader", header, extraData, proof, hint)
}

// CommitValSetHeader is a paid mutator transaction binding the contract method 0x5a810418.
//
// Solidity: function commitValSetHeader((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) header, (bytes32,bytes32)[] extraData, bytes proof, bytes hint) returns()
func (_Master *MasterSession) CommitValSetHeader(header ISettlementValSetHeader, extraData []ISettlementExtraData, proof []byte, hint []byte) (*types.Transaction, error) {
	return _Master.Contract.CommitValSetHeader(&_Master.TransactOpts, header, extraData, proof, hint)
}

// CommitValSetHeader is a paid mutator transaction binding the contract method 0x5a810418.
//
// Solidity: function commitValSetHeader((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) header, (bytes32,bytes32)[] extraData, bytes proof, bytes hint) returns()
func (_Master *MasterTransactorSession) CommitValSetHeader(header ISettlementValSetHeader, extraData []ISettlementExtraData, proof []byte, hint []byte) (*types.Transaction, error) {
	return _Master.Contract.CommitValSetHeader(&_Master.TransactOpts, header, extraData, proof, hint)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Master *MasterTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Master *MasterSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Master.Contract.GrantRole(&_Master.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Master *MasterTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Master.Contract.GrantRole(&_Master.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xa267b54d.
//
// Solidity: function initialize(((address,uint96),(uint48,uint48),(string,string),uint48,uint48,uint8,address) settlementInitParams, ((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]) configProviderInitParams, address defaultAdmin) returns()
func (_Master *MasterTransactor) Initialize(opts *bind.TransactOpts, settlementInitParams ISettlementSettlementInitParams, configProviderInitParams IConfigProviderConfigProviderInitParams, defaultAdmin common.Address) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "initialize", settlementInitParams, configProviderInitParams, defaultAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0xa267b54d.
//
// Solidity: function initialize(((address,uint96),(uint48,uint48),(string,string),uint48,uint48,uint8,address) settlementInitParams, ((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]) configProviderInitParams, address defaultAdmin) returns()
func (_Master *MasterSession) Initialize(settlementInitParams ISettlementSettlementInitParams, configProviderInitParams IConfigProviderConfigProviderInitParams, defaultAdmin common.Address) (*types.Transaction, error) {
	return _Master.Contract.Initialize(&_Master.TransactOpts, settlementInitParams, configProviderInitParams, defaultAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0xa267b54d.
//
// Solidity: function initialize(((address,uint96),(uint48,uint48),(string,string),uint48,uint48,uint8,address) settlementInitParams, ((address,uint64)[],(address,uint64),(address,uint64)[],uint32,uint256,uint256,uint208,uint8[]) configProviderInitParams, address defaultAdmin) returns()
func (_Master *MasterTransactorSession) Initialize(settlementInitParams ISettlementSettlementInitParams, configProviderInitParams IConfigProviderConfigProviderInitParams, defaultAdmin common.Address) (*types.Transaction, error) {
	return _Master.Contract.Initialize(&_Master.TransactOpts, settlementInitParams, configProviderInitParams, defaultAdmin)
}

// Multicall is a paid mutator transaction binding the contract method 0xac9650d8.
//
// Solidity: function multicall(bytes[] data) returns(bytes[] results)
func (_Master *MasterTransactor) Multicall(opts *bind.TransactOpts, data [][]byte) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "multicall", data)
}

// Multicall is a paid mutator transaction binding the contract method 0xac9650d8.
//
// Solidity: function multicall(bytes[] data) returns(bytes[] results)
func (_Master *MasterSession) Multicall(data [][]byte) (*types.Transaction, error) {
	return _Master.Contract.Multicall(&_Master.TransactOpts, data)
}

// Multicall is a paid mutator transaction binding the contract method 0xac9650d8.
//
// Solidity: function multicall(bytes[] data) returns(bytes[] results)
func (_Master *MasterTransactorSession) Multicall(data [][]byte) (*types.Transaction, error) {
	return _Master.Contract.Multicall(&_Master.TransactOpts, data)
}

// RemoveReplica is a paid mutator transaction binding the contract method 0x65f764f0.
//
// Solidity: function removeReplica((address,uint64) replica) returns()
func (_Master *MasterTransactor) RemoveReplica(opts *bind.TransactOpts, replica IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "removeReplica", replica)
}

// RemoveReplica is a paid mutator transaction binding the contract method 0x65f764f0.
//
// Solidity: function removeReplica((address,uint64) replica) returns()
func (_Master *MasterSession) RemoveReplica(replica IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveReplica(&_Master.TransactOpts, replica)
}

// RemoveReplica is a paid mutator transaction binding the contract method 0x65f764f0.
//
// Solidity: function removeReplica((address,uint64) replica) returns()
func (_Master *MasterTransactorSession) RemoveReplica(replica IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveReplica(&_Master.TransactOpts, replica)
}

// RemoveVotingPowerProvider is a paid mutator transaction binding the contract method 0xb01139a1.
//
// Solidity: function removeVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactor) RemoveVotingPowerProvider(opts *bind.TransactOpts, votingPowerProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "removeVotingPowerProvider", votingPowerProvider)
}

// RemoveVotingPowerProvider is a paid mutator transaction binding the contract method 0xb01139a1.
//
// Solidity: function removeVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterSession) RemoveVotingPowerProvider(votingPowerProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// RemoveVotingPowerProvider is a paid mutator transaction binding the contract method 0xb01139a1.
//
// Solidity: function removeVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactorSession) RemoveVotingPowerProvider(votingPowerProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Master *MasterTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Master *MasterSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Master.Contract.RenounceRole(&_Master.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Master *MasterTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Master.Contract.RenounceRole(&_Master.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Master *MasterTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Master *MasterSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Master.Contract.RevokeRole(&_Master.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Master *MasterTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Master.Contract.RevokeRole(&_Master.TransactOpts, role, account)
}

// SetCommitDuration is a paid mutator transaction binding the contract method 0xe5ac1fcd.
//
// Solidity: function setCommitDuration(uint48 commitDuration) returns()
func (_Master *MasterTransactor) SetCommitDuration(opts *bind.TransactOpts, commitDuration *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setCommitDuration", commitDuration)
}

// SetCommitDuration is a paid mutator transaction binding the contract method 0xe5ac1fcd.
//
// Solidity: function setCommitDuration(uint48 commitDuration) returns()
func (_Master *MasterSession) SetCommitDuration(commitDuration *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetCommitDuration(&_Master.TransactOpts, commitDuration)
}

// SetCommitDuration is a paid mutator transaction binding the contract method 0xe5ac1fcd.
//
// Solidity: function setCommitDuration(uint48 commitDuration) returns()
func (_Master *MasterTransactorSession) SetCommitDuration(commitDuration *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetCommitDuration(&_Master.TransactOpts, commitDuration)
}

// SetEpochDuration is a paid mutator transaction binding the contract method 0x2f53d5ff.
//
// Solidity: function setEpochDuration(uint48 epochDuration) returns()
func (_Master *MasterTransactor) SetEpochDuration(opts *bind.TransactOpts, epochDuration *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setEpochDuration", epochDuration)
}

// SetEpochDuration is a paid mutator transaction binding the contract method 0x2f53d5ff.
//
// Solidity: function setEpochDuration(uint48 epochDuration) returns()
func (_Master *MasterSession) SetEpochDuration(epochDuration *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetEpochDuration(&_Master.TransactOpts, epochDuration)
}

// SetEpochDuration is a paid mutator transaction binding the contract method 0x2f53d5ff.
//
// Solidity: function setEpochDuration(uint48 epochDuration) returns()
func (_Master *MasterTransactorSession) SetEpochDuration(epochDuration *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetEpochDuration(&_Master.TransactOpts, epochDuration)
}

// SetGenesis is a paid mutator transaction binding the contract method 0xd67b84bb.
//
// Solidity: function setGenesis((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData) returns()
func (_Master *MasterTransactor) SetGenesis(opts *bind.TransactOpts, valSetHeader ISettlementValSetHeader, extraData []ISettlementExtraData) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setGenesis", valSetHeader, extraData)
}

// SetGenesis is a paid mutator transaction binding the contract method 0xd67b84bb.
//
// Solidity: function setGenesis((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData) returns()
func (_Master *MasterSession) SetGenesis(valSetHeader ISettlementValSetHeader, extraData []ISettlementExtraData) (*types.Transaction, error) {
	return _Master.Contract.SetGenesis(&_Master.TransactOpts, valSetHeader, extraData)
}

// SetGenesis is a paid mutator transaction binding the contract method 0xd67b84bb.
//
// Solidity: function setGenesis((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData) returns()
func (_Master *MasterTransactorSession) SetGenesis(valSetHeader ISettlementValSetHeader, extraData []ISettlementExtraData) (*types.Transaction, error) {
	return _Master.Contract.SetGenesis(&_Master.TransactOpts, valSetHeader, extraData)
}

// SetKeysProvider is a paid mutator transaction binding the contract method 0x48fb6e2b.
//
// Solidity: function setKeysProvider((address,uint64) keysProvider) returns()
func (_Master *MasterTransactor) SetKeysProvider(opts *bind.TransactOpts, keysProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setKeysProvider", keysProvider)
}

// SetKeysProvider is a paid mutator transaction binding the contract method 0x48fb6e2b.
//
// Solidity: function setKeysProvider((address,uint64) keysProvider) returns()
func (_Master *MasterSession) SetKeysProvider(keysProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.SetKeysProvider(&_Master.TransactOpts, keysProvider)
}

// SetKeysProvider is a paid mutator transaction binding the contract method 0x48fb6e2b.
//
// Solidity: function setKeysProvider((address,uint64) keysProvider) returns()
func (_Master *MasterTransactorSession) SetKeysProvider(keysProvider IConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.SetKeysProvider(&_Master.TransactOpts, keysProvider)
}

// SetMaxValidatorsCount is a paid mutator transaction binding the contract method 0xd2384cd3.
//
// Solidity: function setMaxValidatorsCount(uint208 maxValidatorsCount) returns()
func (_Master *MasterTransactor) SetMaxValidatorsCount(opts *bind.TransactOpts, maxValidatorsCount *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setMaxValidatorsCount", maxValidatorsCount)
}

// SetMaxValidatorsCount is a paid mutator transaction binding the contract method 0xd2384cd3.
//
// Solidity: function setMaxValidatorsCount(uint208 maxValidatorsCount) returns()
func (_Master *MasterSession) SetMaxValidatorsCount(maxValidatorsCount *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMaxValidatorsCount(&_Master.TransactOpts, maxValidatorsCount)
}

// SetMaxValidatorsCount is a paid mutator transaction binding the contract method 0xd2384cd3.
//
// Solidity: function setMaxValidatorsCount(uint208 maxValidatorsCount) returns()
func (_Master *MasterTransactorSession) SetMaxValidatorsCount(maxValidatorsCount *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMaxValidatorsCount(&_Master.TransactOpts, maxValidatorsCount)
}

// SetMaxVotingPower is a paid mutator transaction binding the contract method 0xf6af258c.
//
// Solidity: function setMaxVotingPower(uint256 maxVotingPower) returns()
func (_Master *MasterTransactor) SetMaxVotingPower(opts *bind.TransactOpts, maxVotingPower *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setMaxVotingPower", maxVotingPower)
}

// SetMaxVotingPower is a paid mutator transaction binding the contract method 0xf6af258c.
//
// Solidity: function setMaxVotingPower(uint256 maxVotingPower) returns()
func (_Master *MasterSession) SetMaxVotingPower(maxVotingPower *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMaxVotingPower(&_Master.TransactOpts, maxVotingPower)
}

// SetMaxVotingPower is a paid mutator transaction binding the contract method 0xf6af258c.
//
// Solidity: function setMaxVotingPower(uint256 maxVotingPower) returns()
func (_Master *MasterTransactorSession) SetMaxVotingPower(maxVotingPower *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMaxVotingPower(&_Master.TransactOpts, maxVotingPower)
}

// SetMinInclusionVotingPower is a paid mutator transaction binding the contract method 0xfaae42d7.
//
// Solidity: function setMinInclusionVotingPower(uint256 minInclusionVotingPower) returns()
func (_Master *MasterTransactor) SetMinInclusionVotingPower(opts *bind.TransactOpts, minInclusionVotingPower *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setMinInclusionVotingPower", minInclusionVotingPower)
}

// SetMinInclusionVotingPower is a paid mutator transaction binding the contract method 0xfaae42d7.
//
// Solidity: function setMinInclusionVotingPower(uint256 minInclusionVotingPower) returns()
func (_Master *MasterSession) SetMinInclusionVotingPower(minInclusionVotingPower *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMinInclusionVotingPower(&_Master.TransactOpts, minInclusionVotingPower)
}

// SetMinInclusionVotingPower is a paid mutator transaction binding the contract method 0xfaae42d7.
//
// Solidity: function setMinInclusionVotingPower(uint256 minInclusionVotingPower) returns()
func (_Master *MasterTransactorSession) SetMinInclusionVotingPower(minInclusionVotingPower *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMinInclusionVotingPower(&_Master.TransactOpts, minInclusionVotingPower)
}

// SetProlongDuration is a paid mutator transaction binding the contract method 0xbf667981.
//
// Solidity: function setProlongDuration(uint48 prolongDuration) returns()
func (_Master *MasterTransactor) SetProlongDuration(opts *bind.TransactOpts, prolongDuration *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setProlongDuration", prolongDuration)
}

// SetProlongDuration is a paid mutator transaction binding the contract method 0xbf667981.
//
// Solidity: function setProlongDuration(uint48 prolongDuration) returns()
func (_Master *MasterSession) SetProlongDuration(prolongDuration *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetProlongDuration(&_Master.TransactOpts, prolongDuration)
}

// SetProlongDuration is a paid mutator transaction binding the contract method 0xbf667981.
//
// Solidity: function setProlongDuration(uint48 prolongDuration) returns()
func (_Master *MasterTransactorSession) SetProlongDuration(prolongDuration *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetProlongDuration(&_Master.TransactOpts, prolongDuration)
}

// SetRequiredKeyTag is a paid mutator transaction binding the contract method 0xaf09ee17.
//
// Solidity: function setRequiredKeyTag(uint8 requiredKeyTag) returns()
func (_Master *MasterTransactor) SetRequiredKeyTag(opts *bind.TransactOpts, requiredKeyTag uint8) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setRequiredKeyTag", requiredKeyTag)
}

// SetRequiredKeyTag is a paid mutator transaction binding the contract method 0xaf09ee17.
//
// Solidity: function setRequiredKeyTag(uint8 requiredKeyTag) returns()
func (_Master *MasterSession) SetRequiredKeyTag(requiredKeyTag uint8) (*types.Transaction, error) {
	return _Master.Contract.SetRequiredKeyTag(&_Master.TransactOpts, requiredKeyTag)
}

// SetRequiredKeyTag is a paid mutator transaction binding the contract method 0xaf09ee17.
//
// Solidity: function setRequiredKeyTag(uint8 requiredKeyTag) returns()
func (_Master *MasterTransactorSession) SetRequiredKeyTag(requiredKeyTag uint8) (*types.Transaction, error) {
	return _Master.Contract.SetRequiredKeyTag(&_Master.TransactOpts, requiredKeyTag)
}

// SetRequiredKeyTags is a paid mutator transaction binding the contract method 0x4678a284.
//
// Solidity: function setRequiredKeyTags(uint8[] requiredKeyTags) returns()
func (_Master *MasterTransactor) SetRequiredKeyTags(opts *bind.TransactOpts, requiredKeyTags []uint8) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setRequiredKeyTags", requiredKeyTags)
}

// SetRequiredKeyTags is a paid mutator transaction binding the contract method 0x4678a284.
//
// Solidity: function setRequiredKeyTags(uint8[] requiredKeyTags) returns()
func (_Master *MasterSession) SetRequiredKeyTags(requiredKeyTags []uint8) (*types.Transaction, error) {
	return _Master.Contract.SetRequiredKeyTags(&_Master.TransactOpts, requiredKeyTags)
}

// SetRequiredKeyTags is a paid mutator transaction binding the contract method 0x4678a284.
//
// Solidity: function setRequiredKeyTags(uint8[] requiredKeyTags) returns()
func (_Master *MasterTransactorSession) SetRequiredKeyTags(requiredKeyTags []uint8) (*types.Transaction, error) {
	return _Master.Contract.SetRequiredKeyTags(&_Master.TransactOpts, requiredKeyTags)
}

// SetSigVerifier is a paid mutator transaction binding the contract method 0xbd7e9980.
//
// Solidity: function setSigVerifier(address sigVerifier) returns()
func (_Master *MasterTransactor) SetSigVerifier(opts *bind.TransactOpts, sigVerifier common.Address) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setSigVerifier", sigVerifier)
}

// SetSigVerifier is a paid mutator transaction binding the contract method 0xbd7e9980.
//
// Solidity: function setSigVerifier(address sigVerifier) returns()
func (_Master *MasterSession) SetSigVerifier(sigVerifier common.Address) (*types.Transaction, error) {
	return _Master.Contract.SetSigVerifier(&_Master.TransactOpts, sigVerifier)
}

// SetSigVerifier is a paid mutator transaction binding the contract method 0xbd7e9980.
//
// Solidity: function setSigVerifier(address sigVerifier) returns()
func (_Master *MasterTransactorSession) SetSigVerifier(sigVerifier common.Address) (*types.Transaction, error) {
	return _Master.Contract.SetSigVerifier(&_Master.TransactOpts, sigVerifier)
}

// SetVerificationType is a paid mutator transaction binding the contract method 0x7b8ef42d.
//
// Solidity: function setVerificationType(uint32 verificationType) returns()
func (_Master *MasterTransactor) SetVerificationType(opts *bind.TransactOpts, verificationType uint32) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setVerificationType", verificationType)
}

// SetVerificationType is a paid mutator transaction binding the contract method 0x7b8ef42d.
//
// Solidity: function setVerificationType(uint32 verificationType) returns()
func (_Master *MasterSession) SetVerificationType(verificationType uint32) (*types.Transaction, error) {
	return _Master.Contract.SetVerificationType(&_Master.TransactOpts, verificationType)
}

// SetVerificationType is a paid mutator transaction binding the contract method 0x7b8ef42d.
//
// Solidity: function setVerificationType(uint32 verificationType) returns()
func (_Master *MasterTransactorSession) SetVerificationType(verificationType uint32) (*types.Transaction, error) {
	return _Master.Contract.SetVerificationType(&_Master.TransactOpts, verificationType)
}

// StaticDelegateCall is a paid mutator transaction binding the contract method 0x9f86fd85.
//
// Solidity: function staticDelegateCall(address target, bytes data) returns()
func (_Master *MasterTransactor) StaticDelegateCall(opts *bind.TransactOpts, target common.Address, data []byte) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "staticDelegateCall", target, data)
}

// StaticDelegateCall is a paid mutator transaction binding the contract method 0x9f86fd85.
//
// Solidity: function staticDelegateCall(address target, bytes data) returns()
func (_Master *MasterSession) StaticDelegateCall(target common.Address, data []byte) (*types.Transaction, error) {
	return _Master.Contract.StaticDelegateCall(&_Master.TransactOpts, target, data)
}

// StaticDelegateCall is a paid mutator transaction binding the contract method 0x9f86fd85.
//
// Solidity: function staticDelegateCall(address target, bytes data) returns()
func (_Master *MasterTransactorSession) StaticDelegateCall(target common.Address, data []byte) (*types.Transaction, error) {
	return _Master.Contract.StaticDelegateCall(&_Master.TransactOpts, target, data)
}

// MasterAddReplicaIterator is returned from FilterAddReplica and is used to iterate over the raw logs and unpacked data for AddReplica events raised by the Master contract.
type MasterAddReplicaIterator struct {
	Event *MasterAddReplica // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterAddReplicaIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterAddReplica)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterAddReplica)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterAddReplicaIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterAddReplicaIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterAddReplica represents a AddReplica event raised by the Master contract.
type MasterAddReplica struct {
	Replica IConfigProviderCrossChainAddress
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterAddReplica is a free log retrieval operation binding the contract event 0x9c1d65d2b492b934ead150223b8db47248b2143ec0856e59ff4e9ef81fe818c5.
//
// Solidity: event AddReplica((address,uint64) replica)
func (_Master *MasterFilterer) FilterAddReplica(opts *bind.FilterOpts) (*MasterAddReplicaIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "AddReplica")
	if err != nil {
		return nil, err
	}
	return &MasterAddReplicaIterator{contract: _Master.contract, event: "AddReplica", logs: logs, sub: sub}, nil
}

// WatchAddReplica is a free log subscription operation binding the contract event 0x9c1d65d2b492b934ead150223b8db47248b2143ec0856e59ff4e9ef81fe818c5.
//
// Solidity: event AddReplica((address,uint64) replica)
func (_Master *MasterFilterer) WatchAddReplica(opts *bind.WatchOpts, sink chan<- *MasterAddReplica) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "AddReplica")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterAddReplica)
				if err := _Master.contract.UnpackLog(event, "AddReplica", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddReplica is a log parse operation binding the contract event 0x9c1d65d2b492b934ead150223b8db47248b2143ec0856e59ff4e9ef81fe818c5.
//
// Solidity: event AddReplica((address,uint64) replica)
func (_Master *MasterFilterer) ParseAddReplica(log types.Log) (*MasterAddReplica, error) {
	event := new(MasterAddReplica)
	if err := _Master.contract.UnpackLog(event, "AddReplica", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterAddVotingPowerProviderIterator is returned from FilterAddVotingPowerProvider and is used to iterate over the raw logs and unpacked data for AddVotingPowerProvider events raised by the Master contract.
type MasterAddVotingPowerProviderIterator struct {
	Event *MasterAddVotingPowerProvider // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterAddVotingPowerProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterAddVotingPowerProvider)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterAddVotingPowerProvider)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterAddVotingPowerProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterAddVotingPowerProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterAddVotingPowerProvider represents a AddVotingPowerProvider event raised by the Master contract.
type MasterAddVotingPowerProvider struct {
	VotingPowerProvider IConfigProviderCrossChainAddress
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterAddVotingPowerProvider is a free log retrieval operation binding the contract event 0x5b7921a35f3056f811d439a3b44c680d09bbbdc5a1ff8796bcb834a3d3bc7ce4.
//
// Solidity: event AddVotingPowerProvider((address,uint64) votingPowerProvider)
func (_Master *MasterFilterer) FilterAddVotingPowerProvider(opts *bind.FilterOpts) (*MasterAddVotingPowerProviderIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "AddVotingPowerProvider")
	if err != nil {
		return nil, err
	}
	return &MasterAddVotingPowerProviderIterator{contract: _Master.contract, event: "AddVotingPowerProvider", logs: logs, sub: sub}, nil
}

// WatchAddVotingPowerProvider is a free log subscription operation binding the contract event 0x5b7921a35f3056f811d439a3b44c680d09bbbdc5a1ff8796bcb834a3d3bc7ce4.
//
// Solidity: event AddVotingPowerProvider((address,uint64) votingPowerProvider)
func (_Master *MasterFilterer) WatchAddVotingPowerProvider(opts *bind.WatchOpts, sink chan<- *MasterAddVotingPowerProvider) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "AddVotingPowerProvider")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterAddVotingPowerProvider)
				if err := _Master.contract.UnpackLog(event, "AddVotingPowerProvider", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddVotingPowerProvider is a log parse operation binding the contract event 0x5b7921a35f3056f811d439a3b44c680d09bbbdc5a1ff8796bcb834a3d3bc7ce4.
//
// Solidity: event AddVotingPowerProvider((address,uint64) votingPowerProvider)
func (_Master *MasterFilterer) ParseAddVotingPowerProvider(log types.Log) (*MasterAddVotingPowerProvider, error) {
	event := new(MasterAddVotingPowerProvider)
	if err := _Master.contract.UnpackLog(event, "AddVotingPowerProvider", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterCommitValSetHeaderIterator is returned from FilterCommitValSetHeader and is used to iterate over the raw logs and unpacked data for CommitValSetHeader events raised by the Master contract.
type MasterCommitValSetHeaderIterator struct {
	Event *MasterCommitValSetHeader // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterCommitValSetHeaderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterCommitValSetHeader)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterCommitValSetHeader)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterCommitValSetHeaderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterCommitValSetHeaderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterCommitValSetHeader represents a CommitValSetHeader event raised by the Master contract.
type MasterCommitValSetHeader struct {
	ValSetHeader ISettlementValSetHeader
	ExtraData    []ISettlementExtraData
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCommitValSetHeader is a free log retrieval operation binding the contract event 0xb60178009515032571e0da79d8bbc5cd2781acdd8eb0b4e41f7f7479a8fadb65.
//
// Solidity: event CommitValSetHeader((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData)
func (_Master *MasterFilterer) FilterCommitValSetHeader(opts *bind.FilterOpts) (*MasterCommitValSetHeaderIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "CommitValSetHeader")
	if err != nil {
		return nil, err
	}
	return &MasterCommitValSetHeaderIterator{contract: _Master.contract, event: "CommitValSetHeader", logs: logs, sub: sub}, nil
}

// WatchCommitValSetHeader is a free log subscription operation binding the contract event 0xb60178009515032571e0da79d8bbc5cd2781acdd8eb0b4e41f7f7479a8fadb65.
//
// Solidity: event CommitValSetHeader((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData)
func (_Master *MasterFilterer) WatchCommitValSetHeader(opts *bind.WatchOpts, sink chan<- *MasterCommitValSetHeader) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "CommitValSetHeader")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterCommitValSetHeader)
				if err := _Master.contract.UnpackLog(event, "CommitValSetHeader", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCommitValSetHeader is a log parse operation binding the contract event 0xb60178009515032571e0da79d8bbc5cd2781acdd8eb0b4e41f7f7479a8fadb65.
//
// Solidity: event CommitValSetHeader((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData)
func (_Master *MasterFilterer) ParseCommitValSetHeader(log types.Log) (*MasterCommitValSetHeader, error) {
	event := new(MasterCommitValSetHeader)
	if err := _Master.contract.UnpackLog(event, "CommitValSetHeader", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the Master contract.
type MasterEIP712DomainChangedIterator struct {
	Event *MasterEIP712DomainChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterEIP712DomainChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterEIP712DomainChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterEIP712DomainChanged represents a EIP712DomainChanged event raised by the Master contract.
type MasterEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_Master *MasterFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*MasterEIP712DomainChangedIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &MasterEIP712DomainChangedIterator{contract: _Master.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_Master *MasterFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *MasterEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterEIP712DomainChanged)
				if err := _Master.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_Master *MasterFilterer) ParseEIP712DomainChanged(log types.Log) (*MasterEIP712DomainChanged, error) {
	event := new(MasterEIP712DomainChanged)
	if err := _Master.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitCommitDurationIterator is returned from FilterInitCommitDuration and is used to iterate over the raw logs and unpacked data for InitCommitDuration events raised by the Master contract.
type MasterInitCommitDurationIterator struct {
	Event *MasterInitCommitDuration // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitCommitDurationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitCommitDuration)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitCommitDuration)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitCommitDurationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitCommitDurationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitCommitDuration represents a InitCommitDuration event raised by the Master contract.
type MasterInitCommitDuration struct {
	CommitDuration *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterInitCommitDuration is a free log retrieval operation binding the contract event 0x1645b3abdf9aaf0a7cea7c5fa9cb5d186b29d8767b77f32f71535d690d415a3b.
//
// Solidity: event InitCommitDuration(uint48 commitDuration)
func (_Master *MasterFilterer) FilterInitCommitDuration(opts *bind.FilterOpts) (*MasterInitCommitDurationIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitCommitDuration")
	if err != nil {
		return nil, err
	}
	return &MasterInitCommitDurationIterator{contract: _Master.contract, event: "InitCommitDuration", logs: logs, sub: sub}, nil
}

// WatchInitCommitDuration is a free log subscription operation binding the contract event 0x1645b3abdf9aaf0a7cea7c5fa9cb5d186b29d8767b77f32f71535d690d415a3b.
//
// Solidity: event InitCommitDuration(uint48 commitDuration)
func (_Master *MasterFilterer) WatchInitCommitDuration(opts *bind.WatchOpts, sink chan<- *MasterInitCommitDuration) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitCommitDuration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitCommitDuration)
				if err := _Master.contract.UnpackLog(event, "InitCommitDuration", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitCommitDuration is a log parse operation binding the contract event 0x1645b3abdf9aaf0a7cea7c5fa9cb5d186b29d8767b77f32f71535d690d415a3b.
//
// Solidity: event InitCommitDuration(uint48 commitDuration)
func (_Master *MasterFilterer) ParseInitCommitDuration(log types.Log) (*MasterInitCommitDuration, error) {
	event := new(MasterInitCommitDuration)
	if err := _Master.contract.UnpackLog(event, "InitCommitDuration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitEIP712Iterator is returned from FilterInitEIP712 and is used to iterate over the raw logs and unpacked data for InitEIP712 events raised by the Master contract.
type MasterInitEIP712Iterator struct {
	Event *MasterInitEIP712 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitEIP712Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitEIP712)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitEIP712)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitEIP712Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitEIP712Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitEIP712 represents a InitEIP712 event raised by the Master contract.
type MasterInitEIP712 struct {
	Name    string
	Version string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitEIP712 is a free log retrieval operation binding the contract event 0x98790bb3996c909e6f4279ffabdfe70fa6c0d49b8fa04656d6161decfc442e0a.
//
// Solidity: event InitEIP712(string name, string version)
func (_Master *MasterFilterer) FilterInitEIP712(opts *bind.FilterOpts) (*MasterInitEIP712Iterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitEIP712")
	if err != nil {
		return nil, err
	}
	return &MasterInitEIP712Iterator{contract: _Master.contract, event: "InitEIP712", logs: logs, sub: sub}, nil
}

// WatchInitEIP712 is a free log subscription operation binding the contract event 0x98790bb3996c909e6f4279ffabdfe70fa6c0d49b8fa04656d6161decfc442e0a.
//
// Solidity: event InitEIP712(string name, string version)
func (_Master *MasterFilterer) WatchInitEIP712(opts *bind.WatchOpts, sink chan<- *MasterInitEIP712) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitEIP712")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitEIP712)
				if err := _Master.contract.UnpackLog(event, "InitEIP712", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitEIP712 is a log parse operation binding the contract event 0x98790bb3996c909e6f4279ffabdfe70fa6c0d49b8fa04656d6161decfc442e0a.
//
// Solidity: event InitEIP712(string name, string version)
func (_Master *MasterFilterer) ParseInitEIP712(log types.Log) (*MasterInitEIP712, error) {
	event := new(MasterInitEIP712)
	if err := _Master.contract.UnpackLog(event, "InitEIP712", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitEpochDurationIterator is returned from FilterInitEpochDuration and is used to iterate over the raw logs and unpacked data for InitEpochDuration events raised by the Master contract.
type MasterInitEpochDurationIterator struct {
	Event *MasterInitEpochDuration // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitEpochDurationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitEpochDuration)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitEpochDuration)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitEpochDurationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitEpochDurationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitEpochDuration represents a InitEpochDuration event raised by the Master contract.
type MasterInitEpochDuration struct {
	EpochDuration          *big.Int
	EpochDurationTimestamp *big.Int
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterInitEpochDuration is a free log retrieval operation binding the contract event 0xf688b7b02a20c2dda7d7de03a41637b274af7706eb975ea4af45858648370f55.
//
// Solidity: event InitEpochDuration(uint48 epochDuration, uint48 epochDurationTimestamp)
func (_Master *MasterFilterer) FilterInitEpochDuration(opts *bind.FilterOpts) (*MasterInitEpochDurationIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitEpochDuration")
	if err != nil {
		return nil, err
	}
	return &MasterInitEpochDurationIterator{contract: _Master.contract, event: "InitEpochDuration", logs: logs, sub: sub}, nil
}

// WatchInitEpochDuration is a free log subscription operation binding the contract event 0xf688b7b02a20c2dda7d7de03a41637b274af7706eb975ea4af45858648370f55.
//
// Solidity: event InitEpochDuration(uint48 epochDuration, uint48 epochDurationTimestamp)
func (_Master *MasterFilterer) WatchInitEpochDuration(opts *bind.WatchOpts, sink chan<- *MasterInitEpochDuration) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitEpochDuration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitEpochDuration)
				if err := _Master.contract.UnpackLog(event, "InitEpochDuration", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitEpochDuration is a log parse operation binding the contract event 0xf688b7b02a20c2dda7d7de03a41637b274af7706eb975ea4af45858648370f55.
//
// Solidity: event InitEpochDuration(uint48 epochDuration, uint48 epochDurationTimestamp)
func (_Master *MasterFilterer) ParseInitEpochDuration(log types.Log) (*MasterInitEpochDuration, error) {
	event := new(MasterInitEpochDuration)
	if err := _Master.contract.UnpackLog(event, "InitEpochDuration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitProlongDurationIterator is returned from FilterInitProlongDuration and is used to iterate over the raw logs and unpacked data for InitProlongDuration events raised by the Master contract.
type MasterInitProlongDurationIterator struct {
	Event *MasterInitProlongDuration // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitProlongDurationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitProlongDuration)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitProlongDuration)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitProlongDurationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitProlongDurationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitProlongDuration represents a InitProlongDuration event raised by the Master contract.
type MasterInitProlongDuration struct {
	ProlongDuration *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterInitProlongDuration is a free log retrieval operation binding the contract event 0x79728126a42e80f49fc6905f6e1ea6632294402d214398093cdbc526bdeb5608.
//
// Solidity: event InitProlongDuration(uint48 prolongDuration)
func (_Master *MasterFilterer) FilterInitProlongDuration(opts *bind.FilterOpts) (*MasterInitProlongDurationIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitProlongDuration")
	if err != nil {
		return nil, err
	}
	return &MasterInitProlongDurationIterator{contract: _Master.contract, event: "InitProlongDuration", logs: logs, sub: sub}, nil
}

// WatchInitProlongDuration is a free log subscription operation binding the contract event 0x79728126a42e80f49fc6905f6e1ea6632294402d214398093cdbc526bdeb5608.
//
// Solidity: event InitProlongDuration(uint48 prolongDuration)
func (_Master *MasterFilterer) WatchInitProlongDuration(opts *bind.WatchOpts, sink chan<- *MasterInitProlongDuration) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitProlongDuration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitProlongDuration)
				if err := _Master.contract.UnpackLog(event, "InitProlongDuration", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitProlongDuration is a log parse operation binding the contract event 0x79728126a42e80f49fc6905f6e1ea6632294402d214398093cdbc526bdeb5608.
//
// Solidity: event InitProlongDuration(uint48 prolongDuration)
func (_Master *MasterFilterer) ParseInitProlongDuration(log types.Log) (*MasterInitProlongDuration, error) {
	event := new(MasterInitProlongDuration)
	if err := _Master.contract.UnpackLog(event, "InitProlongDuration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitRequiredKeyTagIterator is returned from FilterInitRequiredKeyTag and is used to iterate over the raw logs and unpacked data for InitRequiredKeyTag events raised by the Master contract.
type MasterInitRequiredKeyTagIterator struct {
	Event *MasterInitRequiredKeyTag // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitRequiredKeyTagIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitRequiredKeyTag)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitRequiredKeyTag)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitRequiredKeyTagIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitRequiredKeyTagIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitRequiredKeyTag represents a InitRequiredKeyTag event raised by the Master contract.
type MasterInitRequiredKeyTag struct {
	RequiredKeyTag uint8
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterInitRequiredKeyTag is a free log retrieval operation binding the contract event 0x277cc578e04aceb9647006598e72834d06583784f6c39002b444396f8be20f81.
//
// Solidity: event InitRequiredKeyTag(uint8 requiredKeyTag)
func (_Master *MasterFilterer) FilterInitRequiredKeyTag(opts *bind.FilterOpts) (*MasterInitRequiredKeyTagIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitRequiredKeyTag")
	if err != nil {
		return nil, err
	}
	return &MasterInitRequiredKeyTagIterator{contract: _Master.contract, event: "InitRequiredKeyTag", logs: logs, sub: sub}, nil
}

// WatchInitRequiredKeyTag is a free log subscription operation binding the contract event 0x277cc578e04aceb9647006598e72834d06583784f6c39002b444396f8be20f81.
//
// Solidity: event InitRequiredKeyTag(uint8 requiredKeyTag)
func (_Master *MasterFilterer) WatchInitRequiredKeyTag(opts *bind.WatchOpts, sink chan<- *MasterInitRequiredKeyTag) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitRequiredKeyTag")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitRequiredKeyTag)
				if err := _Master.contract.UnpackLog(event, "InitRequiredKeyTag", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitRequiredKeyTag is a log parse operation binding the contract event 0x277cc578e04aceb9647006598e72834d06583784f6c39002b444396f8be20f81.
//
// Solidity: event InitRequiredKeyTag(uint8 requiredKeyTag)
func (_Master *MasterFilterer) ParseInitRequiredKeyTag(log types.Log) (*MasterInitRequiredKeyTag, error) {
	event := new(MasterInitRequiredKeyTag)
	if err := _Master.contract.UnpackLog(event, "InitRequiredKeyTag", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitSigVerifierIterator is returned from FilterInitSigVerifier and is used to iterate over the raw logs and unpacked data for InitSigVerifier events raised by the Master contract.
type MasterInitSigVerifierIterator struct {
	Event *MasterInitSigVerifier // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitSigVerifierIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitSigVerifier)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitSigVerifier)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitSigVerifierIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitSigVerifierIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitSigVerifier represents a InitSigVerifier event raised by the Master contract.
type MasterInitSigVerifier struct {
	SigVerifier common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitSigVerifier is a free log retrieval operation binding the contract event 0x8c698070f0c9ef92ff032a24e5c83ef7783fd360fde9c6af8ed5fca9fa5abbb7.
//
// Solidity: event InitSigVerifier(address sigVerifier)
func (_Master *MasterFilterer) FilterInitSigVerifier(opts *bind.FilterOpts) (*MasterInitSigVerifierIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitSigVerifier")
	if err != nil {
		return nil, err
	}
	return &MasterInitSigVerifierIterator{contract: _Master.contract, event: "InitSigVerifier", logs: logs, sub: sub}, nil
}

// WatchInitSigVerifier is a free log subscription operation binding the contract event 0x8c698070f0c9ef92ff032a24e5c83ef7783fd360fde9c6af8ed5fca9fa5abbb7.
//
// Solidity: event InitSigVerifier(address sigVerifier)
func (_Master *MasterFilterer) WatchInitSigVerifier(opts *bind.WatchOpts, sink chan<- *MasterInitSigVerifier) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitSigVerifier")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitSigVerifier)
				if err := _Master.contract.UnpackLog(event, "InitSigVerifier", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitSigVerifier is a log parse operation binding the contract event 0x8c698070f0c9ef92ff032a24e5c83ef7783fd360fde9c6af8ed5fca9fa5abbb7.
//
// Solidity: event InitSigVerifier(address sigVerifier)
func (_Master *MasterFilterer) ParseInitSigVerifier(log types.Log) (*MasterInitSigVerifier, error) {
	event := new(MasterInitSigVerifier)
	if err := _Master.contract.UnpackLog(event, "InitSigVerifier", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitSubnetworkIterator is returned from FilterInitSubnetwork and is used to iterate over the raw logs and unpacked data for InitSubnetwork events raised by the Master contract.
type MasterInitSubnetworkIterator struct {
	Event *MasterInitSubnetwork // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitSubnetworkIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitSubnetwork)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitSubnetwork)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitSubnetworkIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitSubnetworkIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitSubnetwork represents a InitSubnetwork event raised by the Master contract.
type MasterInitSubnetwork struct {
	Network      common.Address
	SubnetworkID *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterInitSubnetwork is a free log retrieval operation binding the contract event 0x469c2e982e7d76d34cf5d1e72abee29749bb9971942c180e9023cea09f5f8e83.
//
// Solidity: event InitSubnetwork(address network, uint96 subnetworkID)
func (_Master *MasterFilterer) FilterInitSubnetwork(opts *bind.FilterOpts) (*MasterInitSubnetworkIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "InitSubnetwork")
	if err != nil {
		return nil, err
	}
	return &MasterInitSubnetworkIterator{contract: _Master.contract, event: "InitSubnetwork", logs: logs, sub: sub}, nil
}

// WatchInitSubnetwork is a free log subscription operation binding the contract event 0x469c2e982e7d76d34cf5d1e72abee29749bb9971942c180e9023cea09f5f8e83.
//
// Solidity: event InitSubnetwork(address network, uint96 subnetworkID)
func (_Master *MasterFilterer) WatchInitSubnetwork(opts *bind.WatchOpts, sink chan<- *MasterInitSubnetwork) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "InitSubnetwork")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitSubnetwork)
				if err := _Master.contract.UnpackLog(event, "InitSubnetwork", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitSubnetwork is a log parse operation binding the contract event 0x469c2e982e7d76d34cf5d1e72abee29749bb9971942c180e9023cea09f5f8e83.
//
// Solidity: event InitSubnetwork(address network, uint96 subnetworkID)
func (_Master *MasterFilterer) ParseInitSubnetwork(log types.Log) (*MasterInitSubnetwork, error) {
	event := new(MasterInitSubnetwork)
	if err := _Master.contract.UnpackLog(event, "InitSubnetwork", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Master contract.
type MasterInitializedIterator struct {
	Event *MasterInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterInitialized represents a Initialized event raised by the Master contract.
type MasterInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Master *MasterFilterer) FilterInitialized(opts *bind.FilterOpts) (*MasterInitializedIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &MasterInitializedIterator{contract: _Master.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Master *MasterFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *MasterInitialized) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterInitialized)
				if err := _Master.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Master *MasterFilterer) ParseInitialized(log types.Log) (*MasterInitialized, error) {
	event := new(MasterInitialized)
	if err := _Master.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterRemoveReplicaIterator is returned from FilterRemoveReplica and is used to iterate over the raw logs and unpacked data for RemoveReplica events raised by the Master contract.
type MasterRemoveReplicaIterator struct {
	Event *MasterRemoveReplica // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterRemoveReplicaIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterRemoveReplica)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterRemoveReplica)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterRemoveReplicaIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterRemoveReplicaIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterRemoveReplica represents a RemoveReplica event raised by the Master contract.
type MasterRemoveReplica struct {
	Replica IConfigProviderCrossChainAddress
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRemoveReplica is a free log retrieval operation binding the contract event 0x786c47cb2e925115cd5f876a5d539abae89d23218b6da9607162ace8d3a4d29b.
//
// Solidity: event RemoveReplica((address,uint64) replica)
func (_Master *MasterFilterer) FilterRemoveReplica(opts *bind.FilterOpts) (*MasterRemoveReplicaIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "RemoveReplica")
	if err != nil {
		return nil, err
	}
	return &MasterRemoveReplicaIterator{contract: _Master.contract, event: "RemoveReplica", logs: logs, sub: sub}, nil
}

// WatchRemoveReplica is a free log subscription operation binding the contract event 0x786c47cb2e925115cd5f876a5d539abae89d23218b6da9607162ace8d3a4d29b.
//
// Solidity: event RemoveReplica((address,uint64) replica)
func (_Master *MasterFilterer) WatchRemoveReplica(opts *bind.WatchOpts, sink chan<- *MasterRemoveReplica) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "RemoveReplica")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterRemoveReplica)
				if err := _Master.contract.UnpackLog(event, "RemoveReplica", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRemoveReplica is a log parse operation binding the contract event 0x786c47cb2e925115cd5f876a5d539abae89d23218b6da9607162ace8d3a4d29b.
//
// Solidity: event RemoveReplica((address,uint64) replica)
func (_Master *MasterFilterer) ParseRemoveReplica(log types.Log) (*MasterRemoveReplica, error) {
	event := new(MasterRemoveReplica)
	if err := _Master.contract.UnpackLog(event, "RemoveReplica", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterRemoveVotingPowerProviderIterator is returned from FilterRemoveVotingPowerProvider and is used to iterate over the raw logs and unpacked data for RemoveVotingPowerProvider events raised by the Master contract.
type MasterRemoveVotingPowerProviderIterator struct {
	Event *MasterRemoveVotingPowerProvider // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterRemoveVotingPowerProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterRemoveVotingPowerProvider)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterRemoveVotingPowerProvider)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterRemoveVotingPowerProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterRemoveVotingPowerProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterRemoveVotingPowerProvider represents a RemoveVotingPowerProvider event raised by the Master contract.
type MasterRemoveVotingPowerProvider struct {
	VotingPowerProvider IConfigProviderCrossChainAddress
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterRemoveVotingPowerProvider is a free log retrieval operation binding the contract event 0xfa08af0a4f2329b2baeb1b6c7b9ebc7bf6247abd9efb7e735a644bc98e921075.
//
// Solidity: event RemoveVotingPowerProvider((address,uint64) votingPowerProvider)
func (_Master *MasterFilterer) FilterRemoveVotingPowerProvider(opts *bind.FilterOpts) (*MasterRemoveVotingPowerProviderIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "RemoveVotingPowerProvider")
	if err != nil {
		return nil, err
	}
	return &MasterRemoveVotingPowerProviderIterator{contract: _Master.contract, event: "RemoveVotingPowerProvider", logs: logs, sub: sub}, nil
}

// WatchRemoveVotingPowerProvider is a free log subscription operation binding the contract event 0xfa08af0a4f2329b2baeb1b6c7b9ebc7bf6247abd9efb7e735a644bc98e921075.
//
// Solidity: event RemoveVotingPowerProvider((address,uint64) votingPowerProvider)
func (_Master *MasterFilterer) WatchRemoveVotingPowerProvider(opts *bind.WatchOpts, sink chan<- *MasterRemoveVotingPowerProvider) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "RemoveVotingPowerProvider")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterRemoveVotingPowerProvider)
				if err := _Master.contract.UnpackLog(event, "RemoveVotingPowerProvider", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRemoveVotingPowerProvider is a log parse operation binding the contract event 0xfa08af0a4f2329b2baeb1b6c7b9ebc7bf6247abd9efb7e735a644bc98e921075.
//
// Solidity: event RemoveVotingPowerProvider((address,uint64) votingPowerProvider)
func (_Master *MasterFilterer) ParseRemoveVotingPowerProvider(log types.Log) (*MasterRemoveVotingPowerProvider, error) {
	event := new(MasterRemoveVotingPowerProvider)
	if err := _Master.contract.UnpackLog(event, "RemoveVotingPowerProvider", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Master contract.
type MasterRoleAdminChangedIterator struct {
	Event *MasterRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterRoleAdminChanged represents a RoleAdminChanged event raised by the Master contract.
type MasterRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Master *MasterFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*MasterRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Master.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &MasterRoleAdminChangedIterator{contract: _Master.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Master *MasterFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *MasterRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Master.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterRoleAdminChanged)
				if err := _Master.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Master *MasterFilterer) ParseRoleAdminChanged(log types.Log) (*MasterRoleAdminChanged, error) {
	event := new(MasterRoleAdminChanged)
	if err := _Master.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Master contract.
type MasterRoleGrantedIterator struct {
	Event *MasterRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterRoleGranted represents a RoleGranted event raised by the Master contract.
type MasterRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Master *MasterFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*MasterRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Master.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MasterRoleGrantedIterator{contract: _Master.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Master *MasterFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *MasterRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Master.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterRoleGranted)
				if err := _Master.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Master *MasterFilterer) ParseRoleGranted(log types.Log) (*MasterRoleGranted, error) {
	event := new(MasterRoleGranted)
	if err := _Master.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Master contract.
type MasterRoleRevokedIterator struct {
	Event *MasterRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterRoleRevoked represents a RoleRevoked event raised by the Master contract.
type MasterRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Master *MasterFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*MasterRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Master.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MasterRoleRevokedIterator{contract: _Master.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Master *MasterFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *MasterRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Master.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterRoleRevoked)
				if err := _Master.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Master *MasterFilterer) ParseRoleRevoked(log types.Log) (*MasterRoleRevoked, error) {
	event := new(MasterRoleRevoked)
	if err := _Master.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetCommitDurationIterator is returned from FilterSetCommitDuration and is used to iterate over the raw logs and unpacked data for SetCommitDuration events raised by the Master contract.
type MasterSetCommitDurationIterator struct {
	Event *MasterSetCommitDuration // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetCommitDurationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetCommitDuration)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetCommitDuration)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetCommitDurationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetCommitDurationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetCommitDuration represents a SetCommitDuration event raised by the Master contract.
type MasterSetCommitDuration struct {
	CommitDuration *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetCommitDuration is a free log retrieval operation binding the contract event 0xae8dc913d64bd44cca39779903f29196e6986435d34e6d5341e5c34c29cdb25e.
//
// Solidity: event SetCommitDuration(uint48 commitDuration)
func (_Master *MasterFilterer) FilterSetCommitDuration(opts *bind.FilterOpts) (*MasterSetCommitDurationIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetCommitDuration")
	if err != nil {
		return nil, err
	}
	return &MasterSetCommitDurationIterator{contract: _Master.contract, event: "SetCommitDuration", logs: logs, sub: sub}, nil
}

// WatchSetCommitDuration is a free log subscription operation binding the contract event 0xae8dc913d64bd44cca39779903f29196e6986435d34e6d5341e5c34c29cdb25e.
//
// Solidity: event SetCommitDuration(uint48 commitDuration)
func (_Master *MasterFilterer) WatchSetCommitDuration(opts *bind.WatchOpts, sink chan<- *MasterSetCommitDuration) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetCommitDuration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetCommitDuration)
				if err := _Master.contract.UnpackLog(event, "SetCommitDuration", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetCommitDuration is a log parse operation binding the contract event 0xae8dc913d64bd44cca39779903f29196e6986435d34e6d5341e5c34c29cdb25e.
//
// Solidity: event SetCommitDuration(uint48 commitDuration)
func (_Master *MasterFilterer) ParseSetCommitDuration(log types.Log) (*MasterSetCommitDuration, error) {
	event := new(MasterSetCommitDuration)
	if err := _Master.contract.UnpackLog(event, "SetCommitDuration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetEpochDurationIterator is returned from FilterSetEpochDuration and is used to iterate over the raw logs and unpacked data for SetEpochDuration events raised by the Master contract.
type MasterSetEpochDurationIterator struct {
	Event *MasterSetEpochDuration // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetEpochDurationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetEpochDuration)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetEpochDuration)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetEpochDurationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetEpochDurationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetEpochDuration represents a SetEpochDuration event raised by the Master contract.
type MasterSetEpochDuration struct {
	EpochDuration *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSetEpochDuration is a free log retrieval operation binding the contract event 0xc950f06b73b224f8b32d39245a5905020aebfc426a15833a70ac2e4e2ebe098c.
//
// Solidity: event SetEpochDuration(uint48 epochDuration)
func (_Master *MasterFilterer) FilterSetEpochDuration(opts *bind.FilterOpts) (*MasterSetEpochDurationIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetEpochDuration")
	if err != nil {
		return nil, err
	}
	return &MasterSetEpochDurationIterator{contract: _Master.contract, event: "SetEpochDuration", logs: logs, sub: sub}, nil
}

// WatchSetEpochDuration is a free log subscription operation binding the contract event 0xc950f06b73b224f8b32d39245a5905020aebfc426a15833a70ac2e4e2ebe098c.
//
// Solidity: event SetEpochDuration(uint48 epochDuration)
func (_Master *MasterFilterer) WatchSetEpochDuration(opts *bind.WatchOpts, sink chan<- *MasterSetEpochDuration) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetEpochDuration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetEpochDuration)
				if err := _Master.contract.UnpackLog(event, "SetEpochDuration", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetEpochDuration is a log parse operation binding the contract event 0xc950f06b73b224f8b32d39245a5905020aebfc426a15833a70ac2e4e2ebe098c.
//
// Solidity: event SetEpochDuration(uint48 epochDuration)
func (_Master *MasterFilterer) ParseSetEpochDuration(log types.Log) (*MasterSetEpochDuration, error) {
	event := new(MasterSetEpochDuration)
	if err := _Master.contract.UnpackLog(event, "SetEpochDuration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetGenesisIterator is returned from FilterSetGenesis and is used to iterate over the raw logs and unpacked data for SetGenesis events raised by the Master contract.
type MasterSetGenesisIterator struct {
	Event *MasterSetGenesis // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetGenesisIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetGenesis)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetGenesis)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetGenesisIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetGenesisIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetGenesis represents a SetGenesis event raised by the Master contract.
type MasterSetGenesis struct {
	ValSetHeader ISettlementValSetHeader
	ExtraData    []ISettlementExtraData
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSetGenesis is a free log retrieval operation binding the contract event 0x9bd53a9a031f528c533d7ad36096850d291dd1ddef9a2dc0b41c5623d99fec9f.
//
// Solidity: event SetGenesis((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData)
func (_Master *MasterFilterer) FilterSetGenesis(opts *bind.FilterOpts) (*MasterSetGenesisIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetGenesis")
	if err != nil {
		return nil, err
	}
	return &MasterSetGenesisIterator{contract: _Master.contract, event: "SetGenesis", logs: logs, sub: sub}, nil
}

// WatchSetGenesis is a free log subscription operation binding the contract event 0x9bd53a9a031f528c533d7ad36096850d291dd1ddef9a2dc0b41c5623d99fec9f.
//
// Solidity: event SetGenesis((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData)
func (_Master *MasterFilterer) WatchSetGenesis(opts *bind.WatchOpts, sink chan<- *MasterSetGenesis) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetGenesis")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetGenesis)
				if err := _Master.contract.UnpackLog(event, "SetGenesis", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetGenesis is a log parse operation binding the contract event 0x9bd53a9a031f528c533d7ad36096850d291dd1ddef9a2dc0b41c5623d99fec9f.
//
// Solidity: event SetGenesis((uint8,uint8,uint48,uint48,uint256,bytes32,bytes32) valSetHeader, (bytes32,bytes32)[] extraData)
func (_Master *MasterFilterer) ParseSetGenesis(log types.Log) (*MasterSetGenesis, error) {
	event := new(MasterSetGenesis)
	if err := _Master.contract.UnpackLog(event, "SetGenesis", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetKeysProviderIterator is returned from FilterSetKeysProvider and is used to iterate over the raw logs and unpacked data for SetKeysProvider events raised by the Master contract.
type MasterSetKeysProviderIterator struct {
	Event *MasterSetKeysProvider // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetKeysProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetKeysProvider)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetKeysProvider)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetKeysProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetKeysProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetKeysProvider represents a SetKeysProvider event raised by the Master contract.
type MasterSetKeysProvider struct {
	KeysProvider IConfigProviderCrossChainAddress
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSetKeysProvider is a free log retrieval operation binding the contract event 0xc7810d91e677f4ec97a34d6ec4b8d67e94f7ece239370f2a319b9792abc72251.
//
// Solidity: event SetKeysProvider((address,uint64) keysProvider)
func (_Master *MasterFilterer) FilterSetKeysProvider(opts *bind.FilterOpts) (*MasterSetKeysProviderIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetKeysProvider")
	if err != nil {
		return nil, err
	}
	return &MasterSetKeysProviderIterator{contract: _Master.contract, event: "SetKeysProvider", logs: logs, sub: sub}, nil
}

// WatchSetKeysProvider is a free log subscription operation binding the contract event 0xc7810d91e677f4ec97a34d6ec4b8d67e94f7ece239370f2a319b9792abc72251.
//
// Solidity: event SetKeysProvider((address,uint64) keysProvider)
func (_Master *MasterFilterer) WatchSetKeysProvider(opts *bind.WatchOpts, sink chan<- *MasterSetKeysProvider) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetKeysProvider")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetKeysProvider)
				if err := _Master.contract.UnpackLog(event, "SetKeysProvider", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetKeysProvider is a log parse operation binding the contract event 0xc7810d91e677f4ec97a34d6ec4b8d67e94f7ece239370f2a319b9792abc72251.
//
// Solidity: event SetKeysProvider((address,uint64) keysProvider)
func (_Master *MasterFilterer) ParseSetKeysProvider(log types.Log) (*MasterSetKeysProvider, error) {
	event := new(MasterSetKeysProvider)
	if err := _Master.contract.UnpackLog(event, "SetKeysProvider", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetMaxValidatorsCountIterator is returned from FilterSetMaxValidatorsCount and is used to iterate over the raw logs and unpacked data for SetMaxValidatorsCount events raised by the Master contract.
type MasterSetMaxValidatorsCountIterator struct {
	Event *MasterSetMaxValidatorsCount // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetMaxValidatorsCountIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetMaxValidatorsCount)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetMaxValidatorsCount)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetMaxValidatorsCountIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetMaxValidatorsCountIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetMaxValidatorsCount represents a SetMaxValidatorsCount event raised by the Master contract.
type MasterSetMaxValidatorsCount struct {
	MaxValidatorsCount *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterSetMaxValidatorsCount is a free log retrieval operation binding the contract event 0x37ca3532b507cfa33b11765ae8b499cb6830421b982a7f8837ee71ca5a3119c8.
//
// Solidity: event SetMaxValidatorsCount(uint208 maxValidatorsCount)
func (_Master *MasterFilterer) FilterSetMaxValidatorsCount(opts *bind.FilterOpts) (*MasterSetMaxValidatorsCountIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetMaxValidatorsCount")
	if err != nil {
		return nil, err
	}
	return &MasterSetMaxValidatorsCountIterator{contract: _Master.contract, event: "SetMaxValidatorsCount", logs: logs, sub: sub}, nil
}

// WatchSetMaxValidatorsCount is a free log subscription operation binding the contract event 0x37ca3532b507cfa33b11765ae8b499cb6830421b982a7f8837ee71ca5a3119c8.
//
// Solidity: event SetMaxValidatorsCount(uint208 maxValidatorsCount)
func (_Master *MasterFilterer) WatchSetMaxValidatorsCount(opts *bind.WatchOpts, sink chan<- *MasterSetMaxValidatorsCount) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetMaxValidatorsCount")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetMaxValidatorsCount)
				if err := _Master.contract.UnpackLog(event, "SetMaxValidatorsCount", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetMaxValidatorsCount is a log parse operation binding the contract event 0x37ca3532b507cfa33b11765ae8b499cb6830421b982a7f8837ee71ca5a3119c8.
//
// Solidity: event SetMaxValidatorsCount(uint208 maxValidatorsCount)
func (_Master *MasterFilterer) ParseSetMaxValidatorsCount(log types.Log) (*MasterSetMaxValidatorsCount, error) {
	event := new(MasterSetMaxValidatorsCount)
	if err := _Master.contract.UnpackLog(event, "SetMaxValidatorsCount", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetMaxVotingPowerIterator is returned from FilterSetMaxVotingPower and is used to iterate over the raw logs and unpacked data for SetMaxVotingPower events raised by the Master contract.
type MasterSetMaxVotingPowerIterator struct {
	Event *MasterSetMaxVotingPower // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetMaxVotingPowerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetMaxVotingPower)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetMaxVotingPower)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetMaxVotingPowerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetMaxVotingPowerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetMaxVotingPower represents a SetMaxVotingPower event raised by the Master contract.
type MasterSetMaxVotingPower struct {
	MaxVotingPower *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetMaxVotingPower is a free log retrieval operation binding the contract event 0xe891886eac9e583940fb0844098689693a4d105206ec1f789d119b4314383b95.
//
// Solidity: event SetMaxVotingPower(uint256 maxVotingPower)
func (_Master *MasterFilterer) FilterSetMaxVotingPower(opts *bind.FilterOpts) (*MasterSetMaxVotingPowerIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetMaxVotingPower")
	if err != nil {
		return nil, err
	}
	return &MasterSetMaxVotingPowerIterator{contract: _Master.contract, event: "SetMaxVotingPower", logs: logs, sub: sub}, nil
}

// WatchSetMaxVotingPower is a free log subscription operation binding the contract event 0xe891886eac9e583940fb0844098689693a4d105206ec1f789d119b4314383b95.
//
// Solidity: event SetMaxVotingPower(uint256 maxVotingPower)
func (_Master *MasterFilterer) WatchSetMaxVotingPower(opts *bind.WatchOpts, sink chan<- *MasterSetMaxVotingPower) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetMaxVotingPower")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetMaxVotingPower)
				if err := _Master.contract.UnpackLog(event, "SetMaxVotingPower", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetMaxVotingPower is a log parse operation binding the contract event 0xe891886eac9e583940fb0844098689693a4d105206ec1f789d119b4314383b95.
//
// Solidity: event SetMaxVotingPower(uint256 maxVotingPower)
func (_Master *MasterFilterer) ParseSetMaxVotingPower(log types.Log) (*MasterSetMaxVotingPower, error) {
	event := new(MasterSetMaxVotingPower)
	if err := _Master.contract.UnpackLog(event, "SetMaxVotingPower", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetMinInclusionVotingPowerIterator is returned from FilterSetMinInclusionVotingPower and is used to iterate over the raw logs and unpacked data for SetMinInclusionVotingPower events raised by the Master contract.
type MasterSetMinInclusionVotingPowerIterator struct {
	Event *MasterSetMinInclusionVotingPower // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetMinInclusionVotingPowerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetMinInclusionVotingPower)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetMinInclusionVotingPower)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetMinInclusionVotingPowerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetMinInclusionVotingPowerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetMinInclusionVotingPower represents a SetMinInclusionVotingPower event raised by the Master contract.
type MasterSetMinInclusionVotingPower struct {
	MinInclusionVotingPower *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterSetMinInclusionVotingPower is a free log retrieval operation binding the contract event 0x7ea1f11872caff0567f050bd06f29f128a1407e56e3272abbadef87f6cbb6188.
//
// Solidity: event SetMinInclusionVotingPower(uint256 minInclusionVotingPower)
func (_Master *MasterFilterer) FilterSetMinInclusionVotingPower(opts *bind.FilterOpts) (*MasterSetMinInclusionVotingPowerIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetMinInclusionVotingPower")
	if err != nil {
		return nil, err
	}
	return &MasterSetMinInclusionVotingPowerIterator{contract: _Master.contract, event: "SetMinInclusionVotingPower", logs: logs, sub: sub}, nil
}

// WatchSetMinInclusionVotingPower is a free log subscription operation binding the contract event 0x7ea1f11872caff0567f050bd06f29f128a1407e56e3272abbadef87f6cbb6188.
//
// Solidity: event SetMinInclusionVotingPower(uint256 minInclusionVotingPower)
func (_Master *MasterFilterer) WatchSetMinInclusionVotingPower(opts *bind.WatchOpts, sink chan<- *MasterSetMinInclusionVotingPower) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetMinInclusionVotingPower")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetMinInclusionVotingPower)
				if err := _Master.contract.UnpackLog(event, "SetMinInclusionVotingPower", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetMinInclusionVotingPower is a log parse operation binding the contract event 0x7ea1f11872caff0567f050bd06f29f128a1407e56e3272abbadef87f6cbb6188.
//
// Solidity: event SetMinInclusionVotingPower(uint256 minInclusionVotingPower)
func (_Master *MasterFilterer) ParseSetMinInclusionVotingPower(log types.Log) (*MasterSetMinInclusionVotingPower, error) {
	event := new(MasterSetMinInclusionVotingPower)
	if err := _Master.contract.UnpackLog(event, "SetMinInclusionVotingPower", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetProlongDurationIterator is returned from FilterSetProlongDuration and is used to iterate over the raw logs and unpacked data for SetProlongDuration events raised by the Master contract.
type MasterSetProlongDurationIterator struct {
	Event *MasterSetProlongDuration // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetProlongDurationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetProlongDuration)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetProlongDuration)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetProlongDurationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetProlongDurationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetProlongDuration represents a SetProlongDuration event raised by the Master contract.
type MasterSetProlongDuration struct {
	ProlongDuration *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSetProlongDuration is a free log retrieval operation binding the contract event 0xc0c12bfa4224ca9e540947f48e95231ff35184a4108680b0e7c0c379554f0ae9.
//
// Solidity: event SetProlongDuration(uint48 prolongDuration)
func (_Master *MasterFilterer) FilterSetProlongDuration(opts *bind.FilterOpts) (*MasterSetProlongDurationIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetProlongDuration")
	if err != nil {
		return nil, err
	}
	return &MasterSetProlongDurationIterator{contract: _Master.contract, event: "SetProlongDuration", logs: logs, sub: sub}, nil
}

// WatchSetProlongDuration is a free log subscription operation binding the contract event 0xc0c12bfa4224ca9e540947f48e95231ff35184a4108680b0e7c0c379554f0ae9.
//
// Solidity: event SetProlongDuration(uint48 prolongDuration)
func (_Master *MasterFilterer) WatchSetProlongDuration(opts *bind.WatchOpts, sink chan<- *MasterSetProlongDuration) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetProlongDuration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetProlongDuration)
				if err := _Master.contract.UnpackLog(event, "SetProlongDuration", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetProlongDuration is a log parse operation binding the contract event 0xc0c12bfa4224ca9e540947f48e95231ff35184a4108680b0e7c0c379554f0ae9.
//
// Solidity: event SetProlongDuration(uint48 prolongDuration)
func (_Master *MasterFilterer) ParseSetProlongDuration(log types.Log) (*MasterSetProlongDuration, error) {
	event := new(MasterSetProlongDuration)
	if err := _Master.contract.UnpackLog(event, "SetProlongDuration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetRequiredKeyTagIterator is returned from FilterSetRequiredKeyTag and is used to iterate over the raw logs and unpacked data for SetRequiredKeyTag events raised by the Master contract.
type MasterSetRequiredKeyTagIterator struct {
	Event *MasterSetRequiredKeyTag // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetRequiredKeyTagIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetRequiredKeyTag)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetRequiredKeyTag)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetRequiredKeyTagIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetRequiredKeyTagIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetRequiredKeyTag represents a SetRequiredKeyTag event raised by the Master contract.
type MasterSetRequiredKeyTag struct {
	RequiredKeyTag uint8
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetRequiredKeyTag is a free log retrieval operation binding the contract event 0x0c2213b81cfe15af86bbbff49cc144fa8f808a3733295605e9850388a5793a14.
//
// Solidity: event SetRequiredKeyTag(uint8 requiredKeyTag)
func (_Master *MasterFilterer) FilterSetRequiredKeyTag(opts *bind.FilterOpts) (*MasterSetRequiredKeyTagIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetRequiredKeyTag")
	if err != nil {
		return nil, err
	}
	return &MasterSetRequiredKeyTagIterator{contract: _Master.contract, event: "SetRequiredKeyTag", logs: logs, sub: sub}, nil
}

// WatchSetRequiredKeyTag is a free log subscription operation binding the contract event 0x0c2213b81cfe15af86bbbff49cc144fa8f808a3733295605e9850388a5793a14.
//
// Solidity: event SetRequiredKeyTag(uint8 requiredKeyTag)
func (_Master *MasterFilterer) WatchSetRequiredKeyTag(opts *bind.WatchOpts, sink chan<- *MasterSetRequiredKeyTag) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetRequiredKeyTag")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetRequiredKeyTag)
				if err := _Master.contract.UnpackLog(event, "SetRequiredKeyTag", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetRequiredKeyTag is a log parse operation binding the contract event 0x0c2213b81cfe15af86bbbff49cc144fa8f808a3733295605e9850388a5793a14.
//
// Solidity: event SetRequiredKeyTag(uint8 requiredKeyTag)
func (_Master *MasterFilterer) ParseSetRequiredKeyTag(log types.Log) (*MasterSetRequiredKeyTag, error) {
	event := new(MasterSetRequiredKeyTag)
	if err := _Master.contract.UnpackLog(event, "SetRequiredKeyTag", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetRequiredKeyTagsIterator is returned from FilterSetRequiredKeyTags and is used to iterate over the raw logs and unpacked data for SetRequiredKeyTags events raised by the Master contract.
type MasterSetRequiredKeyTagsIterator struct {
	Event *MasterSetRequiredKeyTags // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetRequiredKeyTagsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetRequiredKeyTags)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetRequiredKeyTags)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetRequiredKeyTagsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetRequiredKeyTagsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetRequiredKeyTags represents a SetRequiredKeyTags event raised by the Master contract.
type MasterSetRequiredKeyTags struct {
	RequiredKeyTags []uint8
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSetRequiredKeyTags is a free log retrieval operation binding the contract event 0x14f8998266f37e593027a05efebf63b8710681d1cdbd39e6d7a156ff7e1485cd.
//
// Solidity: event SetRequiredKeyTags(uint8[] requiredKeyTags)
func (_Master *MasterFilterer) FilterSetRequiredKeyTags(opts *bind.FilterOpts) (*MasterSetRequiredKeyTagsIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetRequiredKeyTags")
	if err != nil {
		return nil, err
	}
	return &MasterSetRequiredKeyTagsIterator{contract: _Master.contract, event: "SetRequiredKeyTags", logs: logs, sub: sub}, nil
}

// WatchSetRequiredKeyTags is a free log subscription operation binding the contract event 0x14f8998266f37e593027a05efebf63b8710681d1cdbd39e6d7a156ff7e1485cd.
//
// Solidity: event SetRequiredKeyTags(uint8[] requiredKeyTags)
func (_Master *MasterFilterer) WatchSetRequiredKeyTags(opts *bind.WatchOpts, sink chan<- *MasterSetRequiredKeyTags) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetRequiredKeyTags")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetRequiredKeyTags)
				if err := _Master.contract.UnpackLog(event, "SetRequiredKeyTags", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetRequiredKeyTags is a log parse operation binding the contract event 0x14f8998266f37e593027a05efebf63b8710681d1cdbd39e6d7a156ff7e1485cd.
//
// Solidity: event SetRequiredKeyTags(uint8[] requiredKeyTags)
func (_Master *MasterFilterer) ParseSetRequiredKeyTags(log types.Log) (*MasterSetRequiredKeyTags, error) {
	event := new(MasterSetRequiredKeyTags)
	if err := _Master.contract.UnpackLog(event, "SetRequiredKeyTags", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetSelectorRoleIterator is returned from FilterSetSelectorRole and is used to iterate over the raw logs and unpacked data for SetSelectorRole events raised by the Master contract.
type MasterSetSelectorRoleIterator struct {
	Event *MasterSetSelectorRole // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetSelectorRoleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetSelectorRole)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetSelectorRole)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetSelectorRoleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetSelectorRoleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetSelectorRole represents a SetSelectorRole event raised by the Master contract.
type MasterSetSelectorRole struct {
	Selector [4]byte
	Role     [32]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSetSelectorRole is a free log retrieval operation binding the contract event 0x205ddee47edfee0f39b93f29e45a801cd7c9cffe0ca9a2da19e547227b2a0504.
//
// Solidity: event SetSelectorRole(bytes4 indexed selector, bytes32 indexed role)
func (_Master *MasterFilterer) FilterSetSelectorRole(opts *bind.FilterOpts, selector [][4]byte, role [][32]byte) (*MasterSetSelectorRoleIterator, error) {

	var selectorRule []interface{}
	for _, selectorItem := range selector {
		selectorRule = append(selectorRule, selectorItem)
	}
	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetSelectorRole", selectorRule, roleRule)
	if err != nil {
		return nil, err
	}
	return &MasterSetSelectorRoleIterator{contract: _Master.contract, event: "SetSelectorRole", logs: logs, sub: sub}, nil
}

// WatchSetSelectorRole is a free log subscription operation binding the contract event 0x205ddee47edfee0f39b93f29e45a801cd7c9cffe0ca9a2da19e547227b2a0504.
//
// Solidity: event SetSelectorRole(bytes4 indexed selector, bytes32 indexed role)
func (_Master *MasterFilterer) WatchSetSelectorRole(opts *bind.WatchOpts, sink chan<- *MasterSetSelectorRole, selector [][4]byte, role [][32]byte) (event.Subscription, error) {

	var selectorRule []interface{}
	for _, selectorItem := range selector {
		selectorRule = append(selectorRule, selectorItem)
	}
	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetSelectorRole", selectorRule, roleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetSelectorRole)
				if err := _Master.contract.UnpackLog(event, "SetSelectorRole", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetSelectorRole is a log parse operation binding the contract event 0x205ddee47edfee0f39b93f29e45a801cd7c9cffe0ca9a2da19e547227b2a0504.
//
// Solidity: event SetSelectorRole(bytes4 indexed selector, bytes32 indexed role)
func (_Master *MasterFilterer) ParseSetSelectorRole(log types.Log) (*MasterSetSelectorRole, error) {
	event := new(MasterSetSelectorRole)
	if err := _Master.contract.UnpackLog(event, "SetSelectorRole", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetSigVerifierIterator is returned from FilterSetSigVerifier and is used to iterate over the raw logs and unpacked data for SetSigVerifier events raised by the Master contract.
type MasterSetSigVerifierIterator struct {
	Event *MasterSetSigVerifier // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetSigVerifierIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetSigVerifier)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetSigVerifier)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetSigVerifierIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetSigVerifierIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetSigVerifier represents a SetSigVerifier event raised by the Master contract.
type MasterSetSigVerifier struct {
	SigVerifier common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSetSigVerifier is a free log retrieval operation binding the contract event 0x3cb2fcfd41e182e933eb967bdeaac4f8ff69c80b6fd24fea9561dfbdec127942.
//
// Solidity: event SetSigVerifier(address sigVerifier)
func (_Master *MasterFilterer) FilterSetSigVerifier(opts *bind.FilterOpts) (*MasterSetSigVerifierIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetSigVerifier")
	if err != nil {
		return nil, err
	}
	return &MasterSetSigVerifierIterator{contract: _Master.contract, event: "SetSigVerifier", logs: logs, sub: sub}, nil
}

// WatchSetSigVerifier is a free log subscription operation binding the contract event 0x3cb2fcfd41e182e933eb967bdeaac4f8ff69c80b6fd24fea9561dfbdec127942.
//
// Solidity: event SetSigVerifier(address sigVerifier)
func (_Master *MasterFilterer) WatchSetSigVerifier(opts *bind.WatchOpts, sink chan<- *MasterSetSigVerifier) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetSigVerifier")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetSigVerifier)
				if err := _Master.contract.UnpackLog(event, "SetSigVerifier", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetSigVerifier is a log parse operation binding the contract event 0x3cb2fcfd41e182e933eb967bdeaac4f8ff69c80b6fd24fea9561dfbdec127942.
//
// Solidity: event SetSigVerifier(address sigVerifier)
func (_Master *MasterFilterer) ParseSetSigVerifier(log types.Log) (*MasterSetSigVerifier, error) {
	event := new(MasterSetSigVerifier)
	if err := _Master.contract.UnpackLog(event, "SetSigVerifier", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MasterSetVerificationTypeIterator is returned from FilterSetVerificationType and is used to iterate over the raw logs and unpacked data for SetVerificationType events raised by the Master contract.
type MasterSetVerificationTypeIterator struct {
	Event *MasterSetVerificationType // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MasterSetVerificationTypeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSetVerificationType)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MasterSetVerificationType)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MasterSetVerificationTypeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSetVerificationTypeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSetVerificationType represents a SetVerificationType event raised by the Master contract.
type MasterSetVerificationType struct {
	VerificationType uint32
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterSetVerificationType is a free log retrieval operation binding the contract event 0x2acc7be3ff5df4b911488f72502071dcf3f4a8f778a8abc351af3220bcd15b7f.
//
// Solidity: event SetVerificationType(uint32 verificationType)
func (_Master *MasterFilterer) FilterSetVerificationType(opts *bind.FilterOpts) (*MasterSetVerificationTypeIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SetVerificationType")
	if err != nil {
		return nil, err
	}
	return &MasterSetVerificationTypeIterator{contract: _Master.contract, event: "SetVerificationType", logs: logs, sub: sub}, nil
}

// WatchSetVerificationType is a free log subscription operation binding the contract event 0x2acc7be3ff5df4b911488f72502071dcf3f4a8f778a8abc351af3220bcd15b7f.
//
// Solidity: event SetVerificationType(uint32 verificationType)
func (_Master *MasterFilterer) WatchSetVerificationType(opts *bind.WatchOpts, sink chan<- *MasterSetVerificationType) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SetVerificationType")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSetVerificationType)
				if err := _Master.contract.UnpackLog(event, "SetVerificationType", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetVerificationType is a log parse operation binding the contract event 0x2acc7be3ff5df4b911488f72502071dcf3f4a8f778a8abc351af3220bcd15b7f.
//
// Solidity: event SetVerificationType(uint32 verificationType)
func (_Master *MasterFilterer) ParseSetVerificationType(log types.Log) (*MasterSetVerificationType, error) {
	event := new(MasterSetVerificationType)
	if err := _Master.contract.UnpackLog(event, "SetVerificationType", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
