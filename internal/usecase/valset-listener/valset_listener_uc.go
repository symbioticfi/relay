package valset_listener

import (
	"context"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/symbiotic/client/evm"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

type signer interface {
	RequestSignature(ctx context.Context, req entity.SignatureRequest) (common.Hash, error)
}

type repo interface {
	GetLatestValidatorSetHeader(_ context.Context) (entity.ValidatorSetHeader, error)
	GetLatestSignedValidatorSetEpoch(ctx context.Context) (entity.Epoch, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch entity.Epoch) (entity.ValidatorSet, error)
	GetValidatorSetMetadata(ctx context.Context, epoch entity.Epoch) (entity.ValidatorSetMetadata, error)
	GetConfigByEpoch(ctx context.Context, epoch entity.Epoch) (entity.NetworkConfig, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (entity.AggregationProof, error)
	SaveLatestSignedValidatorSetEpoch(_ context.Context, valset entity.ValidatorSet) error
	SaveProof(ctx context.Context, aggregationProof entity.AggregationProof) error
	SaveProofCommitPending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error
	GetPendingProofCommitsSinceEpoch(ctx context.Context, epoch entity.Epoch, limit int) ([]entity.ProofCommitKey, error)
	RemoveProofCommitPending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error
	SaveValidatorSetMetadata(ctx context.Context, data entity.ValidatorSetMetadata) error
	SaveConfig(ctx context.Context, config entity.NetworkConfig, epoch entity.Epoch) error
	SaveValidatorSet(ctx context.Context, valset entity.ValidatorSet) error
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch entity.Epoch, config entity.NetworkConfig) (entity.ValidatorSet, error)
	GetNetworkData(ctx context.Context, addr entity.CrossChainAddress) (entity.NetworkData, error)
}

type metrics interface {
	ObserveAggregationProofSize(proofSize int, validatorCount int)
}

type keyProvider interface {
	GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error)
}

type Config struct {
	EvmClient       evm.IEvmClient `validate:"required"`
	Repo            repo           `validate:"required"`
	Deriver         deriver        `validate:"required"`
	PollingInterval time.Duration  `validate:"required,gt=0"`
	Signer          signer         `validate:"required"`
	KeyProvider     keyProvider
	Aggregator      aggregator.Aggregator
	Metrics         metrics
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

// LoadAllMissingEpochs runs tryLoadMissingEpochs until all missing epochs are loaded successfully
func (s *Service) LoadAllMissingEpochs(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "listener")

	slog.InfoContext(ctx, "Loading all missing epochs before starting services")

	const maxRetries = 10
	retryCount := 0
	retryTimer := time.NewTimer(0)
	defer retryTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retryTimer.C:
			if err := s.tryLoadMissingEpochs(ctx); err != nil {
				retryCount++
				if retryCount >= maxRetries {
					return errors.Errorf("failed to load missing epochs after %d retries: %w", maxRetries, err)
				}
				slog.ErrorContext(ctx, "Failed to load missing epochs, retrying", "error", err, "attempt", retryCount, "maxRetries", maxRetries)
				retryTimer.Reset(time.Second * 2)
				continue
			}
			slog.InfoContext(ctx, "Successfully loaded all missing epochs")
			return nil
		}
	}
}

func (s *Service) Start(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "listener")

	slog.InfoContext(ctx, "Starting valset listener service", "pollingInterval", s.cfg.PollingInterval)

	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := s.tryLoadMissingEpochs(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) tryLoadMissingEpochs(ctx context.Context) error {
	// locking up mutex to prevent concurrent processing
	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.DebugContext(ctx, "Checking for missing epochs")

	currentEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	latestHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get latest validator set header: %w", err)
	}

	nextEpoch := entity.Epoch(0)
	if err == nil {
		nextEpoch = latestHeader.Epoch + 1
	}

	for nextEpoch <= currentEpoch {
		epochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, nextEpoch)
		if err != nil {
			return errors.Errorf("failed to get epoch start for epoch %d: %w", nextEpoch, err)
		}

		nextEpochConfig, err := s.cfg.EvmClient.GetConfig(ctx, epochStart)
		if err != nil {
			return errors.Errorf("failed to get network config for epoch %d: %w", nextEpoch, err)
		}

		nextValset, err := s.cfg.Deriver.GetValidatorSet(ctx, nextEpoch, nextEpochConfig)
		if err != nil {
			return errors.Errorf("failed to derive validator set extra for epoch %d: %w", nextEpoch, err)
		}

		if err := s.cfg.Repo.SaveConfig(ctx, nextEpochConfig, nextEpoch); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		if err := s.cfg.Repo.SaveValidatorSet(ctx, nextValset); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		slog.DebugContext(ctx, "Synced validator set", "epoch", nextEpoch, "config", nextEpochConfig, "valset", nextValset)

		if err := s.process(ctx, nextValset, nextEpochConfig); err != nil {
			return errors.Errorf("failed to process validator set for epoch %d: %w", nextEpoch, err)
		}

		nextEpoch = nextValset.Epoch + 1
	}

	slog.DebugContext(ctx, "All missing epochs loaded", "latestProcessedEpoch", currentEpoch)

	return nil
}

func (s *Service) process(ctx context.Context, valSet entity.ValidatorSet, config entity.NetworkConfig) error {
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
	commitmentData, err := s.headerCommitmentData(networkData, header, extraData)
	if err != nil {
		return errors.Errorf("failed to get header commitment hash: %w", err)
	}

	valsetToCheck := valSet

	// process valset signature with previous epoch if not genesis epoch
	if valSet.Epoch > 0 {
		// get previous epoch valset to check if we are a signer
		prevValSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, valSet.Epoch-1)
		if err != nil {
			return errors.Errorf("failed to get previous validator set: %w", err)
		}
		valsetToCheck = prevValSet
	}

	symbPrivate, err := s.cfg.KeyProvider.GetPrivateKey(valsetToCheck.RequiredKeyTag)
	if err != nil {
		return errors.Errorf("failed to get symb private key: %w", err)
	}

	// if we are a signer, sign the commitment, otherwise just save the metadata
	if valsetToCheck.IsSigner(symbPrivate.PublicKey().OnChain()) {
		r := entity.SignatureRequest{
			KeyTag:        valsetToCheck.RequiredKeyTag,
			RequiredEpoch: valsetToCheck.Epoch,
			Message:       commitmentData,
		}
		_, err := s.cfg.Signer.RequestSignature(ctx, r)
		if err != nil {
			return errors.Errorf("failed to sign new validator set extra: %w", err)
		}
	}

	msgHash, err := crypto.HashMessage(valsetToCheck.RequiredKeyTag.Type(), commitmentData)
	if err != nil {
		return errors.Errorf("failed to hash message: %w", err)
	}

	extendedSig := entity.SignatureExtended{
		MessageHash: msgHash,
		KeyTag:      valsetToCheck.RequiredKeyTag,
		Epoch:       valsetToCheck.Epoch,
	}

	metadata := entity.ValidatorSetMetadata{
		RequestID:      extendedSig.RequestID(),
		ExtraData:      extraData,
		Epoch:          valSet.Epoch,
		CommitmentData: commitmentData,
	}

	if err = s.cfg.Repo.SaveValidatorSetMetadata(ctx, metadata); err != nil {
		return errors.Errorf("failed to save validator set metadata: %w", err)
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
			"epoch":         new(big.Int).SetUint64(uint64(header.Epoch)),
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
