package valset_generator

import (
	"context"
	"encoding/hex"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/aggregator"
	"github.com/symbioticfi/relay/pkg/log"
)

const slashableEpochs = 5 // TODO: temp until contracts support

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
	GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error)
}

type repo interface {
	GetLatestValidatorSetHeader(_ context.Context) (entity.ValidatorSetHeader, error)
	GetLatestSignedValidatorSetEpoch(_ context.Context) (uint64, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetConfigByEpoch(ctx context.Context, epoch uint64) (entity.NetworkConfig, error)
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	SaveLatestSignedValidatorSetEpoch(_ context.Context, valset entity.ValidatorSet) error
	SaveAggregationProof(ctx context.Context, reqHash common.Hash, ap entity.AggregationProof) error
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
	KeyProvider     keyprovider.KeyProvider
	Aggregator      aggregator.Aggregator
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	return nil
}

type Service struct {
	cfg   Config
	mutex sync.Mutex
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

	valSet, config, err := s.tryDetectUnsignedValset(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to detect new epoch to commit: %w", err)
	}
	if errors.Is(err, entity.ErrEntityNotFound) {
		// no new validator set extra found, nothing to do
		return nil
	}

	networkData, err := s.getNetworkData(ctx, config)
	if err != nil {
		return errors.Errorf("failed to get network data: %w", err)
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
		RequiredEpoch: entity.Epoch(valSet.Epoch),
		Message:       data,
	}

	slog.DebugContext(ctx, "Signed validator set", "header", header, "extra data", extraData, "hash", hex.EncodeToString(data))
	if err = s.cfg.Repo.SaveLatestSignedValidatorSetEpoch(ctx, valSet); err != nil {
		return errors.Errorf("failed to save latest signed valset epoch: %w", err)
	}

	if err = s.cfg.Signer.Sign(ctx, r); err != nil {
		return errors.Errorf("failed to sign new validator set extra: %w", err)
	}

	return nil
}

func (s *Service) getNetworkData(ctx context.Context, config entity.NetworkConfig) (entity.NetworkData, error) {
	for _, settlement := range config.Settlements {
		networkData, err := s.cfg.Deriver.GetNetworkData(ctx, settlement)
		if err != nil {
			slog.WarnContext(ctx, "Failed to get network data for settlement", "settlement", settlement, "error", err)
			continue
		}
		return networkData, nil
	}

	return entity.NetworkData{}, errors.New("failed to get network data for any settlement")
}

func (s *Service) tryDetectUnsignedValset(ctx context.Context) (entity.ValidatorSet, entity.NetworkConfig, error) {
	slog.DebugContext(ctx, "Trying to detect new epoch to commit")

	epoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get current epoch: %w", err)
	}

	if epoch >= slashableEpochs {
		epoch = epoch - slashableEpochs + 1
	} else {
		epoch = 1
	}

	latestSignedEpoch, err := s.cfg.Repo.GetLatestSignedValidatorSetEpoch(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get latest pending validator set: %w", err)
	}

	if err == nil && latestSignedEpoch >= epoch {
		epoch = latestSignedEpoch + 1
	}

	var valset entity.ValidatorSet

	for {
		valset, err = s.cfg.Repo.GetValidatorSetByEpoch(ctx, epoch)
		if err != nil {
			return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get validator set for epoch %d: %w", epoch, err)
		}

		if valset.Status == entity.HeaderDerived {
			break
		}

		epoch++
	}

	if valset.Status != entity.HeaderDerived {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, entity.ErrEntityNotFound
	}

	config, err := s.cfg.Repo.GetConfigByEpoch(ctx, valset.Epoch)
	if err != nil {
		return entity.ValidatorSet{}, entity.NetworkConfig{}, errors.Errorf("failed to get config for epoch %d: %w", valset.Epoch, err)
	}

	return valset, config, nil
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
