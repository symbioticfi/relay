package valset_generator

import (
	"context"
	"encoding/hex"
	"log/slog"
	"math/big"
	"sync"
	"time"

	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/aggregator"
	"github.com/symbioticfi/relay/pkg/log"
)

type signer interface {
	Sign(ctx context.Context, req entity.SignatureRequest) error
}

type evmClient interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)

	IsValsetHeaderCommittedAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (bool, error)
	CommitValsetHeader(ctx context.Context, addr entity.CrossChainAddress, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) (entity.TxResult, error)
	SetGenesis(ctx context.Context, addr entity.CrossChainAddress, header entity.ValidatorSetHeader, extraData []entity.ExtraData) (entity.TxResult, error)
}

type repo interface {
	GetLatestValidatorSet(ctx context.Context) (entity.ValidatorSet, error)

	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetConfigByEpoch(ctx context.Context, epoch uint64) (entity.NetworkConfig, error)
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	SavePendingValidatorSet(ctx context.Context, reqHash common.Hash, valset entity.ValidatorSet) error
	GetPendingValidatorSet(ctx context.Context, reqHash common.Hash) (entity.ValidatorSet, error)
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
	GetNetworkData(ctx context.Context, addr entity.CrossChainAddress) (entity.NetworkData, error)
}

type Config struct {
	Signer          signer        `validate:"required"`
	EvmClient       evmClient     `validate:"required"`
	Repo            repo          `validate:"required"`
	Deriver         deriver       `validate:"required"`
	PollingInterval time.Duration `validate:"required,gt=0"`
	IsCommitter     bool
	Aggregator      aggregator.Aggregator
	GrowthStrategy  strategyTypes.GrowthStrategy
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	return nil
}

type Service struct {
	cfg            Config
	generatedEpoch uint64
	mutex          sync.Mutex
}

func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Service{
		cfg: cfg,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "generator")

	slog.InfoContext(ctx, "Starting valset generator service", "pollingInterval", s.cfg.PollingInterval)

	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := s.process(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) process(ctx context.Context) error {
	// locking up mutex to prevent concurrent processing
	s.mutex.Lock()
	defer s.mutex.Unlock()

	valSet, config, err := s.tryDetectNewEpochToCommit(ctx)
	if err != nil && !errors.Is(err, entity.ErrValsetAlreadyCommittedForEpoch) {
		return errors.Errorf("failed to detect new epoch to commit: %w", err)
	}
	if errors.Is(err, entity.ErrValsetAlreadyCommittedForEpoch) {
		// no new validator set extra found, nothing to do
		return nil
	}

	if s.generatedEpoch >= valSet.Epoch {
		slog.DebugContext(ctx, "Already signed for this epoch, skipping", "epoch", valSet.Epoch)
		return nil
	}

	networkData, err := s.getNetworkData(ctx, config)
	if err != nil {
		return errors.Errorf("failed to get network data: %w", err)
	}

	latestValset, err := s.cfg.Repo.GetLatestValidatorSet(ctx)
	if err != nil {
		return errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	latestValsetHeader, err := latestValset.GetHeader()
	if err != nil {
		return errors.Errorf("failed to get latest validator set header: %w", err)
	}

	latestValsetHeaderHash, err := latestValsetHeader.Hash()
	if err != nil {
		return errors.Errorf("failed to get latest validator set header hash: %w", err)
	}

	lastCommittedHeaderHash, lastCommittedHeaderEpoch, err := s.cfg.GrowthStrategy.GetLastCommittedHeaderHash(ctx, config)
	if err != nil {
		return errors.Errorf("failed to get last committed header hash: %w", err)
	}

	// waiting for valset listener sync and attaching new valset to growth strategy hash
	if latestValset.Epoch != lastCommittedHeaderEpoch || latestValsetHeaderHash != lastCommittedHeaderHash ||
		valSet.PreviousHeaderHash != lastCommittedHeaderHash {
		slog.WarnContext(ctx, "valset candidate doesn't refer to growth strategy last committed hash", "epoch", valSet.Epoch, "prevEpoch", lastCommittedHeaderEpoch)
		return nil
	}

	if config.MaxMissingEpochs != 0 && latestValset.Epoch-valSet.Epoch > config.MaxMissingEpochs {
		slog.ErrorContext(ctx, "Exceed missing epochs", "latest committed", latestValset.Epoch, "current", valSet.Epoch)
		return errors.New("max missing epochs")
	}

	extraData, err := s.cfg.Aggregator.GenerateExtraData(valSet, config.RequiredKeyTags)
	if err != nil {
		return errors.Errorf("failed to generate extra data: %w", err)
	}

	header, err := valSet.GetHeader()
	if err != nil {
		return errors.Errorf("failed to get validator set header: %w", err)
	}
	data, err := s.headerCommitmentData(networkData, header, extraData)
	if err != nil {
		return errors.Errorf("failed to get header commitment hash: %w", err)
	}

	r := entity.SignatureRequest{
		KeyTag:        entity.ValsetHeaderKeyTag,
		RequiredEpoch: entity.Epoch(latestValset.Epoch),
		Message:       data,
	}

	slog.DebugContext(ctx, "Signed validator set", "header", header, "extra data", extraData, "hash", hex.EncodeToString(data))
	err = s.cfg.Repo.SavePendingValidatorSet(ctx, r.Hash(), valSet)
	if err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.DebugContext(ctx, "Pending valset already exists, skipping save", "requestHash", r.Hash())
			return nil // already exists, nothing to do
		}
		return errors.Errorf("failed to save pending valset: %w", err)
	}

	err = s.cfg.Signer.Sign(ctx, r)
	if err != nil {
		return errors.Errorf("failed to sign new validator set extra: %w", err)
	}

	s.generatedEpoch = header.Epoch

	return nil
}

func (s *Service) getNetworkData(ctx context.Context, config entity.NetworkConfig) (entity.NetworkData, error) {
	for _, replica := range config.Replicas {
		networkData, err := s.cfg.Deriver.GetNetworkData(ctx, replica)
		if err != nil {
			slog.WarnContext(ctx, "Failed to get network data for replica", "replica", replica, "error", err)
			continue
		}
		return networkData, nil
	}

	return entity.NetworkData{}, errors.New("failed to get network data for any replica")
}

func (s *Service) tryDetectNewEpochToCommit(ctx context.Context) (entity.ValidatorSet, entity.NetworkConfig, error) {
	slog.DebugContext(ctx, "Trying to detect new epoch to commit")

	currentOnchainEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get current epoch: %w", err)
	}

	epochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get current epoch start: %w", err)
	}

	config, err := s.cfg.EvmClient.GetConfig(ctx, epochStart)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get network config for current epoch %d: %w", currentOnchainEpoch, err)
	}

	_, isCommitted, err := s.isValsetHeaderCommitted(ctx, config, currentOnchainEpoch)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to check if committed validator set header is committed: %w", err)
	}

	if isCommitted {
		slog.DebugContext(ctx, "Epoch is committed already, skipping", "epoch", currentOnchainEpoch)
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.New(entity.ErrValsetAlreadyCommittedForEpoch)
	}

	newValset, err := s.cfg.Deriver.GetValidatorSet(ctx, currentOnchainEpoch, config)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get validator set extra for epoch %d: %w", currentOnchainEpoch, err)
	}

	return newValset, config, nil
}

func (s *Service) isValsetHeaderCommitted(ctx context.Context, config entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error) {
	for _, addr := range config.Replicas {
		isCommitted, err := s.cfg.EvmClient.IsValsetHeaderCommittedAt(ctx, addr, epoch)
		if err != nil {
			return entity.CrossChainAddress{}, false, errors.Errorf("failed to check if valset header is committed at epoch %d: %w", epoch, err)
		}
		if isCommitted {
			return addr, true, nil
		}
	}
	return entity.CrossChainAddress{}, false, nil
}

func (s *Service) headerCommitmentData(
	networkData entity.NetworkData,
	header entity.ValidatorSetHeader,
	extraData []entity.ExtraData,
) ([]byte, error) {
	headerHash, err := header.Hash()
	if err != nil {
		return nil, errors.Errorf("failed to hash valset header: %w", err)
	}

	extraDataHash, err := entity.ExtraDataList(extraData).Hash()
	if err != nil {
		return nil, errors.Errorf("failed to hash extra data: %w", err)
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
			},
			"ValSetHeaderCommit": []apitypes.Type{
				{Name: "subnetwork", Type: "bytes32"},
				{Name: "epoch", Type: "uint48"},
				{Name: "headerHash", Type: "bytes32"},
				{Name: "extraDataHash", Type: "bytes32"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:    networkData.Eip712Data.Name,
			Version: networkData.Eip712Data.Version,
		},
		PrimaryType: "ValSetHeaderCommit",
		Message: map[string]interface{}{
			"subnetwork":    networkData.Subnetwork,
			"epoch":         new(big.Int).SetUint64(header.Epoch),
			"headerHash":    headerHash,
			"extraDataHash": extraDataHash,
		},
	}

	_, data, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, errors.Errorf("failed to get typed data hash: %w", err)
	}

	return []byte(data), nil
}
