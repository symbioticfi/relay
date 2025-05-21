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

// IBaseKeyManagerKey is an auto generated low-level Go binding around an user-defined struct.
type IBaseKeyManagerKey struct {
	Tag     uint8
	Payload []byte
}

// IEpochManagerEpochManagerInitParams is an auto generated low-level Go binding around an user-defined struct.
type IEpochManagerEpochManagerInitParams struct {
	EpochDuration          *big.Int
	EpochDurationTimestamp *big.Int
}

// IMasterConfigProviderCrossChainAddress is an auto generated low-level Go binding around an user-defined struct.
type IMasterConfigProviderCrossChainAddress struct {
	Addr    common.Address
	ChainId uint64
}

// IMasterConfigProviderMasterConfig is an auto generated low-level Go binding around an user-defined struct.
type IMasterConfigProviderMasterConfig struct {
	VotingPowerProviders []IMasterConfigProviderCrossChainAddress
	KeysProvider         IMasterConfigProviderCrossChainAddress
	Replicas             []IMasterConfigProviderCrossChainAddress
}

// IMasterConfigProviderMasterConfigProviderInitParams is an auto generated low-level Go binding around an user-defined struct.
type IMasterConfigProviderMasterConfigProviderInitParams struct {
	VotingPowerProviders []IMasterConfigProviderCrossChainAddress
	KeysProvider         IMasterConfigProviderCrossChainAddress
	Replicas             []IMasterConfigProviderCrossChainAddress
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

// ISettlementQuorumThreshold is an auto generated low-level Go binding around an user-defined struct.
type ISettlementQuorumThreshold struct {
	KeyTag    uint8
	Threshold *big.Int
}

// ISettlementSettlementInitParams is an auto generated low-level Go binding around an user-defined struct.
type ISettlementSettlementInitParams struct {
	NetworkManagerInitParams INetworkManagerNetworkManagerInitParams
	EpochManagerInitParams   IEpochManagerEpochManagerInitParams
	OzEip712InitParams       IOzEIP712OzEIP712InitParams
	QuorumThresholds         []ISettlementQuorumThreshold
	CommitDuration           *big.Int
	RequiredKeyTag           uint8
	SigVerifier              common.Address
}

// ISettlementValSetHeader is an auto generated low-level Go binding around an user-defined struct.
type ISettlementValSetHeader struct {
	Version                uint8
	ActiveAggregatedKeys   []IBaseKeyManagerKey
	TotalActiveVotingPower *big.Int
	ValidatorsSszMRoot     [32]byte
	ExtraData              []byte
}

// IValSetConfigProviderValSetConfig is an auto generated low-level Go binding around an user-defined struct.
type IValSetConfigProviderValSetConfig struct {
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
}

// IValSetConfigProviderValSetConfigProviderInitParams is an auto generated low-level Go binding around an user-defined struct.
type IValSetConfigProviderValSetConfigProviderInitParams struct {
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
}

// MasterMetaData contains all meta data concerning the Master contract.
var MasterMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"EpochManager_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"MasterConfigProvider_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"NETWORK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"NetworkManager_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"OzAccessControl_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OzEIP712_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PermissionManager_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SUBNETWORK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SUBNETWORK_IDENTIFIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint96\",\"internalType\":\"uint96\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"Settlement_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"VALIDATOR_SET_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"ValSetConfigProvider_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"addReplica\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addVotingPowerProvider\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"commitValSetHeader\",\"inputs\":[{\"name\":\"header\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"activeAggregatedKeys\",\"type\":\"tuple[]\",\"internalType\":\"structIBaseKeyManager.Key[]\",\"components\":[{\"name\":\"tag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"totalActiveVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getActiveAggregatedKeyFromValSetHeader\",\"inputs\":[{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getActiveAggregatedKeyFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getActiveReplicas\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getActiveReplicasAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getActiveVotingPowerProviders\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getActiveVotingPowerProvidersAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCaptureTimestamp\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCommitDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCommitDurationAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpochDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpochStart\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentPhase\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumISettlement.ValSetPhase\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentValSetEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentValSetTimestamp\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochDuration\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochIndex\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochStart\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExtraDataFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExtraDataFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getKeysProvider\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getKeysProviderAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMasterConfig\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.MasterConfig\",\"components\":[{\"name\":\"votingPowerProviders\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"replicas\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMasterConfigAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.MasterConfig\",\"components\":[{\"name\":\"votingPowerProviders\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"replicas\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxValidatorsCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxValidatorsCountAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxVotingPower\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMaxVotingPowerAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMinInclusionVotingPower\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMinInclusionVotingPowerAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNextEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNextEpochStart\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getQuorumThreshold\",\"inputs\":[{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint208\",\"internalType\":\"uint208\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getQuorumThresholdAt\",\"inputs\":[{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint208\",\"internalType\":\"uint208\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTag\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTagAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTags\",\"inputs\":[],\"outputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRequiredKeyTagsAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRole\",\"inputs\":[{\"name\":\"selector\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSigVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSigVerifierAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTotalActiveVotingPowerFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTotalActiveVotingPowerFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetConfig\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIValSetConfigProvider.ValSetConfig\",\"components\":[{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxValidatorsCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetConfigAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIValSetConfigProvider.ValSetConfig\",\"components\":[{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxValidatorsCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"header\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"activeAggregatedKeys\",\"type\":\"tuple[]\",\"internalType\":\"structIBaseKeyManager.Key[]\",\"components\":[{\"name\":\"tag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"totalActiveVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"activeAggregatedKeys\",\"type\":\"tuple[]\",\"internalType\":\"structIBaseKeyManager.Key[]\",\"components\":[{\"name\":\"tag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"totalActiveVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValidatorsSszMRootFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getValidatorsSszMRootFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVersionFromValSetHeader\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVersionFromValSetHeaderAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4CrossChain\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"settlementInitParams\",\"type\":\"tuple\",\"internalType\":\"structISettlement.SettlementInitParams\",\"components\":[{\"name\":\"networkManagerInitParams\",\"type\":\"tuple\",\"internalType\":\"structINetworkManager.NetworkManagerInitParams\",\"components\":[{\"name\":\"network\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"subnetworkID\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]},{\"name\":\"epochManagerInitParams\",\"type\":\"tuple\",\"internalType\":\"structIEpochManager.EpochManagerInitParams\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"epochDurationTimestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]},{\"name\":\"ozEip712InitParams\",\"type\":\"tuple\",\"internalType\":\"structIOzEIP712.OzEIP712InitParams\",\"components\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"name\":\"quorumThresholds\",\"type\":\"tuple[]\",\"internalType\":\"structISettlement.QuorumThreshold[]\",\"components\":[{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"threshold\",\"type\":\"uint208\",\"internalType\":\"uint208\"}]},{\"name\":\"commitDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"sigVerifier\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"name\":\"valSetConfigProviderInitParams\",\"type\":\"tuple\",\"internalType\":\"structIValSetConfigProvider.ValSetConfigProviderInitParams\",\"components\":[{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxValidatorsCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}]},{\"name\":\"masterConfigProviderInitParams\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.MasterConfigProviderInitParams\",\"components\":[{\"name\":\"votingPowerProviders\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"replicas\",\"type\":\"tuple[]\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}]},{\"name\":\"defaultAdmin\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isReplicaActive\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isReplicaActiveAt\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isValSetHeaderSubmitted\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isValSetHeaderSubmittedAt\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isVotingPowerProviderActive\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isVotingPowerProviderActiveAt\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"multicall\",\"inputs\":[{\"name\":\"data\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"results\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeReplica\",\"inputs\":[{\"name\":\"replica\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeVotingPowerProvider\",\"inputs\":[{\"name\":\"votingPowerProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setCommitDuration\",\"inputs\":[{\"name\":\"commitDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setEpochDuration\",\"inputs\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGenesis\",\"inputs\":[{\"name\":\"valSetHeader\",\"type\":\"tuple\",\"internalType\":\"structISettlement.ValSetHeader\",\"components\":[{\"name\":\"version\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"activeAggregatedKeys\",\"type\":\"tuple[]\",\"internalType\":\"structIBaseKeyManager.Key[]\",\"components\":[{\"name\":\"tag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"totalActiveVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"validatorsSszMRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setKeysProvider\",\"inputs\":[{\"name\":\"keysProvider\",\"type\":\"tuple\",\"internalType\":\"structIMasterConfigProvider.CrossChainAddress\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxValidatorsCount\",\"inputs\":[{\"name\":\"maxValidatorsCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxVotingPower\",\"inputs\":[{\"name\":\"maxVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinInclusionVotingPower\",\"inputs\":[{\"name\":\"minInclusionVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setQuorumThreshold\",\"inputs\":[{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"quorumThreshold\",\"type\":\"uint208\",\"internalType\":\"uint208\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRequiredKeyTag\",\"inputs\":[{\"name\":\"requiredKeyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRequiredKeyTags\",\"inputs\":[{\"name\":\"requiredKeyTags\",\"type\":\"uint8[]\",\"internalType\":\"uint8[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSigVerifier\",\"inputs\":[{\"name\":\"sigVerifier\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"staticDelegateCall\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifyQuorumSig\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"keyTag\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"quorumThreshold\",\"type\":\"uint208\",\"internalType\":\"uint208\"},{\"name\":\"proof\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SelectorRoleSet\",\"inputs\":[{\"name\":\"selector\",\"type\":\"bytes4\",\"indexed\":false,\"internalType\":\"bytes4\"},{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"EpochManager_InvalidEpochDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_InvalidEpochDurationIndex\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_InvalidEpochDurationTimestamp\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager_NoCheckpoint\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MasterConfigProvider_AlreadyAdded\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MasterConfigProvider_NotAdded\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_Duplicate\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_EpochDurationTooShort\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidPhase\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_InvalidVersion\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Settlement_VerificationFailed\",\"inputs\":[]}]",
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

// MasterConfigProviderVERSION is a free data retrieval call binding the contract method 0xd1c5a7c4.
//
// Solidity: function MasterConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterCaller) MasterConfigProviderVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "MasterConfigProvider_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MasterConfigProviderVERSION is a free data retrieval call binding the contract method 0xd1c5a7c4.
//
// Solidity: function MasterConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterSession) MasterConfigProviderVERSION() (uint64, error) {
	return _Master.Contract.MasterConfigProviderVERSION(&_Master.CallOpts)
}

// MasterConfigProviderVERSION is a free data retrieval call binding the contract method 0xd1c5a7c4.
//
// Solidity: function MasterConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterCallerSession) MasterConfigProviderVERSION() (uint64, error) {
	return _Master.Contract.MasterConfigProviderVERSION(&_Master.CallOpts)
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

// ValSetConfigProviderVERSION is a free data retrieval call binding the contract method 0xecaad2a1.
//
// Solidity: function ValSetConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterCaller) ValSetConfigProviderVERSION(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "ValSetConfigProvider_VERSION")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// ValSetConfigProviderVERSION is a free data retrieval call binding the contract method 0xecaad2a1.
//
// Solidity: function ValSetConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterSession) ValSetConfigProviderVERSION() (uint64, error) {
	return _Master.Contract.ValSetConfigProviderVERSION(&_Master.CallOpts)
}

// ValSetConfigProviderVERSION is a free data retrieval call binding the contract method 0xecaad2a1.
//
// Solidity: function ValSetConfigProvider_VERSION() pure returns(uint64)
func (_Master *MasterCallerSession) ValSetConfigProviderVERSION() (uint64, error) {
	return _Master.Contract.ValSetConfigProviderVERSION(&_Master.CallOpts)
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

// GetActiveAggregatedKeyFromValSetHeader is a free data retrieval call binding the contract method 0x521fbb67.
//
// Solidity: function getActiveAggregatedKeyFromValSetHeader(uint8 keyTag) view returns(bytes)
func (_Master *MasterCaller) GetActiveAggregatedKeyFromValSetHeader(opts *bind.CallOpts, keyTag uint8) ([]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getActiveAggregatedKeyFromValSetHeader", keyTag)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetActiveAggregatedKeyFromValSetHeader is a free data retrieval call binding the contract method 0x521fbb67.
//
// Solidity: function getActiveAggregatedKeyFromValSetHeader(uint8 keyTag) view returns(bytes)
func (_Master *MasterSession) GetActiveAggregatedKeyFromValSetHeader(keyTag uint8) ([]byte, error) {
	return _Master.Contract.GetActiveAggregatedKeyFromValSetHeader(&_Master.CallOpts, keyTag)
}

// GetActiveAggregatedKeyFromValSetHeader is a free data retrieval call binding the contract method 0x521fbb67.
//
// Solidity: function getActiveAggregatedKeyFromValSetHeader(uint8 keyTag) view returns(bytes)
func (_Master *MasterCallerSession) GetActiveAggregatedKeyFromValSetHeader(keyTag uint8) ([]byte, error) {
	return _Master.Contract.GetActiveAggregatedKeyFromValSetHeader(&_Master.CallOpts, keyTag)
}

// GetActiveAggregatedKeyFromValSetHeaderAt is a free data retrieval call binding the contract method 0xa7f291bd.
//
// Solidity: function getActiveAggregatedKeyFromValSetHeaderAt(uint48 epoch, uint8 keyTag) view returns(bytes)
func (_Master *MasterCaller) GetActiveAggregatedKeyFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int, keyTag uint8) ([]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getActiveAggregatedKeyFromValSetHeaderAt", epoch, keyTag)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetActiveAggregatedKeyFromValSetHeaderAt is a free data retrieval call binding the contract method 0xa7f291bd.
//
// Solidity: function getActiveAggregatedKeyFromValSetHeaderAt(uint48 epoch, uint8 keyTag) view returns(bytes)
func (_Master *MasterSession) GetActiveAggregatedKeyFromValSetHeaderAt(epoch *big.Int, keyTag uint8) ([]byte, error) {
	return _Master.Contract.GetActiveAggregatedKeyFromValSetHeaderAt(&_Master.CallOpts, epoch, keyTag)
}

// GetActiveAggregatedKeyFromValSetHeaderAt is a free data retrieval call binding the contract method 0xa7f291bd.
//
// Solidity: function getActiveAggregatedKeyFromValSetHeaderAt(uint48 epoch, uint8 keyTag) view returns(bytes)
func (_Master *MasterCallerSession) GetActiveAggregatedKeyFromValSetHeaderAt(epoch *big.Int, keyTag uint8) ([]byte, error) {
	return _Master.Contract.GetActiveAggregatedKeyFromValSetHeaderAt(&_Master.CallOpts, epoch, keyTag)
}

// GetActiveReplicas is a free data retrieval call binding the contract method 0x5cc43174.
//
// Solidity: function getActiveReplicas() view returns((address,uint64)[])
func (_Master *MasterCaller) GetActiveReplicas(opts *bind.CallOpts) ([]IMasterConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getActiveReplicas")

	if err != nil {
		return *new([]IMasterConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IMasterConfigProviderCrossChainAddress)).(*[]IMasterConfigProviderCrossChainAddress)

	return out0, err

}

// GetActiveReplicas is a free data retrieval call binding the contract method 0x5cc43174.
//
// Solidity: function getActiveReplicas() view returns((address,uint64)[])
func (_Master *MasterSession) GetActiveReplicas() ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveReplicas(&_Master.CallOpts)
}

// GetActiveReplicas is a free data retrieval call binding the contract method 0x5cc43174.
//
// Solidity: function getActiveReplicas() view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetActiveReplicas() ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveReplicas(&_Master.CallOpts)
}

// GetActiveReplicasAt is a free data retrieval call binding the contract method 0x02eb1000.
//
// Solidity: function getActiveReplicasAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCaller) GetActiveReplicasAt(opts *bind.CallOpts, timestamp *big.Int, hints [][]byte) ([]IMasterConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getActiveReplicasAt", timestamp, hints)

	if err != nil {
		return *new([]IMasterConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IMasterConfigProviderCrossChainAddress)).(*[]IMasterConfigProviderCrossChainAddress)

	return out0, err

}

// GetActiveReplicasAt is a free data retrieval call binding the contract method 0x02eb1000.
//
// Solidity: function getActiveReplicasAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterSession) GetActiveReplicasAt(timestamp *big.Int, hints [][]byte) ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveReplicasAt(&_Master.CallOpts, timestamp, hints)
}

// GetActiveReplicasAt is a free data retrieval call binding the contract method 0x02eb1000.
//
// Solidity: function getActiveReplicasAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetActiveReplicasAt(timestamp *big.Int, hints [][]byte) ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveReplicasAt(&_Master.CallOpts, timestamp, hints)
}

// GetActiveVotingPowerProviders is a free data retrieval call binding the contract method 0xfda10c48.
//
// Solidity: function getActiveVotingPowerProviders() view returns((address,uint64)[])
func (_Master *MasterCaller) GetActiveVotingPowerProviders(opts *bind.CallOpts) ([]IMasterConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getActiveVotingPowerProviders")

	if err != nil {
		return *new([]IMasterConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IMasterConfigProviderCrossChainAddress)).(*[]IMasterConfigProviderCrossChainAddress)

	return out0, err

}

// GetActiveVotingPowerProviders is a free data retrieval call binding the contract method 0xfda10c48.
//
// Solidity: function getActiveVotingPowerProviders() view returns((address,uint64)[])
func (_Master *MasterSession) GetActiveVotingPowerProviders() ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveVotingPowerProviders(&_Master.CallOpts)
}

// GetActiveVotingPowerProviders is a free data retrieval call binding the contract method 0xfda10c48.
//
// Solidity: function getActiveVotingPowerProviders() view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetActiveVotingPowerProviders() ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveVotingPowerProviders(&_Master.CallOpts)
}

// GetActiveVotingPowerProvidersAt is a free data retrieval call binding the contract method 0xd74bc508.
//
// Solidity: function getActiveVotingPowerProvidersAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCaller) GetActiveVotingPowerProvidersAt(opts *bind.CallOpts, timestamp *big.Int, hints [][]byte) ([]IMasterConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getActiveVotingPowerProvidersAt", timestamp, hints)

	if err != nil {
		return *new([]IMasterConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new([]IMasterConfigProviderCrossChainAddress)).(*[]IMasterConfigProviderCrossChainAddress)

	return out0, err

}

// GetActiveVotingPowerProvidersAt is a free data retrieval call binding the contract method 0xd74bc508.
//
// Solidity: function getActiveVotingPowerProvidersAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterSession) GetActiveVotingPowerProvidersAt(timestamp *big.Int, hints [][]byte) ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveVotingPowerProvidersAt(&_Master.CallOpts, timestamp, hints)
}

// GetActiveVotingPowerProvidersAt is a free data retrieval call binding the contract method 0xd74bc508.
//
// Solidity: function getActiveVotingPowerProvidersAt(uint48 timestamp, bytes[] hints) view returns((address,uint64)[])
func (_Master *MasterCallerSession) GetActiveVotingPowerProvidersAt(timestamp *big.Int, hints [][]byte) ([]IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetActiveVotingPowerProvidersAt(&_Master.CallOpts, timestamp, hints)
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
// Solidity: function getCommitDurationAt(uint48 epoch, bytes hint) view returns(uint48)
func (_Master *MasterCaller) GetCommitDurationAt(opts *bind.CallOpts, epoch *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getCommitDurationAt", epoch, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommitDurationAt is a free data retrieval call binding the contract method 0x8f7a22c6.
//
// Solidity: function getCommitDurationAt(uint48 epoch, bytes hint) view returns(uint48)
func (_Master *MasterSession) GetCommitDurationAt(epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetCommitDurationAt(&_Master.CallOpts, epoch, hint)
}

// GetCommitDurationAt is a free data retrieval call binding the contract method 0x8f7a22c6.
//
// Solidity: function getCommitDurationAt(uint48 epoch, bytes hint) view returns(uint48)
func (_Master *MasterCallerSession) GetCommitDurationAt(epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetCommitDurationAt(&_Master.CallOpts, epoch, hint)
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

// GetExtraDataFromValSetHeader is a free data retrieval call binding the contract method 0x756e2b11.
//
// Solidity: function getExtraDataFromValSetHeader() view returns(bytes)
func (_Master *MasterCaller) GetExtraDataFromValSetHeader(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getExtraDataFromValSetHeader")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetExtraDataFromValSetHeader is a free data retrieval call binding the contract method 0x756e2b11.
//
// Solidity: function getExtraDataFromValSetHeader() view returns(bytes)
func (_Master *MasterSession) GetExtraDataFromValSetHeader() ([]byte, error) {
	return _Master.Contract.GetExtraDataFromValSetHeader(&_Master.CallOpts)
}

// GetExtraDataFromValSetHeader is a free data retrieval call binding the contract method 0x756e2b11.
//
// Solidity: function getExtraDataFromValSetHeader() view returns(bytes)
func (_Master *MasterCallerSession) GetExtraDataFromValSetHeader() ([]byte, error) {
	return _Master.Contract.GetExtraDataFromValSetHeader(&_Master.CallOpts)
}

// GetExtraDataFromValSetHeaderAt is a free data retrieval call binding the contract method 0xf30bbaf2.
//
// Solidity: function getExtraDataFromValSetHeaderAt(uint48 epoch) view returns(bytes)
func (_Master *MasterCaller) GetExtraDataFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) ([]byte, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getExtraDataFromValSetHeaderAt", epoch)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetExtraDataFromValSetHeaderAt is a free data retrieval call binding the contract method 0xf30bbaf2.
//
// Solidity: function getExtraDataFromValSetHeaderAt(uint48 epoch) view returns(bytes)
func (_Master *MasterSession) GetExtraDataFromValSetHeaderAt(epoch *big.Int) ([]byte, error) {
	return _Master.Contract.GetExtraDataFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetExtraDataFromValSetHeaderAt is a free data retrieval call binding the contract method 0xf30bbaf2.
//
// Solidity: function getExtraDataFromValSetHeaderAt(uint48 epoch) view returns(bytes)
func (_Master *MasterCallerSession) GetExtraDataFromValSetHeaderAt(epoch *big.Int) ([]byte, error) {
	return _Master.Contract.GetExtraDataFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetKeysProvider is a free data retrieval call binding the contract method 0x297d29b8.
//
// Solidity: function getKeysProvider() view returns((address,uint64))
func (_Master *MasterCaller) GetKeysProvider(opts *bind.CallOpts) (IMasterConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getKeysProvider")

	if err != nil {
		return *new(IMasterConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new(IMasterConfigProviderCrossChainAddress)).(*IMasterConfigProviderCrossChainAddress)

	return out0, err

}

// GetKeysProvider is a free data retrieval call binding the contract method 0x297d29b8.
//
// Solidity: function getKeysProvider() view returns((address,uint64))
func (_Master *MasterSession) GetKeysProvider() (IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProvider(&_Master.CallOpts)
}

// GetKeysProvider is a free data retrieval call binding the contract method 0x297d29b8.
//
// Solidity: function getKeysProvider() view returns((address,uint64))
func (_Master *MasterCallerSession) GetKeysProvider() (IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProvider(&_Master.CallOpts)
}

// GetKeysProviderAt is a free data retrieval call binding the contract method 0xb405b818.
//
// Solidity: function getKeysProviderAt(uint48 timestamp, bytes hint) view returns((address,uint64))
func (_Master *MasterCaller) GetKeysProviderAt(opts *bind.CallOpts, timestamp *big.Int, hint []byte) (IMasterConfigProviderCrossChainAddress, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getKeysProviderAt", timestamp, hint)

	if err != nil {
		return *new(IMasterConfigProviderCrossChainAddress), err
	}

	out0 := *abi.ConvertType(out[0], new(IMasterConfigProviderCrossChainAddress)).(*IMasterConfigProviderCrossChainAddress)

	return out0, err

}

// GetKeysProviderAt is a free data retrieval call binding the contract method 0xb405b818.
//
// Solidity: function getKeysProviderAt(uint48 timestamp, bytes hint) view returns((address,uint64))
func (_Master *MasterSession) GetKeysProviderAt(timestamp *big.Int, hint []byte) (IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProviderAt(&_Master.CallOpts, timestamp, hint)
}

// GetKeysProviderAt is a free data retrieval call binding the contract method 0xb405b818.
//
// Solidity: function getKeysProviderAt(uint48 timestamp, bytes hint) view returns((address,uint64))
func (_Master *MasterCallerSession) GetKeysProviderAt(timestamp *big.Int, hint []byte) (IMasterConfigProviderCrossChainAddress, error) {
	return _Master.Contract.GetKeysProviderAt(&_Master.CallOpts, timestamp, hint)
}

// GetMasterConfig is a free data retrieval call binding the contract method 0x063c0c5c.
//
// Solidity: function getMasterConfig() view returns(((address,uint64)[],(address,uint64),(address,uint64)[]))
func (_Master *MasterCaller) GetMasterConfig(opts *bind.CallOpts) (IMasterConfigProviderMasterConfig, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMasterConfig")

	if err != nil {
		return *new(IMasterConfigProviderMasterConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IMasterConfigProviderMasterConfig)).(*IMasterConfigProviderMasterConfig)

	return out0, err

}

// GetMasterConfig is a free data retrieval call binding the contract method 0x063c0c5c.
//
// Solidity: function getMasterConfig() view returns(((address,uint64)[],(address,uint64),(address,uint64)[]))
func (_Master *MasterSession) GetMasterConfig() (IMasterConfigProviderMasterConfig, error) {
	return _Master.Contract.GetMasterConfig(&_Master.CallOpts)
}

// GetMasterConfig is a free data retrieval call binding the contract method 0x063c0c5c.
//
// Solidity: function getMasterConfig() view returns(((address,uint64)[],(address,uint64),(address,uint64)[]))
func (_Master *MasterCallerSession) GetMasterConfig() (IMasterConfigProviderMasterConfig, error) {
	return _Master.Contract.GetMasterConfig(&_Master.CallOpts)
}

// GetMasterConfigAt is a free data retrieval call binding the contract method 0x9288968d.
//
// Solidity: function getMasterConfigAt(uint48 timestamp, bytes hints) view returns(((address,uint64)[],(address,uint64),(address,uint64)[]))
func (_Master *MasterCaller) GetMasterConfigAt(opts *bind.CallOpts, timestamp *big.Int, hints []byte) (IMasterConfigProviderMasterConfig, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getMasterConfigAt", timestamp, hints)

	if err != nil {
		return *new(IMasterConfigProviderMasterConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IMasterConfigProviderMasterConfig)).(*IMasterConfigProviderMasterConfig)

	return out0, err

}

// GetMasterConfigAt is a free data retrieval call binding the contract method 0x9288968d.
//
// Solidity: function getMasterConfigAt(uint48 timestamp, bytes hints) view returns(((address,uint64)[],(address,uint64),(address,uint64)[]))
func (_Master *MasterSession) GetMasterConfigAt(timestamp *big.Int, hints []byte) (IMasterConfigProviderMasterConfig, error) {
	return _Master.Contract.GetMasterConfigAt(&_Master.CallOpts, timestamp, hints)
}

// GetMasterConfigAt is a free data retrieval call binding the contract method 0x9288968d.
//
// Solidity: function getMasterConfigAt(uint48 timestamp, bytes hints) view returns(((address,uint64)[],(address,uint64),(address,uint64)[]))
func (_Master *MasterCallerSession) GetMasterConfigAt(timestamp *big.Int, hints []byte) (IMasterConfigProviderMasterConfig, error) {
	return _Master.Contract.GetMasterConfigAt(&_Master.CallOpts, timestamp, hints)
}

// GetMaxValidatorsCount is a free data retrieval call binding the contract method 0x06ce894d.
//
// Solidity: function getMaxValidatorsCount() view returns(uint256)
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
// Solidity: function getMaxValidatorsCount() view returns(uint256)
func (_Master *MasterSession) GetMaxValidatorsCount() (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCount(&_Master.CallOpts)
}

// GetMaxValidatorsCount is a free data retrieval call binding the contract method 0x06ce894d.
//
// Solidity: function getMaxValidatorsCount() view returns(uint256)
func (_Master *MasterCallerSession) GetMaxValidatorsCount() (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCount(&_Master.CallOpts)
}

// GetMaxValidatorsCountAt is a free data retrieval call binding the contract method 0x2e1f3b08.
//
// Solidity: function getMaxValidatorsCountAt(uint48 timestamp, bytes hint) view returns(uint256)
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
// Solidity: function getMaxValidatorsCountAt(uint48 timestamp, bytes hint) view returns(uint256)
func (_Master *MasterSession) GetMaxValidatorsCountAt(timestamp *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetMaxValidatorsCountAt(&_Master.CallOpts, timestamp, hint)
}

// GetMaxValidatorsCountAt is a free data retrieval call binding the contract method 0x2e1f3b08.
//
// Solidity: function getMaxValidatorsCountAt(uint48 timestamp, bytes hint) view returns(uint256)
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

// GetQuorumThreshold is a free data retrieval call binding the contract method 0x98439923.
//
// Solidity: function getQuorumThreshold(uint8 keyTag) view returns(uint208)
func (_Master *MasterCaller) GetQuorumThreshold(opts *bind.CallOpts, keyTag uint8) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getQuorumThreshold", keyTag)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetQuorumThreshold is a free data retrieval call binding the contract method 0x98439923.
//
// Solidity: function getQuorumThreshold(uint8 keyTag) view returns(uint208)
func (_Master *MasterSession) GetQuorumThreshold(keyTag uint8) (*big.Int, error) {
	return _Master.Contract.GetQuorumThreshold(&_Master.CallOpts, keyTag)
}

// GetQuorumThreshold is a free data retrieval call binding the contract method 0x98439923.
//
// Solidity: function getQuorumThreshold(uint8 keyTag) view returns(uint208)
func (_Master *MasterCallerSession) GetQuorumThreshold(keyTag uint8) (*big.Int, error) {
	return _Master.Contract.GetQuorumThreshold(&_Master.CallOpts, keyTag)
}

// GetQuorumThresholdAt is a free data retrieval call binding the contract method 0x24826162.
//
// Solidity: function getQuorumThresholdAt(uint8 keyTag, uint48 epoch, bytes hint) view returns(uint208)
func (_Master *MasterCaller) GetQuorumThresholdAt(opts *bind.CallOpts, keyTag uint8, epoch *big.Int, hint []byte) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getQuorumThresholdAt", keyTag, epoch, hint)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetQuorumThresholdAt is a free data retrieval call binding the contract method 0x24826162.
//
// Solidity: function getQuorumThresholdAt(uint8 keyTag, uint48 epoch, bytes hint) view returns(uint208)
func (_Master *MasterSession) GetQuorumThresholdAt(keyTag uint8, epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetQuorumThresholdAt(&_Master.CallOpts, keyTag, epoch, hint)
}

// GetQuorumThresholdAt is a free data retrieval call binding the contract method 0x24826162.
//
// Solidity: function getQuorumThresholdAt(uint8 keyTag, uint48 epoch, bytes hint) view returns(uint208)
func (_Master *MasterCallerSession) GetQuorumThresholdAt(keyTag uint8, epoch *big.Int, hint []byte) (*big.Int, error) {
	return _Master.Contract.GetQuorumThresholdAt(&_Master.CallOpts, keyTag, epoch, hint)
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
// Solidity: function getRequiredKeyTagAt(uint48 epoch, bytes hint) view returns(uint8)
func (_Master *MasterCaller) GetRequiredKeyTagAt(opts *bind.CallOpts, epoch *big.Int, hint []byte) (uint8, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getRequiredKeyTagAt", epoch, hint)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRequiredKeyTagAt is a free data retrieval call binding the contract method 0x5b1bcecc.
//
// Solidity: function getRequiredKeyTagAt(uint48 epoch, bytes hint) view returns(uint8)
func (_Master *MasterSession) GetRequiredKeyTagAt(epoch *big.Int, hint []byte) (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagAt(&_Master.CallOpts, epoch, hint)
}

// GetRequiredKeyTagAt is a free data retrieval call binding the contract method 0x5b1bcecc.
//
// Solidity: function getRequiredKeyTagAt(uint48 epoch, bytes hint) view returns(uint8)
func (_Master *MasterCallerSession) GetRequiredKeyTagAt(epoch *big.Int, hint []byte) (uint8, error) {
	return _Master.Contract.GetRequiredKeyTagAt(&_Master.CallOpts, epoch, hint)
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
// Solidity: function getSigVerifierAt(uint48 epoch, bytes hint) view returns(address)
func (_Master *MasterCaller) GetSigVerifierAt(opts *bind.CallOpts, epoch *big.Int, hint []byte) (common.Address, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getSigVerifierAt", epoch, hint)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetSigVerifierAt is a free data retrieval call binding the contract method 0xa54ce263.
//
// Solidity: function getSigVerifierAt(uint48 epoch, bytes hint) view returns(address)
func (_Master *MasterSession) GetSigVerifierAt(epoch *big.Int, hint []byte) (common.Address, error) {
	return _Master.Contract.GetSigVerifierAt(&_Master.CallOpts, epoch, hint)
}

// GetSigVerifierAt is a free data retrieval call binding the contract method 0xa54ce263.
//
// Solidity: function getSigVerifierAt(uint48 epoch, bytes hint) view returns(address)
func (_Master *MasterCallerSession) GetSigVerifierAt(epoch *big.Int, hint []byte) (common.Address, error) {
	return _Master.Contract.GetSigVerifierAt(&_Master.CallOpts, epoch, hint)
}

// GetTotalActiveVotingPowerFromValSetHeader is a free data retrieval call binding the contract method 0x57a01f29.
//
// Solidity: function getTotalActiveVotingPowerFromValSetHeader() view returns(uint256)
func (_Master *MasterCaller) GetTotalActiveVotingPowerFromValSetHeader(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getTotalActiveVotingPowerFromValSetHeader")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalActiveVotingPowerFromValSetHeader is a free data retrieval call binding the contract method 0x57a01f29.
//
// Solidity: function getTotalActiveVotingPowerFromValSetHeader() view returns(uint256)
func (_Master *MasterSession) GetTotalActiveVotingPowerFromValSetHeader() (*big.Int, error) {
	return _Master.Contract.GetTotalActiveVotingPowerFromValSetHeader(&_Master.CallOpts)
}

// GetTotalActiveVotingPowerFromValSetHeader is a free data retrieval call binding the contract method 0x57a01f29.
//
// Solidity: function getTotalActiveVotingPowerFromValSetHeader() view returns(uint256)
func (_Master *MasterCallerSession) GetTotalActiveVotingPowerFromValSetHeader() (*big.Int, error) {
	return _Master.Contract.GetTotalActiveVotingPowerFromValSetHeader(&_Master.CallOpts)
}

// GetTotalActiveVotingPowerFromValSetHeaderAt is a free data retrieval call binding the contract method 0x7e1642f7.
//
// Solidity: function getTotalActiveVotingPowerFromValSetHeaderAt(uint48 epoch) view returns(uint256)
func (_Master *MasterCaller) GetTotalActiveVotingPowerFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getTotalActiveVotingPowerFromValSetHeaderAt", epoch)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalActiveVotingPowerFromValSetHeaderAt is a free data retrieval call binding the contract method 0x7e1642f7.
//
// Solidity: function getTotalActiveVotingPowerFromValSetHeaderAt(uint48 epoch) view returns(uint256)
func (_Master *MasterSession) GetTotalActiveVotingPowerFromValSetHeaderAt(epoch *big.Int) (*big.Int, error) {
	return _Master.Contract.GetTotalActiveVotingPowerFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetTotalActiveVotingPowerFromValSetHeaderAt is a free data retrieval call binding the contract method 0x7e1642f7.
//
// Solidity: function getTotalActiveVotingPowerFromValSetHeaderAt(uint48 epoch) view returns(uint256)
func (_Master *MasterCallerSession) GetTotalActiveVotingPowerFromValSetHeaderAt(epoch *big.Int) (*big.Int, error) {
	return _Master.Contract.GetTotalActiveVotingPowerFromValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetValSetConfig is a free data retrieval call binding the contract method 0x991bac99.
//
// Solidity: function getValSetConfig() view returns((uint256,uint256,uint256,uint8[]))
func (_Master *MasterCaller) GetValSetConfig(opts *bind.CallOpts) (IValSetConfigProviderValSetConfig, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValSetConfig")

	if err != nil {
		return *new(IValSetConfigProviderValSetConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IValSetConfigProviderValSetConfig)).(*IValSetConfigProviderValSetConfig)

	return out0, err

}

// GetValSetConfig is a free data retrieval call binding the contract method 0x991bac99.
//
// Solidity: function getValSetConfig() view returns((uint256,uint256,uint256,uint8[]))
func (_Master *MasterSession) GetValSetConfig() (IValSetConfigProviderValSetConfig, error) {
	return _Master.Contract.GetValSetConfig(&_Master.CallOpts)
}

// GetValSetConfig is a free data retrieval call binding the contract method 0x991bac99.
//
// Solidity: function getValSetConfig() view returns((uint256,uint256,uint256,uint8[]))
func (_Master *MasterCallerSession) GetValSetConfig() (IValSetConfigProviderValSetConfig, error) {
	return _Master.Contract.GetValSetConfig(&_Master.CallOpts)
}

// GetValSetConfigAt is a free data retrieval call binding the contract method 0x9029c8fe.
//
// Solidity: function getValSetConfigAt(uint48 timestamp, bytes hints) view returns((uint256,uint256,uint256,uint8[]))
func (_Master *MasterCaller) GetValSetConfigAt(opts *bind.CallOpts, timestamp *big.Int, hints []byte) (IValSetConfigProviderValSetConfig, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "getValSetConfigAt", timestamp, hints)

	if err != nil {
		return *new(IValSetConfigProviderValSetConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IValSetConfigProviderValSetConfig)).(*IValSetConfigProviderValSetConfig)

	return out0, err

}

// GetValSetConfigAt is a free data retrieval call binding the contract method 0x9029c8fe.
//
// Solidity: function getValSetConfigAt(uint48 timestamp, bytes hints) view returns((uint256,uint256,uint256,uint8[]))
func (_Master *MasterSession) GetValSetConfigAt(timestamp *big.Int, hints []byte) (IValSetConfigProviderValSetConfig, error) {
	return _Master.Contract.GetValSetConfigAt(&_Master.CallOpts, timestamp, hints)
}

// GetValSetConfigAt is a free data retrieval call binding the contract method 0x9029c8fe.
//
// Solidity: function getValSetConfigAt(uint48 timestamp, bytes hints) view returns((uint256,uint256,uint256,uint8[]))
func (_Master *MasterCallerSession) GetValSetConfigAt(timestamp *big.Int, hints []byte) (IValSetConfigProviderValSetConfig, error) {
	return _Master.Contract.GetValSetConfigAt(&_Master.CallOpts, timestamp, hints)
}

// GetValSetHeader is a free data retrieval call binding the contract method 0xadc91fc8.
//
// Solidity: function getValSetHeader() view returns((uint8,(uint8,bytes)[],uint256,bytes32,bytes) header)
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
// Solidity: function getValSetHeader() view returns((uint8,(uint8,bytes)[],uint256,bytes32,bytes) header)
func (_Master *MasterSession) GetValSetHeader() (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeader(&_Master.CallOpts)
}

// GetValSetHeader is a free data retrieval call binding the contract method 0xadc91fc8.
//
// Solidity: function getValSetHeader() view returns((uint8,(uint8,bytes)[],uint256,bytes32,bytes) header)
func (_Master *MasterCallerSession) GetValSetHeader() (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeader(&_Master.CallOpts)
}

// GetValSetHeaderAt is a free data retrieval call binding the contract method 0x4addaee7.
//
// Solidity: function getValSetHeaderAt(uint48 epoch) view returns((uint8,(uint8,bytes)[],uint256,bytes32,bytes))
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
// Solidity: function getValSetHeaderAt(uint48 epoch) view returns((uint8,(uint8,bytes)[],uint256,bytes32,bytes))
func (_Master *MasterSession) GetValSetHeaderAt(epoch *big.Int) (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeaderAt(&_Master.CallOpts, epoch)
}

// GetValSetHeaderAt is a free data retrieval call binding the contract method 0x4addaee7.
//
// Solidity: function getValSetHeaderAt(uint48 epoch) view returns((uint8,(uint8,bytes)[],uint256,bytes32,bytes))
func (_Master *MasterCallerSession) GetValSetHeaderAt(epoch *big.Int) (ISettlementValSetHeader, error) {
	return _Master.Contract.GetValSetHeaderAt(&_Master.CallOpts, epoch)
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

// IsReplicaActive is a free data retrieval call binding the contract method 0x545ae7bb.
//
// Solidity: function isReplicaActive((address,uint64) replica) view returns(bool)
func (_Master *MasterCaller) IsReplicaActive(opts *bind.CallOpts, replica IMasterConfigProviderCrossChainAddress) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isReplicaActive", replica)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReplicaActive is a free data retrieval call binding the contract method 0x545ae7bb.
//
// Solidity: function isReplicaActive((address,uint64) replica) view returns(bool)
func (_Master *MasterSession) IsReplicaActive(replica IMasterConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsReplicaActive(&_Master.CallOpts, replica)
}

// IsReplicaActive is a free data retrieval call binding the contract method 0x545ae7bb.
//
// Solidity: function isReplicaActive((address,uint64) replica) view returns(bool)
func (_Master *MasterCallerSession) IsReplicaActive(replica IMasterConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsReplicaActive(&_Master.CallOpts, replica)
}

// IsReplicaActiveAt is a free data retrieval call binding the contract method 0x53cd1864.
//
// Solidity: function isReplicaActiveAt((address,uint64) replica, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCaller) IsReplicaActiveAt(opts *bind.CallOpts, replica IMasterConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isReplicaActiveAt", replica, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsReplicaActiveAt is a free data retrieval call binding the contract method 0x53cd1864.
//
// Solidity: function isReplicaActiveAt((address,uint64) replica, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterSession) IsReplicaActiveAt(replica IMasterConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsReplicaActiveAt(&_Master.CallOpts, replica, timestamp, hint)
}

// IsReplicaActiveAt is a free data retrieval call binding the contract method 0x53cd1864.
//
// Solidity: function isReplicaActiveAt((address,uint64) replica, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCallerSession) IsReplicaActiveAt(replica IMasterConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsReplicaActiveAt(&_Master.CallOpts, replica, timestamp, hint)
}

// IsValSetHeaderSubmitted is a free data retrieval call binding the contract method 0x29c29c9c.
//
// Solidity: function isValSetHeaderSubmitted() view returns(bool)
func (_Master *MasterCaller) IsValSetHeaderSubmitted(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isValSetHeaderSubmitted")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValSetHeaderSubmitted is a free data retrieval call binding the contract method 0x29c29c9c.
//
// Solidity: function isValSetHeaderSubmitted() view returns(bool)
func (_Master *MasterSession) IsValSetHeaderSubmitted() (bool, error) {
	return _Master.Contract.IsValSetHeaderSubmitted(&_Master.CallOpts)
}

// IsValSetHeaderSubmitted is a free data retrieval call binding the contract method 0x29c29c9c.
//
// Solidity: function isValSetHeaderSubmitted() view returns(bool)
func (_Master *MasterCallerSession) IsValSetHeaderSubmitted() (bool, error) {
	return _Master.Contract.IsValSetHeaderSubmitted(&_Master.CallOpts)
}

// IsValSetHeaderSubmittedAt is a free data retrieval call binding the contract method 0xf9fa4038.
//
// Solidity: function isValSetHeaderSubmittedAt(uint48 epoch) view returns(bool)
func (_Master *MasterCaller) IsValSetHeaderSubmittedAt(opts *bind.CallOpts, epoch *big.Int) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isValSetHeaderSubmittedAt", epoch)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValSetHeaderSubmittedAt is a free data retrieval call binding the contract method 0xf9fa4038.
//
// Solidity: function isValSetHeaderSubmittedAt(uint48 epoch) view returns(bool)
func (_Master *MasterSession) IsValSetHeaderSubmittedAt(epoch *big.Int) (bool, error) {
	return _Master.Contract.IsValSetHeaderSubmittedAt(&_Master.CallOpts, epoch)
}

// IsValSetHeaderSubmittedAt is a free data retrieval call binding the contract method 0xf9fa4038.
//
// Solidity: function isValSetHeaderSubmittedAt(uint48 epoch) view returns(bool)
func (_Master *MasterCallerSession) IsValSetHeaderSubmittedAt(epoch *big.Int) (bool, error) {
	return _Master.Contract.IsValSetHeaderSubmittedAt(&_Master.CallOpts, epoch)
}

// IsVotingPowerProviderActive is a free data retrieval call binding the contract method 0x9290f2de.
//
// Solidity: function isVotingPowerProviderActive((address,uint64) votingPowerProvider) view returns(bool)
func (_Master *MasterCaller) IsVotingPowerProviderActive(opts *bind.CallOpts, votingPowerProvider IMasterConfigProviderCrossChainAddress) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isVotingPowerProviderActive", votingPowerProvider)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsVotingPowerProviderActive is a free data retrieval call binding the contract method 0x9290f2de.
//
// Solidity: function isVotingPowerProviderActive((address,uint64) votingPowerProvider) view returns(bool)
func (_Master *MasterSession) IsVotingPowerProviderActive(votingPowerProvider IMasterConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderActive(&_Master.CallOpts, votingPowerProvider)
}

// IsVotingPowerProviderActive is a free data retrieval call binding the contract method 0x9290f2de.
//
// Solidity: function isVotingPowerProviderActive((address,uint64) votingPowerProvider) view returns(bool)
func (_Master *MasterCallerSession) IsVotingPowerProviderActive(votingPowerProvider IMasterConfigProviderCrossChainAddress) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderActive(&_Master.CallOpts, votingPowerProvider)
}

// IsVotingPowerProviderActiveAt is a free data retrieval call binding the contract method 0x01054538.
//
// Solidity: function isVotingPowerProviderActiveAt((address,uint64) votingPowerProvider, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCaller) IsVotingPowerProviderActiveAt(opts *bind.CallOpts, votingPowerProvider IMasterConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "isVotingPowerProviderActiveAt", votingPowerProvider, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsVotingPowerProviderActiveAt is a free data retrieval call binding the contract method 0x01054538.
//
// Solidity: function isVotingPowerProviderActiveAt((address,uint64) votingPowerProvider, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterSession) IsVotingPowerProviderActiveAt(votingPowerProvider IMasterConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderActiveAt(&_Master.CallOpts, votingPowerProvider, timestamp, hint)
}

// IsVotingPowerProviderActiveAt is a free data retrieval call binding the contract method 0x01054538.
//
// Solidity: function isVotingPowerProviderActiveAt((address,uint64) votingPowerProvider, uint48 timestamp, bytes hint) view returns(bool)
func (_Master *MasterCallerSession) IsVotingPowerProviderActiveAt(votingPowerProvider IMasterConfigProviderCrossChainAddress, timestamp *big.Int, hint []byte) (bool, error) {
	return _Master.Contract.IsVotingPowerProviderActiveAt(&_Master.CallOpts, votingPowerProvider, timestamp, hint)
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

// VerifyQuorumSig is a free data retrieval call binding the contract method 0x71ef1c11.
//
// Solidity: function verifyQuorumSig(uint48 epoch, bytes message, uint8 keyTag, uint208 quorumThreshold, bytes proof) view returns(bool)
func (_Master *MasterCaller) VerifyQuorumSig(opts *bind.CallOpts, epoch *big.Int, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte) (bool, error) {
	var out []interface{}
	err := _Master.contract.Call(opts, &out, "verifyQuorumSig", epoch, message, keyTag, quorumThreshold, proof)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyQuorumSig is a free data retrieval call binding the contract method 0x71ef1c11.
//
// Solidity: function verifyQuorumSig(uint48 epoch, bytes message, uint8 keyTag, uint208 quorumThreshold, bytes proof) view returns(bool)
func (_Master *MasterSession) VerifyQuorumSig(epoch *big.Int, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte) (bool, error) {
	return _Master.Contract.VerifyQuorumSig(&_Master.CallOpts, epoch, message, keyTag, quorumThreshold, proof)
}

// VerifyQuorumSig is a free data retrieval call binding the contract method 0x71ef1c11.
//
// Solidity: function verifyQuorumSig(uint48 epoch, bytes message, uint8 keyTag, uint208 quorumThreshold, bytes proof) view returns(bool)
func (_Master *MasterCallerSession) VerifyQuorumSig(epoch *big.Int, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte) (bool, error) {
	return _Master.Contract.VerifyQuorumSig(&_Master.CallOpts, epoch, message, keyTag, quorumThreshold, proof)
}

// AddReplica is a paid mutator transaction binding the contract method 0x5c0bc730.
//
// Solidity: function addReplica((address,uint64) replica) returns()
func (_Master *MasterTransactor) AddReplica(opts *bind.TransactOpts, replica IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "addReplica", replica)
}

// AddReplica is a paid mutator transaction binding the contract method 0x5c0bc730.
//
// Solidity: function addReplica((address,uint64) replica) returns()
func (_Master *MasterSession) AddReplica(replica IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddReplica(&_Master.TransactOpts, replica)
}

// AddReplica is a paid mutator transaction binding the contract method 0x5c0bc730.
//
// Solidity: function addReplica((address,uint64) replica) returns()
func (_Master *MasterTransactorSession) AddReplica(replica IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddReplica(&_Master.TransactOpts, replica)
}

// AddVotingPowerProvider is a paid mutator transaction binding the contract method 0x88514824.
//
// Solidity: function addVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactor) AddVotingPowerProvider(opts *bind.TransactOpts, votingPowerProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "addVotingPowerProvider", votingPowerProvider)
}

// AddVotingPowerProvider is a paid mutator transaction binding the contract method 0x88514824.
//
// Solidity: function addVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterSession) AddVotingPowerProvider(votingPowerProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// AddVotingPowerProvider is a paid mutator transaction binding the contract method 0x88514824.
//
// Solidity: function addVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactorSession) AddVotingPowerProvider(votingPowerProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.AddVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// CommitValSetHeader is a paid mutator transaction binding the contract method 0x81408a08.
//
// Solidity: function commitValSetHeader((uint8,(uint8,bytes)[],uint256,bytes32,bytes) header, bytes proof) returns()
func (_Master *MasterTransactor) CommitValSetHeader(opts *bind.TransactOpts, header ISettlementValSetHeader, proof []byte) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "commitValSetHeader", header, proof)
}

// CommitValSetHeader is a paid mutator transaction binding the contract method 0x81408a08.
//
// Solidity: function commitValSetHeader((uint8,(uint8,bytes)[],uint256,bytes32,bytes) header, bytes proof) returns()
func (_Master *MasterSession) CommitValSetHeader(header ISettlementValSetHeader, proof []byte) (*types.Transaction, error) {
	return _Master.Contract.CommitValSetHeader(&_Master.TransactOpts, header, proof)
}

// CommitValSetHeader is a paid mutator transaction binding the contract method 0x81408a08.
//
// Solidity: function commitValSetHeader((uint8,(uint8,bytes)[],uint256,bytes32,bytes) header, bytes proof) returns()
func (_Master *MasterTransactorSession) CommitValSetHeader(header ISettlementValSetHeader, proof []byte) (*types.Transaction, error) {
	return _Master.Contract.CommitValSetHeader(&_Master.TransactOpts, header, proof)
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

// Initialize is a paid mutator transaction binding the contract method 0xed2293cf.
//
// Solidity: function initialize(((address,uint96),(uint48,uint48),(string,string),(uint8,uint208)[],uint48,uint8,address) settlementInitParams, (uint256,uint256,uint256,uint8[]) valSetConfigProviderInitParams, ((address,uint64)[],(address,uint64),(address,uint64)[]) masterConfigProviderInitParams, address defaultAdmin) returns()
func (_Master *MasterTransactor) Initialize(opts *bind.TransactOpts, settlementInitParams ISettlementSettlementInitParams, valSetConfigProviderInitParams IValSetConfigProviderValSetConfigProviderInitParams, masterConfigProviderInitParams IMasterConfigProviderMasterConfigProviderInitParams, defaultAdmin common.Address) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "initialize", settlementInitParams, valSetConfigProviderInitParams, masterConfigProviderInitParams, defaultAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0xed2293cf.
//
// Solidity: function initialize(((address,uint96),(uint48,uint48),(string,string),(uint8,uint208)[],uint48,uint8,address) settlementInitParams, (uint256,uint256,uint256,uint8[]) valSetConfigProviderInitParams, ((address,uint64)[],(address,uint64),(address,uint64)[]) masterConfigProviderInitParams, address defaultAdmin) returns()
func (_Master *MasterSession) Initialize(settlementInitParams ISettlementSettlementInitParams, valSetConfigProviderInitParams IValSetConfigProviderValSetConfigProviderInitParams, masterConfigProviderInitParams IMasterConfigProviderMasterConfigProviderInitParams, defaultAdmin common.Address) (*types.Transaction, error) {
	return _Master.Contract.Initialize(&_Master.TransactOpts, settlementInitParams, valSetConfigProviderInitParams, masterConfigProviderInitParams, defaultAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0xed2293cf.
//
// Solidity: function initialize(((address,uint96),(uint48,uint48),(string,string),(uint8,uint208)[],uint48,uint8,address) settlementInitParams, (uint256,uint256,uint256,uint8[]) valSetConfigProviderInitParams, ((address,uint64)[],(address,uint64),(address,uint64)[]) masterConfigProviderInitParams, address defaultAdmin) returns()
func (_Master *MasterTransactorSession) Initialize(settlementInitParams ISettlementSettlementInitParams, valSetConfigProviderInitParams IValSetConfigProviderValSetConfigProviderInitParams, masterConfigProviderInitParams IMasterConfigProviderMasterConfigProviderInitParams, defaultAdmin common.Address) (*types.Transaction, error) {
	return _Master.Contract.Initialize(&_Master.TransactOpts, settlementInitParams, valSetConfigProviderInitParams, masterConfigProviderInitParams, defaultAdmin)
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
func (_Master *MasterTransactor) RemoveReplica(opts *bind.TransactOpts, replica IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "removeReplica", replica)
}

// RemoveReplica is a paid mutator transaction binding the contract method 0x65f764f0.
//
// Solidity: function removeReplica((address,uint64) replica) returns()
func (_Master *MasterSession) RemoveReplica(replica IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveReplica(&_Master.TransactOpts, replica)
}

// RemoveReplica is a paid mutator transaction binding the contract method 0x65f764f0.
//
// Solidity: function removeReplica((address,uint64) replica) returns()
func (_Master *MasterTransactorSession) RemoveReplica(replica IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveReplica(&_Master.TransactOpts, replica)
}

// RemoveVotingPowerProvider is a paid mutator transaction binding the contract method 0xb01139a1.
//
// Solidity: function removeVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactor) RemoveVotingPowerProvider(opts *bind.TransactOpts, votingPowerProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "removeVotingPowerProvider", votingPowerProvider)
}

// RemoveVotingPowerProvider is a paid mutator transaction binding the contract method 0xb01139a1.
//
// Solidity: function removeVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterSession) RemoveVotingPowerProvider(votingPowerProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.RemoveVotingPowerProvider(&_Master.TransactOpts, votingPowerProvider)
}

// RemoveVotingPowerProvider is a paid mutator transaction binding the contract method 0xb01139a1.
//
// Solidity: function removeVotingPowerProvider((address,uint64) votingPowerProvider) returns()
func (_Master *MasterTransactorSession) RemoveVotingPowerProvider(votingPowerProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
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

// SetGenesis is a paid mutator transaction binding the contract method 0x7eaa4932.
//
// Solidity: function setGenesis((uint8,(uint8,bytes)[],uint256,bytes32,bytes) valSetHeader) returns()
func (_Master *MasterTransactor) SetGenesis(opts *bind.TransactOpts, valSetHeader ISettlementValSetHeader) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setGenesis", valSetHeader)
}

// SetGenesis is a paid mutator transaction binding the contract method 0x7eaa4932.
//
// Solidity: function setGenesis((uint8,(uint8,bytes)[],uint256,bytes32,bytes) valSetHeader) returns()
func (_Master *MasterSession) SetGenesis(valSetHeader ISettlementValSetHeader) (*types.Transaction, error) {
	return _Master.Contract.SetGenesis(&_Master.TransactOpts, valSetHeader)
}

// SetGenesis is a paid mutator transaction binding the contract method 0x7eaa4932.
//
// Solidity: function setGenesis((uint8,(uint8,bytes)[],uint256,bytes32,bytes) valSetHeader) returns()
func (_Master *MasterTransactorSession) SetGenesis(valSetHeader ISettlementValSetHeader) (*types.Transaction, error) {
	return _Master.Contract.SetGenesis(&_Master.TransactOpts, valSetHeader)
}

// SetKeysProvider is a paid mutator transaction binding the contract method 0x48fb6e2b.
//
// Solidity: function setKeysProvider((address,uint64) keysProvider) returns()
func (_Master *MasterTransactor) SetKeysProvider(opts *bind.TransactOpts, keysProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setKeysProvider", keysProvider)
}

// SetKeysProvider is a paid mutator transaction binding the contract method 0x48fb6e2b.
//
// Solidity: function setKeysProvider((address,uint64) keysProvider) returns()
func (_Master *MasterSession) SetKeysProvider(keysProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.SetKeysProvider(&_Master.TransactOpts, keysProvider)
}

// SetKeysProvider is a paid mutator transaction binding the contract method 0x48fb6e2b.
//
// Solidity: function setKeysProvider((address,uint64) keysProvider) returns()
func (_Master *MasterTransactorSession) SetKeysProvider(keysProvider IMasterConfigProviderCrossChainAddress) (*types.Transaction, error) {
	return _Master.Contract.SetKeysProvider(&_Master.TransactOpts, keysProvider)
}

// SetMaxValidatorsCount is a paid mutator transaction binding the contract method 0x582dffec.
//
// Solidity: function setMaxValidatorsCount(uint256 maxValidatorsCount) returns()
func (_Master *MasterTransactor) SetMaxValidatorsCount(opts *bind.TransactOpts, maxValidatorsCount *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setMaxValidatorsCount", maxValidatorsCount)
}

// SetMaxValidatorsCount is a paid mutator transaction binding the contract method 0x582dffec.
//
// Solidity: function setMaxValidatorsCount(uint256 maxValidatorsCount) returns()
func (_Master *MasterSession) SetMaxValidatorsCount(maxValidatorsCount *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetMaxValidatorsCount(&_Master.TransactOpts, maxValidatorsCount)
}

// SetMaxValidatorsCount is a paid mutator transaction binding the contract method 0x582dffec.
//
// Solidity: function setMaxValidatorsCount(uint256 maxValidatorsCount) returns()
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

// SetQuorumThreshold is a paid mutator transaction binding the contract method 0x6d34b673.
//
// Solidity: function setQuorumThreshold(uint8 keyTag, uint208 quorumThreshold) returns()
func (_Master *MasterTransactor) SetQuorumThreshold(opts *bind.TransactOpts, keyTag uint8, quorumThreshold *big.Int) (*types.Transaction, error) {
	return _Master.contract.Transact(opts, "setQuorumThreshold", keyTag, quorumThreshold)
}

// SetQuorumThreshold is a paid mutator transaction binding the contract method 0x6d34b673.
//
// Solidity: function setQuorumThreshold(uint8 keyTag, uint208 quorumThreshold) returns()
func (_Master *MasterSession) SetQuorumThreshold(keyTag uint8, quorumThreshold *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetQuorumThreshold(&_Master.TransactOpts, keyTag, quorumThreshold)
}

// SetQuorumThreshold is a paid mutator transaction binding the contract method 0x6d34b673.
//
// Solidity: function setQuorumThreshold(uint8 keyTag, uint208 quorumThreshold) returns()
func (_Master *MasterTransactorSession) SetQuorumThreshold(keyTag uint8, quorumThreshold *big.Int) (*types.Transaction, error) {
	return _Master.Contract.SetQuorumThreshold(&_Master.TransactOpts, keyTag, quorumThreshold)
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

// MasterSelectorRoleSetIterator is returned from FilterSelectorRoleSet and is used to iterate over the raw logs and unpacked data for SelectorRoleSet events raised by the Master contract.
type MasterSelectorRoleSetIterator struct {
	Event *MasterSelectorRoleSet // Event containing the contract specifics and raw log

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
func (it *MasterSelectorRoleSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MasterSelectorRoleSet)
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
		it.Event = new(MasterSelectorRoleSet)
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
func (it *MasterSelectorRoleSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MasterSelectorRoleSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MasterSelectorRoleSet represents a SelectorRoleSet event raised by the Master contract.
type MasterSelectorRoleSet struct {
	Selector [4]byte
	Role     [32]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSelectorRoleSet is a free log retrieval operation binding the contract event 0xb579d5e7e95ac8795a9c9ecce0ee2e2d189dce9827bac2e35ebbd3a68be7d423.
//
// Solidity: event SelectorRoleSet(bytes4 selector, bytes32 role)
func (_Master *MasterFilterer) FilterSelectorRoleSet(opts *bind.FilterOpts) (*MasterSelectorRoleSetIterator, error) {

	logs, sub, err := _Master.contract.FilterLogs(opts, "SelectorRoleSet")
	if err != nil {
		return nil, err
	}
	return &MasterSelectorRoleSetIterator{contract: _Master.contract, event: "SelectorRoleSet", logs: logs, sub: sub}, nil
}

// WatchSelectorRoleSet is a free log subscription operation binding the contract event 0xb579d5e7e95ac8795a9c9ecce0ee2e2d189dce9827bac2e35ebbd3a68be7d423.
//
// Solidity: event SelectorRoleSet(bytes4 selector, bytes32 role)
func (_Master *MasterFilterer) WatchSelectorRoleSet(opts *bind.WatchOpts, sink chan<- *MasterSelectorRoleSet) (event.Subscription, error) {

	logs, sub, err := _Master.contract.WatchLogs(opts, "SelectorRoleSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MasterSelectorRoleSet)
				if err := _Master.contract.UnpackLog(event, "SelectorRoleSet", log); err != nil {
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

// ParseSelectorRoleSet is a log parse operation binding the contract event 0xb579d5e7e95ac8795a9c9ecce0ee2e2d189dce9827bac2e35ebbd3a68be7d423.
//
// Solidity: event SelectorRoleSet(bytes4 selector, bytes32 role)
func (_Master *MasterFilterer) ParseSelectorRoleSet(log types.Log) (*MasterSelectorRoleSet, error) {
	event := new(MasterSelectorRoleSet)
	if err := _Master.contract.UnpackLog(event, "SelectorRoleSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
