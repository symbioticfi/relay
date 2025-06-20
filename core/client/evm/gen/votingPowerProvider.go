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

// IVotingPowerProviderOperatorVotingPower is an auto generated low-level Go binding around an user-defined struct.
type IVotingPowerProviderOperatorVotingPower struct {
	Operator common.Address
	Vaults   []IVotingPowerProviderVaultVotingPower
}

// IVotingPowerProviderVaultVotingPower is an auto generated low-level Go binding around an user-defined struct.
type IVotingPowerProviderVaultVotingPower struct {
	Vault       common.Address
	VotingPower *big.Int
}

// IVotingPowerProviderMetaData contains all meta data concerning the IVotingPowerProvider contract.
var IVotingPowerProviderMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"NETWORK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OPERATOR_REGISTRY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SUBNETWORK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SUBNETWORK_IDENTIFIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint96\",\"internalType\":\"uint96\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VAULT_FACTORY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorStake\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorStakeAt\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVaults\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVaultsAt\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVaultsLength\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVotingPower\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVotingPowerAt\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hints\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVotingPowers\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIVotingPowerProvider.VaultVotingPower[]\",\"components\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorVotingPowersAt\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIVotingPowerProvider.VaultVotingPower[]\",\"components\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperators\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorsAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorsLength\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSharedVaults\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSharedVaultsAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSharedVaultsLength\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSlashingWindow\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTokens\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTokensAt\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTokensLength\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVotingPowers\",\"inputs\":[{\"name\":\"extraData\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIVotingPowerProvider.OperatorVotingPower[]\",\"components\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vaults\",\"type\":\"tuple[]\",\"internalType\":\"structIVotingPowerProvider.VaultVotingPower[]\",\"components\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVotingPowersAt\",\"inputs\":[{\"name\":\"extraData\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIVotingPowerProvider.OperatorVotingPower[]\",\"components\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vaults\",\"type\":\"tuple[]\",\"internalType\":\"structIVotingPowerProvider.VaultVotingPower[]\",\"components\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashTypedDataV4CrossChain\",\"inputs\":[{\"name\":\"structHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"invalidateOldSignatures\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isOperatorRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorRegisteredAt\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorVaultRegistered\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorVaultRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorVaultRegisteredAt\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorVaultRegisteredAt\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isSharedVaultRegistered\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isSharedVaultRegisteredAt\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isTokenRegistered\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isTokenRegisteredAt\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nonces\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerOperatorWithSignature\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stakeToVotingPower\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"stake\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"power\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"stakeToVotingPowerAt\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"stake\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"extraData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"outputs\":[{\"name\":\"power\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"staticDelegateCall\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unregisterOperator\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unregisterOperatorWithSignature\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitEIP712\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InitSubnetwork\",\"inputs\":[{\"name\":\"network\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"subnetworkID\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegisterOperatorVault\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegisterSharedVault\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegisterToken\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetSlashingWindow\",\"inputs\":[{\"name\":\"slashingWindow\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnregisterOperatorVault\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnregisterSharedVault\",\"inputs\":[{\"name\":\"vault\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnregisterToken\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidAccountNonce\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"currentNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NetworkManager_InvalidNetwork\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_InvalidOperator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_InvalidOperatorVault\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_InvalidSharedVault\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_InvalidToken\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_InvalidVault\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_NoSlasher\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_NonVetoSlasher\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_OperatorNotOptedIn\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_OperatorVaultAlreadyIsRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_OperatorVaultNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_SharedVaultAlreadyIsRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_SharedVaultNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_SlashingWindowTooLarge\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_TokenAlreadyIsRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_TokenNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_UnknownSlasherType\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_UnregisteredOperatorSlash\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"VotingPowerProvider_UnregisteredVaultSlash\",\"inputs\":[]}]",
}

// IVotingPowerProviderABI is the input ABI used to generate the binding from.
// Deprecated: Use IVotingPowerProviderMetaData.ABI instead.
var IVotingPowerProviderABI = IVotingPowerProviderMetaData.ABI

// IVotingPowerProvider is an auto generated Go binding around an Ethereum contract.
type IVotingPowerProvider struct {
	IVotingPowerProviderCaller     // Read-only binding to the contract
	IVotingPowerProviderTransactor // Write-only binding to the contract
	IVotingPowerProviderFilterer   // Log filterer for contract events
}

// IVotingPowerProviderCaller is an auto generated read-only Go binding around an Ethereum contract.
type IVotingPowerProviderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IVotingPowerProviderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IVotingPowerProviderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IVotingPowerProviderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IVotingPowerProviderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IVotingPowerProviderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IVotingPowerProviderSession struct {
	Contract     *IVotingPowerProvider // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// IVotingPowerProviderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IVotingPowerProviderCallerSession struct {
	Contract *IVotingPowerProviderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// IVotingPowerProviderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IVotingPowerProviderTransactorSession struct {
	Contract     *IVotingPowerProviderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// IVotingPowerProviderRaw is an auto generated low-level Go binding around an Ethereum contract.
type IVotingPowerProviderRaw struct {
	Contract *IVotingPowerProvider // Generic contract binding to access the raw methods on
}

// IVotingPowerProviderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IVotingPowerProviderCallerRaw struct {
	Contract *IVotingPowerProviderCaller // Generic read-only contract binding to access the raw methods on
}

// IVotingPowerProviderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IVotingPowerProviderTransactorRaw struct {
	Contract *IVotingPowerProviderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIVotingPowerProvider creates a new instance of IVotingPowerProvider, bound to a specific deployed contract.
func NewIVotingPowerProvider(address common.Address, backend bind.ContractBackend) (*IVotingPowerProvider, error) {
	contract, err := bindIVotingPowerProvider(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProvider{IVotingPowerProviderCaller: IVotingPowerProviderCaller{contract: contract}, IVotingPowerProviderTransactor: IVotingPowerProviderTransactor{contract: contract}, IVotingPowerProviderFilterer: IVotingPowerProviderFilterer{contract: contract}}, nil
}

// NewIVotingPowerProviderCaller creates a new read-only instance of IVotingPowerProvider, bound to a specific deployed contract.
func NewIVotingPowerProviderCaller(address common.Address, caller bind.ContractCaller) (*IVotingPowerProviderCaller, error) {
	contract, err := bindIVotingPowerProvider(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderCaller{contract: contract}, nil
}

// NewIVotingPowerProviderTransactor creates a new write-only instance of IVotingPowerProvider, bound to a specific deployed contract.
func NewIVotingPowerProviderTransactor(address common.Address, transactor bind.ContractTransactor) (*IVotingPowerProviderTransactor, error) {
	contract, err := bindIVotingPowerProvider(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderTransactor{contract: contract}, nil
}

// NewIVotingPowerProviderFilterer creates a new log filterer instance of IVotingPowerProvider, bound to a specific deployed contract.
func NewIVotingPowerProviderFilterer(address common.Address, filterer bind.ContractFilterer) (*IVotingPowerProviderFilterer, error) {
	contract, err := bindIVotingPowerProvider(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderFilterer{contract: contract}, nil
}

// bindIVotingPowerProvider binds a generic wrapper to an already deployed contract.
func bindIVotingPowerProvider(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IVotingPowerProviderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IVotingPowerProvider *IVotingPowerProviderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IVotingPowerProvider.Contract.IVotingPowerProviderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IVotingPowerProvider *IVotingPowerProviderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.IVotingPowerProviderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IVotingPowerProvider *IVotingPowerProviderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.IVotingPowerProviderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IVotingPowerProvider *IVotingPowerProviderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IVotingPowerProvider.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IVotingPowerProvider *IVotingPowerProviderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IVotingPowerProvider *IVotingPowerProviderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.contract.Transact(opts, method, params...)
}

// NETWORK is a free data retrieval call binding the contract method 0x8759e6d1.
//
// Solidity: function NETWORK() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) NETWORK(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "NETWORK")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NETWORK is a free data retrieval call binding the contract method 0x8759e6d1.
//
// Solidity: function NETWORK() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderSession) NETWORK() (common.Address, error) {
	return _IVotingPowerProvider.Contract.NETWORK(&_IVotingPowerProvider.CallOpts)
}

// NETWORK is a free data retrieval call binding the contract method 0x8759e6d1.
//
// Solidity: function NETWORK() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) NETWORK() (common.Address, error) {
	return _IVotingPowerProvider.Contract.NETWORK(&_IVotingPowerProvider.CallOpts)
}

// OPERATORREGISTRY is a free data retrieval call binding the contract method 0x83ce0322.
//
// Solidity: function OPERATOR_REGISTRY() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) OPERATORREGISTRY(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "OPERATOR_REGISTRY")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OPERATORREGISTRY is a free data retrieval call binding the contract method 0x83ce0322.
//
// Solidity: function OPERATOR_REGISTRY() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderSession) OPERATORREGISTRY() (common.Address, error) {
	return _IVotingPowerProvider.Contract.OPERATORREGISTRY(&_IVotingPowerProvider.CallOpts)
}

// OPERATORREGISTRY is a free data retrieval call binding the contract method 0x83ce0322.
//
// Solidity: function OPERATOR_REGISTRY() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) OPERATORREGISTRY() (common.Address, error) {
	return _IVotingPowerProvider.Contract.OPERATORREGISTRY(&_IVotingPowerProvider.CallOpts)
}

// SUBNETWORK is a free data retrieval call binding the contract method 0x773e6b54.
//
// Solidity: function SUBNETWORK() view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) SUBNETWORK(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "SUBNETWORK")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SUBNETWORK is a free data retrieval call binding the contract method 0x773e6b54.
//
// Solidity: function SUBNETWORK() view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderSession) SUBNETWORK() ([32]byte, error) {
	return _IVotingPowerProvider.Contract.SUBNETWORK(&_IVotingPowerProvider.CallOpts)
}

// SUBNETWORK is a free data retrieval call binding the contract method 0x773e6b54.
//
// Solidity: function SUBNETWORK() view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) SUBNETWORK() ([32]byte, error) {
	return _IVotingPowerProvider.Contract.SUBNETWORK(&_IVotingPowerProvider.CallOpts)
}

// SUBNETWORKIDENTIFIER is a free data retrieval call binding the contract method 0xabacb807.
//
// Solidity: function SUBNETWORK_IDENTIFIER() view returns(uint96)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) SUBNETWORKIDENTIFIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "SUBNETWORK_IDENTIFIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SUBNETWORKIDENTIFIER is a free data retrieval call binding the contract method 0xabacb807.
//
// Solidity: function SUBNETWORK_IDENTIFIER() view returns(uint96)
func (_IVotingPowerProvider *IVotingPowerProviderSession) SUBNETWORKIDENTIFIER() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.SUBNETWORKIDENTIFIER(&_IVotingPowerProvider.CallOpts)
}

// SUBNETWORKIDENTIFIER is a free data retrieval call binding the contract method 0xabacb807.
//
// Solidity: function SUBNETWORK_IDENTIFIER() view returns(uint96)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) SUBNETWORKIDENTIFIER() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.SUBNETWORKIDENTIFIER(&_IVotingPowerProvider.CallOpts)
}

// VAULTFACTORY is a free data retrieval call binding the contract method 0x103f2907.
//
// Solidity: function VAULT_FACTORY() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) VAULTFACTORY(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "VAULT_FACTORY")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VAULTFACTORY is a free data retrieval call binding the contract method 0x103f2907.
//
// Solidity: function VAULT_FACTORY() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderSession) VAULTFACTORY() (common.Address, error) {
	return _IVotingPowerProvider.Contract.VAULTFACTORY(&_IVotingPowerProvider.CallOpts)
}

// VAULTFACTORY is a free data retrieval call binding the contract method 0x103f2907.
//
// Solidity: function VAULT_FACTORY() view returns(address)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) VAULTFACTORY() (common.Address, error) {
	return _IVotingPowerProvider.Contract.VAULTFACTORY(&_IVotingPowerProvider.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "eip712Domain")

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
func (_IVotingPowerProvider *IVotingPowerProviderSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _IVotingPowerProvider.Contract.Eip712Domain(&_IVotingPowerProvider.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _IVotingPowerProvider.Contract.Eip712Domain(&_IVotingPowerProvider.CallOpts)
}

// GetOperatorStake is a free data retrieval call binding the contract method 0xff8726e7.
//
// Solidity: function getOperatorStake(address vault, address operator) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorStake(opts *bind.CallOpts, vault common.Address, operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorStake", vault, operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorStake is a free data retrieval call binding the contract method 0xff8726e7.
//
// Solidity: function getOperatorStake(address vault, address operator) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorStake(vault common.Address, operator common.Address) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorStake(&_IVotingPowerProvider.CallOpts, vault, operator)
}

// GetOperatorStake is a free data retrieval call binding the contract method 0xff8726e7.
//
// Solidity: function getOperatorStake(address vault, address operator) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorStake(vault common.Address, operator common.Address) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorStake(&_IVotingPowerProvider.CallOpts, vault, operator)
}

// GetOperatorStakeAt is a free data retrieval call binding the contract method 0xaadd3367.
//
// Solidity: function getOperatorStakeAt(address vault, address operator, uint48 timestamp, bytes hints) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorStakeAt(opts *bind.CallOpts, vault common.Address, operator common.Address, timestamp *big.Int, hints []byte) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorStakeAt", vault, operator, timestamp, hints)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorStakeAt is a free data retrieval call binding the contract method 0xaadd3367.
//
// Solidity: function getOperatorStakeAt(address vault, address operator, uint48 timestamp, bytes hints) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorStakeAt(vault common.Address, operator common.Address, timestamp *big.Int, hints []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorStakeAt(&_IVotingPowerProvider.CallOpts, vault, operator, timestamp, hints)
}

// GetOperatorStakeAt is a free data retrieval call binding the contract method 0xaadd3367.
//
// Solidity: function getOperatorStakeAt(address vault, address operator, uint48 timestamp, bytes hints) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorStakeAt(vault common.Address, operator common.Address, timestamp *big.Int, hints []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorStakeAt(&_IVotingPowerProvider.CallOpts, vault, operator, timestamp, hints)
}

// GetOperatorVaults is a free data retrieval call binding the contract method 0x14d7e25b.
//
// Solidity: function getOperatorVaults(address operator) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVaults(opts *bind.CallOpts, operator common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVaults", operator)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperatorVaults is a free data retrieval call binding the contract method 0x14d7e25b.
//
// Solidity: function getOperatorVaults(address operator) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVaults(operator common.Address) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVaults(&_IVotingPowerProvider.CallOpts, operator)
}

// GetOperatorVaults is a free data retrieval call binding the contract method 0x14d7e25b.
//
// Solidity: function getOperatorVaults(address operator) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVaults(operator common.Address) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVaults(&_IVotingPowerProvider.CallOpts, operator)
}

// GetOperatorVaultsAt is a free data retrieval call binding the contract method 0x49f993ec.
//
// Solidity: function getOperatorVaultsAt(address operator, uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVaultsAt(opts *bind.CallOpts, operator common.Address, timestamp *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVaultsAt", operator, timestamp)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperatorVaultsAt is a free data retrieval call binding the contract method 0x49f993ec.
//
// Solidity: function getOperatorVaultsAt(address operator, uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVaultsAt(operator common.Address, timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVaultsAt(&_IVotingPowerProvider.CallOpts, operator, timestamp)
}

// GetOperatorVaultsAt is a free data retrieval call binding the contract method 0x49f993ec.
//
// Solidity: function getOperatorVaultsAt(address operator, uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVaultsAt(operator common.Address, timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVaultsAt(&_IVotingPowerProvider.CallOpts, operator, timestamp)
}

// GetOperatorVaultsLength is a free data retrieval call binding the contract method 0x743a461b.
//
// Solidity: function getOperatorVaultsLength(address operator) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVaultsLength(opts *bind.CallOpts, operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVaultsLength", operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorVaultsLength is a free data retrieval call binding the contract method 0x743a461b.
//
// Solidity: function getOperatorVaultsLength(address operator) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVaultsLength(operator common.Address) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVaultsLength(&_IVotingPowerProvider.CallOpts, operator)
}

// GetOperatorVaultsLength is a free data retrieval call binding the contract method 0x743a461b.
//
// Solidity: function getOperatorVaultsLength(address operator) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVaultsLength(operator common.Address) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVaultsLength(&_IVotingPowerProvider.CallOpts, operator)
}

// GetOperatorVotingPower is a free data retrieval call binding the contract method 0x4130e52e.
//
// Solidity: function getOperatorVotingPower(address operator, address vault, bytes extraData) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVotingPower(opts *bind.CallOpts, operator common.Address, vault common.Address, extraData []byte) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVotingPower", operator, vault, extraData)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorVotingPower is a free data retrieval call binding the contract method 0x4130e52e.
//
// Solidity: function getOperatorVotingPower(address operator, address vault, bytes extraData) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVotingPower(operator common.Address, vault common.Address, extraData []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPower(&_IVotingPowerProvider.CallOpts, operator, vault, extraData)
}

// GetOperatorVotingPower is a free data retrieval call binding the contract method 0x4130e52e.
//
// Solidity: function getOperatorVotingPower(address operator, address vault, bytes extraData) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVotingPower(operator common.Address, vault common.Address, extraData []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPower(&_IVotingPowerProvider.CallOpts, operator, vault, extraData)
}

// GetOperatorVotingPowerAt is a free data retrieval call binding the contract method 0xd38ef044.
//
// Solidity: function getOperatorVotingPowerAt(address operator, address vault, bytes extraData, uint48 timestamp, bytes hints) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVotingPowerAt(opts *bind.CallOpts, operator common.Address, vault common.Address, extraData []byte, timestamp *big.Int, hints []byte) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVotingPowerAt", operator, vault, extraData, timestamp, hints)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorVotingPowerAt is a free data retrieval call binding the contract method 0xd38ef044.
//
// Solidity: function getOperatorVotingPowerAt(address operator, address vault, bytes extraData, uint48 timestamp, bytes hints) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVotingPowerAt(operator common.Address, vault common.Address, extraData []byte, timestamp *big.Int, hints []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPowerAt(&_IVotingPowerProvider.CallOpts, operator, vault, extraData, timestamp, hints)
}

// GetOperatorVotingPowerAt is a free data retrieval call binding the contract method 0xd38ef044.
//
// Solidity: function getOperatorVotingPowerAt(address operator, address vault, bytes extraData, uint48 timestamp, bytes hints) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVotingPowerAt(operator common.Address, vault common.Address, extraData []byte, timestamp *big.Int, hints []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPowerAt(&_IVotingPowerProvider.CallOpts, operator, vault, extraData, timestamp, hints)
}

// GetOperatorVotingPowers is a free data retrieval call binding the contract method 0x63ff1140.
//
// Solidity: function getOperatorVotingPowers(address operator, bytes extraData) view returns((address,uint256)[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVotingPowers(opts *bind.CallOpts, operator common.Address, extraData []byte) ([]IVotingPowerProviderVaultVotingPower, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVotingPowers", operator, extraData)

	if err != nil {
		return *new([]IVotingPowerProviderVaultVotingPower), err
	}

	out0 := *abi.ConvertType(out[0], new([]IVotingPowerProviderVaultVotingPower)).(*[]IVotingPowerProviderVaultVotingPower)

	return out0, err

}

// GetOperatorVotingPowers is a free data retrieval call binding the contract method 0x63ff1140.
//
// Solidity: function getOperatorVotingPowers(address operator, bytes extraData) view returns((address,uint256)[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVotingPowers(operator common.Address, extraData []byte) ([]IVotingPowerProviderVaultVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPowers(&_IVotingPowerProvider.CallOpts, operator, extraData)
}

// GetOperatorVotingPowers is a free data retrieval call binding the contract method 0x63ff1140.
//
// Solidity: function getOperatorVotingPowers(address operator, bytes extraData) view returns((address,uint256)[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVotingPowers(operator common.Address, extraData []byte) ([]IVotingPowerProviderVaultVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPowers(&_IVotingPowerProvider.CallOpts, operator, extraData)
}

// GetOperatorVotingPowersAt is a free data retrieval call binding the contract method 0x380f9945.
//
// Solidity: function getOperatorVotingPowersAt(address operator, bytes extraData, uint48 timestamp) view returns((address,uint256)[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorVotingPowersAt(opts *bind.CallOpts, operator common.Address, extraData []byte, timestamp *big.Int) ([]IVotingPowerProviderVaultVotingPower, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorVotingPowersAt", operator, extraData, timestamp)

	if err != nil {
		return *new([]IVotingPowerProviderVaultVotingPower), err
	}

	out0 := *abi.ConvertType(out[0], new([]IVotingPowerProviderVaultVotingPower)).(*[]IVotingPowerProviderVaultVotingPower)

	return out0, err

}

// GetOperatorVotingPowersAt is a free data retrieval call binding the contract method 0x380f9945.
//
// Solidity: function getOperatorVotingPowersAt(address operator, bytes extraData, uint48 timestamp) view returns((address,uint256)[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorVotingPowersAt(operator common.Address, extraData []byte, timestamp *big.Int) ([]IVotingPowerProviderVaultVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPowersAt(&_IVotingPowerProvider.CallOpts, operator, extraData, timestamp)
}

// GetOperatorVotingPowersAt is a free data retrieval call binding the contract method 0x380f9945.
//
// Solidity: function getOperatorVotingPowersAt(address operator, bytes extraData, uint48 timestamp) view returns((address,uint256)[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorVotingPowersAt(operator common.Address, extraData []byte, timestamp *big.Int) ([]IVotingPowerProviderVaultVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetOperatorVotingPowersAt(&_IVotingPowerProvider.CallOpts, operator, extraData, timestamp)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperators")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperators() ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperators(&_IVotingPowerProvider.CallOpts)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperators() ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperators(&_IVotingPowerProvider.CallOpts)
}

// GetOperatorsAt is a free data retrieval call binding the contract method 0xa2e33009.
//
// Solidity: function getOperatorsAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorsAt(opts *bind.CallOpts, timestamp *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorsAt", timestamp)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperatorsAt is a free data retrieval call binding the contract method 0xa2e33009.
//
// Solidity: function getOperatorsAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorsAt(timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperatorsAt(&_IVotingPowerProvider.CallOpts, timestamp)
}

// GetOperatorsAt is a free data retrieval call binding the contract method 0xa2e33009.
//
// Solidity: function getOperatorsAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorsAt(timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetOperatorsAt(&_IVotingPowerProvider.CallOpts, timestamp)
}

// GetOperatorsLength is a free data retrieval call binding the contract method 0xba24ecab.
//
// Solidity: function getOperatorsLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetOperatorsLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getOperatorsLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorsLength is a free data retrieval call binding the contract method 0xba24ecab.
//
// Solidity: function getOperatorsLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetOperatorsLength() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorsLength(&_IVotingPowerProvider.CallOpts)
}

// GetOperatorsLength is a free data retrieval call binding the contract method 0xba24ecab.
//
// Solidity: function getOperatorsLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetOperatorsLength() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetOperatorsLength(&_IVotingPowerProvider.CallOpts)
}

// GetSharedVaults is a free data retrieval call binding the contract method 0xc28474cd.
//
// Solidity: function getSharedVaults() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetSharedVaults(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getSharedVaults")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetSharedVaults is a free data retrieval call binding the contract method 0xc28474cd.
//
// Solidity: function getSharedVaults() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetSharedVaults() ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetSharedVaults(&_IVotingPowerProvider.CallOpts)
}

// GetSharedVaults is a free data retrieval call binding the contract method 0xc28474cd.
//
// Solidity: function getSharedVaults() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetSharedVaults() ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetSharedVaults(&_IVotingPowerProvider.CallOpts)
}

// GetSharedVaultsAt is a free data retrieval call binding the contract method 0x4a0c7c17.
//
// Solidity: function getSharedVaultsAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetSharedVaultsAt(opts *bind.CallOpts, timestamp *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getSharedVaultsAt", timestamp)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetSharedVaultsAt is a free data retrieval call binding the contract method 0x4a0c7c17.
//
// Solidity: function getSharedVaultsAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetSharedVaultsAt(timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetSharedVaultsAt(&_IVotingPowerProvider.CallOpts, timestamp)
}

// GetSharedVaultsAt is a free data retrieval call binding the contract method 0x4a0c7c17.
//
// Solidity: function getSharedVaultsAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetSharedVaultsAt(timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetSharedVaultsAt(&_IVotingPowerProvider.CallOpts, timestamp)
}

// GetSharedVaultsLength is a free data retrieval call binding the contract method 0x0bfb3713.
//
// Solidity: function getSharedVaultsLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetSharedVaultsLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getSharedVaultsLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSharedVaultsLength is a free data retrieval call binding the contract method 0x0bfb3713.
//
// Solidity: function getSharedVaultsLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetSharedVaultsLength() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetSharedVaultsLength(&_IVotingPowerProvider.CallOpts)
}

// GetSharedVaultsLength is a free data retrieval call binding the contract method 0x0bfb3713.
//
// Solidity: function getSharedVaultsLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetSharedVaultsLength() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetSharedVaultsLength(&_IVotingPowerProvider.CallOpts)
}

// GetSlashingWindow is a free data retrieval call binding the contract method 0x4a4807b7.
//
// Solidity: function getSlashingWindow() view returns(uint48)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetSlashingWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getSlashingWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlashingWindow is a free data retrieval call binding the contract method 0x4a4807b7.
//
// Solidity: function getSlashingWindow() view returns(uint48)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetSlashingWindow() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetSlashingWindow(&_IVotingPowerProvider.CallOpts)
}

// GetSlashingWindow is a free data retrieval call binding the contract method 0x4a4807b7.
//
// Solidity: function getSlashingWindow() view returns(uint48)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetSlashingWindow() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetSlashingWindow(&_IVotingPowerProvider.CallOpts)
}

// GetTokens is a free data retrieval call binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetTokens(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getTokens")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetTokens is a free data retrieval call binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetTokens() ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetTokens(&_IVotingPowerProvider.CallOpts)
}

// GetTokens is a free data retrieval call binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetTokens() ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetTokens(&_IVotingPowerProvider.CallOpts)
}

// GetTokensAt is a free data retrieval call binding the contract method 0x1796df1b.
//
// Solidity: function getTokensAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetTokensAt(opts *bind.CallOpts, timestamp *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getTokensAt", timestamp)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetTokensAt is a free data retrieval call binding the contract method 0x1796df1b.
//
// Solidity: function getTokensAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetTokensAt(timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetTokensAt(&_IVotingPowerProvider.CallOpts, timestamp)
}

// GetTokensAt is a free data retrieval call binding the contract method 0x1796df1b.
//
// Solidity: function getTokensAt(uint48 timestamp) view returns(address[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetTokensAt(timestamp *big.Int) ([]common.Address, error) {
	return _IVotingPowerProvider.Contract.GetTokensAt(&_IVotingPowerProvider.CallOpts, timestamp)
}

// GetTokensLength is a free data retrieval call binding the contract method 0xb0c26ecf.
//
// Solidity: function getTokensLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetTokensLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getTokensLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokensLength is a free data retrieval call binding the contract method 0xb0c26ecf.
//
// Solidity: function getTokensLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetTokensLength() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetTokensLength(&_IVotingPowerProvider.CallOpts)
}

// GetTokensLength is a free data retrieval call binding the contract method 0xb0c26ecf.
//
// Solidity: function getTokensLength() view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetTokensLength() (*big.Int, error) {
	return _IVotingPowerProvider.Contract.GetTokensLength(&_IVotingPowerProvider.CallOpts)
}

// GetVotingPowers is a free data retrieval call binding the contract method 0xff7cd71c.
//
// Solidity: function getVotingPowers(bytes[] extraData) view returns((address,(address,uint256)[])[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetVotingPowers(opts *bind.CallOpts, extraData [][]byte) ([]IVotingPowerProviderOperatorVotingPower, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getVotingPowers", extraData)

	if err != nil {
		return *new([]IVotingPowerProviderOperatorVotingPower), err
	}

	out0 := *abi.ConvertType(out[0], new([]IVotingPowerProviderOperatorVotingPower)).(*[]IVotingPowerProviderOperatorVotingPower)

	return out0, err

}

// GetVotingPowers is a free data retrieval call binding the contract method 0xff7cd71c.
//
// Solidity: function getVotingPowers(bytes[] extraData) view returns((address,(address,uint256)[])[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetVotingPowers(extraData [][]byte) ([]IVotingPowerProviderOperatorVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetVotingPowers(&_IVotingPowerProvider.CallOpts, extraData)
}

// GetVotingPowers is a free data retrieval call binding the contract method 0xff7cd71c.
//
// Solidity: function getVotingPowers(bytes[] extraData) view returns((address,(address,uint256)[])[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetVotingPowers(extraData [][]byte) ([]IVotingPowerProviderOperatorVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetVotingPowers(&_IVotingPowerProvider.CallOpts, extraData)
}

// GetVotingPowersAt is a free data retrieval call binding the contract method 0x77adea5f.
//
// Solidity: function getVotingPowersAt(bytes[] extraData, uint48 timestamp) view returns((address,(address,uint256)[])[])
func (_IVotingPowerProvider *IVotingPowerProviderCaller) GetVotingPowersAt(opts *bind.CallOpts, extraData [][]byte, timestamp *big.Int) ([]IVotingPowerProviderOperatorVotingPower, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "getVotingPowersAt", extraData, timestamp)

	if err != nil {
		return *new([]IVotingPowerProviderOperatorVotingPower), err
	}

	out0 := *abi.ConvertType(out[0], new([]IVotingPowerProviderOperatorVotingPower)).(*[]IVotingPowerProviderOperatorVotingPower)

	return out0, err

}

// GetVotingPowersAt is a free data retrieval call binding the contract method 0x77adea5f.
//
// Solidity: function getVotingPowersAt(bytes[] extraData, uint48 timestamp) view returns((address,(address,uint256)[])[])
func (_IVotingPowerProvider *IVotingPowerProviderSession) GetVotingPowersAt(extraData [][]byte, timestamp *big.Int) ([]IVotingPowerProviderOperatorVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetVotingPowersAt(&_IVotingPowerProvider.CallOpts, extraData, timestamp)
}

// GetVotingPowersAt is a free data retrieval call binding the contract method 0x77adea5f.
//
// Solidity: function getVotingPowersAt(bytes[] extraData, uint48 timestamp) view returns((address,(address,uint256)[])[])
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) GetVotingPowersAt(extraData [][]byte, timestamp *big.Int) ([]IVotingPowerProviderOperatorVotingPower, error) {
	return _IVotingPowerProvider.Contract.GetVotingPowersAt(&_IVotingPowerProvider.CallOpts, extraData, timestamp)
}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) HashTypedDataV4(opts *bind.CallOpts, structHash [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "hashTypedDataV4", structHash)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderSession) HashTypedDataV4(structHash [32]byte) ([32]byte, error) {
	return _IVotingPowerProvider.Contract.HashTypedDataV4(&_IVotingPowerProvider.CallOpts, structHash)
}

// HashTypedDataV4 is a free data retrieval call binding the contract method 0x4980f288.
//
// Solidity: function hashTypedDataV4(bytes32 structHash) view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) HashTypedDataV4(structHash [32]byte) ([32]byte, error) {
	return _IVotingPowerProvider.Contract.HashTypedDataV4(&_IVotingPowerProvider.CallOpts, structHash)
}

// HashTypedDataV4CrossChain is a free data retrieval call binding the contract method 0x518dcf3b.
//
// Solidity: function hashTypedDataV4CrossChain(bytes32 structHash) view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) HashTypedDataV4CrossChain(opts *bind.CallOpts, structHash [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "hashTypedDataV4CrossChain", structHash)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashTypedDataV4CrossChain is a free data retrieval call binding the contract method 0x518dcf3b.
//
// Solidity: function hashTypedDataV4CrossChain(bytes32 structHash) view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderSession) HashTypedDataV4CrossChain(structHash [32]byte) ([32]byte, error) {
	return _IVotingPowerProvider.Contract.HashTypedDataV4CrossChain(&_IVotingPowerProvider.CallOpts, structHash)
}

// HashTypedDataV4CrossChain is a free data retrieval call binding the contract method 0x518dcf3b.
//
// Solidity: function hashTypedDataV4CrossChain(bytes32 structHash) view returns(bytes32)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) HashTypedDataV4CrossChain(structHash [32]byte) ([32]byte, error) {
	return _IVotingPowerProvider.Contract.HashTypedDataV4CrossChain(&_IVotingPowerProvider.CallOpts, structHash)
}

// IsOperatorRegistered is a free data retrieval call binding the contract method 0x6b1906f8.
//
// Solidity: function isOperatorRegistered(address operator) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsOperatorRegistered(opts *bind.CallOpts, operator common.Address) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isOperatorRegistered", operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorRegistered is a free data retrieval call binding the contract method 0x6b1906f8.
//
// Solidity: function isOperatorRegistered(address operator) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsOperatorRegistered(operator common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorRegistered(&_IVotingPowerProvider.CallOpts, operator)
}

// IsOperatorRegistered is a free data retrieval call binding the contract method 0x6b1906f8.
//
// Solidity: function isOperatorRegistered(address operator) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsOperatorRegistered(operator common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorRegistered(&_IVotingPowerProvider.CallOpts, operator)
}

// IsOperatorRegisteredAt is a free data retrieval call binding the contract method 0xf85ccf07.
//
// Solidity: function isOperatorRegisteredAt(address operator, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsOperatorRegisteredAt(opts *bind.CallOpts, operator common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isOperatorRegisteredAt", operator, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorRegisteredAt is a free data retrieval call binding the contract method 0xf85ccf07.
//
// Solidity: function isOperatorRegisteredAt(address operator, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsOperatorRegisteredAt(operator common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorRegisteredAt(&_IVotingPowerProvider.CallOpts, operator, timestamp, hint)
}

// IsOperatorRegisteredAt is a free data retrieval call binding the contract method 0xf85ccf07.
//
// Solidity: function isOperatorRegisteredAt(address operator, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsOperatorRegisteredAt(operator common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorRegisteredAt(&_IVotingPowerProvider.CallOpts, operator, timestamp, hint)
}

// IsOperatorVaultRegistered is a free data retrieval call binding the contract method 0x0f6e0743.
//
// Solidity: function isOperatorVaultRegistered(address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsOperatorVaultRegistered(opts *bind.CallOpts, vault common.Address) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isOperatorVaultRegistered", vault)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorVaultRegistered is a free data retrieval call binding the contract method 0x0f6e0743.
//
// Solidity: function isOperatorVaultRegistered(address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsOperatorVaultRegistered(vault common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegistered(&_IVotingPowerProvider.CallOpts, vault)
}

// IsOperatorVaultRegistered is a free data retrieval call binding the contract method 0x0f6e0743.
//
// Solidity: function isOperatorVaultRegistered(address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsOperatorVaultRegistered(vault common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegistered(&_IVotingPowerProvider.CallOpts, vault)
}

// IsOperatorVaultRegistered0 is a free data retrieval call binding the contract method 0x669fa8c7.
//
// Solidity: function isOperatorVaultRegistered(address operator, address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsOperatorVaultRegistered0(opts *bind.CallOpts, operator common.Address, vault common.Address) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isOperatorVaultRegistered0", operator, vault)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorVaultRegistered0 is a free data retrieval call binding the contract method 0x669fa8c7.
//
// Solidity: function isOperatorVaultRegistered(address operator, address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsOperatorVaultRegistered0(operator common.Address, vault common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegistered0(&_IVotingPowerProvider.CallOpts, operator, vault)
}

// IsOperatorVaultRegistered0 is a free data retrieval call binding the contract method 0x669fa8c7.
//
// Solidity: function isOperatorVaultRegistered(address operator, address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsOperatorVaultRegistered0(operator common.Address, vault common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegistered0(&_IVotingPowerProvider.CallOpts, operator, vault)
}

// IsOperatorVaultRegisteredAt is a free data retrieval call binding the contract method 0x713b524c.
//
// Solidity: function isOperatorVaultRegisteredAt(address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsOperatorVaultRegisteredAt(opts *bind.CallOpts, vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isOperatorVaultRegisteredAt", vault, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorVaultRegisteredAt is a free data retrieval call binding the contract method 0x713b524c.
//
// Solidity: function isOperatorVaultRegisteredAt(address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsOperatorVaultRegisteredAt(vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegisteredAt(&_IVotingPowerProvider.CallOpts, vault, timestamp, hint)
}

// IsOperatorVaultRegisteredAt is a free data retrieval call binding the contract method 0x713b524c.
//
// Solidity: function isOperatorVaultRegisteredAt(address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsOperatorVaultRegisteredAt(vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegisteredAt(&_IVotingPowerProvider.CallOpts, vault, timestamp, hint)
}

// IsOperatorVaultRegisteredAt0 is a free data retrieval call binding the contract method 0x846f139e.
//
// Solidity: function isOperatorVaultRegisteredAt(address operator, address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsOperatorVaultRegisteredAt0(opts *bind.CallOpts, operator common.Address, vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isOperatorVaultRegisteredAt0", operator, vault, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorVaultRegisteredAt0 is a free data retrieval call binding the contract method 0x846f139e.
//
// Solidity: function isOperatorVaultRegisteredAt(address operator, address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsOperatorVaultRegisteredAt0(operator common.Address, vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegisteredAt0(&_IVotingPowerProvider.CallOpts, operator, vault, timestamp, hint)
}

// IsOperatorVaultRegisteredAt0 is a free data retrieval call binding the contract method 0x846f139e.
//
// Solidity: function isOperatorVaultRegisteredAt(address operator, address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsOperatorVaultRegisteredAt0(operator common.Address, vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsOperatorVaultRegisteredAt0(&_IVotingPowerProvider.CallOpts, operator, vault, timestamp, hint)
}

// IsSharedVaultRegistered is a free data retrieval call binding the contract method 0x9a1ebee9.
//
// Solidity: function isSharedVaultRegistered(address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsSharedVaultRegistered(opts *bind.CallOpts, vault common.Address) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isSharedVaultRegistered", vault)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSharedVaultRegistered is a free data retrieval call binding the contract method 0x9a1ebee9.
//
// Solidity: function isSharedVaultRegistered(address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsSharedVaultRegistered(vault common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsSharedVaultRegistered(&_IVotingPowerProvider.CallOpts, vault)
}

// IsSharedVaultRegistered is a free data retrieval call binding the contract method 0x9a1ebee9.
//
// Solidity: function isSharedVaultRegistered(address vault) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsSharedVaultRegistered(vault common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsSharedVaultRegistered(&_IVotingPowerProvider.CallOpts, vault)
}

// IsSharedVaultRegisteredAt is a free data retrieval call binding the contract method 0xb827209d.
//
// Solidity: function isSharedVaultRegisteredAt(address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsSharedVaultRegisteredAt(opts *bind.CallOpts, vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isSharedVaultRegisteredAt", vault, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSharedVaultRegisteredAt is a free data retrieval call binding the contract method 0xb827209d.
//
// Solidity: function isSharedVaultRegisteredAt(address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsSharedVaultRegisteredAt(vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsSharedVaultRegisteredAt(&_IVotingPowerProvider.CallOpts, vault, timestamp, hint)
}

// IsSharedVaultRegisteredAt is a free data retrieval call binding the contract method 0xb827209d.
//
// Solidity: function isSharedVaultRegisteredAt(address vault, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsSharedVaultRegisteredAt(vault common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsSharedVaultRegisteredAt(&_IVotingPowerProvider.CallOpts, vault, timestamp, hint)
}

// IsTokenRegistered is a free data retrieval call binding the contract method 0x26aa101f.
//
// Solidity: function isTokenRegistered(address token) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsTokenRegistered(opts *bind.CallOpts, token common.Address) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isTokenRegistered", token)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTokenRegistered is a free data retrieval call binding the contract method 0x26aa101f.
//
// Solidity: function isTokenRegistered(address token) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsTokenRegistered(token common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsTokenRegistered(&_IVotingPowerProvider.CallOpts, token)
}

// IsTokenRegistered is a free data retrieval call binding the contract method 0x26aa101f.
//
// Solidity: function isTokenRegistered(address token) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsTokenRegistered(token common.Address) (bool, error) {
	return _IVotingPowerProvider.Contract.IsTokenRegistered(&_IVotingPowerProvider.CallOpts, token)
}

// IsTokenRegisteredAt is a free data retrieval call binding the contract method 0x1a416e7e.
//
// Solidity: function isTokenRegisteredAt(address token, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) IsTokenRegisteredAt(opts *bind.CallOpts, token common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "isTokenRegisteredAt", token, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTokenRegisteredAt is a free data retrieval call binding the contract method 0x1a416e7e.
//
// Solidity: function isTokenRegisteredAt(address token, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderSession) IsTokenRegisteredAt(token common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsTokenRegisteredAt(&_IVotingPowerProvider.CallOpts, token, timestamp, hint)
}

// IsTokenRegisteredAt is a free data retrieval call binding the contract method 0x1a416e7e.
//
// Solidity: function isTokenRegisteredAt(address token, uint48 timestamp, bytes hint) view returns(bool)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) IsTokenRegisteredAt(token common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _IVotingPowerProvider.Contract.IsTokenRegisteredAt(&_IVotingPowerProvider.CallOpts, token, timestamp, hint)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) Nonces(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "nonces", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderSession) Nonces(owner common.Address) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.Nonces(&_IVotingPowerProvider.CallOpts, owner)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) Nonces(owner common.Address) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.Nonces(&_IVotingPowerProvider.CallOpts, owner)
}

// StakeToVotingPower is a free data retrieval call binding the contract method 0x039b8dd0.
//
// Solidity: function stakeToVotingPower(address vault, uint256 stake, bytes extraData) view returns(uint256 power)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) StakeToVotingPower(opts *bind.CallOpts, vault common.Address, stake *big.Int, extraData []byte) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "stakeToVotingPower", vault, stake, extraData)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeToVotingPower is a free data retrieval call binding the contract method 0x039b8dd0.
//
// Solidity: function stakeToVotingPower(address vault, uint256 stake, bytes extraData) view returns(uint256 power)
func (_IVotingPowerProvider *IVotingPowerProviderSession) StakeToVotingPower(vault common.Address, stake *big.Int, extraData []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.StakeToVotingPower(&_IVotingPowerProvider.CallOpts, vault, stake, extraData)
}

// StakeToVotingPower is a free data retrieval call binding the contract method 0x039b8dd0.
//
// Solidity: function stakeToVotingPower(address vault, uint256 stake, bytes extraData) view returns(uint256 power)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) StakeToVotingPower(vault common.Address, stake *big.Int, extraData []byte) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.StakeToVotingPower(&_IVotingPowerProvider.CallOpts, vault, stake, extraData)
}

// StakeToVotingPowerAt is a free data retrieval call binding the contract method 0x52936362.
//
// Solidity: function stakeToVotingPowerAt(address vault, uint256 stake, bytes extraData, uint48 timestamp) view returns(uint256 power)
func (_IVotingPowerProvider *IVotingPowerProviderCaller) StakeToVotingPowerAt(opts *bind.CallOpts, vault common.Address, stake *big.Int, extraData []byte, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _IVotingPowerProvider.contract.Call(opts, &out, "stakeToVotingPowerAt", vault, stake, extraData, timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeToVotingPowerAt is a free data retrieval call binding the contract method 0x52936362.
//
// Solidity: function stakeToVotingPowerAt(address vault, uint256 stake, bytes extraData, uint48 timestamp) view returns(uint256 power)
func (_IVotingPowerProvider *IVotingPowerProviderSession) StakeToVotingPowerAt(vault common.Address, stake *big.Int, extraData []byte, timestamp *big.Int) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.StakeToVotingPowerAt(&_IVotingPowerProvider.CallOpts, vault, stake, extraData, timestamp)
}

// StakeToVotingPowerAt is a free data retrieval call binding the contract method 0x52936362.
//
// Solidity: function stakeToVotingPowerAt(address vault, uint256 stake, bytes extraData, uint48 timestamp) view returns(uint256 power)
func (_IVotingPowerProvider *IVotingPowerProviderCallerSession) StakeToVotingPowerAt(vault common.Address, stake *big.Int, extraData []byte, timestamp *big.Int) (*big.Int, error) {
	return _IVotingPowerProvider.Contract.StakeToVotingPowerAt(&_IVotingPowerProvider.CallOpts, vault, stake, extraData, timestamp)
}

// InvalidateOldSignatures is a paid mutator transaction binding the contract method 0x622e4dba.
//
// Solidity: function invalidateOldSignatures() returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactor) InvalidateOldSignatures(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IVotingPowerProvider.contract.Transact(opts, "invalidateOldSignatures")
}

// InvalidateOldSignatures is a paid mutator transaction binding the contract method 0x622e4dba.
//
// Solidity: function invalidateOldSignatures() returns()
func (_IVotingPowerProvider *IVotingPowerProviderSession) InvalidateOldSignatures() (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.InvalidateOldSignatures(&_IVotingPowerProvider.TransactOpts)
}

// InvalidateOldSignatures is a paid mutator transaction binding the contract method 0x622e4dba.
//
// Solidity: function invalidateOldSignatures() returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactorSession) InvalidateOldSignatures() (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.InvalidateOldSignatures(&_IVotingPowerProvider.TransactOpts)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x2acde098.
//
// Solidity: function registerOperator() returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactor) RegisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IVotingPowerProvider.contract.Transact(opts, "registerOperator")
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x2acde098.
//
// Solidity: function registerOperator() returns()
func (_IVotingPowerProvider *IVotingPowerProviderSession) RegisterOperator() (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.RegisterOperator(&_IVotingPowerProvider.TransactOpts)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x2acde098.
//
// Solidity: function registerOperator() returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactorSession) RegisterOperator() (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.RegisterOperator(&_IVotingPowerProvider.TransactOpts)
}

// RegisterOperatorWithSignature is a paid mutator transaction binding the contract method 0xeb5e940d.
//
// Solidity: function registerOperatorWithSignature(address operator, bytes signature) returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactor) RegisterOperatorWithSignature(opts *bind.TransactOpts, operator common.Address, signature []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.contract.Transact(opts, "registerOperatorWithSignature", operator, signature)
}

// RegisterOperatorWithSignature is a paid mutator transaction binding the contract method 0xeb5e940d.
//
// Solidity: function registerOperatorWithSignature(address operator, bytes signature) returns()
func (_IVotingPowerProvider *IVotingPowerProviderSession) RegisterOperatorWithSignature(operator common.Address, signature []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.RegisterOperatorWithSignature(&_IVotingPowerProvider.TransactOpts, operator, signature)
}

// RegisterOperatorWithSignature is a paid mutator transaction binding the contract method 0xeb5e940d.
//
// Solidity: function registerOperatorWithSignature(address operator, bytes signature) returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactorSession) RegisterOperatorWithSignature(operator common.Address, signature []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.RegisterOperatorWithSignature(&_IVotingPowerProvider.TransactOpts, operator, signature)
}

// StaticDelegateCall is a paid mutator transaction binding the contract method 0x9f86fd85.
//
// Solidity: function staticDelegateCall(address target, bytes data) returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactor) StaticDelegateCall(opts *bind.TransactOpts, target common.Address, data []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.contract.Transact(opts, "staticDelegateCall", target, data)
}

// StaticDelegateCall is a paid mutator transaction binding the contract method 0x9f86fd85.
//
// Solidity: function staticDelegateCall(address target, bytes data) returns()
func (_IVotingPowerProvider *IVotingPowerProviderSession) StaticDelegateCall(target common.Address, data []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.StaticDelegateCall(&_IVotingPowerProvider.TransactOpts, target, data)
}

// StaticDelegateCall is a paid mutator transaction binding the contract method 0x9f86fd85.
//
// Solidity: function staticDelegateCall(address target, bytes data) returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactorSession) StaticDelegateCall(target common.Address, data []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.StaticDelegateCall(&_IVotingPowerProvider.TransactOpts, target, data)
}

// UnregisterOperator is a paid mutator transaction binding the contract method 0xa876b89a.
//
// Solidity: function unregisterOperator() returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactor) UnregisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IVotingPowerProvider.contract.Transact(opts, "unregisterOperator")
}

// UnregisterOperator is a paid mutator transaction binding the contract method 0xa876b89a.
//
// Solidity: function unregisterOperator() returns()
func (_IVotingPowerProvider *IVotingPowerProviderSession) UnregisterOperator() (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.UnregisterOperator(&_IVotingPowerProvider.TransactOpts)
}

// UnregisterOperator is a paid mutator transaction binding the contract method 0xa876b89a.
//
// Solidity: function unregisterOperator() returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactorSession) UnregisterOperator() (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.UnregisterOperator(&_IVotingPowerProvider.TransactOpts)
}

// UnregisterOperatorWithSignature is a paid mutator transaction binding the contract method 0xf96d1946.
//
// Solidity: function unregisterOperatorWithSignature(address operator, bytes signature) returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactor) UnregisterOperatorWithSignature(opts *bind.TransactOpts, operator common.Address, signature []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.contract.Transact(opts, "unregisterOperatorWithSignature", operator, signature)
}

// UnregisterOperatorWithSignature is a paid mutator transaction binding the contract method 0xf96d1946.
//
// Solidity: function unregisterOperatorWithSignature(address operator, bytes signature) returns()
func (_IVotingPowerProvider *IVotingPowerProviderSession) UnregisterOperatorWithSignature(operator common.Address, signature []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.UnregisterOperatorWithSignature(&_IVotingPowerProvider.TransactOpts, operator, signature)
}

// UnregisterOperatorWithSignature is a paid mutator transaction binding the contract method 0xf96d1946.
//
// Solidity: function unregisterOperatorWithSignature(address operator, bytes signature) returns()
func (_IVotingPowerProvider *IVotingPowerProviderTransactorSession) UnregisterOperatorWithSignature(operator common.Address, signature []byte) (*types.Transaction, error) {
	return _IVotingPowerProvider.Contract.UnregisterOperatorWithSignature(&_IVotingPowerProvider.TransactOpts, operator, signature)
}

// IVotingPowerProviderEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderEIP712DomainChangedIterator struct {
	Event *IVotingPowerProviderEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderEIP712DomainChanged)
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
		it.Event = new(IVotingPowerProviderEIP712DomainChanged)
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
func (it *IVotingPowerProviderEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderEIP712DomainChanged represents a EIP712DomainChanged event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*IVotingPowerProviderEIP712DomainChangedIterator, error) {

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderEIP712DomainChangedIterator{contract: _IVotingPowerProvider.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderEIP712DomainChanged)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseEIP712DomainChanged(log types.Log) (*IVotingPowerProviderEIP712DomainChanged, error) {
	event := new(IVotingPowerProviderEIP712DomainChanged)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderInitEIP712Iterator is returned from FilterInitEIP712 and is used to iterate over the raw logs and unpacked data for InitEIP712 events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderInitEIP712Iterator struct {
	Event *IVotingPowerProviderInitEIP712 // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderInitEIP712Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderInitEIP712)
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
		it.Event = new(IVotingPowerProviderInitEIP712)
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
func (it *IVotingPowerProviderInitEIP712Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderInitEIP712Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderInitEIP712 represents a InitEIP712 event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderInitEIP712 struct {
	Name    string
	Version string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitEIP712 is a free log retrieval operation binding the contract event 0x98790bb3996c909e6f4279ffabdfe70fa6c0d49b8fa04656d6161decfc442e0a.
//
// Solidity: event InitEIP712(string name, string version)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterInitEIP712(opts *bind.FilterOpts) (*IVotingPowerProviderInitEIP712Iterator, error) {

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "InitEIP712")
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderInitEIP712Iterator{contract: _IVotingPowerProvider.contract, event: "InitEIP712", logs: logs, sub: sub}, nil
}

// WatchInitEIP712 is a free log subscription operation binding the contract event 0x98790bb3996c909e6f4279ffabdfe70fa6c0d49b8fa04656d6161decfc442e0a.
//
// Solidity: event InitEIP712(string name, string version)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchInitEIP712(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderInitEIP712) (event.Subscription, error) {

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "InitEIP712")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderInitEIP712)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "InitEIP712", log); err != nil {
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
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseInitEIP712(log types.Log) (*IVotingPowerProviderInitEIP712, error) {
	event := new(IVotingPowerProviderInitEIP712)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "InitEIP712", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderInitSubnetworkIterator is returned from FilterInitSubnetwork and is used to iterate over the raw logs and unpacked data for InitSubnetwork events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderInitSubnetworkIterator struct {
	Event *IVotingPowerProviderInitSubnetwork // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderInitSubnetworkIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderInitSubnetwork)
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
		it.Event = new(IVotingPowerProviderInitSubnetwork)
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
func (it *IVotingPowerProviderInitSubnetworkIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderInitSubnetworkIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderInitSubnetwork represents a InitSubnetwork event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderInitSubnetwork struct {
	Network      common.Address
	SubnetworkID *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterInitSubnetwork is a free log retrieval operation binding the contract event 0x469c2e982e7d76d34cf5d1e72abee29749bb9971942c180e9023cea09f5f8e83.
//
// Solidity: event InitSubnetwork(address network, uint96 subnetworkID)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterInitSubnetwork(opts *bind.FilterOpts) (*IVotingPowerProviderInitSubnetworkIterator, error) {

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "InitSubnetwork")
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderInitSubnetworkIterator{contract: _IVotingPowerProvider.contract, event: "InitSubnetwork", logs: logs, sub: sub}, nil
}

// WatchInitSubnetwork is a free log subscription operation binding the contract event 0x469c2e982e7d76d34cf5d1e72abee29749bb9971942c180e9023cea09f5f8e83.
//
// Solidity: event InitSubnetwork(address network, uint96 subnetworkID)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchInitSubnetwork(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderInitSubnetwork) (event.Subscription, error) {

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "InitSubnetwork")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderInitSubnetwork)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "InitSubnetwork", log); err != nil {
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
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseInitSubnetwork(log types.Log) (*IVotingPowerProviderInitSubnetwork, error) {
	event := new(IVotingPowerProviderInitSubnetwork)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "InitSubnetwork", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderInitializedIterator struct {
	Event *IVotingPowerProviderInitialized // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderInitialized)
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
		it.Event = new(IVotingPowerProviderInitialized)
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
func (it *IVotingPowerProviderInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderInitialized represents a Initialized event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterInitialized(opts *bind.FilterOpts) (*IVotingPowerProviderInitializedIterator, error) {

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderInitializedIterator{contract: _IVotingPowerProvider.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderInitialized) (event.Subscription, error) {

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderInitialized)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseInitialized(log types.Log) (*IVotingPowerProviderInitialized, error) {
	event := new(IVotingPowerProviderInitialized)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderRegisterOperatorIterator is returned from FilterRegisterOperator and is used to iterate over the raw logs and unpacked data for RegisterOperator events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterOperatorIterator struct {
	Event *IVotingPowerProviderRegisterOperator // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderRegisterOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderRegisterOperator)
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
		it.Event = new(IVotingPowerProviderRegisterOperator)
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
func (it *IVotingPowerProviderRegisterOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderRegisterOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderRegisterOperator represents a RegisterOperator event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterOperator struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRegisterOperator is a free log retrieval operation binding the contract event 0xdfd9e0392912bee97777ec588d2ff7ae010ea24202d153a0bff1b30aed643daa.
//
// Solidity: event RegisterOperator(address indexed operator)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterRegisterOperator(opts *bind.FilterOpts, operator []common.Address) (*IVotingPowerProviderRegisterOperatorIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "RegisterOperator", operatorRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderRegisterOperatorIterator{contract: _IVotingPowerProvider.contract, event: "RegisterOperator", logs: logs, sub: sub}, nil
}

// WatchRegisterOperator is a free log subscription operation binding the contract event 0xdfd9e0392912bee97777ec588d2ff7ae010ea24202d153a0bff1b30aed643daa.
//
// Solidity: event RegisterOperator(address indexed operator)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchRegisterOperator(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderRegisterOperator, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "RegisterOperator", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderRegisterOperator)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterOperator", log); err != nil {
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

// ParseRegisterOperator is a log parse operation binding the contract event 0xdfd9e0392912bee97777ec588d2ff7ae010ea24202d153a0bff1b30aed643daa.
//
// Solidity: event RegisterOperator(address indexed operator)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseRegisterOperator(log types.Log) (*IVotingPowerProviderRegisterOperator, error) {
	event := new(IVotingPowerProviderRegisterOperator)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterOperator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderRegisterOperatorVaultIterator is returned from FilterRegisterOperatorVault and is used to iterate over the raw logs and unpacked data for RegisterOperatorVault events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterOperatorVaultIterator struct {
	Event *IVotingPowerProviderRegisterOperatorVault // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderRegisterOperatorVaultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderRegisterOperatorVault)
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
		it.Event = new(IVotingPowerProviderRegisterOperatorVault)
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
func (it *IVotingPowerProviderRegisterOperatorVaultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderRegisterOperatorVaultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderRegisterOperatorVault represents a RegisterOperatorVault event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterOperatorVault struct {
	Operator common.Address
	Vault    common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRegisterOperatorVault is a free log retrieval operation binding the contract event 0x6db8d1ad7903329250db9b7a653d3aa009807c85daa2281a75e063808bceefdc.
//
// Solidity: event RegisterOperatorVault(address indexed operator, address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterRegisterOperatorVault(opts *bind.FilterOpts, operator []common.Address, vault []common.Address) (*IVotingPowerProviderRegisterOperatorVaultIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "RegisterOperatorVault", operatorRule, vaultRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderRegisterOperatorVaultIterator{contract: _IVotingPowerProvider.contract, event: "RegisterOperatorVault", logs: logs, sub: sub}, nil
}

// WatchRegisterOperatorVault is a free log subscription operation binding the contract event 0x6db8d1ad7903329250db9b7a653d3aa009807c85daa2281a75e063808bceefdc.
//
// Solidity: event RegisterOperatorVault(address indexed operator, address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchRegisterOperatorVault(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderRegisterOperatorVault, operator []common.Address, vault []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "RegisterOperatorVault", operatorRule, vaultRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderRegisterOperatorVault)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterOperatorVault", log); err != nil {
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

// ParseRegisterOperatorVault is a log parse operation binding the contract event 0x6db8d1ad7903329250db9b7a653d3aa009807c85daa2281a75e063808bceefdc.
//
// Solidity: event RegisterOperatorVault(address indexed operator, address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseRegisterOperatorVault(log types.Log) (*IVotingPowerProviderRegisterOperatorVault, error) {
	event := new(IVotingPowerProviderRegisterOperatorVault)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterOperatorVault", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderRegisterSharedVaultIterator is returned from FilterRegisterSharedVault and is used to iterate over the raw logs and unpacked data for RegisterSharedVault events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterSharedVaultIterator struct {
	Event *IVotingPowerProviderRegisterSharedVault // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderRegisterSharedVaultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderRegisterSharedVault)
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
		it.Event = new(IVotingPowerProviderRegisterSharedVault)
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
func (it *IVotingPowerProviderRegisterSharedVaultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderRegisterSharedVaultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderRegisterSharedVault represents a RegisterSharedVault event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterSharedVault struct {
	Vault common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRegisterSharedVault is a free log retrieval operation binding the contract event 0x99528065e654d6d4b95447d6787148a84b7e98a95e752784e99da056b403b25c.
//
// Solidity: event RegisterSharedVault(address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterRegisterSharedVault(opts *bind.FilterOpts, vault []common.Address) (*IVotingPowerProviderRegisterSharedVaultIterator, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "RegisterSharedVault", vaultRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderRegisterSharedVaultIterator{contract: _IVotingPowerProvider.contract, event: "RegisterSharedVault", logs: logs, sub: sub}, nil
}

// WatchRegisterSharedVault is a free log subscription operation binding the contract event 0x99528065e654d6d4b95447d6787148a84b7e98a95e752784e99da056b403b25c.
//
// Solidity: event RegisterSharedVault(address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchRegisterSharedVault(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderRegisterSharedVault, vault []common.Address) (event.Subscription, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "RegisterSharedVault", vaultRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderRegisterSharedVault)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterSharedVault", log); err != nil {
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

// ParseRegisterSharedVault is a log parse operation binding the contract event 0x99528065e654d6d4b95447d6787148a84b7e98a95e752784e99da056b403b25c.
//
// Solidity: event RegisterSharedVault(address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseRegisterSharedVault(log types.Log) (*IVotingPowerProviderRegisterSharedVault, error) {
	event := new(IVotingPowerProviderRegisterSharedVault)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterSharedVault", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderRegisterTokenIterator is returned from FilterRegisterToken and is used to iterate over the raw logs and unpacked data for RegisterToken events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterTokenIterator struct {
	Event *IVotingPowerProviderRegisterToken // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderRegisterTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderRegisterToken)
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
		it.Event = new(IVotingPowerProviderRegisterToken)
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
func (it *IVotingPowerProviderRegisterTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderRegisterTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderRegisterToken represents a RegisterToken event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderRegisterToken struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRegisterToken is a free log retrieval operation binding the contract event 0xf7fe8023cb2e36bde1d59a88ac5763a8c11be6d25e6819f71bb7e23e5bf0dc16.
//
// Solidity: event RegisterToken(address indexed token)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterRegisterToken(opts *bind.FilterOpts, token []common.Address) (*IVotingPowerProviderRegisterTokenIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "RegisterToken", tokenRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderRegisterTokenIterator{contract: _IVotingPowerProvider.contract, event: "RegisterToken", logs: logs, sub: sub}, nil
}

// WatchRegisterToken is a free log subscription operation binding the contract event 0xf7fe8023cb2e36bde1d59a88ac5763a8c11be6d25e6819f71bb7e23e5bf0dc16.
//
// Solidity: event RegisterToken(address indexed token)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchRegisterToken(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderRegisterToken, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "RegisterToken", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderRegisterToken)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterToken", log); err != nil {
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

// ParseRegisterToken is a log parse operation binding the contract event 0xf7fe8023cb2e36bde1d59a88ac5763a8c11be6d25e6819f71bb7e23e5bf0dc16.
//
// Solidity: event RegisterToken(address indexed token)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseRegisterToken(log types.Log) (*IVotingPowerProviderRegisterToken, error) {
	event := new(IVotingPowerProviderRegisterToken)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "RegisterToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderSetSlashingWindowIterator is returned from FilterSetSlashingWindow and is used to iterate over the raw logs and unpacked data for SetSlashingWindow events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderSetSlashingWindowIterator struct {
	Event *IVotingPowerProviderSetSlashingWindow // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderSetSlashingWindowIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderSetSlashingWindow)
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
		it.Event = new(IVotingPowerProviderSetSlashingWindow)
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
func (it *IVotingPowerProviderSetSlashingWindowIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderSetSlashingWindowIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderSetSlashingWindow represents a SetSlashingWindow event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderSetSlashingWindow struct {
	SlashingWindow *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetSlashingWindow is a free log retrieval operation binding the contract event 0x43a1beccb3450080dfa04b4af037ea2115e75a5749538480b77e933bdbff91f1.
//
// Solidity: event SetSlashingWindow(uint48 slashingWindow)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterSetSlashingWindow(opts *bind.FilterOpts) (*IVotingPowerProviderSetSlashingWindowIterator, error) {

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "SetSlashingWindow")
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderSetSlashingWindowIterator{contract: _IVotingPowerProvider.contract, event: "SetSlashingWindow", logs: logs, sub: sub}, nil
}

// WatchSetSlashingWindow is a free log subscription operation binding the contract event 0x43a1beccb3450080dfa04b4af037ea2115e75a5749538480b77e933bdbff91f1.
//
// Solidity: event SetSlashingWindow(uint48 slashingWindow)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchSetSlashingWindow(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderSetSlashingWindow) (event.Subscription, error) {

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "SetSlashingWindow")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderSetSlashingWindow)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "SetSlashingWindow", log); err != nil {
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

// ParseSetSlashingWindow is a log parse operation binding the contract event 0x43a1beccb3450080dfa04b4af037ea2115e75a5749538480b77e933bdbff91f1.
//
// Solidity: event SetSlashingWindow(uint48 slashingWindow)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseSetSlashingWindow(log types.Log) (*IVotingPowerProviderSetSlashingWindow, error) {
	event := new(IVotingPowerProviderSetSlashingWindow)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "SetSlashingWindow", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderUnregisterOperatorIterator is returned from FilterUnregisterOperator and is used to iterate over the raw logs and unpacked data for UnregisterOperator events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterOperatorIterator struct {
	Event *IVotingPowerProviderUnregisterOperator // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderUnregisterOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderUnregisterOperator)
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
		it.Event = new(IVotingPowerProviderUnregisterOperator)
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
func (it *IVotingPowerProviderUnregisterOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderUnregisterOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderUnregisterOperator represents a UnregisterOperator event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterOperator struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUnregisterOperator is a free log retrieval operation binding the contract event 0xd1b48d1e49885298af5dc8adc7777836ef804b38af88eabf4e079c04ee1538a7.
//
// Solidity: event UnregisterOperator(address indexed operator)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterUnregisterOperator(opts *bind.FilterOpts, operator []common.Address) (*IVotingPowerProviderUnregisterOperatorIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "UnregisterOperator", operatorRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderUnregisterOperatorIterator{contract: _IVotingPowerProvider.contract, event: "UnregisterOperator", logs: logs, sub: sub}, nil
}

// WatchUnregisterOperator is a free log subscription operation binding the contract event 0xd1b48d1e49885298af5dc8adc7777836ef804b38af88eabf4e079c04ee1538a7.
//
// Solidity: event UnregisterOperator(address indexed operator)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchUnregisterOperator(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderUnregisterOperator, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "UnregisterOperator", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderUnregisterOperator)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterOperator", log); err != nil {
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

// ParseUnregisterOperator is a log parse operation binding the contract event 0xd1b48d1e49885298af5dc8adc7777836ef804b38af88eabf4e079c04ee1538a7.
//
// Solidity: event UnregisterOperator(address indexed operator)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseUnregisterOperator(log types.Log) (*IVotingPowerProviderUnregisterOperator, error) {
	event := new(IVotingPowerProviderUnregisterOperator)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterOperator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderUnregisterOperatorVaultIterator is returned from FilterUnregisterOperatorVault and is used to iterate over the raw logs and unpacked data for UnregisterOperatorVault events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterOperatorVaultIterator struct {
	Event *IVotingPowerProviderUnregisterOperatorVault // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderUnregisterOperatorVaultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderUnregisterOperatorVault)
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
		it.Event = new(IVotingPowerProviderUnregisterOperatorVault)
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
func (it *IVotingPowerProviderUnregisterOperatorVaultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderUnregisterOperatorVaultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderUnregisterOperatorVault represents a UnregisterOperatorVault event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterOperatorVault struct {
	Operator common.Address
	Vault    common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUnregisterOperatorVault is a free log retrieval operation binding the contract event 0x3455b6128675eff843703027879cc9b52d6ce684ddc6077cbe0d191ad98b255e.
//
// Solidity: event UnregisterOperatorVault(address indexed operator, address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterUnregisterOperatorVault(opts *bind.FilterOpts, operator []common.Address, vault []common.Address) (*IVotingPowerProviderUnregisterOperatorVaultIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "UnregisterOperatorVault", operatorRule, vaultRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderUnregisterOperatorVaultIterator{contract: _IVotingPowerProvider.contract, event: "UnregisterOperatorVault", logs: logs, sub: sub}, nil
}

// WatchUnregisterOperatorVault is a free log subscription operation binding the contract event 0x3455b6128675eff843703027879cc9b52d6ce684ddc6077cbe0d191ad98b255e.
//
// Solidity: event UnregisterOperatorVault(address indexed operator, address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchUnregisterOperatorVault(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderUnregisterOperatorVault, operator []common.Address, vault []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "UnregisterOperatorVault", operatorRule, vaultRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderUnregisterOperatorVault)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterOperatorVault", log); err != nil {
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

// ParseUnregisterOperatorVault is a log parse operation binding the contract event 0x3455b6128675eff843703027879cc9b52d6ce684ddc6077cbe0d191ad98b255e.
//
// Solidity: event UnregisterOperatorVault(address indexed operator, address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseUnregisterOperatorVault(log types.Log) (*IVotingPowerProviderUnregisterOperatorVault, error) {
	event := new(IVotingPowerProviderUnregisterOperatorVault)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterOperatorVault", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderUnregisterSharedVaultIterator is returned from FilterUnregisterSharedVault and is used to iterate over the raw logs and unpacked data for UnregisterSharedVault events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterSharedVaultIterator struct {
	Event *IVotingPowerProviderUnregisterSharedVault // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderUnregisterSharedVaultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderUnregisterSharedVault)
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
		it.Event = new(IVotingPowerProviderUnregisterSharedVault)
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
func (it *IVotingPowerProviderUnregisterSharedVaultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderUnregisterSharedVaultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderUnregisterSharedVault represents a UnregisterSharedVault event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterSharedVault struct {
	Vault common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterUnregisterSharedVault is a free log retrieval operation binding the contract event 0xead83f8482d0fa5de2b5c28fb39ee288392076d150db7020e10a92954aea82ee.
//
// Solidity: event UnregisterSharedVault(address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterUnregisterSharedVault(opts *bind.FilterOpts, vault []common.Address) (*IVotingPowerProviderUnregisterSharedVaultIterator, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "UnregisterSharedVault", vaultRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderUnregisterSharedVaultIterator{contract: _IVotingPowerProvider.contract, event: "UnregisterSharedVault", logs: logs, sub: sub}, nil
}

// WatchUnregisterSharedVault is a free log subscription operation binding the contract event 0xead83f8482d0fa5de2b5c28fb39ee288392076d150db7020e10a92954aea82ee.
//
// Solidity: event UnregisterSharedVault(address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchUnregisterSharedVault(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderUnregisterSharedVault, vault []common.Address) (event.Subscription, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "UnregisterSharedVault", vaultRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderUnregisterSharedVault)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterSharedVault", log); err != nil {
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

// ParseUnregisterSharedVault is a log parse operation binding the contract event 0xead83f8482d0fa5de2b5c28fb39ee288392076d150db7020e10a92954aea82ee.
//
// Solidity: event UnregisterSharedVault(address indexed vault)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseUnregisterSharedVault(log types.Log) (*IVotingPowerProviderUnregisterSharedVault, error) {
	event := new(IVotingPowerProviderUnregisterSharedVault)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterSharedVault", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IVotingPowerProviderUnregisterTokenIterator is returned from FilterUnregisterToken and is used to iterate over the raw logs and unpacked data for UnregisterToken events raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterTokenIterator struct {
	Event *IVotingPowerProviderUnregisterToken // Event containing the contract specifics and raw log

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
func (it *IVotingPowerProviderUnregisterTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IVotingPowerProviderUnregisterToken)
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
		it.Event = new(IVotingPowerProviderUnregisterToken)
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
func (it *IVotingPowerProviderUnregisterTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IVotingPowerProviderUnregisterTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IVotingPowerProviderUnregisterToken represents a UnregisterToken event raised by the IVotingPowerProvider contract.
type IVotingPowerProviderUnregisterToken struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterUnregisterToken is a free log retrieval operation binding the contract event 0xca2a890939276223a9122217752c67608466faee388aff53f077d00a186a389b.
//
// Solidity: event UnregisterToken(address indexed token)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) FilterUnregisterToken(opts *bind.FilterOpts, token []common.Address) (*IVotingPowerProviderUnregisterTokenIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.FilterLogs(opts, "UnregisterToken", tokenRule)
	if err != nil {
		return nil, err
	}
	return &IVotingPowerProviderUnregisterTokenIterator{contract: _IVotingPowerProvider.contract, event: "UnregisterToken", logs: logs, sub: sub}, nil
}

// WatchUnregisterToken is a free log subscription operation binding the contract event 0xca2a890939276223a9122217752c67608466faee388aff53f077d00a186a389b.
//
// Solidity: event UnregisterToken(address indexed token)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) WatchUnregisterToken(opts *bind.WatchOpts, sink chan<- *IVotingPowerProviderUnregisterToken, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _IVotingPowerProvider.contract.WatchLogs(opts, "UnregisterToken", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IVotingPowerProviderUnregisterToken)
				if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterToken", log); err != nil {
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

// ParseUnregisterToken is a log parse operation binding the contract event 0xca2a890939276223a9122217752c67608466faee388aff53f077d00a186a389b.
//
// Solidity: event UnregisterToken(address indexed token)
func (_IVotingPowerProvider *IVotingPowerProviderFilterer) ParseUnregisterToken(log types.Log) (*IVotingPowerProviderUnregisterToken, error) {
	event := new(IVotingPowerProviderUnregisterToken)
	if err := _IVotingPowerProvider.contract.UnpackLog(event, "UnregisterToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
