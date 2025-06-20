package evm

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/hex"
	"math/big"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/core/client/evm/gen"
	"middleware-offchain/core/entity"
)

type Config struct {
	Chains         []entity.ChainURL        `validate:"required"`
	DriverAddress  entity.CrossChainAddress `validate:"required"`
	RequestTimeout time.Duration            `validate:"required,gt=0"`
	PrivateKey     []byte
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
	cfg Config

	conns  map[uint64]*ethclient.Client
	driver *gen.IValSetDriverCaller

	masterPK *ecdsa.PrivateKey // could be nil for read-only access
}

func NewEVMClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	conns := make(map[uint64]*ethclient.Client)

	for _, chainURL := range cfg.Chains {
		client, err := ethclient.DialContext(ctx, chainURL.RPCURL)
		if err != nil {
			return nil, errors.Errorf("failed to connect to Ethereum client: %w", err)
		}
		chainID, err := client.ChainID(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to get chain ID: %w", err)
		}
		if chainID.Uint64() != chainURL.ChainID {
			return nil, errors.Errorf("chain ID mismatch: expected %d, got %d", chainURL.ChainID, chainID.Uint64())
		}

		conns[chainURL.ChainID] = client
	}

	var pk *ecdsa.PrivateKey
	if cfg.PrivateKey != nil {
		var err error
		pk, err = crypto.ToECDSA(cfg.PrivateKey)
		if err != nil {
			return nil, errors.Errorf("failed to convert private key: %w", err)
		}
	}

	driver, err := gen.NewIValSetDriverCaller(cfg.DriverAddress.Address, conns[cfg.DriverAddress.ChainId])
	if err != nil {
		return nil, errors.Errorf("failed to create driver contract: %w", err)
	}

	return &Client{
		cfg:      cfg,
		conns:    conns,
		driver:   driver,
		masterPK: pk,
	}, nil
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
		MaxVotingPower:          entity.ToVotingPower(dtoConfig.MaxVotingPower),
		MinInclusionVotingPower: entity.ToVotingPower(dtoConfig.MinInclusionVotingPower),
		MaxValidatorsCount:      entity.ToVotingPower(dtoConfig.MaxValidatorsCount),
		RequiredKeyTags: lo.Map(dtoConfig.RequiredKeyTags, func(v uint8, _ int) entity.KeyTag {
			return entity.KeyTag(v)
		}),
		RequiredHeaderKeyTag: entity.KeyTag(dtoConfig.RequiredHeaderKeyTag),
		QuorumThresholds: lo.Map(dtoConfig.QuorumThresholds, func(v gen.IValSetDriverQuorumThreshold, _ int) entity.QuorumThreshold {
			return entity.QuorumThreshold{
				KeyTag:          entity.KeyTag(v.KeyTag),
				QuorumThreshold: entity.ToQuorumThresholdPct(v.QuorumThreshold),
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

func (e *Client) IsValsetHeaderCommittedAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (bool, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return false, errors.Errorf("failed to get settlement contract: %w", err)
	}

	ok, err := settlement.IsValSetHeaderCommittedAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return false, errors.Errorf("failed to call isValsetHeaderCommittedAt: %w", err)
	}
	return ok, nil
}

func (e *Client) GetPreviousHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	hash, err := settlement.GetPreviousHeaderHashFromValSetHeader(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getPreviousHeaderHash: %w", err)
	}

	return hash, nil
}

func (e *Client) GetPreviousHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	hash, err := settlement.GetPreviousHeaderHashFromValSetHeaderAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getPreviousHeaderHashAt: %w", err)
	}

	return hash, nil
}

func (e *Client) GetHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	hash, err := settlement.GetValSetHeaderHash(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getValSetHeaderHash: %w", err)
	}

	return hash, nil
}

func (e *Client) GetHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	hash, err := settlement.GetValSetHeaderHashAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getValSetHeaderHashAt: %w", err)
	}

	return hash, nil
}

func (e *Client) GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return 0, errors.Errorf("failed to get settlement contract: %w", err)
	}

	epoch, err := settlement.GetLastCommittedHeaderEpoch(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return 0, errors.Errorf("failed to call getValSetHeaderHashAt: %w", err)
	}

	// todo if zero epoch need to check if it's committed or not

	return epoch.Uint64(), nil
}

func (e *Client) GetCaptureTimestampFromValsetHeaderAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (uint64, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return 0, errors.Errorf("failed to get settlement contract: %w", err)
	}

	timestamp, err := settlement.GetCaptureTimestampFromValSetHeaderAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch))
	if err != nil {
		return 0, errors.Errorf("failed to call getCaptureTimestampFromValSetHeaderAt: %w", err)
	}

	return timestamp.Uint64(), nil
}

func (e *Client) GetValSetHeaderAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (entity.ValidatorSetHeader, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	header, err := settlement.GetValSetHeaderAt(&bind.CallOpts{
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
		QuorumThreshold:    entity.ToVotingPower(header.QuorumThreshold),
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetValSetHeader(ctx context.Context, addr entity.CrossChainAddress) (entity.ValidatorSetHeader, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	header, err := settlement.GetValSetHeader(&bind.CallOpts{
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
		QuorumThreshold:    entity.ToVotingPower(header.QuorumThreshold),
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetEip712Domain(ctx context.Context, addr entity.CrossChainAddress) (entity.Eip712Domain, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return entity.Eip712Domain{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	eip712Domain, err := settlement.Eip712Domain(&bind.CallOpts{
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
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	votingPowerProvider, err := e.getVotingPowerProviderContract(address)
	if err != nil {
		return nil, errors.Errorf("failed to create voting power provider contract: %w", err)
	}

	votingPowersAt, err := votingPowerProvider.GetVotingPowersAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, [][]byte{}, new(big.Int).SetUint64(timestamp))
	if err != nil {
		return nil, errors.Errorf("failed to call getVotingPowersAt: %w", e.formatEVMContractError(gen.IVotingPowerProviderMetaData, err))
	}

	return lo.Map(votingPowersAt, func(v gen.IVotingPowerProviderOperatorVotingPower, _ int) entity.OperatorVotingPower {
		return entity.OperatorVotingPower{
			Operator: v.Operator,
			Vaults: lo.Map(v.Vaults, func(v gen.IVotingPowerProviderVaultVotingPower, _ int) entity.VaultVotingPower {
				return entity.VaultVotingPower{
					Vault:       v.Vault,
					VotingPower: entity.ToVotingPower(v.VotingPower),
				}
			}),
		}
	}), nil
}

func (e *Client) GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	keyRegistry, err := e.getKeyRegistryContract(address)
	if err != nil {
		return nil, errors.Errorf("failed to create key registry contract: %w", err)
	}

	keys, err := keyRegistry.GetKeysAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(timestamp))
	if err != nil {
		return nil, errors.Errorf("failed to call getKeysAt: %w", e.formatEVMContractError(gen.IKeyRegistryMetaData, err))
	}

	return lo.Map(keys, func(v gen.IKeyRegistryOperatorWithKeys, _ int) entity.OperatorWithKeys {
		return entity.OperatorWithKeys{
			Operator: v.Operator,
			Keys: lo.Map(v.Keys, func(v gen.IKeyRegistryKey, _ int) entity.ValidatorKey {
				return entity.ValidatorKey{
					Tag:     entity.KeyTag(v.Tag),
					Payload: v.Payload,
				}
			}),
		}
	}), nil
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

func (e *Client) getSettlementContract(addr entity.CrossChainAddress) (*gen.ISettlement, error) {
	client, ok := e.conns[addr.ChainId]
	if !ok {
		return nil, errors.Errorf("no connection for chain ID %d: %w", addr.ChainId, entity.ErrChainNotFound)
	}

	return gen.NewISettlement(addr.Address, client)
}

func (e *Client) getVotingPowerProviderContract(addr entity.CrossChainAddress) (*gen.IVotingPowerProviderCaller, error) {
	client, ok := e.conns[addr.ChainId]
	if !ok {
		return nil, errors.Errorf("no connection for chain ID %d: %w", addr.ChainId, entity.ErrChainNotFound)
	}

	return gen.NewIVotingPowerProviderCaller(addr.Address, client)
}

func (e *Client) getKeyRegistryContract(addr entity.CrossChainAddress) (*gen.IKeyRegistryCaller, error) {
	client, ok := e.conns[addr.ChainId]
	if !ok {
		return nil, errors.Errorf("no connection for chain ID %d: %w", addr.ChainId, entity.ErrChainNotFound)
	}

	return gen.NewIKeyRegistryCaller(addr.Address, client)
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
