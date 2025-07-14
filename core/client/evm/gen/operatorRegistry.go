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

// OperatorRegistryMetaData contains all meta data concerning the OperatorRegistry contract.
var OperatorRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"entity\",\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isEntity\",\"inputs\":[{\"name\":\"entity_\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"totalEntities\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"AddEntity\",\"inputs\":[{\"name\":\"entity\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"EntityNotExist\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]}]",
}

// OperatorRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use OperatorRegistryMetaData.ABI instead.
var OperatorRegistryABI = OperatorRegistryMetaData.ABI

// OperatorRegistry is an auto generated Go binding around an Ethereum contract.
type OperatorRegistry struct {
	OperatorRegistryCaller     // Read-only binding to the contract
	OperatorRegistryTransactor // Write-only binding to the contract
	OperatorRegistryFilterer   // Log filterer for contract events
}

// OperatorRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type OperatorRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperatorRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OperatorRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperatorRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OperatorRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperatorRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OperatorRegistrySession struct {
	Contract     *OperatorRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OperatorRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OperatorRegistryCallerSession struct {
	Contract *OperatorRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// OperatorRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OperatorRegistryTransactorSession struct {
	Contract     *OperatorRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// OperatorRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type OperatorRegistryRaw struct {
	Contract *OperatorRegistry // Generic contract binding to access the raw methods on
}

// OperatorRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OperatorRegistryCallerRaw struct {
	Contract *OperatorRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// OperatorRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OperatorRegistryTransactorRaw struct {
	Contract *OperatorRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOperatorRegistry creates a new instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistry(address common.Address, backend bind.ContractBackend) (*OperatorRegistry, error) {
	contract, err := bindOperatorRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistry{OperatorRegistryCaller: OperatorRegistryCaller{contract: contract}, OperatorRegistryTransactor: OperatorRegistryTransactor{contract: contract}, OperatorRegistryFilterer: OperatorRegistryFilterer{contract: contract}}, nil
}

// NewOperatorRegistryCaller creates a new read-only instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistryCaller(address common.Address, caller bind.ContractCaller) (*OperatorRegistryCaller, error) {
	contract, err := bindOperatorRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryCaller{contract: contract}, nil
}

// NewOperatorRegistryTransactor creates a new write-only instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*OperatorRegistryTransactor, error) {
	contract, err := bindOperatorRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryTransactor{contract: contract}, nil
}

// NewOperatorRegistryFilterer creates a new log filterer instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*OperatorRegistryFilterer, error) {
	contract, err := bindOperatorRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryFilterer{contract: contract}, nil
}

// bindOperatorRegistry binds a generic wrapper to an already deployed contract.
func bindOperatorRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OperatorRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperatorRegistry *OperatorRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperatorRegistry.Contract.OperatorRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperatorRegistry *OperatorRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OperatorRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperatorRegistry *OperatorRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OperatorRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperatorRegistry *OperatorRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperatorRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperatorRegistry *OperatorRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperatorRegistry *OperatorRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.contract.Transact(opts, method, params...)
}

// Entity is a free data retrieval call binding the contract method 0xb42ba2a2.
//
// Solidity: function entity(uint256 index) view returns(address)
func (_OperatorRegistry *OperatorRegistryCaller) Entity(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "entity", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Entity is a free data retrieval call binding the contract method 0xb42ba2a2.
//
// Solidity: function entity(uint256 index) view returns(address)
func (_OperatorRegistry *OperatorRegistrySession) Entity(index *big.Int) (common.Address, error) {
	return _OperatorRegistry.Contract.Entity(&_OperatorRegistry.CallOpts, index)
}

// Entity is a free data retrieval call binding the contract method 0xb42ba2a2.
//
// Solidity: function entity(uint256 index) view returns(address)
func (_OperatorRegistry *OperatorRegistryCallerSession) Entity(index *big.Int) (common.Address, error) {
	return _OperatorRegistry.Contract.Entity(&_OperatorRegistry.CallOpts, index)
}

// IsEntity is a free data retrieval call binding the contract method 0x14887c58.
//
// Solidity: function isEntity(address entity_) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) IsEntity(opts *bind.CallOpts, entity_ common.Address) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "isEntity", entity_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsEntity is a free data retrieval call binding the contract method 0x14887c58.
//
// Solidity: function isEntity(address entity_) view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) IsEntity(entity_ common.Address) (bool, error) {
	return _OperatorRegistry.Contract.IsEntity(&_OperatorRegistry.CallOpts, entity_)
}

// IsEntity is a free data retrieval call binding the contract method 0x14887c58.
//
// Solidity: function isEntity(address entity_) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) IsEntity(entity_ common.Address) (bool, error) {
	return _OperatorRegistry.Contract.IsEntity(&_OperatorRegistry.CallOpts, entity_)
}

// TotalEntities is a free data retrieval call binding the contract method 0x5cd8b15e.
//
// Solidity: function totalEntities() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) TotalEntities(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "totalEntities")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalEntities is a free data retrieval call binding the contract method 0x5cd8b15e.
//
// Solidity: function totalEntities() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) TotalEntities() (*big.Int, error) {
	return _OperatorRegistry.Contract.TotalEntities(&_OperatorRegistry.CallOpts)
}

// TotalEntities is a free data retrieval call binding the contract method 0x5cd8b15e.
//
// Solidity: function totalEntities() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) TotalEntities() (*big.Int, error) {
	return _OperatorRegistry.Contract.TotalEntities(&_OperatorRegistry.CallOpts)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x2acde098.
//
// Solidity: function registerOperator() returns()
func (_OperatorRegistry *OperatorRegistryTransactor) RegisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "registerOperator")
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x2acde098.
//
// Solidity: function registerOperator() returns()
func (_OperatorRegistry *OperatorRegistrySession) RegisterOperator() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RegisterOperator(&_OperatorRegistry.TransactOpts)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x2acde098.
//
// Solidity: function registerOperator() returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) RegisterOperator() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RegisterOperator(&_OperatorRegistry.TransactOpts)
}

// OperatorRegistryAddEntityIterator is returned from FilterAddEntity and is used to iterate over the raw logs and unpacked data for AddEntity events raised by the OperatorRegistry contract.
type OperatorRegistryAddEntityIterator struct {
	Event *OperatorRegistryAddEntity // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryAddEntityIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryAddEntity)
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
		it.Event = new(OperatorRegistryAddEntity)
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
func (it *OperatorRegistryAddEntityIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryAddEntityIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryAddEntity represents a AddEntity event raised by the OperatorRegistry contract.
type OperatorRegistryAddEntity struct {
	Entity common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAddEntity is a free log retrieval operation binding the contract event 0xb919910dcefbf753bfd926ab3b1d3f85d877190c3d01ba1bd585047b99b99f0b.
//
// Solidity: event AddEntity(address indexed entity)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterAddEntity(opts *bind.FilterOpts, entity []common.Address) (*OperatorRegistryAddEntityIterator, error) {

	var entityRule []interface{}
	for _, entityItem := range entity {
		entityRule = append(entityRule, entityItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "AddEntity", entityRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryAddEntityIterator{contract: _OperatorRegistry.contract, event: "AddEntity", logs: logs, sub: sub}, nil
}

// WatchAddEntity is a free log subscription operation binding the contract event 0xb919910dcefbf753bfd926ab3b1d3f85d877190c3d01ba1bd585047b99b99f0b.
//
// Solidity: event AddEntity(address indexed entity)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchAddEntity(opts *bind.WatchOpts, sink chan<- *OperatorRegistryAddEntity, entity []common.Address) (event.Subscription, error) {

	var entityRule []interface{}
	for _, entityItem := range entity {
		entityRule = append(entityRule, entityItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "AddEntity", entityRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryAddEntity)
				if err := _OperatorRegistry.contract.UnpackLog(event, "AddEntity", log); err != nil {
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

// ParseAddEntity is a log parse operation binding the contract event 0xb919910dcefbf753bfd926ab3b1d3f85d877190c3d01ba1bd585047b99b99f0b.
//
// Solidity: event AddEntity(address indexed entity)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseAddEntity(log types.Log) (*OperatorRegistryAddEntity, error) {
	event := new(OperatorRegistryAddEntity)
	if err := _OperatorRegistry.contract.UnpackLog(event, "AddEntity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
