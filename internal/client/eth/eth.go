package eth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	_ "embed"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"

	"middleware-offchain/internal/client/eth/gen"
	"middleware-offchain/internal/entity"
)

//go:embed Master.abi.json
var masterAbiJSON []byte
var masterABI abi.ABI

//go:embed VaultManager.abi.json
var vaultManagerAbiJSON []byte
var vaultManagerABI abi.ABI

//go:embed KeyRegistry.abi.json
var keyRegistryAbiJSON []byte
var keyRegistryABI abi.ABI

var (
	getConfigFunction                 = "getConfigAt"
	getIsGenesisSetFunction           = "isGenesisSet"
	getCurrentEpochFunction           = "getCurrentEpoch"
	getCurrentPhaseFunction           = "getCurrentPhase"
	getCurrentValsetTimestampFunction = "getCurrentValSetTimestamp"
	getCaptureTimestampFunction       = "getCaptureTimestamp"
	getVotingPowersFunction           = "getVotingPowersAt"
	getKeysFunction                   = "getKeysAt"
	getRequiredKeyTagFunction         = "getRequiredKeyTagAt"
	getQuorumThresholdFunction        = "getQuorumThresholdAt"
	getSubnetworkFunction             = "SUBNETWORK"
	getEip712DomainFunction           = "eip712Domain"
	verifyQuorumSigFunction           = "verifyQuorumSig"
)

type Config struct {
	MasterRPCURL   string `validate:"required"`
	MasterAddress  string `validate:"required"`
	PrivateKey     []byte
	RequestTimeout time.Duration `validate:"required,gt=0"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	if c.PrivateKey != nil {
		_, err := crypto.ToECDSA(c.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to convert private key: %w", err)
		}
	}

	return nil
}

type Client struct {
	client                *ethclient.Client
	masterContractAddress common.Address
	cfg                   Config

	masterPK *ecdsa.PrivateKey // could be nil for read-only access
	master   *gen.Master       // could be nil for read-only access
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

	master, err := gen.NewMaster(common.HexToAddress(cfg.MasterAddress), client)
	if err != nil {
		return nil, errors.Errorf("failed to create new master contract: %w", err)
	}

	var pk *ecdsa.PrivateKey
	if cfg.PrivateKey != nil {
		pk, err = crypto.ToECDSA(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert private key: %w", err)
		}
	}

	return &Client{
		client:                client,
		masterContractAddress: common.HexToAddress(cfg.MasterAddress),
		masterPK:              pk,
		cfg:                   cfg,
		master:                master,
	}, nil
}

func initABI() error {
	var err error
	masterABI, err = abi.JSON(bytes.NewReader(masterAbiJSON))
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

type configDTO struct {
	VotingPowerProviders    []crossChainAddressDTO `json:"votingPowerProviders"`
	KeysProvider            crossChainAddressDTO   `json:"keysProvider"`
	Replicas                []crossChainAddressDTO `json:"replicas"`
	VerificationType        uint32                 `json:"verificationType"`
	MaxVotingPower          *big.Int               `json:"maxVotingPower"`
	MinInclusionVotingPower *big.Int               `json:"minInclusionVotingPower"`
	MaxValidatorsCount      *big.Int               `json:"maxValidatorsCount"`
	RequiredKeyTags         []uint8                `json:"requiredKeyTags"`
}

func (c configDTO) toEntity() entity.Config {
	return entity.Config{
		VotingPowerProviders: lo.Map(c.VotingPowerProviders, func(v crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				Address: v.Addr,
				ChainId: v.ChainId,
			}
		}),
		KeysProvider: entity.CrossChainAddress{
			Address: c.KeysProvider.Addr,
			ChainId: c.KeysProvider.ChainId,
		},
		Replicas: lo.Map(c.Replicas, func(v crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				Address: v.Addr,
				ChainId: v.ChainId,
			}
		}),
		VerificationType:        c.VerificationType,
		MaxVotingPower:          c.MaxVotingPower,
		MinInclusionVotingPower: c.MinInclusionVotingPower,
		MaxValidatorsCount:      c.MaxValidatorsCount,
		RequiredKeyTags:         c.RequiredKeyTags,
	}
}

type configContainer struct {
	Config configDTO
}

func (e *Client) GetConfig(ctx context.Context, timestamp *big.Int) (entity.Config, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getConfigFunction, timestamp, []byte{})
	if err != nil {
		return entity.Config{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	fmt.Println("callMsg>>>", callMsg)

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.Config{}, errors.Errorf("failed to call contract: %w", err)
	}

	fmt.Println("result>>>", result)

	c := configContainer{}
	err = masterABI.UnpackIntoInterface(&c, getConfigFunction, result)
	if err != nil {
		return entity.Config{}, errors.Errorf("failed to unpack config: %w", err)
	}

	return c.Config.toEntity(), nil
}

func (e *Client) GetIsGenesisSet(ctx context.Context) (bool, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getIsGenesisSetFunction)
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
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getCurrentEpochFunction)
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
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getCurrentPhaseFunction)
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
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getCurrentValsetTimestampFunction)
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
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getCaptureTimestampFunction)
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

type operatorVotingPowerDTO struct {
	Operator common.Address
	Vaults   []vaultVotingPowerDTO
}

type vaultVotingPowerDTO struct {
	Vault       common.Address
	VotingPower *big.Int
}

func (e *Client) GetVotingPowers(ctx context.Context, address common.Address, timestamp *big.Int) ([]entity.OperatorVotingPower, error) {
	callMsg, err := constructCallMsg(address, vaultManagerABI, getVotingPowersFunction, [][]byte{}, timestamp, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var votingPowers []operatorVotingPowerDTO
	err = vaultManagerABI.UnpackIntoInterface(&votingPowers, getVotingPowersFunction, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack voting powers: %w", err)
	}

	return lo.Map(votingPowers, func(item operatorVotingPowerDTO, index int) entity.OperatorVotingPower {
		return entity.OperatorVotingPower{
			Operator: item.Operator,
			Vaults: lo.Map(item.Vaults, func(vault vaultVotingPowerDTO, index int) entity.VaultVotingPower {
				return entity.VaultVotingPower{
					Vault:       vault.Vault,
					VotingPower: vault.VotingPower,
				}
			}),
		}
	}), nil
}

func (e *Client) GetKeys(ctx context.Context, address common.Address, timestamp *big.Int) ([]entity.OperatorWithKeys, error) {
	callMsg, err := constructCallMsg(address, keyRegistryABI, getKeysFunction, timestamp, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var keys []entity.OperatorWithKeys
	err = keyRegistryABI.UnpackIntoInterface(&keys, getKeysFunction, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack keys: %w", err)
	}

	return keys, nil
}

func (e *Client) GetRequiredKeyTag(ctx context.Context, timestamp *big.Int) (uint8, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getRequiredKeyTagFunction, timestamp, []byte{})
	if err != nil {
		return 0, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	var keyTag uint8
	err = masterABI.UnpackIntoInterface(&keyTag, getRequiredKeyTagFunction, result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack key tag: %w", err)
	}

	return keyTag, nil
}

func (e *Client) GetQuorumThreshold(ctx context.Context, timestamp *big.Int, keyTag uint8) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getQuorumThresholdFunction, keyTag, timestamp, []byte{})
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
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getSubnetworkFunction)
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return result, nil
}

func (e *Client) GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, getEip712DomainFunction)
	if err != nil {
		return entity.Eip712Domain{}, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.Eip712Domain{}, fmt.Errorf("failed to call contract: %w", err)
	}

	var eip712Domain entity.Eip712Domain

	out, err := masterABI.Unpack(getEip712DomainFunction, result)
	if err != nil {
		return entity.Eip712Domain{}, fmt.Errorf("failed to unpack eip712 domain: %w", err)
	}

	eip712Domain.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	eip712Domain.Name = *abi.ConvertType(out[1], new(string)).(*string)
	eip712Domain.Version = *abi.ConvertType(out[2], new(string)).(*string)
	eip712Domain.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	eip712Domain.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	salt := *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	eip712Domain.Salt = new(big.Int).SetBytes(salt[:])
	eip712Domain.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return eip712Domain, nil
}

func (e *Client) callContract(ctx context.Context, callMsg ethereum.CallMsg) (result []byte, err error) {
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	result, err = e.client.CallContract(tmCtx, callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return result, nil
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
