package valset_generator

import (
	"context"
	"encoding/hex"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
)

type signer interface {
	Sign(ctx context.Context, req entity.SignatureRequest) error
}

type eth interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetCurrentPhase(ctx context.Context) (entity.Phase, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)

	IsValsetHeaderCommittedAt(ctx context.Context, epoch uint64) (bool, error)
	CommitValsetHeader(ctx context.Context, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) (entity.CommitValsetHeaderResult, error)
	VerifyQuorumSig(ctx context.Context, epoch uint64, message []byte, keyTag entity.KeyTag, threshold *big.Int, proof []byte) (bool, error)
}

type repo interface {
	GetLatestSignedValset(_ context.Context) (entity.ValidatorSet, error)
	GetLatestValset(ctx context.Context) (entity.ValidatorSet, error)

	GetValsetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetConfigByEpoch(ctx context.Context, epoch uint64) (entity.NetworkConfig, error)
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	SavePendingValset(ctx context.Context, reqHash common.Hash, valset entity.ValidatorSet) error
	GetPendingValset(ctx context.Context, reqHash common.Hash) (entity.ValidatorSet, error)
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
	GetNetworkData(ctx context.Context) (entity.NetworkData, error)
	GenerateExtraData(valset entity.ValidatorSet, config entity.NetworkConfig) ([]entity.ExtraData, error)
}

type Config struct {
	Signer          signer        `validate:"required"`
	Eth             eth           `validate:"required"`
	Repo            repo          `validate:"required"`
	Deriver         deriver       `validate:"required"`
	PollingInterval time.Duration `validate:"required,gt=0"`
	IsCommitter     bool
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
	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := s.process(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) process(ctx context.Context) error {
	valSet, config, err := s.tryDetectNewEpochToCommit(ctx)
	if err != nil {
		return errors.Errorf("failed to detect new epoch to commit: %w", err)
	}
	if valSet == nil || config == nil {
		// no new validator set extra found, nothing to do
		return nil
	}

	if s.generatedEpoch >= valSet.Epoch {
		slog.DebugContext(ctx, "no new epoch to commit, already generated for this epoch", "epoch", valSet.Epoch)
		return nil
	}

	networkData, err := s.cfg.Deriver.GetNetworkData(ctx)
	if err != nil {
		return errors.Errorf("failed to get network data: %w", err)
	}

	extraData, err := s.cfg.Deriver.GenerateExtraData(*valSet, *config)
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

	slog.DebugContext(ctx, "generated commitment data", "hash", hex.EncodeToString(data))

	latestValset, err := s.cfg.Repo.GetLatestValset(ctx)
	if err != nil {
		return errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	if latestValset.Epoch < valSet.Epoch-10 {
		slog.WarnContext(ctx, "Header is not committed for much epochs", "latest committed", latestValset.Epoch, "current", valSet.Epoch)
	}
	r := entity.SignatureRequest{
		KeyTag:        entity.ValsetHeaderKeyTag,
		RequiredEpoch: latestValset.Epoch,
		Message:       data,
	}

	slog.DebugContext(ctx, "Signed header", "header", header)
	slog.DebugContext(ctx, "Signed extra data", "extraData", extraData)
	err = s.cfg.Repo.SavePendingValset(ctx, r.Hash(), *valSet)
	if err != nil {
		return errors.Errorf("failed to save pending valset: %w", err)
	}

	err = s.cfg.Signer.Sign(ctx, r)
	if err != nil {
		return errors.Errorf("failed to sign new validator set extra: %w", err)
	}

	s.generatedEpoch = header.Epoch

	return nil
}

func (s *Service) tryDetectNewEpochToCommit(ctx context.Context) (*entity.ValidatorSet, *entity.NetworkConfig, error) {
	currentOnchainEpoch, err := s.cfg.Eth.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get current epoch: %w", err)
	}

	isCommitted, err := s.cfg.Eth.IsValsetHeaderCommittedAt(ctx, currentOnchainEpoch)
	if err != nil {
		return nil, nil, errors.Errorf("failed to check if committed validator set header is committed: %w", err)
	}

	if isCommitted {
		slog.DebugContext(ctx, "Epoch is committed already, skipping", "epoch", currentOnchainEpoch)
		return nil, nil, nil
	}

	epochStart, err := s.cfg.Eth.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get epoch start for epoch %d: %w", currentOnchainEpoch, err)
	}

	config, err := s.cfg.Eth.GetConfig(ctx, epochStart)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get network config for epoch %d: %w", currentOnchainEpoch, err)
	}

	newValset, err := s.cfg.Deriver.GetValidatorSet(ctx, currentOnchainEpoch, config)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get validator set extra for epoch %d: %w", currentOnchainEpoch, err)
	}

	return &newValset, &config, nil
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
