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

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type signer interface {
	RequestSignature(ctx context.Context, req symbiotic.SignatureRequest) (common.Hash, error)
}

type repo interface {
	GetLatestValidatorSetHeader(_ context.Context) (symbiotic.ValidatorSetHeader, error)
	GetOldestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetLatestSignedValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error)
	GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error)
	SaveLatestSignedValidatorSetEpoch(_ context.Context, valset symbiotic.ValidatorSet) error
	SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error
	SaveProofCommitPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error
	GetPendingProofCommitsSinceEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]symbiotic.ProofCommitKey, error)
	RemoveProofCommitPending(ctx context.Context, epoch symbiotic.Epoch) error
	SaveValidatorSetMetadata(ctx context.Context, data symbiotic.ValidatorSetMetadata) error
	SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error
	SaveValidatorSet(ctx context.Context, valset symbiotic.ValidatorSet) error
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch symbiotic.Epoch, config symbiotic.NetworkConfig) (symbiotic.ValidatorSet, error)
	GetNetworkData(ctx context.Context, addr symbiotic.CrossChainAddress) (symbiotic.NetworkData, error)
}

type metrics interface {
	ObserveAggregationProofSize(proofSize int, validatorCount int)
}

type keyProvider interface {
	GetPrivateKey(keyTag symbiotic.KeyTag) (crypto.PrivateKey, error)
	GetOnchainKeyFromCache(keyTag symbiotic.KeyTag) (symbiotic.CompactPublicKey, error)
}

type Config struct {
	EvmClient           evm.IEvmClient                          `validate:"required"`
	Repo                repo                                    `validate:"required"`
	Deriver             deriver                                 `validate:"required"`
	PollingInterval     time.Duration                           `validate:"required,gt=0"`
	Signer              signer                                  `validate:"required"`
	ValidatorSet        *signals.Signal[symbiotic.ValidatorSet] `validate:"required"`
	KeyProvider         keyProvider
	Aggregator          aggregator.Aggregator
	Metrics             metrics `validate:"required"`
	ForceCommitter      bool
	EpochRetentionCount uint64
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
			fromEpoch, toEpoch, err := s.determineStartupSyncRange(ctx)
			if err != nil {
				retryCount++
				if retryCount >= maxRetries {
					return errors.Errorf("failed to load missing epochs after %d retries: %w", maxRetries, err)
				}
				slog.ErrorContext(ctx, "Failed to load missing epochs, retrying", "error", err, "attempt", retryCount, "maxRetries", maxRetries)
				retryTimer.Reset(time.Second * 2)
				continue
			}

			if err := s.tryLoadMissingEpochs(ctx, fromEpoch, toEpoch); err != nil {
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
			fromEpoch, toEpoch, err := s.determineSteadySyncRange(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to determine sync range", "error", err)
				timer.Reset(s.cfg.PollingInterval)
				continue
			}

			if err := s.tryLoadMissingEpochs(ctx, fromEpoch, toEpoch); err != nil {
				slog.ErrorContext(ctx, "Failed to process epochs", "error", err)
			}

			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) determineStartupSyncRange(ctx context.Context) (from symbiotic.Epoch, to symbiotic.Epoch, err error) {
	latestHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return 0, 0, errors.Errorf("failed to get latest validator set header: %w", err)
	}
	freshNode := errors.Is(err, entity.ErrEntityNotFound)

	currentEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return 0, 0, errors.Errorf("failed to get current epoch: %w", err)
	}

	if !freshNode {
		return s.determineSyncRangeFromLatestWithHeader(ctx, latestHeader, currentEpoch)
	}

	nextEpoch := symbiotic.Epoch(0)
	if s.cfg.EpochRetentionCount == 0 {
		slog.InfoContext(ctx, "Fresh node syncing from genesis (unlimited retention)")
		return nextEpoch, currentEpoch, nil
	}

	if currentEpoch >= symbiotic.Epoch(s.cfg.EpochRetentionCount) {
		nextEpoch = currentEpoch - symbiotic.Epoch(s.cfg.EpochRetentionCount) + 1
		slog.InfoContext(ctx, "Fresh node with epoch retention enabled",
			"epochRetentionCount", s.cfg.EpochRetentionCount,
			"currentEpoch", currentEpoch,
			"startEpoch", nextEpoch,
			"skippedEpochs", nextEpoch,
		)
	} else {
		slog.InfoContext(ctx, "Fresh node syncing from genesis (network younger than retention)",
			"epochRetentionCount", s.cfg.EpochRetentionCount,
			"currentEpoch", currentEpoch,
		)
	}

	return nextEpoch, currentEpoch, nil
}

func (s *Service) determineSteadySyncRange(ctx context.Context) (from symbiotic.Epoch, to symbiotic.Epoch, err error) {
	latestHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return 0, 0, errors.New("latest validator set header not found during steady-state sync")
		}
		return 0, 0, errors.Errorf("failed to get latest validator set header: %w", err)
	}

	currentEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return 0, 0, errors.Errorf("failed to get current epoch: %w", err)
	}

	return s.determineSyncRangeFromLatestWithHeader(ctx, latestHeader, currentEpoch)
}

func (s *Service) determineSyncRangeFromLatestWithHeader(
	ctx context.Context,
	latestHeader symbiotic.ValidatorSetHeader,
	currentEpoch symbiotic.Epoch,
) (from symbiotic.Epoch, to symbiotic.Epoch, err error) {
	headerEpochConfig, err := s.cfg.Repo.GetConfigByEpoch(ctx, latestHeader.Epoch)
	if err != nil {
		return 0, 0, errors.Errorf("failed to get network config for epoch %d: %w", latestHeader.Epoch, err)
	}

	if time.Unix(int64(latestHeader.CaptureTimestamp), 0).Add(time.Duration(headerEpochConfig.EpochDuration) * time.Second).After(time.Now()) {
		slog.DebugContext(ctx, "Last epoch is still ongoing, no new valset to process", "lastEpoch", latestHeader.Epoch)
		return latestHeader.Epoch + 1, currentEpoch, nil
	}

	nextEpoch := latestHeader.Epoch + 1

	if s.cfg.EpochRetentionCount > 0 {
		expectedStartEpoch := symbiotic.Epoch(0)
		if currentEpoch >= symbiotic.Epoch(s.cfg.EpochRetentionCount) {
			expectedStartEpoch = currentEpoch - symbiotic.Epoch(s.cfg.EpochRetentionCount) + 1
		}

		oldestEpoch, err := s.cfg.Repo.GetOldestValidatorSetEpoch(ctx)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return 0, 0, errors.Errorf("failed to get oldest validator set epoch: %w", err)
		}

		if err == nil && oldestEpoch > expectedStartEpoch {
			slog.WarnContext(ctx,
				"Detected epoch gap due to retention configuration change. Gap will persist until data is pruned or retention is decreased. Increasing retention on existing nodes does not backfill historical data.",
				"epochRetentionCount", s.cfg.EpochRetentionCount,
				"currentEpoch", currentEpoch,
				"expectedStartEpoch", expectedStartEpoch,
				"oldestStoredEpoch", oldestEpoch,
			)
		}
	}

	return nextEpoch, currentEpoch, nil
}

func (s *Service) tryLoadMissingEpochs(ctx context.Context, nextEpoch, currentEpoch symbiotic.Epoch) error {
	// locking up mutex to prevent concurrent processing
	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.DebugContext(ctx, "Checking for missing epochs")

	for nextEpoch <= currentEpoch {
		epochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, nextEpoch)
		if err != nil {
			return errors.Errorf("failed to get epoch start for epoch %d: %w", nextEpoch, err)
		}

		nextEpochConfig, err := s.cfg.EvmClient.GetConfig(ctx, epochStart, nextEpoch)
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

		if err = s.cfg.ValidatorSet.Emit(nextValset); err != nil {
			slog.ErrorContext(ctx, "Failed to emit validator set", "error", err)
		}

		nextEpoch = nextValset.Epoch + 1
	}

	slog.DebugContext(ctx, "All missing epochs loaded", "latestProcessedEpoch", currentEpoch)

	return nil
}

func (s *Service) process(ctx context.Context, valSet symbiotic.ValidatorSet, config symbiotic.NetworkConfig) error {
	ctx = log.WithAttrs(ctx, slog.Uint64("epoch", uint64(valSet.Epoch)))
	slog.DebugContext(ctx, "Started processing valset for epoch")

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

	onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(valsetToCheck.RequiredKeyTag)
	if err != nil {
		return errors.Errorf("failed to get onchain symb key from cache: %w", err)
	}

	// if we are a signer, sign the commitment, otherwise just save the metadata
	if valsetToCheck.IsSigner(onchainKey) {
		r := symbiotic.SignatureRequest{
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

	extendedSig := symbiotic.Signature{
		MessageHash: msgHash,
		KeyTag:      valsetToCheck.RequiredKeyTag,
		Epoch:       valsetToCheck.Epoch,
	}

	metadata := symbiotic.ValidatorSetMetadata{
		RequestID:      extendedSig.RequestID(),
		ExtraData:      extraData,
		Epoch:          valSet.Epoch,
		CommitmentData: commitmentData,
	}

	if err = s.cfg.Repo.SaveValidatorSetMetadata(ctx, metadata); err != nil {
		return errors.Errorf("failed to save validator set metadata: %w", err)
	}

	// save pending proof commit here
	// we store pending commit request for all nodes and not just current commiters because
	// if committers of this epoch fail then commiters for next epoch should still try to commit old proofs
	if err := s.cfg.Repo.SaveProofCommitPending(ctx, valSet.Epoch, extendedSig.RequestID()); err != nil {
		if !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to mark proof commit as pending: %w", err)
		}
		slog.DebugContext(ctx, "Skipped proof commit, already pending", "epoch", valSet.Epoch)
		return nil
	}
	slog.DebugContext(ctx, "Marked proof commit as pending", "epoch", valSet.Epoch, "requestId", extendedSig.RequestID().Hex())
	return nil
}

func (s *Service) getNetworkData(ctx context.Context, config symbiotic.NetworkConfig) (symbiotic.NetworkData, error) {
	for _, settlement := range config.Settlements {
		networkData, err := s.cfg.Deriver.GetNetworkData(ctx, settlement)
		if err != nil {
			slog.WarnContext(ctx, "Failed to get network data for settlement", "settlement", settlement, "error", err)
			continue
		}
		return networkData, nil
	}

	return symbiotic.NetworkData{}, errors.New("failed to get network data for any settlement")
}

func (s *Service) headerCommitmentData(
	networkData symbiotic.NetworkData,
	header symbiotic.ValidatorSetHeader,
	extraData []symbiotic.ExtraData,
) ([]byte, error) {
	headerHash, err := header.Hash()
	if err != nil {
		return nil, errors.Errorf("failed to hash valset header: %w", err)
	}

	extraDataHash, err := symbiotic.ExtraDataList(extraData).Hash()
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
