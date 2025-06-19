package core

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	"middleware-offchain/pkg/proof"
)

type prover interface {
	Prove(proveInput proof.ProveInput) (proof.ProofData, error)
	Verify(valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error)
}

type Config struct {
	MasterRPCURL   string `validate:"required"`
	DriverAddress  string `validate:"required"`
	PrivateKey     []byte
	RequestTimeout time.Duration `validate:"required,gt=0"`
	Prover         prover        `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

// Symbiotic is a facade that provides a unified interface for interacting with the Symbiotic middleware.
type Symbiotic struct {
	evmClient  *evm.Client
	aggregator *aggregator.Aggregator
	deriver    *valsetDeriver.Deriver
}

func NewSymbiotic(cfg Config) (*Symbiotic, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	evmClient, err := evm.NewEVMClient(evm.Config{
		MasterRPCURL:   cfg.MasterRPCURL,
		DriverAddress:  cfg.DriverAddress,
		PrivateKey:     cfg.PrivateKey,
		RequestTimeout: cfg.RequestTimeout,
	})
	if err != nil {
		return nil, errors.Errorf("failed to create EVM client: %w", err)
	}

	agg := aggregator.NewAggregator(cfg.Prover)
	deriver, err := valsetDeriver.NewDeriver(evmClient)
	if err != nil {
		return nil, errors.Errorf("failed to create validator set deriver: %w", err)
	}
	return &Symbiotic{
		evmClient:  evmClient,
		aggregator: agg,
		deriver:    deriver,
	}, nil
}

// ========== Aggregator methods ==========

func (s *Symbiotic) Aggregate(valset entity.ValidatorSet, keyTag entity.KeyTag, verificationType entity.VerificationType, messageHash []byte, signatures []entity.SignatureExtended) (entity.AggregationProof, error) {
	return s.aggregator.Aggregate(valset, keyTag, verificationType, messageHash, signatures)
}

func (s *Symbiotic) VerifyAggregated(
	valset entity.ValidatorSet,
	keyTag entity.KeyTag,
	aggregationProof entity.AggregationProof,
) (bool, error) {
	return s.aggregator.Verify(valset, keyTag, aggregationProof)
}

// ========== Deriver methods ==========

func (s *Symbiotic) GetNetworkData(ctx context.Context) (entity.NetworkData, error) {
	return s.deriver.GetNetworkData(ctx)
}

func (s *Symbiotic) GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error) {
	return s.deriver.GetValidatorSet(ctx, epoch, config)
}

// ========= EVM Client methods ==========

func (s *Symbiotic) GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error) {
	return s.evmClient.GetConfig(ctx, timestamp)
}

func (s *Symbiotic) IsValsetHeaderCommittedAt(ctx context.Context, epoch uint64) (bool, error) {
	return s.evmClient.IsValsetHeaderCommittedAt(ctx, epoch)
}

func (s *Symbiotic) GetCurrentEpoch(ctx context.Context) (uint64, error) {
	return s.evmClient.GetCurrentEpoch(ctx)
}

func (s *Symbiotic) GetPreviousHeaderHash(ctx context.Context) (common.Hash, error) {
	return s.evmClient.GetPreviousHeaderHash(ctx)
}

func (s *Symbiotic) GetPreviousHeaderHashAt(ctx context.Context, epoch uint64) (common.Hash, error) {
	return s.evmClient.GetPreviousHeaderHashAt(ctx, epoch)
}

func (s *Symbiotic) GetHeaderHash(ctx context.Context) (common.Hash, error) {
	return s.evmClient.GetHeaderHash(ctx)
}

func (s *Symbiotic) GetHeaderHashAt(ctx context.Context, epoch uint64) (common.Hash, error) {
	return s.evmClient.GetHeaderHashAt(ctx, epoch)
}

func (s *Symbiotic) GetEpochStart(ctx context.Context, epoch uint64) (uint64, error) {
	return s.evmClient.GetEpochStart(ctx, epoch)
}

func (s *Symbiotic) GetLastCommittedHeaderEpoch(ctx context.Context) (uint64, error) {
	return s.evmClient.GetLastCommittedHeaderEpoch(ctx)
}

func (s *Symbiotic) GetCaptureTimestampFromValsetHeaderAt(ctx context.Context, epoch uint64) (uint64, error) {
	return s.evmClient.GetCaptureTimestampFromValsetHeaderAt(ctx, epoch)
}

func (s *Symbiotic) GetVotingPowers(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error) {
	return s.evmClient.GetVotingPowers(ctx, address, timestamp)
}

func (s *Symbiotic) GetKeys(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorWithKeys, error) {
	return s.evmClient.GetKeys(ctx, address, timestamp)
}

func (s *Symbiotic) GetSubnetwork(ctx context.Context) (common.Hash, error) {
	return s.evmClient.GetSubnetwork(ctx)
}

func (s *Symbiotic) GetNetworkAddress(ctx context.Context) (*common.Address, error) {
	return s.evmClient.GetNetworkAddress(ctx)
}

func (s *Symbiotic) GetValSetHeaderAt(ctx context.Context, epoch uint64) (entity.ValidatorSetHeader, error) {
	return s.evmClient.GetValSetHeaderAt(ctx, epoch)
}

func (s *Symbiotic) GetEip712Domain(ctx context.Context) (entity.Eip712Domain, error) {
	return s.evmClient.GetEip712Domain(ctx)
}

func (s *Symbiotic) CommitValsetHeader(ctx context.Context, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) (entity.TxResult, error) {
	return s.evmClient.CommitValsetHeader(ctx, header, extraData, proof)
}

func (s *Symbiotic) VerifyQuorumSig(ctx context.Context, epoch uint64, message []byte, keyTag entity.KeyTag, threshold *big.Int, proof []byte) (bool, error) {
	return s.evmClient.VerifyQuorumSig(ctx, epoch, message, keyTag, threshold, proof)
}
