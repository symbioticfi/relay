package eth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	_ "embed"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"

	"middleware-offchain/internal/entity"
)

//go:embed contract.abi.json
var contractAbiJSON []byte
var contractABI abi.ABI

//go:embed VaultManager.abi.json
var vaultManagerAbiJSON []byte
var vaultManagerABI abi.ABI

//go:embed KeyRegistry.abi.json
var keyRegistryAbiJSON []byte
var keyRegistryABI abi.ABI

var (
	GET_MASTER_CONFIG_FUNCTION            = "getMasterConfigAt"
	GET_VALSET_CONFIG_FUNCTION            = "getValSetConfigAt"
	GET_IS_GENESIS_SET_FUNCTION           = "isGenesisSet"
	GET_CURRENT_EPOCH_FUNCTION            = "getCurrentEpoch"
	GET_CURRENT_PHASE_FUNCTION            = "getCurrentPhase"
	GET_CURRENT_VALSET_TIMESTAMP_FUNCTION = "getCurrentValsetTimestamp"
	GET_CAPTURE_TIMESTAMP_FUNCTION        = "getCaptureTimestamp"
	GET_VOTING_POWERS_FUNCTION            = "getVotingPowersAt"
	GET_KEYS_FUNCTION                     = "getKeysAt"
	GET_REQUIRED_KEY_TAG_FUNCTION         = "getRequiredKeyTagAt"
	GET_QUORUM_THRESHOLD_FUNCTION         = "getQuorumThresholdAt"
	GET_SUBNETWORK_FUNCTION               = "SUBNETWORK"
	GET_EIP_712_DOMAIN_FUNCTION           = "eip712Domain"
)

type Config struct {
	MasterRPCURL  string `validate:"required"`
	MasterAddress string `validate:"required"`
	PrivateKey    []byte `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	_, err := crypto.ToECDSA(c.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key: %w", err)
	}

	return nil
}

type Client struct {
	client                *ethclient.Client
	masterContractAddress common.Address
	privateKey            *ecdsa.PrivateKey // could be nil for read-only access
}

func NewEthClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	if err := initABI(); err != nil {
		return nil, fmt.Errorf("failed to initialize ABI: %w", err)
	}

	client, err := ethclient.Dial(cfg.MasterRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	pk, err := crypto.ToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key: %w", err)
	}

	return &Client{
		client:                client,
		masterContractAddress: common.HexToAddress(cfg.MasterAddress),
		privateKey:            pk,
	}, nil
}

func initABI() error {
	var err error
	contractABI, err = abi.JSON(bytes.NewReader(contractAbiJSON))
	if err != nil {
		return fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	vaultManagerABI, err = abi.JSON(bytes.NewReader(vaultManagerAbiJSON))
	if err != nil {
		return fmt.Errorf("failed to parse vault manager ABI: %w", err)
	}

	keyRegistryABI, err = abi.JSON(bytes.NewReader(keyRegistryAbiJSON))
	if err != nil {
		return fmt.Errorf("failed to parse key registry ABI: %w", err)
	}

	return nil
}

func GeneratePrivateKey() ([]byte, error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return crypto.FromECDSA(pk), nil
}

func (e *Client) Commit(messageHash string, signature []byte) error {
	return nil
}

type crossChainAddressDTO struct {
	Addr    common.Address `json:"addr"`
	ChainId uint64         `json:"chainId"`
}

type masterConfigDTO struct {
	VotingPowerProviders []crossChainAddressDTO `json:"votingPowerProviders"`
	KeysProvider         crossChainAddressDTO   `json:"keysProvider"`
	Replicas             []crossChainAddressDTO `json:"replicas"`
}

func (c masterConfigDTO) toEntity() entity.MasterConfig {
	return entity.MasterConfig{
		VotingPowerProviders: lo.Map(c.VotingPowerProviders, func(v crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				Address: v.Addr,
				ChainID: v.ChainId,
			}
		}),
		KeysProvider: entity.CrossChainAddress{
			Address: c.KeysProvider.Addr,
			ChainID: c.KeysProvider.ChainId,
		},
		Replicas: lo.Map(c.Replicas, func(v crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				Address: v.Addr,
				ChainID: v.ChainId,
			}
		}),
	}
}

type masterConfigContainer struct {
	MasterConfig masterConfigDTO
}

func (e *Client) GetMasterConfig(ctx context.Context, timestamp *big.Int) (entity.MasterConfig, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_MASTER_CONFIG_FUNCTION, timestamp, []byte{})
	if err != nil {
		return entity.MasterConfig{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.MasterConfig{}, errors.Errorf("failed to call contract: %w", err)
	}

	mcc := masterConfigContainer{}
	err = contractABI.UnpackIntoInterface(&mcc, GET_MASTER_CONFIG_FUNCTION, result)
	if err != nil {
		return entity.MasterConfig{}, errors.Errorf("failed to unpack master config: %w", err)
	}

	return mcc.MasterConfig.toEntity(), nil
}

type valSetConfigDTO struct {
	MaxVotingPower          *big.Int `json:"max_voting_power"`
	MinInclusionVotingPower *big.Int `json:"min_inclusion_voting_power"`
	MaxValidatorsCount      *big.Int `json:"max_validators_count"`
	RequiredKeyTags         []byte   `json:"required_key_tags"`
}

type valSetConfigContainer struct {
	ValSetConfig valSetConfigDTO
}

func (c valSetConfigDTO) toEntity() entity.ValSetConfig {
	return entity.ValSetConfig{
		MaxVotingPower:          c.MaxVotingPower,
		MinInclusionVotingPower: c.MinInclusionVotingPower,
		MaxValidatorsCount:      c.MaxValidatorsCount,
		RequiredKeyTags:         c.RequiredKeyTags,
	}
}

func (e *Client) GetValSetConfig(ctx context.Context, timestamp *big.Int) (entity.ValSetConfig, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_VALSET_CONFIG_FUNCTION, timestamp, []byte{})
	if err != nil {
		return entity.ValSetConfig{}, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.ValSetConfig{}, fmt.Errorf("failed to call contract: %w", err)
	}

	var valSetConfig valSetConfigContainer
	err = contractABI.UnpackIntoInterface(&valSetConfig, GET_VALSET_CONFIG_FUNCTION, result)
	if err != nil {
		return entity.ValSetConfig{}, fmt.Errorf("failed to unpack val set config: %w", err)
	}

	return valSetConfig.ValSetConfig.toEntity(), nil
}

func (e *Client) GetIsGenesisSet(ctx context.Context) (bool, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_IS_GENESIS_SET_FUNCTION)
	if err != nil {
		return false, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return false, fmt.Errorf("failed to call contract: %w", err)
	}

	isGenesisSet := new(big.Int).SetBytes(result).Uint64()
	return isGenesisSet == 1, nil
}

func (e *Client) GetCurrentEpoch(ctx context.Context) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_CURRENT_EPOCH_FUNCTION)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	epoch := new(big.Int).SetBytes(result)
	return epoch, nil
}

func (e *Client) GetCurrentPhase(ctx context.Context) (entity.Phase, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_CURRENT_PHASE_FUNCTION)
	if err != nil {
		return 0, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	phase := new(big.Int).SetBytes(result).Uint64()
	return entity.Phase(phase), nil
}

func (e *Client) GetCurrentValsetTimestamp(ctx context.Context) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_CURRENT_VALSET_TIMESTAMP_FUNCTION)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	timestamp := new(big.Int).SetBytes(result)
	return timestamp, nil
}

func (e *Client) GetCaptureTimestamp(ctx context.Context) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_CAPTURE_TIMESTAMP_FUNCTION)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	timestamp := new(big.Int).SetBytes(result)
	return timestamp, nil
}

func (e *Client) GetVotingPowers(ctx context.Context, address common.Address, timestamp *big.Int) ([]entity.OperatorVotingPower, error) {
	callMsg, err := constructCallMsg(address, vaultManagerABI, GET_VOTING_POWERS_FUNCTION, [][]byte{}, timestamp, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var votingPowers []entity.OperatorVotingPower
	err = vaultManagerABI.UnpackIntoInterface(&votingPowers, GET_VOTING_POWERS_FUNCTION, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack voting powers: %w", err)
	}

	return votingPowers, nil
}

func (e *Client) GetKeys(ctx context.Context, address common.Address, timestamp *big.Int) ([]entity.OperatorWithKeys, error) {
	callMsg, err := constructCallMsg(address, keyRegistryABI, GET_KEYS_FUNCTION, timestamp, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var keys []entity.OperatorWithKeys
	err = keyRegistryABI.UnpackIntoInterface(&keys, GET_KEYS_FUNCTION, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack keys: %w", err)
	}

	return keys, nil
}

func (e *Client) GetRequiredKeyTag(ctx context.Context, timestamp *big.Int) (uint8, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_REQUIRED_KEY_TAG_FUNCTION, timestamp, []byte{})
	if err != nil {
		return 0, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	var keyTag uint8
	err = contractABI.UnpackIntoInterface(&keyTag, GET_REQUIRED_KEY_TAG_FUNCTION, result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack key tag: %w", err)
	}

	return keyTag, nil
}

func (e *Client) GetQuorumThreshold(ctx context.Context, timestamp *big.Int, keyTag uint8) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_QUORUM_THRESHOLD_FUNCTION, keyTag, timestamp, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return new(big.Int).SetBytes(result), nil
}

func (e *Client) GetSubnetwork(ctx context.Context) ([]byte, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_SUBNETWORK_FUNCTION)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return result, nil
}

func (e *Client) GetEip712Domain(ctx context.Context) (*entity.Eip712Domain, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, contractABI, GET_EIP_712_DOMAIN_FUNCTION)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var eip712Domain entity.Eip712Domain
	err = contractABI.UnpackIntoInterface(&eip712Domain, GET_EIP_712_DOMAIN_FUNCTION, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack eip712 domain: %w", err)
	}

	return &eip712Domain, nil
}

func (e *Client) callContract(ctx context.Context, callMsg ethereum.CallMsg) (result []byte, err error) {
	return e.client.CallContract(ctx, callMsg, nil)
}

func constructCallMsg(contractAddress common.Address, abi abi.ABI, method string, args ...interface{}) (ethereum.CallMsg, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return ethereum.CallMsg{}, fmt.Errorf("failed to pack method: %w", err)
	}

	return ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}, nil
}
