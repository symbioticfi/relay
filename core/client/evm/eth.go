package evm

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

//go:embed abi/IVotingPowerProvider.abi.json
var votingPowerProviderAbiJSON []byte
var votingPowerProviderABI abi.ABI

//go:embed abi/IKeyRegistry.abi.json
var keyRegistryAbiJSON []byte
var keyRegistryABI abi.ABI

var (
	getVotingPowersFunction = "getVotingPowersAt"
	getKeysFunction         = "getKeysAt"
)

type Config struct {
	MasterRPCURL   string `validate:"required"`
	DriverAddress  string `validate:"required"`
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

	driver     *gen.IValSetDriverCaller
	settlement *gen.ISettlement
	masterPK   *ecdsa.PrivateKey // could be nil for read-only access
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

	driver, err := gen.NewIValSetDriverCaller(common.HexToAddress(cfg.DriverAddress), client)
	if err != nil {
		return nil, errors.Errorf("failed to create new driver contract: %w", err)
	}

	// todo multiple settlements instead of it
	netConfig, err := driver.GetConfig(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
	})
	if err != nil {
		return nil, errors.Errorf("failed to get network configuration: %w", err)
	}
	if len(netConfig.Replicas) != 1 {
		return nil, errors.Errorf("suppported only 1 replica in configuration, but got %d", len(netConfig.Replicas))
	}
	settlement, err := gen.NewISettlement(netConfig.Replicas[0].Addr, client)
	if err != nil {
		return nil, errors.Errorf("failed to create settlement contract: %w", err)
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
			Address: common.HexToAddress(cfg.DriverAddress),
			ChainId: 111,
		},
		masterPK:   pk,
		cfg:        cfg,
		driver:     driver,
		settlement: settlement,
	}, nil
}

func initABI() error {
	var err error

	votingPowerProviderABI, err = abi.JSON(bytes.NewReader(votingPowerProviderAbiJSON))
	if err != nil {
		return errors.Errorf("failed to parse vault manager ABI: %w", err)
	}

	keyRegistryABI, err = abi.JSON(bytes.NewReader(keyRegistryAbiJSON))
	if err != nil {
		return errors.Errorf("failed to parse key registry ABI: %w", err)
	}

	return nil
}

func (e *Client) GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	dtoConfig, err := e.driver.GetConfigAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(timestamp))
	if err != nil {
		return entity.NetworkConfig{}, errors.Errorf("failed to call getConfigAt: %w", err)
	}

	return entity.NetworkConfig{
		VotingPowerProviders: lo.Map(dtoConfig.VotingPowerProviders, func(v gen.IValSetDriverCrossChainAddress, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				ChainId: v.ChainId,
				Address: v.Addr,
			}
		}),
		KeysProvider: entity.CrossChainAddress{
			Address: dtoConfig.KeysProvider.Addr,
			ChainId: dtoConfig.KeysProvider.ChainId,
		},
		Replicas: lo.Map(dtoConfig.Replicas, func(v gen.IValSetDriverCrossChainAddress, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				ChainId: v.ChainId,
				Address: v.Addr,
			}
		}),
		VerificationType:        entity.VerificationType(dtoConfig.VerificationType),
		MaxVotingPower:          dtoConfig.MaxVotingPower,
		MinInclusionVotingPower: dtoConfig.MinInclusionVotingPower,
		MaxValidatorsCount:      dtoConfig.MaxValidatorsCount,
		RequiredKeyTags: lo.Map(dtoConfig.RequiredKeyTags, func(v uint8, _ int) entity.KeyTag {
			return entity.KeyTag(v)
		}),
		RequiredHeaderKeyTag: entity.KeyTag(dtoConfig.RequiredHeaderKeyTag),
		QuorumThresholds: lo.Map(dtoConfig.QuorumThresholds, func(v gen.IValSetDriverQuorumThreshold, _ int) entity.QuorumThreshold {
			return entity.QuorumThreshold{
				KeyTag:          entity.KeyTag(v.KeyTag),
				QuorumThreshold: v.QuorumThreshold,
			}
		}),
	}, nil
}

func (e *Client) GetCurrentEpoch(ctx context.Context) (uint64, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	epoch, err := e.driver.GetCurrentEpoch(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return 0, errors.Errorf("failed to call getCurrentEpoch: %w", e.formatEVMContractError(gen.IValSetDriverMetaData, err))
	}
	return epoch.Uint64(), nil
}

func (e *Client) GetEpochStart(ctx context.Context, epoch uint64) (uint64, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	epochStart, err := e.driver.GetEpochStart(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch), []byte{})
	if err != nil {
		return 0, errors.Errorf("failed to call getEpochStart: %w", e.formatEVMContractError(gen.IValSetDriverMetaData, err))
	}
	return epochStart.Uint64(), nil
}

func (e *Client) GetSubnetwork(ctx context.Context) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	subnetwork, err := e.driver.SUBNETWORK(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getSubnetwork: %w", err)
	}

	return subnetwork, nil
}

func (e *Client) GetNetworkAddress(ctx context.Context) (*common.Address, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	networkAddress, err := e.driver.NETWORK(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return nil, errors.Errorf("failed to call getSubnetwork: %w", err)
	}

	return &networkAddress, nil
}

func (e *Client) IsValsetHeaderCommittedAt(ctx context.Context, epoch uint64) (bool, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	ok, err := e.settlement.IsValSetHeaderCommittedAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return false, errors.Errorf("failed to call isValsetHeaderCommittedAt: %w", err)
	}
	return ok, nil
}

func (e *Client) GetPreviousHeaderHash(ctx context.Context) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	hash, err := e.settlement.GetPreviousHeaderHashFromValSetHeader(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getPreviousHeaderHash: %w", err)
	}

	return hash, nil
}

func (e *Client) GetPreviousHeaderHashAt(ctx context.Context, epoch uint64) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	hash, err := e.settlement.GetPreviousHeaderHashFromValSetHeaderAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getPreviousHeaderHashAt: %w", err)
	}

	return hash, nil
}

func (e *Client) GetHeaderHash(ctx context.Context) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	hash, err := e.settlement.GetValSetHeaderHash(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getValSetHeaderHash: %w", err)
	}

	return hash, nil
}

func (e *Client) GetHeaderHashAt(ctx context.Context, epoch uint64) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	hash, err := e.settlement.GetValSetHeaderHashAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getValSetHeaderHashAt: %w", err)
	}

	return hash, nil
}

func (e *Client) GetLastCommittedHeaderEpoch(ctx context.Context) (uint64, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	epoch, err := e.settlement.GetLastCommittedHeaderEpoch(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return 0, errors.Errorf("failed to call getValSetHeaderHashAt: %w", err)
	}

	// todo if zero epoch need to check if it's committed or not

	return epoch.Uint64(), nil
}

func (e *Client) GetCaptureTimestampFromValsetHeaderAt(ctx context.Context, epoch uint64) (uint64, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	timestamp, err := e.settlement.GetCaptureTimestampFromValSetHeaderAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return 0, errors.Errorf("failed to call getCaptureTimestampFromValSetHeaderAt: %w", err)
	}

	return timestamp.Uint64(), nil
}

func (e *Client) GetValSetHeaderAt(ctx context.Context, epoch uint64) (entity.ValidatorSetHeader, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	header, err := e.settlement.GetValSetHeaderAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to call getValSetHeaderAt: %w", err)
	}

	return entity.ValidatorSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     entity.KeyTag(header.RequiredKeyTag),
		Epoch:              header.Epoch.Uint64(),
		CaptureTimestamp:   header.CaptureTimestamp.Uint64(),
		QuorumThreshold:    header.QuorumThreshold,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetValSetHeader(ctx context.Context) (entity.ValidatorSetHeader, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	header, err := e.settlement.GetValSetHeader(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to call getValSetHeader: %w", err)
	}

	return entity.ValidatorSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     entity.KeyTag(header.RequiredKeyTag),
		Epoch:              header.Epoch.Uint64(),
		CaptureTimestamp:   header.CaptureTimestamp.Uint64(),
		QuorumThreshold:    header.QuorumThreshold,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	eip712Domain, err := e.settlement.Eip712Domain(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return entity.Eip712Domain{}, errors.Errorf("failed to call Eip712Domain: %w", err)
	}

	return entity.Eip712Domain{
		Fields:            eip712Domain.Fields,
		Name:              eip712Domain.Name,
		Version:           eip712Domain.Version,
		ChainId:           eip712Domain.ChainId,
		VerifyingContract: eip712Domain.VerifyingContract,
		Salt:              new(big.Int).SetBytes(eip712Domain.Salt[:]),
		Extensions:        eip712Domain.Extensions,
	}, nil
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

	callMsg, err := constructCallMsg(address.Address, votingPowerProviderABI, getVotingPowersFunction, [][]byte{}, new(big.Int).SetUint64(timestamp))
	if err != nil {
		return nil, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return nil, errors.Errorf("failed to call contract: %w", err)
	}

	var dto []dtoOperatorVotingPower
	err = votingPowerProviderABI.UnpackIntoInterface(&dto, getVotingPowersFunction, result)
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

	callMsg, err := constructCallMsg(address.Address, keyRegistryABI, getKeysFunction, new(big.Int).SetUint64(timestamp))
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

type metadata interface {
	GetAbi() (*abi.ABI, error)
}

func (e *Client) formatEVMContractError(meta metadata, err error) error {
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

	parsedAbi, err := meta.GetAbi()
	if err != nil {
		return err
	}

	hexSelector, err := hexutil.Decode(matches[0])
	if err != nil {
		return err
	}

	if len(hexSelector) < 4 {
		return errors.New("too short hex selector")
	}

	contractError, err := parsedAbi.ErrorByID([4]byte(hexSelector[:4]))
	if err != nil {
		return err
	}

	return errors.Errorf("%w: %s", err, contractError.String())
}

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
	// todo handle error
	settlementAbi, _ := gen.ISettlementMetaData.GetAbi()

	for _, errDef := range settlementAbi.Errors {
		selector := hex.EncodeToString(crypto.Keccak256([]byte(errDef.Sig))[:4])
		if "0x"+selector == errSelector {
			return errDef, true
		}
	}

	return abi.Error{}, false
}
