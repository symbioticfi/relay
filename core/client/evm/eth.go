package evm

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/hex"
	"math/big"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/core/client/evm/gen"
	"middleware-offchain/core/entity"
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
	getConfigFunction                             = "getConfigAt"
	getIsGenesisSetFunction                       = "isGenesisSet"
	getCurrentEpochFunction                       = "getCurrentEpoch"
	getEpochStartFunction                         = "getEpochStart"
	getCurrentPhaseFunction                       = "getCurrentPhase"
	getCurrentValsetTimestampFunction             = "getCurrentValSetTimestamp"
	getLastCommittedHeaderEpochFunction           = "getLastCommittedHeaderEpoch"
	getCaptureTimestampFunction                   = "getCaptureTimestamp"
	getCaptureTimestampFromValsetHeaderAtFunction = "getCaptureTimestampFromValSetHeaderAt"
	getVotingPowersFunction                       = "getVotingPowersAt"
	getKeysFunction                               = "getKeysAt"
	getRequiredKeyTagFunction                     = "getRequiredKeyTagAt"
	getQuorumThresholdFunction                    = "getQuorumThresholdAt"
	getSubnetworkFunction                         = "SUBNETWORK"
	getNetworkFunction                            = "NETWORK"
	getPreviousHeaderHashFunction                 = "getPreviousHeaderHashFromValSetHeader"
	getPreviousHeaderHashAtFunction               = "getPreviousHeaderHashFromValSetHeaderAt"
	getEip712DomainFunction                       = "eip712Domain"
	verifyQuorumSigFunction                       = "verifyQuorumSig"
	getLatestHeaderHashFunction                   = "getValSetHeaderHash"
	getLatestHeaderHashAtFunction                 = "getValSetHeaderHashAt"
	isValSetHeaderCommittedAtFunction             = "isValSetHeaderCommittedAt"
	getValSetHeaderAtFunction                     = "getValSetHeaderAt"
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
			return errors.Errorf("failed to convert private key: %w", err)
		}
	}

	return nil
}

type Client struct {
	client        *ethclient.Client
	masterAddress entity.CrossChainAddress
	cfg           Config

	masterPK *ecdsa.PrivateKey // could be nil for read-only access
	master   *gen.Master       // could be nil for read-only access
}

func NewEVMClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	if err := initABI(); err != nil {
		return nil, errors.Errorf("failed to initialize ABI: %w", err)
	}

	client, err := ethclient.Dial(cfg.MasterRPCURL)
	if err != nil {
		return nil, errors.Errorf("failed to connect to Ethereum client: %w", err)
	}

	master, err := gen.NewMaster(common.HexToAddress(cfg.MasterAddress), client)
	if err != nil {
		return nil, errors.Errorf("failed to create new master contract: %w", err)
	}

	var pk *ecdsa.PrivateKey
	if cfg.PrivateKey != nil {
		pk, err = crypto.ToECDSA(cfg.PrivateKey)
		if err != nil {
			return nil, errors.Errorf("failed to convert private key: %w", err)
		}
	}

	return &Client{
		client: client,
		masterAddress: entity.CrossChainAddress{
			Address: common.HexToAddress(cfg.MasterAddress),
			ChainId: 111,
		},
		masterPK: pk,
		cfg:      cfg,
		master:   master,
	}, nil
}

func initABI() error {
	var err error
	masterABI, err = abi.JSON(bytes.NewReader(masterAbiJSON))
	if err != nil {
		return errors.Errorf("failed to parse contract ABI: %w", err)
	}

	vaultManagerABI, err = abi.JSON(bytes.NewReader(vaultManagerAbiJSON))
	if err != nil {
		return errors.Errorf("failed to parse vault manager ABI: %w", err)
	}

	keyRegistryABI, err = abi.JSON(bytes.NewReader(keyRegistryAbiJSON))
	if err != nil {
		return errors.Errorf("failed to parse key registry ABI: %w", err)
	}

	return nil
}

type dtoCrossChainAddress struct {
	Address common.Address `json:"addr"`
	ChainId uint64         `json:"chainId"`
}

func (d dtoCrossChainAddress) toEntity() entity.CrossChainAddress {
	return entity.CrossChainAddress{
		ChainId: d.ChainId,
		Address: d.Address,
	}
}

func (e *Client) GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error) {
	type dtoNetworkConfig struct {
		VotingPowerProviders    []dtoCrossChainAddress `json:"votingPowerProviders"`
		KeysProvider            dtoCrossChainAddress   `json:"keysProvider"`
		Replicas                []dtoCrossChainAddress `json:"replicas"`
		VerificationType        uint32                 `json:"verificationType"`
		MaxVotingPower          *big.Int               `json:"maxVotingPower"`
		MinInclusionVotingPower *big.Int               `json:"minInclusionVotingPower"`
		MaxValidatorsCount      *big.Int               `json:"maxValidatorsCount"`
		RequiredKeyTags         []uint8                `json:"requiredKeyTags"`
	}
	type dtoConfigContainer struct {
		Config dtoNetworkConfig
	}
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getConfigFunction, new(big.Int).SetUint64(timestamp), []byte{})
	if err != nil {
		return entity.NetworkConfig{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.NetworkConfig{}, errors.Errorf("failed to call contract: %w", err)
	}

	configContainer := dtoConfigContainer{}
	err = masterABI.UnpackIntoInterface(&configContainer, getConfigFunction, result)
	if err != nil {
		return entity.NetworkConfig{}, errors.Errorf("failed to unpack config: %w", err)
	}

	dtoConfig := configContainer.Config

	return entity.NetworkConfig{
		VotingPowerProviders: lo.Map(dtoConfig.VotingPowerProviders, func(v dtoCrossChainAddress, _ int) entity.CrossChainAddress {
			return v.toEntity()
		}),
		KeysProvider: entity.CrossChainAddress{
			Address: dtoConfig.KeysProvider.Address,
			ChainId: dtoConfig.KeysProvider.ChainId,
		},
		Replicas: lo.Map(dtoConfig.Replicas, func(v dtoCrossChainAddress, _ int) entity.CrossChainAddress {
			return v.toEntity()
		}),
		VerificationType:        entity.VerificationType(dtoConfig.VerificationType),
		MaxVotingPower:          dtoConfig.MaxVotingPower,
		MinInclusionVotingPower: dtoConfig.MinInclusionVotingPower,
		MaxValidatorsCount:      dtoConfig.MaxValidatorsCount,
		RequiredKeyTags: lo.Map(dtoConfig.RequiredKeyTags, func(v uint8, _ int) entity.KeyTag {
			return entity.KeyTag(v)
		}),
	}, nil
}

func (e *Client) GetIsGenesisSet(ctx context.Context) (bool, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getIsGenesisSetFunction)
	if err != nil {
		return false, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return false, errors.Errorf("failed to call contract: %w", err)
	}

	isGenesisSet := new(big.Int).SetBytes(result).Uint64()
	return isGenesisSet == 1, nil
}

func (e *Client) IsValsetHeaderCommittedAt(ctx context.Context, epoch uint64) (bool, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, isValSetHeaderCommittedAtFunction, new(big.Int).SetUint64(epoch))
	if err != nil {
		return false, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return false, errors.Errorf("failed to call contract: %w", err)
	}

	flag := new(big.Int).SetBytes(result).Uint64()
	return flag == 1, nil
}

func (e *Client) GetCurrentEpoch(ctx context.Context) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getCurrentEpochFunction)
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	epoch := new(big.Int).SetBytes(result).Uint64()
	return epoch, nil
}

func (e *Client) GetPreviousHeaderHash(ctx context.Context) ([32]byte, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getPreviousHeaderHashFunction)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to call contract: %w", err)
	}

	return [32]byte(result), nil
}

func (e *Client) GetPreviousHeaderHashAt(ctx context.Context, epoch uint64) ([32]byte, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getPreviousHeaderHashAtFunction, new(big.Int).SetUint64(epoch))
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to call contract: %w", err)
	}

	return [32]byte(result), nil
}

func (e *Client) GetLatestHeaderHash(ctx context.Context) ([32]byte, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getLatestHeaderHashFunction)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to call contract: %w", err)
	}

	return [32]byte(result), nil
}

func (e *Client) GetHeaderHashAt(ctx context.Context, epoch uint64) ([32]byte, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getLatestHeaderHashAtFunction, new(big.Int).SetUint64(epoch))
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to call contract: %w", err)
	}

	return [32]byte(result), nil
}

func (e *Client) GetEpochStart(ctx context.Context, epoch uint64) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getEpochStartFunction, new(big.Int).SetUint64(epoch), []byte{})
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	timestamp := new(big.Int).SetBytes(result).Uint64()
	return timestamp, nil
}

func (e *Client) GetLastCommittedHeaderEpoch(ctx context.Context) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getLastCommittedHeaderEpochFunction)
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	epoch := new(big.Int).SetBytes(result).Uint64()
	return epoch, nil
}

func (e *Client) GetCurrentPhase(ctx context.Context) (entity.Phase, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getCurrentPhaseFunction)
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	phase := new(big.Int).SetBytes(result).Uint64()
	return entity.Phase(phase), nil
}

func (e *Client) GetCurrentValsetTimestamp(ctx context.Context) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getCurrentValsetTimestampFunction)
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	timestamp := new(big.Int).SetBytes(result).Uint64()
	return timestamp, nil
}

func (e *Client) GetCaptureTimestamp(ctx context.Context) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getCaptureTimestampFunction)
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	timestamp := new(big.Int).SetBytes(result).Uint64()
	return timestamp, nil
}

func (e *Client) GetCaptureTimestampFromValsetHeaderAt(ctx context.Context, epoch uint64) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getCaptureTimestampFromValsetHeaderAtFunction, epoch)
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	timestamp := new(big.Int).SetBytes(result).Uint64()
	return timestamp, nil
}

func (e *Client) GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error) {
	type dtoVaultVotingPower struct {
		Vault       common.Address
		VotingPower *big.Int
	}
	type dtoOperatorVotingPower struct {
		Operator common.Address
		Vaults   []dtoVaultVotingPower
	}

	callMsg, err := constructCallMsg(address.Address, vaultManagerABI, getVotingPowersFunction, [][]byte{}, new(big.Int).SetUint64(timestamp), []byte{})
	if err != nil {
		return nil, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, errors.Errorf("failed to call contract: %w", err)
	}

	var dto []dtoOperatorVotingPower
	err = vaultManagerABI.UnpackIntoInterface(&dto, getVotingPowersFunction, result)
	if err != nil {
		return nil, errors.Errorf("failed to unpack voting powers: %w", err)
	}

	return lo.Map(dto, func(v dtoOperatorVotingPower, _ int) entity.OperatorVotingPower {
		return entity.OperatorVotingPower{
			Operator: v.Operator,
			Vaults: lo.Map(v.Vaults, func(v dtoVaultVotingPower, _ int) entity.VaultVotingPower {
				return entity.VaultVotingPower{
					Vault:       v.Vault,
					VotingPower: v.VotingPower,
				}
			}),
		}
	}), nil
}

func (e *Client) GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error) {
	type dtoKey struct {
		Tag     uint8
		Payload []byte
	}

	type dtoOperatorWithKeys struct {
		Operator common.Address
		Keys     []dtoKey
	}

	callMsg, err := constructCallMsg(address.Address, keyRegistryABI, getKeysFunction, new(big.Int).SetUint64(timestamp), []byte{})
	if err != nil {
		return nil, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, errors.Errorf("failed to call contract: %w", err)
	}

	var dto []dtoOperatorWithKeys

	err = keyRegistryABI.UnpackIntoInterface(&dto, getKeysFunction, result)
	if err != nil {
		return nil, errors.Errorf("failed to unpack keys: %w", err)
	}

	return lo.Map(dto, func(v dtoOperatorWithKeys, _ int) entity.OperatorWithKeys {
		return entity.OperatorWithKeys{
			Operator: v.Operator,
			Keys: lo.Map(v.Keys, func(v dtoKey, _ int) entity.Key {
				return entity.Key{
					Tag:     entity.KeyTag(v.Tag),
					Payload: v.Payload,
				}
			}),
		}
	}), nil
}

func (e *Client) GetRequiredKeyTag(ctx context.Context, timestamp uint64) (entity.KeyTag, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getRequiredKeyTagFunction, new(big.Int).SetUint64(timestamp), []byte{})
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	var keyTag uint8
	err = masterABI.UnpackIntoInterface(&keyTag, getRequiredKeyTagFunction, result)
	if err != nil {
		return 0, errors.Errorf("failed to unpack key tag: %w", err)
	}

	return entity.KeyTag(keyTag), nil
}

func (e *Client) GetQuorumThreshold(ctx context.Context, timestamp uint64, keyTag entity.KeyTag) (uint64, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getQuorumThresholdFunction, keyTag, timestamp, []byte{})
	if err != nil {
		return 0, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return 0, errors.Errorf("failed to call contract: %w", err)
	}

	return new(big.Int).SetBytes(result).Uint64(), nil
}

func (e *Client) GetSubnetwork(ctx context.Context) ([32]byte, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getSubnetworkFunction)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to call contract: %w", err)
	}

	return [32]byte(result), nil
}

func (e *Client) GetNetworkAddress(ctx context.Context) (*common.Address, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getNetworkFunction)
	if err != nil {
		return nil, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, errors.Errorf("failed to call contract: %w", err)
	}

	return abi.ConvertType(result, new(common.Address)).(*common.Address), nil
}

func (e *Client) GetValSetHeaderAt(ctx context.Context, epoch uint64) (entity.ValidatorSetHeader, error) {
	type dtoValsetHeader struct {
		Version            uint8
		RequiredKeyTag     uint8
		Epoch              *big.Int
		CaptureTimestamp   *big.Int
		QuorumThreshold    *big.Int
		ValidatorsSszMRoot [32]byte
		PreviousHeaderHash [32]byte
	}
	type dtoValsetHeaderContainer struct {
		DTOValsetHeader dtoValsetHeader
	}
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getValSetHeaderAtFunction, new(big.Int).SetUint64(epoch))
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to call contract: %w", err)
	}

	dtoContainer := dtoValsetHeaderContainer{}
	err = masterABI.UnpackIntoInterface(&dtoContainer, getValSetHeaderAtFunction, result)
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to unpack config: %w", err)
	}

	valsetHeader := dtoContainer.DTOValsetHeader

	return entity.ValidatorSetHeader{
		Version:            valsetHeader.Version,
		RequiredKeyTag:     entity.KeyTag(valsetHeader.RequiredKeyTag),
		Epoch:              valsetHeader.Epoch.Uint64(),
		CaptureTimestamp:   valsetHeader.CaptureTimestamp.Uint64(),
		QuorumThreshold:    valsetHeader.QuorumThreshold,
		ValidatorsSszMRoot: valsetHeader.ValidatorsSszMRoot,
		PreviousHeaderHash: valsetHeader.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error) {
	callMsg, err := constructCallMsg(e.masterAddress.Address, masterABI, getEip712DomainFunction)
	if err != nil {
		return entity.Eip712Domain{}, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return entity.Eip712Domain{}, errors.Errorf("failed to call contract: %w", err)
	}

	var eip712Domain entity.Eip712Domain

	out, err := masterABI.Unpack(getEip712DomainFunction, result)
	if err != nil {
		return entity.Eip712Domain{}, errors.Errorf("failed to unpack eip712 domain: %w", err)
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

	result, err = e.client.CallContract(tmCtx, callMsg, new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()))
	if err != nil {
		return nil, errors.Errorf("failed to call contract: %w", err)
	}
	if len(result) == 0 {
		// most probably we use incorrect contract address
		return nil, errors.New("no data returned from contract call")
	}

	return result, nil
}

func constructCallMsg(contractAddress common.Address, abi abi.ABI, method string, args ...interface{}) (ethereum.CallMsg, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return ethereum.CallMsg{}, errors.Errorf("failed to pack method: %w", err)
	}

	return ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}, nil
}

var customErrRegExp = regexp.MustCompile(`0x[0-9a-fA-F]{8}`)

func (e *Client) formatEVMError(err error) error {
	type jsonError interface {
		Error() string
		ErrorData() interface{}
		ErrorCode() int
	}
	var errData jsonError
	if !errors.As(err, &errData) {
		return err
	}
	if errData.ErrorCode() != 3 && errData.ErrorData() == nil {
		return err
	}

	matches := customErrRegExp.FindStringSubmatch(errData.Error())
	if len(matches) < 1 {
		return err
	}

	errDef, ok := findErrorBySelector(matches[0])
	if !ok {
		return err
	}

	return errors.Errorf("%w: %s", err, errDef.String())
}

func findErrorBySelector(errSelector string) (abi.Error, bool) {
	for _, errDef := range masterABI.Errors {
		selector := hex.EncodeToString(crypto.Keccak256([]byte(errDef.Sig))[:4])
		if "0x"+selector == errSelector {
			return errDef, true
		}
	}

	return abi.Error{}, false
}
