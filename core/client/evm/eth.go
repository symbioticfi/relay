package evm

import (
	"context"
	_ "embed"
	"encoding/hex"
	"log/slog"
	"math/big"
	"regexp"
	"time"

	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"

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

	"github.com/symbioticfi/relay/core/client/evm/gen"
	"github.com/symbioticfi/relay/core/entity"
)

type metrics interface {
	ObserveEVMMethodCall(method, status string, d time.Duration)
	ObserveCommitValsetHeaderParams(chainID uint64, gasUsed uint64, effectiveGasPrice *big.Int)
}

// IEvmClient defines the interface for EVM client operations
type IEvmClient interface {
	GetChains() []uint64
	GetSubnetwork(ctx context.Context) (common.Hash, error)
	GetNetworkAddress(ctx context.Context) (common.Address, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEip712Domain(ctx context.Context, addr entity.CrossChainAddress) (entity.Eip712Domain, error)
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetCurrentEpochDuration(ctx context.Context) (uint64, error)
	GetEpochDuration(ctx context.Context, epoch uint64) (uint64, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
	IsValsetHeaderCommittedAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (bool, error)
	GetPreviousHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (common.Hash, error)
	GetPreviousHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
	GetHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (common.Hash, error)
	GetHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error)
	GetCaptureTimestampFromValsetHeaderAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (uint64, error)
	GetValSetHeaderAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (entity.ValidatorSetHeader, error)
	GetValSetHeader(ctx context.Context, addr entity.CrossChainAddress) (entity.ValidatorSetHeader, error)
	GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error)
	GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error)
	CommitValsetHeader(ctx context.Context, addr entity.CrossChainAddress, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) (entity.TxResult, error)
	RegisterOperator(ctx context.Context, addr entity.CrossChainAddress) (entity.TxResult, error)
	RegisterKey(ctx context.Context, addr entity.CrossChainAddress, keyTag entity.KeyTag, key entity.CompactPublicKey, signature entity.RawSignature, extraData []byte) (entity.TxResult, error)
	SetGenesis(ctx context.Context, addr entity.CrossChainAddress, header entity.ValidatorSetHeader, extraData []entity.ExtraData) (entity.TxResult, error)
	VerifyQuorumSig(ctx context.Context, addr entity.CrossChainAddress, epoch uint64, message []byte, keyTag entity.KeyTag, threshold *big.Int, proof []byte) (bool, error)
	IsValsetHeaderCommittedAtEpochs(ctx context.Context, addr entity.CrossChainAddress, epochs []uint64) ([]bool, error)
}

type Config struct {
	ChainURLs      []string                 `validate:"required"`
	DriverAddress  entity.CrossChainAddress `validate:"required"`
	RequestTimeout time.Duration            `validate:"required,gt=0"`
	KeyProvider    keyprovider.KeyProvider
	Metrics        metrics
	MaxCalls       int
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type Client struct {
	cfg Config

	conns  map[uint64]*ethclient.Client
	driver *gen.IValSetDriverCaller

	metrics metrics
}

func NewEvmClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	conns := make(map[uint64]*ethclient.Client)

	for _, chainURL := range cfg.ChainURLs {
		client, err := ethclient.DialContext(ctx, chainURL)
		if err != nil {
			return nil, errors.Errorf("failed to connect to Ethereum client: %w", err)
		}
		chainID, err := client.ChainID(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to get chain ID: %w", err)
		}

		conns[chainID.Uint64()] = client
	}

	if _, found := conns[cfg.DriverAddress.ChainId]; !found {
		return nil, errors.Errorf("driver's chain rpc url omitted")
	}

	driver, err := gen.NewIValSetDriverCaller(cfg.DriverAddress.Address, conns[cfg.DriverAddress.ChainId])
	if err != nil {
		return nil, errors.Errorf("failed to create driver contract: %w", err)
	}

	return &Client{
		cfg:     cfg,
		conns:   conns,
		driver:  driver,
		metrics: cfg.Metrics,
	}, nil
}

func (e *Client) GetChains() []uint64 {
	chainIds := make([]uint64, 0, len(e.conns))
	for chainId := range e.conns {
		chainIds = append(chainIds, chainId)
	}
	return chainIds
}

func (e *Client) GetConfig(ctx context.Context, timestamp uint64) (_ entity.NetworkConfig, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetConfigAt", err, now)
	}(time.Now())

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
		MaxMissingEpochs: 0,                          // TODO set it after core update
		GrowthStrategy:   entity.GrowthStrategyAsync, // TODO same
	}, nil
}

func (e *Client) GetCurrentEpoch(ctx context.Context) (_ uint64, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetCurrentEpoch", err, now)
	}(time.Now())

	epoch, err := e.driver.GetCurrentEpoch(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return 0, errors.Errorf("failed to call getCurrentEpoch: %w", e.formatEVMContractError(gen.IValSetDriverMetaData, err))
	}
	return epoch.Uint64(), nil
}

func (e *Client) GetCurrentEpochDuration(ctx context.Context) (_ uint64, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetCurrentEpochDuration", err, now)
	}(time.Now())

	epochDuration, err := e.driver.GetCurrentEpochDuration(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return 0, errors.Errorf("failed to call getCurrentEpochDuration: %w", e.formatEVMContractError(gen.IValSetDriverMetaData, err))
	}
	return epochDuration.Uint64(), nil
}

func (e *Client) GetEpochDuration(ctx context.Context, epoch uint64) (_ uint64, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetEpochDuration", err, now)
	}(time.Now())

	epochDuration, err := e.driver.GetEpochDuration(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch), []byte{})
	if err != nil {
		return 0, errors.Errorf("failed to call getEpochDuration: %w", e.formatEVMContractError(gen.IValSetDriverMetaData, err))
	}
	return epochDuration.Uint64(), nil
}

func (e *Client) GetEpochStart(ctx context.Context, epoch uint64) (_ uint64, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetEpochStart", err, now)
	}(time.Now())

	epochStart, err := e.driver.GetEpochStart(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(epoch), []byte{})
	if err != nil {
		return 0, errors.Errorf("failed to call getEpochStart: %w", e.formatEVMContractError(gen.IValSetDriverMetaData, err))
	}
	return epochStart.Uint64(), nil
}

func (e *Client) GetSubnetwork(ctx context.Context) (_ common.Hash, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("SUBNETWORK", err, now)
	}(time.Now())

	subnetwork, err := e.driver.SUBNETWORK(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to call getSubnetwork: %w", err)
	}

	return subnetwork, nil
}

func (e *Client) GetNetworkAddress(ctx context.Context) (_ common.Address, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("NETWORK", err, now)
	}(time.Now())

	networkAddress, err := e.driver.NETWORK(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	})
	if err != nil {
		return common.Address{}, errors.Errorf("failed to call getSubnetwork: %w", err)
	}

	return networkAddress, nil
}

func (e *Client) IsValsetHeaderCommittedAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (_ bool, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("IsValSetHeaderCommittedAt", err, now)
	}(time.Now())

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

func (e *Client) GetPreviousHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (_ common.Hash, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetPreviousHeaderHashFromValSetHeader", err, now)
	}(time.Now())

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

func (e *Client) GetPreviousHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (_ common.Hash, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetPreviousHeaderHashFromValSetHeaderAt", err, now)
	}(time.Now())

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

func (e *Client) GetHeaderHash(ctx context.Context, addr entity.CrossChainAddress) (_ common.Hash, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetValSetHeaderHash", err, now)
	}(time.Now())

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

func (e *Client) GetHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (_ common.Hash, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetValSetHeaderHashAt", err, now)
	}(time.Now())

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

func (e *Client) GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (_ uint64, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetLastCommittedHeaderEpoch", err, now)
	}(time.Now())

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

func (e *Client) GetCaptureTimestampFromValsetHeaderAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (_ uint64, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetCaptureTimestampFromValSetHeaderAt", err, now)
	}(time.Now())

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

func (e *Client) GetValSetHeaderAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (_ entity.ValidatorSetHeader, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetValSetHeaderAt", err, now)
	}(time.Now())

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
		Epoch:              entity.Epoch(header.Epoch.Uint64()),
		CaptureTimestamp:   header.CaptureTimestamp.Uint64(),
		QuorumThreshold:    entity.ToVotingPower(header.QuorumThreshold),
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetValSetHeader(ctx context.Context, addr entity.CrossChainAddress) (_ entity.ValidatorSetHeader, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetValSetHeader", err, now)
	}(time.Now())

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
		Epoch:              entity.Epoch(header.Epoch.Uint64()),
		CaptureTimestamp:   header.CaptureTimestamp.Uint64(),
		QuorumThreshold:    entity.ToVotingPower(header.QuorumThreshold),
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}, nil
}

func (e *Client) GetEip712Domain(ctx context.Context, addr entity.CrossChainAddress) (_ entity.Eip712Domain, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("Eip712Domain", err, now)
	}(time.Now())

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

func (e *Client) GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) (_ []entity.OperatorVotingPower, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetVotingPowersAt", err, now)
	}(time.Now())

	multicallExists, err := e.multicallExists(ctx, address.ChainId)
	if err != nil {
		return nil, errors.Errorf("multicall check failed: %v", err)
	}

	if multicallExists {
		return e.getVotingPowersMulticall(ctx, address, timestamp)
	}

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

func (e *Client) GetOperators(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) (_ []common.Address, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetOperators", err, now)
	}(time.Now())

	votingPowerProvider, err := e.getVotingPowerProviderContract(address)
	if err != nil {
		return nil, errors.Errorf("failed to create voting power provider contract: %w", err)
	}

	operators, err := votingPowerProvider.GetOperatorsAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(timestamp))
	if err != nil {
		return nil, errors.Errorf("failed to call getOperatorsAt: %w", e.formatEVMContractError(gen.IVotingPowerProviderMetaData, err))
	}

	return operators, nil
}

func (e *Client) GetKeysOperators(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) (_ []common.Address, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetKeysOperators", err, now)
	}(time.Now())

	keyRegistry, err := e.getKeyRegistryContract(address)
	if err != nil {
		return nil, errors.Errorf("failed to create voting power provider contract: %w", err)
	}

	operators, err := keyRegistry.GetKeysOperatorsAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, new(big.Int).SetUint64(timestamp))
	if err != nil {
		return nil, errors.Errorf("failed to call getKeysOperatorsAt: %w", e.formatEVMContractError(gen.IKeyRegistryMetaData, err))
	}

	return operators, nil
}

func (e *Client) GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) (_ []entity.OperatorWithKeys, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("GetKeysAt", err, now)
	}(time.Now())

	multicallExists, err := e.multicallExists(ctx, address.ChainId)
	if err != nil {
		return nil, errors.Errorf("multicall check failed: %v", err)
	}

	if multicallExists {
		return e.getKeysMulticall(ctx, address, timestamp)
	}

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

func (e *Client) IsValsetHeaderCommittedAtEpochs(ctx context.Context, addr entity.CrossChainAddress, epochs []uint64) (_ []bool, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("IsValSetHeaderCommittedAt", err, now)
	}(time.Now())

	multicallExists, err := e.multicallExists(ctx, addr.ChainId)
	if err != nil {
		return nil, errors.Errorf("multicall check failed: %v", err)
	}
	if !multicallExists {
		return nil, errors.New("multicall not available on this chain")
	}

	abi, err := gen.ISettlementMetaData.GetAbi()
	if err != nil {
		return nil, errors.Errorf("failed to get ABI: %v", err)
	}

	isCommitted := make([]bool, 0, len(epochs))
	calls := make([]Call, 0, len(epochs))

	for _, epoch := range epochs {
		bytes, err := abi.Pack("isValSetHeaderCommittedAt", big.NewInt(int64(epoch)))
		if err != nil {
			return nil, errors.Errorf("failed to get bytes: %v", err)
		}

		calls = append(calls, Call{
			Target:       addr.Address,
			CallData:     bytes,
			AllowFailure: false,
		})
	}

	outs, err := e.multicall(toCtx, addr.ChainId, calls)
	if err != nil {
		return nil, errors.Errorf("multicall failed: %v", err)
	}

	if len(outs) != len(calls) {
		return nil, errors.Errorf("multicall failed: expected %d calls, got %d", len(calls), len(outs))
	}

	for _, out := range outs {
		var res bool

		if err := abi.UnpackIntoInterface(&res, "isValSetHeaderCommittedAt", out.ReturnData); err != nil {
			return nil, errors.Errorf("failed to unpack isValSetHeaderCommittedAt: %v", err)
		}

		isCommitted = append(isCommitted, res)
	}

	return isCommitted, nil
}

var customErrRegExp = regexp.MustCompile(`0x[0-9a-fA-F]{8}`)

type metadata interface {
	GetAbi() (*abi.ABI, error)
}

func (e *Client) formatEVMContractError(meta metadata, originalErr error) error {
	type jsonError interface {
		Error() string
		ErrorData() interface{}
		ErrorCode() int
	}
	var errData jsonError
	if !errors.As(originalErr, &errData) {
		return originalErr
	}
	if errData.ErrorCode() != 3 && errData.ErrorData() == nil {
		return originalErr
	}

	matches := customErrRegExp.FindStringSubmatch(errData.Error())
	if len(matches) < 1 {
		return originalErr
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

	return errors.Errorf("%w: %s", originalErr, contractError.String())
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

func (e *Client) getKeyRegistryContract(addr entity.CrossChainAddress) (*gen.IKeyRegistry, error) {
	client, ok := e.conns[addr.ChainId]
	if !ok {
		return nil, errors.Errorf("no connection for chain ID %d: %w", addr.ChainId, entity.ErrChainNotFound)
	}

	return gen.NewIKeyRegistry(addr.Address, client)
}

func (e *Client) getOperatorRegistryContract(addr entity.CrossChainAddress) (*gen.OperatorRegistry, error) {
	client, ok := e.conns[addr.ChainId]
	if !ok {
		return nil, errors.Errorf("no connection for chain ID %d: %w", addr.ChainId, entity.ErrChainNotFound)
	}

	return gen.NewOperatorRegistry(addr.Address, client)
}

func findErrorBySelector(errSelector string) (abi.Error, bool) {
	settlementAbi, err := gen.ISettlementMetaData.GetAbi()
	if err != nil {
		slog.Warn("Failed to get settlement ABI", "error", err)
		return abi.Error{}, false
	}

	for _, errDef := range settlementAbi.Errors {
		selector := hex.EncodeToString(crypto.Keccak256([]byte(errDef.Sig))[:4])
		if "0x"+selector == errSelector {
			return errDef, true
		}
	}

	return abi.Error{}, false
}

func (e *Client) observeMetrics(method string, err error, start time.Time) {
	if e.metrics != nil {
		status := lo.Ternary(err != nil, "error", "success")
		e.metrics.ObserveEVMMethodCall(method, status, time.Since(start))
	}
}
