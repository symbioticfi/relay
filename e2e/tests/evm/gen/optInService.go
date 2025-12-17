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

// OptInServiceMetaData contains all meta data concerning the OptInService contract.
var OptInServiceMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"WHERE_REGISTRY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"WHO_REGISTRY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"increaseNonce\",\"inputs\":[{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isOptedIn\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOptedInAt\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"hint\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nonces\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"optIn\",\"inputs\":[{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"optIn\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"deadline\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"optOut\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"deadline\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"optOut\",\"inputs\":[{\"name\":\"where\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"IncreaseNonce\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OptIn\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OptOut\",\"inputs\":[{\"name\":\"who\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"where\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AlreadyOptedIn\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpiredSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotOptedIn\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotWhereEntity\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotWho\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OptOutCooldown\",\"inputs\":[]}]",
}

// OptInServiceABI is the input ABI used to generate the binding from.
// Deprecated: Use OptInServiceMetaData.ABI instead.
var OptInServiceABI = OptInServiceMetaData.ABI

// OptInService is an auto generated Go binding around an Ethereum contract.
type OptInService struct {
	OptInServiceCaller     // Read-only binding to the contract
	OptInServiceTransactor // Write-only binding to the contract
	OptInServiceFilterer   // Log filterer for contract events
}

// OptInServiceCaller is an auto generated read-only Go binding around an Ethereum contract.
type OptInServiceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptInServiceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OptInServiceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptInServiceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OptInServiceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptInServiceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OptInServiceSession struct {
	Contract     *OptInService     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OptInServiceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OptInServiceCallerSession struct {
	Contract *OptInServiceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// OptInServiceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OptInServiceTransactorSession struct {
	Contract     *OptInServiceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// OptInServiceRaw is an auto generated low-level Go binding around an Ethereum contract.
type OptInServiceRaw struct {
	Contract *OptInService // Generic contract binding to access the raw methods on
}

// OptInServiceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OptInServiceCallerRaw struct {
	Contract *OptInServiceCaller // Generic read-only contract binding to access the raw methods on
}

// OptInServiceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OptInServiceTransactorRaw struct {
	Contract *OptInServiceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOptInService creates a new instance of OptInService, bound to a specific deployed contract.
func NewOptInService(address common.Address, backend bind.ContractBackend) (*OptInService, error) {
	contract, err := bindOptInService(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OptInService{OptInServiceCaller: OptInServiceCaller{contract: contract}, OptInServiceTransactor: OptInServiceTransactor{contract: contract}, OptInServiceFilterer: OptInServiceFilterer{contract: contract}}, nil
}

// NewOptInServiceCaller creates a new read-only instance of OptInService, bound to a specific deployed contract.
func NewOptInServiceCaller(address common.Address, caller bind.ContractCaller) (*OptInServiceCaller, error) {
	contract, err := bindOptInService(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OptInServiceCaller{contract: contract}, nil
}

// NewOptInServiceTransactor creates a new write-only instance of OptInService, bound to a specific deployed contract.
func NewOptInServiceTransactor(address common.Address, transactor bind.ContractTransactor) (*OptInServiceTransactor, error) {
	contract, err := bindOptInService(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OptInServiceTransactor{contract: contract}, nil
}

// NewOptInServiceFilterer creates a new log filterer instance of OptInService, bound to a specific deployed contract.
func NewOptInServiceFilterer(address common.Address, filterer bind.ContractFilterer) (*OptInServiceFilterer, error) {
	contract, err := bindOptInService(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OptInServiceFilterer{contract: contract}, nil
}

// bindOptInService binds a generic wrapper to an already deployed contract.
func bindOptInService(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OptInServiceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OptInService *OptInServiceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OptInService.Contract.OptInServiceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OptInService *OptInServiceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptInService.Contract.OptInServiceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OptInService *OptInServiceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OptInService.Contract.OptInServiceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OptInService *OptInServiceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OptInService.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OptInService *OptInServiceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptInService.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OptInService *OptInServiceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OptInService.Contract.contract.Transact(opts, method, params...)
}

// WHEREREGISTRY is a free data retrieval call binding the contract method 0x6d0fae4a.
//
// Solidity: function WHERE_REGISTRY() view returns(address)
func (_OptInService *OptInServiceCaller) WHEREREGISTRY(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptInService.contract.Call(opts, &out, "WHERE_REGISTRY")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WHEREREGISTRY is a free data retrieval call binding the contract method 0x6d0fae4a.
//
// Solidity: function WHERE_REGISTRY() view returns(address)
func (_OptInService *OptInServiceSession) WHEREREGISTRY() (common.Address, error) {
	return _OptInService.Contract.WHEREREGISTRY(&_OptInService.CallOpts)
}

// WHEREREGISTRY is a free data retrieval call binding the contract method 0x6d0fae4a.
//
// Solidity: function WHERE_REGISTRY() view returns(address)
func (_OptInService *OptInServiceCallerSession) WHEREREGISTRY() (common.Address, error) {
	return _OptInService.Contract.WHEREREGISTRY(&_OptInService.CallOpts)
}

// WHOREGISTRY is a free data retrieval call binding the contract method 0xc4b5d62c.
//
// Solidity: function WHO_REGISTRY() view returns(address)
func (_OptInService *OptInServiceCaller) WHOREGISTRY(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptInService.contract.Call(opts, &out, "WHO_REGISTRY")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WHOREGISTRY is a free data retrieval call binding the contract method 0xc4b5d62c.
//
// Solidity: function WHO_REGISTRY() view returns(address)
func (_OptInService *OptInServiceSession) WHOREGISTRY() (common.Address, error) {
	return _OptInService.Contract.WHOREGISTRY(&_OptInService.CallOpts)
}

// WHOREGISTRY is a free data retrieval call binding the contract method 0xc4b5d62c.
//
// Solidity: function WHO_REGISTRY() view returns(address)
func (_OptInService *OptInServiceCallerSession) WHOREGISTRY() (common.Address, error) {
	return _OptInService.Contract.WHOREGISTRY(&_OptInService.CallOpts)
}

// IsOptedIn is a free data retrieval call binding the contract method 0x220d32d4.
//
// Solidity: function isOptedIn(address who, address where) view returns(bool)
func (_OptInService *OptInServiceCaller) IsOptedIn(opts *bind.CallOpts, who common.Address, where common.Address) (bool, error) {
	var out []interface{}
	err := _OptInService.contract.Call(opts, &out, "isOptedIn", who, where)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOptedIn is a free data retrieval call binding the contract method 0x220d32d4.
//
// Solidity: function isOptedIn(address who, address where) view returns(bool)
func (_OptInService *OptInServiceSession) IsOptedIn(who common.Address, where common.Address) (bool, error) {
	return _OptInService.Contract.IsOptedIn(&_OptInService.CallOpts, who, where)
}

// IsOptedIn is a free data retrieval call binding the contract method 0x220d32d4.
//
// Solidity: function isOptedIn(address who, address where) view returns(bool)
func (_OptInService *OptInServiceCallerSession) IsOptedIn(who common.Address, where common.Address) (bool, error) {
	return _OptInService.Contract.IsOptedIn(&_OptInService.CallOpts, who, where)
}

// IsOptedInAt is a free data retrieval call binding the contract method 0x530e1d43.
//
// Solidity: function isOptedInAt(address who, address where, uint48 timestamp, bytes hint) view returns(bool)
func (_OptInService *OptInServiceCaller) IsOptedInAt(opts *bind.CallOpts, who common.Address, where common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	var out []interface{}
	err := _OptInService.contract.Call(opts, &out, "isOptedInAt", who, where, timestamp, hint)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOptedInAt is a free data retrieval call binding the contract method 0x530e1d43.
//
// Solidity: function isOptedInAt(address who, address where, uint48 timestamp, bytes hint) view returns(bool)
func (_OptInService *OptInServiceSession) IsOptedInAt(who common.Address, where common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _OptInService.Contract.IsOptedInAt(&_OptInService.CallOpts, who, where, timestamp, hint)
}

// IsOptedInAt is a free data retrieval call binding the contract method 0x530e1d43.
//
// Solidity: function isOptedInAt(address who, address where, uint48 timestamp, bytes hint) view returns(bool)
func (_OptInService *OptInServiceCallerSession) IsOptedInAt(who common.Address, where common.Address, timestamp *big.Int, hint []byte) (bool, error) {
	return _OptInService.Contract.IsOptedInAt(&_OptInService.CallOpts, who, where, timestamp, hint)
}

// Nonces is a free data retrieval call binding the contract method 0x9333fbda.
//
// Solidity: function nonces(address who, address where) view returns(uint256)
func (_OptInService *OptInServiceCaller) Nonces(opts *bind.CallOpts, who common.Address, where common.Address) (*big.Int, error) {
	var out []interface{}
	err := _OptInService.contract.Call(opts, &out, "nonces", who, where)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x9333fbda.
//
// Solidity: function nonces(address who, address where) view returns(uint256)
func (_OptInService *OptInServiceSession) Nonces(who common.Address, where common.Address) (*big.Int, error) {
	return _OptInService.Contract.Nonces(&_OptInService.CallOpts, who, where)
}

// Nonces is a free data retrieval call binding the contract method 0x9333fbda.
//
// Solidity: function nonces(address who, address where) view returns(uint256)
func (_OptInService *OptInServiceCallerSession) Nonces(who common.Address, where common.Address) (*big.Int, error) {
	return _OptInService.Contract.Nonces(&_OptInService.CallOpts, who, where)
}

// IncreaseNonce is a paid mutator transaction binding the contract method 0x8247a97c.
//
// Solidity: function increaseNonce(address where) returns()
func (_OptInService *OptInServiceTransactor) IncreaseNonce(opts *bind.TransactOpts, where common.Address) (*types.Transaction, error) {
	return _OptInService.contract.Transact(opts, "increaseNonce", where)
}

// IncreaseNonce is a paid mutator transaction binding the contract method 0x8247a97c.
//
// Solidity: function increaseNonce(address where) returns()
func (_OptInService *OptInServiceSession) IncreaseNonce(where common.Address) (*types.Transaction, error) {
	return _OptInService.Contract.IncreaseNonce(&_OptInService.TransactOpts, where)
}

// IncreaseNonce is a paid mutator transaction binding the contract method 0x8247a97c.
//
// Solidity: function increaseNonce(address where) returns()
func (_OptInService *OptInServiceTransactorSession) IncreaseNonce(where common.Address) (*types.Transaction, error) {
	return _OptInService.Contract.IncreaseNonce(&_OptInService.TransactOpts, where)
}

// OptIn is a paid mutator transaction binding the contract method 0xb1138ad1.
//
// Solidity: function optIn(address where) returns()
func (_OptInService *OptInServiceTransactor) OptIn(opts *bind.TransactOpts, where common.Address) (*types.Transaction, error) {
	return _OptInService.contract.Transact(opts, "optIn", where)
}

// OptIn is a paid mutator transaction binding the contract method 0xb1138ad1.
//
// Solidity: function optIn(address where) returns()
func (_OptInService *OptInServiceSession) OptIn(where common.Address) (*types.Transaction, error) {
	return _OptInService.Contract.OptIn(&_OptInService.TransactOpts, where)
}

// OptIn is a paid mutator transaction binding the contract method 0xb1138ad1.
//
// Solidity: function optIn(address where) returns()
func (_OptInService *OptInServiceTransactorSession) OptIn(where common.Address) (*types.Transaction, error) {
	return _OptInService.Contract.OptIn(&_OptInService.TransactOpts, where)
}

// OptIn0 is a paid mutator transaction binding the contract method 0xced44ba7.
//
// Solidity: function optIn(address who, address where, uint48 deadline, bytes signature) returns()
func (_OptInService *OptInServiceTransactor) OptIn0(opts *bind.TransactOpts, who common.Address, where common.Address, deadline *big.Int, signature []byte) (*types.Transaction, error) {
	return _OptInService.contract.Transact(opts, "optIn0", who, where, deadline, signature)
}

// OptIn0 is a paid mutator transaction binding the contract method 0xced44ba7.
//
// Solidity: function optIn(address who, address where, uint48 deadline, bytes signature) returns()
func (_OptInService *OptInServiceSession) OptIn0(who common.Address, where common.Address, deadline *big.Int, signature []byte) (*types.Transaction, error) {
	return _OptInService.Contract.OptIn0(&_OptInService.TransactOpts, who, where, deadline, signature)
}

// OptIn0 is a paid mutator transaction binding the contract method 0xced44ba7.
//
// Solidity: function optIn(address who, address where, uint48 deadline, bytes signature) returns()
func (_OptInService *OptInServiceTransactorSession) OptIn0(who common.Address, where common.Address, deadline *big.Int, signature []byte) (*types.Transaction, error) {
	return _OptInService.Contract.OptIn0(&_OptInService.TransactOpts, who, where, deadline, signature)
}

// OptOut is a paid mutator transaction binding the contract method 0x93f79bc3.
//
// Solidity: function optOut(address who, address where, uint48 deadline, bytes signature) returns()
func (_OptInService *OptInServiceTransactor) OptOut(opts *bind.TransactOpts, who common.Address, where common.Address, deadline *big.Int, signature []byte) (*types.Transaction, error) {
	return _OptInService.contract.Transact(opts, "optOut", who, where, deadline, signature)
}

// OptOut is a paid mutator transaction binding the contract method 0x93f79bc3.
//
// Solidity: function optOut(address who, address where, uint48 deadline, bytes signature) returns()
func (_OptInService *OptInServiceSession) OptOut(who common.Address, where common.Address, deadline *big.Int, signature []byte) (*types.Transaction, error) {
	return _OptInService.Contract.OptOut(&_OptInService.TransactOpts, who, where, deadline, signature)
}

// OptOut is a paid mutator transaction binding the contract method 0x93f79bc3.
//
// Solidity: function optOut(address who, address where, uint48 deadline, bytes signature) returns()
func (_OptInService *OptInServiceTransactorSession) OptOut(who common.Address, where common.Address, deadline *big.Int, signature []byte) (*types.Transaction, error) {
	return _OptInService.Contract.OptOut(&_OptInService.TransactOpts, who, where, deadline, signature)
}

// OptOut0 is a paid mutator transaction binding the contract method 0xd4610483.
//
// Solidity: function optOut(address where) returns()
func (_OptInService *OptInServiceTransactor) OptOut0(opts *bind.TransactOpts, where common.Address) (*types.Transaction, error) {
	return _OptInService.contract.Transact(opts, "optOut0", where)
}

// OptOut0 is a paid mutator transaction binding the contract method 0xd4610483.
//
// Solidity: function optOut(address where) returns()
func (_OptInService *OptInServiceSession) OptOut0(where common.Address) (*types.Transaction, error) {
	return _OptInService.Contract.OptOut0(&_OptInService.TransactOpts, where)
}

// OptOut0 is a paid mutator transaction binding the contract method 0xd4610483.
//
// Solidity: function optOut(address where) returns()
func (_OptInService *OptInServiceTransactorSession) OptOut0(where common.Address) (*types.Transaction, error) {
	return _OptInService.Contract.OptOut0(&_OptInService.TransactOpts, where)
}

// OptInServiceIncreaseNonceIterator is returned from FilterIncreaseNonce and is used to iterate over the raw logs and unpacked data for IncreaseNonce events raised by the OptInService contract.
type OptInServiceIncreaseNonceIterator struct {
	Event *OptInServiceIncreaseNonce // Event containing the contract specifics and raw log

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
func (it *OptInServiceIncreaseNonceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptInServiceIncreaseNonce)
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
		it.Event = new(OptInServiceIncreaseNonce)
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
func (it *OptInServiceIncreaseNonceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptInServiceIncreaseNonceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptInServiceIncreaseNonce represents a IncreaseNonce event raised by the OptInService contract.
type OptInServiceIncreaseNonce struct {
	Who   common.Address
	Where common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterIncreaseNonce is a free log retrieval operation binding the contract event 0x8ed32926585579e6191b145240df788165f4957e1135f30a00e08ee8feb9d680.
//
// Solidity: event IncreaseNonce(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) FilterIncreaseNonce(opts *bind.FilterOpts, who []common.Address, where []common.Address) (*OptInServiceIncreaseNonceIterator, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var whereRule []interface{}
	for _, whereItem := range where {
		whereRule = append(whereRule, whereItem)
	}

	logs, sub, err := _OptInService.contract.FilterLogs(opts, "IncreaseNonce", whoRule, whereRule)
	if err != nil {
		return nil, err
	}
	return &OptInServiceIncreaseNonceIterator{contract: _OptInService.contract, event: "IncreaseNonce", logs: logs, sub: sub}, nil
}

// WatchIncreaseNonce is a free log subscription operation binding the contract event 0x8ed32926585579e6191b145240df788165f4957e1135f30a00e08ee8feb9d680.
//
// Solidity: event IncreaseNonce(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) WatchIncreaseNonce(opts *bind.WatchOpts, sink chan<- *OptInServiceIncreaseNonce, who []common.Address, where []common.Address) (event.Subscription, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var whereRule []interface{}
	for _, whereItem := range where {
		whereRule = append(whereRule, whereItem)
	}

	logs, sub, err := _OptInService.contract.WatchLogs(opts, "IncreaseNonce", whoRule, whereRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptInServiceIncreaseNonce)
				if err := _OptInService.contract.UnpackLog(event, "IncreaseNonce", log); err != nil {
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

// ParseIncreaseNonce is a log parse operation binding the contract event 0x8ed32926585579e6191b145240df788165f4957e1135f30a00e08ee8feb9d680.
//
// Solidity: event IncreaseNonce(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) ParseIncreaseNonce(log types.Log) (*OptInServiceIncreaseNonce, error) {
	event := new(OptInServiceIncreaseNonce)
	if err := _OptInService.contract.UnpackLog(event, "IncreaseNonce", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OptInServiceOptInIterator is returned from FilterOptIn and is used to iterate over the raw logs and unpacked data for OptIn events raised by the OptInService contract.
type OptInServiceOptInIterator struct {
	Event *OptInServiceOptIn // Event containing the contract specifics and raw log

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
func (it *OptInServiceOptInIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptInServiceOptIn)
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
		it.Event = new(OptInServiceOptIn)
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
func (it *OptInServiceOptInIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptInServiceOptInIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptInServiceOptIn represents a OptIn event raised by the OptInService contract.
type OptInServiceOptIn struct {
	Who   common.Address
	Where common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOptIn is a free log retrieval operation binding the contract event 0x9b730d5b907ee509de729817a2bb37e404418ba569b3a50f36192372f973cb41.
//
// Solidity: event OptIn(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) FilterOptIn(opts *bind.FilterOpts, who []common.Address, where []common.Address) (*OptInServiceOptInIterator, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var whereRule []interface{}
	for _, whereItem := range where {
		whereRule = append(whereRule, whereItem)
	}

	logs, sub, err := _OptInService.contract.FilterLogs(opts, "OptIn", whoRule, whereRule)
	if err != nil {
		return nil, err
	}
	return &OptInServiceOptInIterator{contract: _OptInService.contract, event: "OptIn", logs: logs, sub: sub}, nil
}

// WatchOptIn is a free log subscription operation binding the contract event 0x9b730d5b907ee509de729817a2bb37e404418ba569b3a50f36192372f973cb41.
//
// Solidity: event OptIn(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) WatchOptIn(opts *bind.WatchOpts, sink chan<- *OptInServiceOptIn, who []common.Address, where []common.Address) (event.Subscription, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var whereRule []interface{}
	for _, whereItem := range where {
		whereRule = append(whereRule, whereItem)
	}

	logs, sub, err := _OptInService.contract.WatchLogs(opts, "OptIn", whoRule, whereRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptInServiceOptIn)
				if err := _OptInService.contract.UnpackLog(event, "OptIn", log); err != nil {
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

// ParseOptIn is a log parse operation binding the contract event 0x9b730d5b907ee509de729817a2bb37e404418ba569b3a50f36192372f973cb41.
//
// Solidity: event OptIn(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) ParseOptIn(log types.Log) (*OptInServiceOptIn, error) {
	event := new(OptInServiceOptIn)
	if err := _OptInService.contract.UnpackLog(event, "OptIn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OptInServiceOptOutIterator is returned from FilterOptOut and is used to iterate over the raw logs and unpacked data for OptOut events raised by the OptInService contract.
type OptInServiceOptOutIterator struct {
	Event *OptInServiceOptOut // Event containing the contract specifics and raw log

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
func (it *OptInServiceOptOutIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptInServiceOptOut)
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
		it.Event = new(OptInServiceOptOut)
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
func (it *OptInServiceOptOutIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptInServiceOptOutIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptInServiceOptOut represents a OptOut event raised by the OptInService contract.
type OptInServiceOptOut struct {
	Who   common.Address
	Where common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOptOut is a free log retrieval operation binding the contract event 0x1629cd9ad365627cf8408d19c50224af8f3213c1a18ae48062d92e22bddf7de5.
//
// Solidity: event OptOut(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) FilterOptOut(opts *bind.FilterOpts, who []common.Address, where []common.Address) (*OptInServiceOptOutIterator, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var whereRule []interface{}
	for _, whereItem := range where {
		whereRule = append(whereRule, whereItem)
	}

	logs, sub, err := _OptInService.contract.FilterLogs(opts, "OptOut", whoRule, whereRule)
	if err != nil {
		return nil, err
	}
	return &OptInServiceOptOutIterator{contract: _OptInService.contract, event: "OptOut", logs: logs, sub: sub}, nil
}

// WatchOptOut is a free log subscription operation binding the contract event 0x1629cd9ad365627cf8408d19c50224af8f3213c1a18ae48062d92e22bddf7de5.
//
// Solidity: event OptOut(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) WatchOptOut(opts *bind.WatchOpts, sink chan<- *OptInServiceOptOut, who []common.Address, where []common.Address) (event.Subscription, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var whereRule []interface{}
	for _, whereItem := range where {
		whereRule = append(whereRule, whereItem)
	}

	logs, sub, err := _OptInService.contract.WatchLogs(opts, "OptOut", whoRule, whereRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptInServiceOptOut)
				if err := _OptInService.contract.UnpackLog(event, "OptOut", log); err != nil {
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

// ParseOptOut is a log parse operation binding the contract event 0x1629cd9ad365627cf8408d19c50224af8f3213c1a18ae48062d92e22bddf7de5.
//
// Solidity: event OptOut(address indexed who, address indexed where)
func (_OptInService *OptInServiceFilterer) ParseOptOut(log types.Log) (*OptInServiceOptOut, error) {
	event := new(OptInServiceOptOut)
	if err := _OptInService.contract.UnpackLog(event, "OptOut", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
