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

// IOpNetVaultAutoDeployAutoDeployConfig is an auto generated low-level Go binding around an user-defined struct.
type IOpNetVaultAutoDeployAutoDeployConfig struct {
	EpochDuration *big.Int
	Collateral    common.Address
	Burner        common.Address
	WithSlasher   bool
	IsBurnerHook  bool
}

// IVaultInitParams is an auto generated low-level Go binding around an user-defined struct.
type IVaultInitParams struct {
	Collateral                    common.Address
	Burner                        common.Address
	EpochDuration                 *big.Int
	DepositWhitelist              bool
	IsDepositLimit                bool
	DepositLimit                  *big.Int
	DefaultAdminRoleHolder        common.Address
	DepositWhitelistSetRoleHolder common.Address
	DepositorWhitelistRoleHolder  common.Address
	IsDepositLimitSetRoleHolder   common.Address
	DepositLimitSetRoleHolder     common.Address
}

// OpNetVaultAutoDeployLogicMetaData contains all meta data concerning the OpNetVaultAutoDeployLogic contract.
var OpNetVaultAutoDeployLogicMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"_validateConfig\",\"inputs\":[{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structIOpNetVaultAutoDeploy.AutoDeployConfig\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withSlasher\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAutoDeployConfig\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIOpNetVaultAutoDeploy.AutoDeployConfig\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withSlasher\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAutoDeployedVault\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getDelegatorParams\",\"inputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIOpNetVaultAutoDeploy.AutoDeployConfig\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withSlasher\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}]},{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorNetworkSpecificDelegatorParams\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"defaultAdminRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"hook\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"hookSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSlasherParams\",\"inputs\":[{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSlasherParams\",\"inputs\":[{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structIOpNetVaultAutoDeploy.AutoDeployConfig\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withSlasher\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVaultParams\",\"inputs\":[{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structIOpNetVaultAutoDeploy.AutoDeployConfig\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withSlasher\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVaultParams\",\"inputs\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structIVault.InitParams\",\"components\":[{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"depositWhitelist\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isDepositLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"depositLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"defaultAdminRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"depositWhitelistSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"depositorWhitelistRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isDepositLimitSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"depositLimitSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVaultTokenizedParams\",\"inputs\":[{\"name\":\"baseParams\",\"type\":\"tuple\",\"internalType\":\"structIVault.InitParams\",\"components\":[{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"depositWhitelist\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isDepositLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"depositLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"defaultAdminRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"depositWhitelistSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"depositorWhitelistRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isDepositLimitSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"depositLimitSetRoleHolder\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"symbol\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVetoSlasherParams\",\"inputs\":[{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"vetoDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"resolverSetEpochsDelay\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isAutoDeployEnabled\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isSetMaxNetworkLimitHookEnabled\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"SetAutoDeployConfig\",\"inputs\":[{\"name\":\"config\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIOpNetVaultAutoDeploy.AutoDeployConfig\",\"components\":[{\"name\":\"epochDuration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"collateral\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"burner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withSlasher\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isBurnerHook\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAutoDeployStatus\",\"inputs\":[{\"name\":\"status\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetSetMaxNetworkLimitHookStatus\",\"inputs\":[{\"name\":\"status\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"OpNetVaultAutoDeploy_InvalidBurnerHook\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OpNetVaultAutoDeploy_InvalidCollateral\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OpNetVaultAutoDeploy_InvalidEpochDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OpNetVaultAutoDeploy_InvalidWithSlasher\",\"inputs\":[]}]",
}

// OpNetVaultAutoDeployLogicABI is the input ABI used to generate the binding from.
// Deprecated: Use OpNetVaultAutoDeployLogicMetaData.ABI instead.
var OpNetVaultAutoDeployLogicABI = OpNetVaultAutoDeployLogicMetaData.ABI

// OpNetVaultAutoDeployLogic is an auto generated Go binding around an Ethereum contract.
type OpNetVaultAutoDeployLogic struct {
	OpNetVaultAutoDeployLogicCaller     // Read-only binding to the contract
	OpNetVaultAutoDeployLogicTransactor // Write-only binding to the contract
	OpNetVaultAutoDeployLogicFilterer   // Log filterer for contract events
}

// OpNetVaultAutoDeployLogicCaller is an auto generated read-only Go binding around an Ethereum contract.
type OpNetVaultAutoDeployLogicCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OpNetVaultAutoDeployLogicTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OpNetVaultAutoDeployLogicTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OpNetVaultAutoDeployLogicFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OpNetVaultAutoDeployLogicFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OpNetVaultAutoDeployLogicSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OpNetVaultAutoDeployLogicSession struct {
	Contract     *OpNetVaultAutoDeployLogic // Generic contract binding to set the session for
	CallOpts     bind.CallOpts              // Call options to use throughout this session
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// OpNetVaultAutoDeployLogicCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OpNetVaultAutoDeployLogicCallerSession struct {
	Contract *OpNetVaultAutoDeployLogicCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                    // Call options to use throughout this session
}

// OpNetVaultAutoDeployLogicTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OpNetVaultAutoDeployLogicTransactorSession struct {
	Contract     *OpNetVaultAutoDeployLogicTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// OpNetVaultAutoDeployLogicRaw is an auto generated low-level Go binding around an Ethereum contract.
type OpNetVaultAutoDeployLogicRaw struct {
	Contract *OpNetVaultAutoDeployLogic // Generic contract binding to access the raw methods on
}

// OpNetVaultAutoDeployLogicCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OpNetVaultAutoDeployLogicCallerRaw struct {
	Contract *OpNetVaultAutoDeployLogicCaller // Generic read-only contract binding to access the raw methods on
}

// OpNetVaultAutoDeployLogicTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OpNetVaultAutoDeployLogicTransactorRaw struct {
	Contract *OpNetVaultAutoDeployLogicTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOpNetVaultAutoDeployLogic creates a new instance of OpNetVaultAutoDeployLogic, bound to a specific deployed contract.
func NewOpNetVaultAutoDeployLogic(address common.Address, backend bind.ContractBackend) (*OpNetVaultAutoDeployLogic, error) {
	contract, err := bindOpNetVaultAutoDeployLogic(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogic{OpNetVaultAutoDeployLogicCaller: OpNetVaultAutoDeployLogicCaller{contract: contract}, OpNetVaultAutoDeployLogicTransactor: OpNetVaultAutoDeployLogicTransactor{contract: contract}, OpNetVaultAutoDeployLogicFilterer: OpNetVaultAutoDeployLogicFilterer{contract: contract}}, nil
}

// NewOpNetVaultAutoDeployLogicCaller creates a new read-only instance of OpNetVaultAutoDeployLogic, bound to a specific deployed contract.
func NewOpNetVaultAutoDeployLogicCaller(address common.Address, caller bind.ContractCaller) (*OpNetVaultAutoDeployLogicCaller, error) {
	contract, err := bindOpNetVaultAutoDeployLogic(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogicCaller{contract: contract}, nil
}

// NewOpNetVaultAutoDeployLogicTransactor creates a new write-only instance of OpNetVaultAutoDeployLogic, bound to a specific deployed contract.
func NewOpNetVaultAutoDeployLogicTransactor(address common.Address, transactor bind.ContractTransactor) (*OpNetVaultAutoDeployLogicTransactor, error) {
	contract, err := bindOpNetVaultAutoDeployLogic(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogicTransactor{contract: contract}, nil
}

// NewOpNetVaultAutoDeployLogicFilterer creates a new log filterer instance of OpNetVaultAutoDeployLogic, bound to a specific deployed contract.
func NewOpNetVaultAutoDeployLogicFilterer(address common.Address, filterer bind.ContractFilterer) (*OpNetVaultAutoDeployLogicFilterer, error) {
	contract, err := bindOpNetVaultAutoDeployLogic(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogicFilterer{contract: contract}, nil
}

// bindOpNetVaultAutoDeployLogic binds a generic wrapper to an already deployed contract.
func bindOpNetVaultAutoDeployLogic(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OpNetVaultAutoDeployLogicMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OpNetVaultAutoDeployLogic.Contract.OpNetVaultAutoDeployLogicCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpNetVaultAutoDeployLogic.Contract.OpNetVaultAutoDeployLogicTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OpNetVaultAutoDeployLogic.Contract.OpNetVaultAutoDeployLogicTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OpNetVaultAutoDeployLogic.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpNetVaultAutoDeployLogic.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OpNetVaultAutoDeployLogic.Contract.contract.Transact(opts, method, params...)
}

// ValidateConfig is a free data retrieval call binding the contract method 0x97c1ad57.
//
// Solidity: function _validateConfig((uint48,address,address,bool,bool) config) view returns()
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) ValidateConfig(opts *bind.CallOpts, config IOpNetVaultAutoDeployAutoDeployConfig) error {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "_validateConfig", config)

	if err != nil {
		return err
	}

	return err

}

// ValidateConfig is a free data retrieval call binding the contract method 0x97c1ad57.
//
// Solidity: function _validateConfig((uint48,address,address,bool,bool) config) view returns()
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) ValidateConfig(config IOpNetVaultAutoDeployAutoDeployConfig) error {
	return _OpNetVaultAutoDeployLogic.Contract.ValidateConfig(&_OpNetVaultAutoDeployLogic.CallOpts, config)
}

// ValidateConfig is a free data retrieval call binding the contract method 0x97c1ad57.
//
// Solidity: function _validateConfig((uint48,address,address,bool,bool) config) view returns()
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) ValidateConfig(config IOpNetVaultAutoDeployAutoDeployConfig) error {
	return _OpNetVaultAutoDeployLogic.Contract.ValidateConfig(&_OpNetVaultAutoDeployLogic.CallOpts, config)
}

// GetAutoDeployConfig is a free data retrieval call binding the contract method 0xa149c987.
//
// Solidity: function getAutoDeployConfig() view returns((uint48,address,address,bool,bool))
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetAutoDeployConfig(opts *bind.CallOpts) (IOpNetVaultAutoDeployAutoDeployConfig, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getAutoDeployConfig")

	if err != nil {
		return *new(IOpNetVaultAutoDeployAutoDeployConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(IOpNetVaultAutoDeployAutoDeployConfig)).(*IOpNetVaultAutoDeployAutoDeployConfig)

	return out0, err

}

// GetAutoDeployConfig is a free data retrieval call binding the contract method 0xa149c987.
//
// Solidity: function getAutoDeployConfig() view returns((uint48,address,address,bool,bool))
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetAutoDeployConfig() (IOpNetVaultAutoDeployAutoDeployConfig, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetAutoDeployConfig(&_OpNetVaultAutoDeployLogic.CallOpts)
}

// GetAutoDeployConfig is a free data retrieval call binding the contract method 0xa149c987.
//
// Solidity: function getAutoDeployConfig() view returns((uint48,address,address,bool,bool))
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetAutoDeployConfig() (IOpNetVaultAutoDeployAutoDeployConfig, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetAutoDeployConfig(&_OpNetVaultAutoDeployLogic.CallOpts)
}

// GetAutoDeployedVault is a free data retrieval call binding the contract method 0x2b2fd015.
//
// Solidity: function getAutoDeployedVault(address operator) view returns(address)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetAutoDeployedVault(opts *bind.CallOpts, operator common.Address) (common.Address, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getAutoDeployedVault", operator)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAutoDeployedVault is a free data retrieval call binding the contract method 0x2b2fd015.
//
// Solidity: function getAutoDeployedVault(address operator) view returns(address)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetAutoDeployedVault(operator common.Address) (common.Address, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetAutoDeployedVault(&_OpNetVaultAutoDeployLogic.CallOpts, operator)
}

// GetAutoDeployedVault is a free data retrieval call binding the contract method 0x2b2fd015.
//
// Solidity: function getAutoDeployedVault(address operator) view returns(address)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetAutoDeployedVault(operator common.Address) (common.Address, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetAutoDeployedVault(&_OpNetVaultAutoDeployLogic.CallOpts, operator)
}

// GetDelegatorParams is a free data retrieval call binding the contract method 0x4c0098f2.
//
// Solidity: function getDelegatorParams((uint48,address,address,bool,bool) , address operator) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetDelegatorParams(opts *bind.CallOpts, arg0 IOpNetVaultAutoDeployAutoDeployConfig, operator common.Address) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getDelegatorParams", arg0, operator)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetDelegatorParams is a free data retrieval call binding the contract method 0x4c0098f2.
//
// Solidity: function getDelegatorParams((uint48,address,address,bool,bool) , address operator) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetDelegatorParams(arg0 IOpNetVaultAutoDeployAutoDeployConfig, operator common.Address) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetDelegatorParams(&_OpNetVaultAutoDeployLogic.CallOpts, arg0, operator)
}

// GetDelegatorParams is a free data retrieval call binding the contract method 0x4c0098f2.
//
// Solidity: function getDelegatorParams((uint48,address,address,bool,bool) , address operator) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetDelegatorParams(arg0 IOpNetVaultAutoDeployAutoDeployConfig, operator common.Address) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetDelegatorParams(&_OpNetVaultAutoDeployLogic.CallOpts, arg0, operator)
}

// GetOperatorNetworkSpecificDelegatorParams is a free data retrieval call binding the contract method 0xb21fe3ed.
//
// Solidity: function getOperatorNetworkSpecificDelegatorParams(address operator, address defaultAdminRoleHolder, address hook, address hookSetRoleHolder) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetOperatorNetworkSpecificDelegatorParams(opts *bind.CallOpts, operator common.Address, defaultAdminRoleHolder common.Address, hook common.Address, hookSetRoleHolder common.Address) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getOperatorNetworkSpecificDelegatorParams", operator, defaultAdminRoleHolder, hook, hookSetRoleHolder)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetOperatorNetworkSpecificDelegatorParams is a free data retrieval call binding the contract method 0xb21fe3ed.
//
// Solidity: function getOperatorNetworkSpecificDelegatorParams(address operator, address defaultAdminRoleHolder, address hook, address hookSetRoleHolder) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetOperatorNetworkSpecificDelegatorParams(operator common.Address, defaultAdminRoleHolder common.Address, hook common.Address, hookSetRoleHolder common.Address) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetOperatorNetworkSpecificDelegatorParams(&_OpNetVaultAutoDeployLogic.CallOpts, operator, defaultAdminRoleHolder, hook, hookSetRoleHolder)
}

// GetOperatorNetworkSpecificDelegatorParams is a free data retrieval call binding the contract method 0xb21fe3ed.
//
// Solidity: function getOperatorNetworkSpecificDelegatorParams(address operator, address defaultAdminRoleHolder, address hook, address hookSetRoleHolder) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetOperatorNetworkSpecificDelegatorParams(operator common.Address, defaultAdminRoleHolder common.Address, hook common.Address, hookSetRoleHolder common.Address) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetOperatorNetworkSpecificDelegatorParams(&_OpNetVaultAutoDeployLogic.CallOpts, operator, defaultAdminRoleHolder, hook, hookSetRoleHolder)
}

// GetSlasherParams is a free data retrieval call binding the contract method 0x2b45a2f9.
//
// Solidity: function getSlasherParams(bool isBurnerHook) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetSlasherParams(opts *bind.CallOpts, isBurnerHook bool) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getSlasherParams", isBurnerHook)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetSlasherParams is a free data retrieval call binding the contract method 0x2b45a2f9.
//
// Solidity: function getSlasherParams(bool isBurnerHook) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetSlasherParams(isBurnerHook bool) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetSlasherParams(&_OpNetVaultAutoDeployLogic.CallOpts, isBurnerHook)
}

// GetSlasherParams is a free data retrieval call binding the contract method 0x2b45a2f9.
//
// Solidity: function getSlasherParams(bool isBurnerHook) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetSlasherParams(isBurnerHook bool) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetSlasherParams(&_OpNetVaultAutoDeployLogic.CallOpts, isBurnerHook)
}

// GetSlasherParams0 is a free data retrieval call binding the contract method 0xac296ec2.
//
// Solidity: function getSlasherParams((uint48,address,address,bool,bool) config) view returns(bool, uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetSlasherParams0(opts *bind.CallOpts, config IOpNetVaultAutoDeployAutoDeployConfig) (bool, uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getSlasherParams0", config)

	if err != nil {
		return *new(bool), *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	out1 := *abi.ConvertType(out[1], new(uint64)).(*uint64)
	out2 := *abi.ConvertType(out[2], new([]byte)).(*[]byte)

	return out0, out1, out2, err

}

// GetSlasherParams0 is a free data retrieval call binding the contract method 0xac296ec2.
//
// Solidity: function getSlasherParams((uint48,address,address,bool,bool) config) view returns(bool, uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetSlasherParams0(config IOpNetVaultAutoDeployAutoDeployConfig) (bool, uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetSlasherParams0(&_OpNetVaultAutoDeployLogic.CallOpts, config)
}

// GetSlasherParams0 is a free data retrieval call binding the contract method 0xac296ec2.
//
// Solidity: function getSlasherParams((uint48,address,address,bool,bool) config) view returns(bool, uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetSlasherParams0(config IOpNetVaultAutoDeployAutoDeployConfig) (bool, uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetSlasherParams0(&_OpNetVaultAutoDeployLogic.CallOpts, config)
}

// GetVaultParams is a free data retrieval call binding the contract method 0xea8ea5c7.
//
// Solidity: function getVaultParams((uint48,address,address,bool,bool) config) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetVaultParams(opts *bind.CallOpts, config IOpNetVaultAutoDeployAutoDeployConfig) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getVaultParams", config)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetVaultParams is a free data retrieval call binding the contract method 0xea8ea5c7.
//
// Solidity: function getVaultParams((uint48,address,address,bool,bool) config) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetVaultParams(config IOpNetVaultAutoDeployAutoDeployConfig) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVaultParams(&_OpNetVaultAutoDeployLogic.CallOpts, config)
}

// GetVaultParams is a free data retrieval call binding the contract method 0xea8ea5c7.
//
// Solidity: function getVaultParams((uint48,address,address,bool,bool) config) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetVaultParams(config IOpNetVaultAutoDeployAutoDeployConfig) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVaultParams(&_OpNetVaultAutoDeployLogic.CallOpts, config)
}

// GetVaultParams0 is a free data retrieval call binding the contract method 0x168432da.
//
// Solidity: function getVaultParams((address,address,uint48,bool,bool,uint256,address,address,address,address,address) params) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetVaultParams0(opts *bind.CallOpts, params IVaultInitParams) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getVaultParams0", params)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetVaultParams0 is a free data retrieval call binding the contract method 0x168432da.
//
// Solidity: function getVaultParams((address,address,uint48,bool,bool,uint256,address,address,address,address,address) params) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetVaultParams0(params IVaultInitParams) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVaultParams0(&_OpNetVaultAutoDeployLogic.CallOpts, params)
}

// GetVaultParams0 is a free data retrieval call binding the contract method 0x168432da.
//
// Solidity: function getVaultParams((address,address,uint48,bool,bool,uint256,address,address,address,address,address) params) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetVaultParams0(params IVaultInitParams) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVaultParams0(&_OpNetVaultAutoDeployLogic.CallOpts, params)
}

// GetVaultTokenizedParams is a free data retrieval call binding the contract method 0x49b6cfc6.
//
// Solidity: function getVaultTokenizedParams((address,address,uint48,bool,bool,uint256,address,address,address,address,address) baseParams, string name, string symbol) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetVaultTokenizedParams(opts *bind.CallOpts, baseParams IVaultInitParams, name string, symbol string) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getVaultTokenizedParams", baseParams, name, symbol)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetVaultTokenizedParams is a free data retrieval call binding the contract method 0x49b6cfc6.
//
// Solidity: function getVaultTokenizedParams((address,address,uint48,bool,bool,uint256,address,address,address,address,address) baseParams, string name, string symbol) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetVaultTokenizedParams(baseParams IVaultInitParams, name string, symbol string) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVaultTokenizedParams(&_OpNetVaultAutoDeployLogic.CallOpts, baseParams, name, symbol)
}

// GetVaultTokenizedParams is a free data retrieval call binding the contract method 0x49b6cfc6.
//
// Solidity: function getVaultTokenizedParams((address,address,uint48,bool,bool,uint256,address,address,address,address,address) baseParams, string name, string symbol) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetVaultTokenizedParams(baseParams IVaultInitParams, name string, symbol string) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVaultTokenizedParams(&_OpNetVaultAutoDeployLogic.CallOpts, baseParams, name, symbol)
}

// GetVetoSlasherParams is a free data retrieval call binding the contract method 0xae500700.
//
// Solidity: function getVetoSlasherParams(bool isBurnerHook, uint48 vetoDuration, uint256 resolverSetEpochsDelay) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) GetVetoSlasherParams(opts *bind.CallOpts, isBurnerHook bool, vetoDuration *big.Int, resolverSetEpochsDelay *big.Int) (uint64, []byte, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "getVetoSlasherParams", isBurnerHook, vetoDuration, resolverSetEpochsDelay)

	if err != nil {
		return *new(uint64), *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)
	out1 := *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return out0, out1, err

}

// GetVetoSlasherParams is a free data retrieval call binding the contract method 0xae500700.
//
// Solidity: function getVetoSlasherParams(bool isBurnerHook, uint48 vetoDuration, uint256 resolverSetEpochsDelay) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) GetVetoSlasherParams(isBurnerHook bool, vetoDuration *big.Int, resolverSetEpochsDelay *big.Int) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVetoSlasherParams(&_OpNetVaultAutoDeployLogic.CallOpts, isBurnerHook, vetoDuration, resolverSetEpochsDelay)
}

// GetVetoSlasherParams is a free data retrieval call binding the contract method 0xae500700.
//
// Solidity: function getVetoSlasherParams(bool isBurnerHook, uint48 vetoDuration, uint256 resolverSetEpochsDelay) view returns(uint64, bytes)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) GetVetoSlasherParams(isBurnerHook bool, vetoDuration *big.Int, resolverSetEpochsDelay *big.Int) (uint64, []byte, error) {
	return _OpNetVaultAutoDeployLogic.Contract.GetVetoSlasherParams(&_OpNetVaultAutoDeployLogic.CallOpts, isBurnerHook, vetoDuration, resolverSetEpochsDelay)
}

// IsAutoDeployEnabled is a free data retrieval call binding the contract method 0xdeb018dc.
//
// Solidity: function isAutoDeployEnabled() view returns(bool)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) IsAutoDeployEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "isAutoDeployEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAutoDeployEnabled is a free data retrieval call binding the contract method 0xdeb018dc.
//
// Solidity: function isAutoDeployEnabled() view returns(bool)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) IsAutoDeployEnabled() (bool, error) {
	return _OpNetVaultAutoDeployLogic.Contract.IsAutoDeployEnabled(&_OpNetVaultAutoDeployLogic.CallOpts)
}

// IsAutoDeployEnabled is a free data retrieval call binding the contract method 0xdeb018dc.
//
// Solidity: function isAutoDeployEnabled() view returns(bool)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) IsAutoDeployEnabled() (bool, error) {
	return _OpNetVaultAutoDeployLogic.Contract.IsAutoDeployEnabled(&_OpNetVaultAutoDeployLogic.CallOpts)
}

// IsSetMaxNetworkLimitHookEnabled is a free data retrieval call binding the contract method 0xe77b136d.
//
// Solidity: function isSetMaxNetworkLimitHookEnabled() view returns(bool)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCaller) IsSetMaxNetworkLimitHookEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _OpNetVaultAutoDeployLogic.contract.Call(opts, &out, "isSetMaxNetworkLimitHookEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSetMaxNetworkLimitHookEnabled is a free data retrieval call binding the contract method 0xe77b136d.
//
// Solidity: function isSetMaxNetworkLimitHookEnabled() view returns(bool)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicSession) IsSetMaxNetworkLimitHookEnabled() (bool, error) {
	return _OpNetVaultAutoDeployLogic.Contract.IsSetMaxNetworkLimitHookEnabled(&_OpNetVaultAutoDeployLogic.CallOpts)
}

// IsSetMaxNetworkLimitHookEnabled is a free data retrieval call binding the contract method 0xe77b136d.
//
// Solidity: function isSetMaxNetworkLimitHookEnabled() view returns(bool)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicCallerSession) IsSetMaxNetworkLimitHookEnabled() (bool, error) {
	return _OpNetVaultAutoDeployLogic.Contract.IsSetMaxNetworkLimitHookEnabled(&_OpNetVaultAutoDeployLogic.CallOpts)
}

// OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator is returned from FilterSetAutoDeployConfig and is used to iterate over the raw logs and unpacked data for SetAutoDeployConfig events raised by the OpNetVaultAutoDeployLogic contract.
type OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator struct {
	Event *OpNetVaultAutoDeployLogicSetAutoDeployConfig // Event containing the contract specifics and raw log

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
func (it *OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OpNetVaultAutoDeployLogicSetAutoDeployConfig)
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
		it.Event = new(OpNetVaultAutoDeployLogicSetAutoDeployConfig)
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
func (it *OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OpNetVaultAutoDeployLogicSetAutoDeployConfig represents a SetAutoDeployConfig event raised by the OpNetVaultAutoDeployLogic contract.
type OpNetVaultAutoDeployLogicSetAutoDeployConfig struct {
	Config IOpNetVaultAutoDeployAutoDeployConfig
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSetAutoDeployConfig is a free log retrieval operation binding the contract event 0x77e47da1f6025186b00adae5351f623eba1ab5151f7d15bc44c6a39be86e6c05.
//
// Solidity: event SetAutoDeployConfig((uint48,address,address,bool,bool) config)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) FilterSetAutoDeployConfig(opts *bind.FilterOpts) (*OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator, error) {

	logs, sub, err := _OpNetVaultAutoDeployLogic.contract.FilterLogs(opts, "SetAutoDeployConfig")
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogicSetAutoDeployConfigIterator{contract: _OpNetVaultAutoDeployLogic.contract, event: "SetAutoDeployConfig", logs: logs, sub: sub}, nil
}

// WatchSetAutoDeployConfig is a free log subscription operation binding the contract event 0x77e47da1f6025186b00adae5351f623eba1ab5151f7d15bc44c6a39be86e6c05.
//
// Solidity: event SetAutoDeployConfig((uint48,address,address,bool,bool) config)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) WatchSetAutoDeployConfig(opts *bind.WatchOpts, sink chan<- *OpNetVaultAutoDeployLogicSetAutoDeployConfig) (event.Subscription, error) {

	logs, sub, err := _OpNetVaultAutoDeployLogic.contract.WatchLogs(opts, "SetAutoDeployConfig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OpNetVaultAutoDeployLogicSetAutoDeployConfig)
				if err := _OpNetVaultAutoDeployLogic.contract.UnpackLog(event, "SetAutoDeployConfig", log); err != nil {
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

// ParseSetAutoDeployConfig is a log parse operation binding the contract event 0x77e47da1f6025186b00adae5351f623eba1ab5151f7d15bc44c6a39be86e6c05.
//
// Solidity: event SetAutoDeployConfig((uint48,address,address,bool,bool) config)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) ParseSetAutoDeployConfig(log types.Log) (*OpNetVaultAutoDeployLogicSetAutoDeployConfig, error) {
	event := new(OpNetVaultAutoDeployLogicSetAutoDeployConfig)
	if err := _OpNetVaultAutoDeployLogic.contract.UnpackLog(event, "SetAutoDeployConfig", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator is returned from FilterSetAutoDeployStatus and is used to iterate over the raw logs and unpacked data for SetAutoDeployStatus events raised by the OpNetVaultAutoDeployLogic contract.
type OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator struct {
	Event *OpNetVaultAutoDeployLogicSetAutoDeployStatus // Event containing the contract specifics and raw log

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
func (it *OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OpNetVaultAutoDeployLogicSetAutoDeployStatus)
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
		it.Event = new(OpNetVaultAutoDeployLogicSetAutoDeployStatus)
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
func (it *OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OpNetVaultAutoDeployLogicSetAutoDeployStatus represents a SetAutoDeployStatus event raised by the OpNetVaultAutoDeployLogic contract.
type OpNetVaultAutoDeployLogicSetAutoDeployStatus struct {
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSetAutoDeployStatus is a free log retrieval operation binding the contract event 0x8951c46d8957e007c4f4222e768ee8e59bb367b6c72569e92e337a5b194bf04d.
//
// Solidity: event SetAutoDeployStatus(bool status)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) FilterSetAutoDeployStatus(opts *bind.FilterOpts) (*OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator, error) {

	logs, sub, err := _OpNetVaultAutoDeployLogic.contract.FilterLogs(opts, "SetAutoDeployStatus")
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogicSetAutoDeployStatusIterator{contract: _OpNetVaultAutoDeployLogic.contract, event: "SetAutoDeployStatus", logs: logs, sub: sub}, nil
}

// WatchSetAutoDeployStatus is a free log subscription operation binding the contract event 0x8951c46d8957e007c4f4222e768ee8e59bb367b6c72569e92e337a5b194bf04d.
//
// Solidity: event SetAutoDeployStatus(bool status)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) WatchSetAutoDeployStatus(opts *bind.WatchOpts, sink chan<- *OpNetVaultAutoDeployLogicSetAutoDeployStatus) (event.Subscription, error) {

	logs, sub, err := _OpNetVaultAutoDeployLogic.contract.WatchLogs(opts, "SetAutoDeployStatus")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OpNetVaultAutoDeployLogicSetAutoDeployStatus)
				if err := _OpNetVaultAutoDeployLogic.contract.UnpackLog(event, "SetAutoDeployStatus", log); err != nil {
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

// ParseSetAutoDeployStatus is a log parse operation binding the contract event 0x8951c46d8957e007c4f4222e768ee8e59bb367b6c72569e92e337a5b194bf04d.
//
// Solidity: event SetAutoDeployStatus(bool status)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) ParseSetAutoDeployStatus(log types.Log) (*OpNetVaultAutoDeployLogicSetAutoDeployStatus, error) {
	event := new(OpNetVaultAutoDeployLogicSetAutoDeployStatus)
	if err := _OpNetVaultAutoDeployLogic.contract.UnpackLog(event, "SetAutoDeployStatus", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator is returned from FilterSetSetMaxNetworkLimitHookStatus and is used to iterate over the raw logs and unpacked data for SetSetMaxNetworkLimitHookStatus events raised by the OpNetVaultAutoDeployLogic contract.
type OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator struct {
	Event *OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus // Event containing the contract specifics and raw log

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
func (it *OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus)
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
		it.Event = new(OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus)
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
func (it *OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus represents a SetSetMaxNetworkLimitHookStatus event raised by the OpNetVaultAutoDeployLogic contract.
type OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus struct {
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSetSetMaxNetworkLimitHookStatus is a free log retrieval operation binding the contract event 0x8bd71bb92871c7cb65d4ba7554dadeb02abcf4d9e99aff8367714c5a15bd019c.
//
// Solidity: event SetSetMaxNetworkLimitHookStatus(bool status)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) FilterSetSetMaxNetworkLimitHookStatus(opts *bind.FilterOpts) (*OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator, error) {

	logs, sub, err := _OpNetVaultAutoDeployLogic.contract.FilterLogs(opts, "SetSetMaxNetworkLimitHookStatus")
	if err != nil {
		return nil, err
	}
	return &OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatusIterator{contract: _OpNetVaultAutoDeployLogic.contract, event: "SetSetMaxNetworkLimitHookStatus", logs: logs, sub: sub}, nil
}

// WatchSetSetMaxNetworkLimitHookStatus is a free log subscription operation binding the contract event 0x8bd71bb92871c7cb65d4ba7554dadeb02abcf4d9e99aff8367714c5a15bd019c.
//
// Solidity: event SetSetMaxNetworkLimitHookStatus(bool status)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) WatchSetSetMaxNetworkLimitHookStatus(opts *bind.WatchOpts, sink chan<- *OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus) (event.Subscription, error) {

	logs, sub, err := _OpNetVaultAutoDeployLogic.contract.WatchLogs(opts, "SetSetMaxNetworkLimitHookStatus")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus)
				if err := _OpNetVaultAutoDeployLogic.contract.UnpackLog(event, "SetSetMaxNetworkLimitHookStatus", log); err != nil {
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

// ParseSetSetMaxNetworkLimitHookStatus is a log parse operation binding the contract event 0x8bd71bb92871c7cb65d4ba7554dadeb02abcf4d9e99aff8367714c5a15bd019c.
//
// Solidity: event SetSetMaxNetworkLimitHookStatus(bool status)
func (_OpNetVaultAutoDeployLogic *OpNetVaultAutoDeployLogicFilterer) ParseSetSetMaxNetworkLimitHookStatus(log types.Log) (*OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus, error) {
	event := new(OpNetVaultAutoDeployLogicSetSetMaxNetworkLimitHookStatus)
	if err := _OpNetVaultAutoDeployLogic.contract.UnpackLog(event, "SetSetMaxNetworkLimitHookStatus", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
