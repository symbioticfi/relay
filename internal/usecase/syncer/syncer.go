package syncer

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
)

// SignatureProcessingStats contains detailed statistics for processing received signatures
type SignatureProcessingStats struct {
	ProcessedCount             int // Successfully processed signatures
	UnrequestedSignatureCount  int // Signatures for validators we didn't request
	UnrequestedHashCount       int // Signatures for hashes we didn't request
	SignatureRequestErrorCount int // Failed to get signature request
	PublicKeyErrorCount        int // Failed to create public key from signature
	ValidatorInfoErrorCount    int // Failed to get validator info
	ProcessingErrorCount       int // Failed to process signature
	AlreadyExistCount          int // Signature already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s SignatureProcessingStats) TotalErrors() int {
	return s.UnrequestedSignatureCount + s.UnrequestedHashCount + s.SignatureRequestErrorCount +
		s.PublicKeyErrorCount + s.ValidatorInfoErrorCount + s.ProcessingErrorCount + s.AlreadyExistCount
}

type repo interface {
	GetSignatureRequestsByEpochPending(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error)
	GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (uint64, error)
	GetActiveValidatorCountByEpoch(ctx context.Context, epoch uint64) (uint32, error)
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetAllSignatures(ctx context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
}

type p2pService interface {
	SendWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
}

type signatureProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
}

type Config struct {
	Repo                        repo               `validate:"required"`
	P2PService                  p2pService         `validate:"required"`
	SignatureProcessor          signatureProcessor `validate:"required"`
	EpochsToSync                int                `validate:"gte=0"`
	SyncPeriod                  time.Duration      `validate:"gt=0"`
	SyncTimeout                 time.Duration      `validate:"gt=0"`
	MaxSignatureRequestsPerSync int                `validate:"gt=0"`
	MaxResponseSignatureCount   int                `validate:"gt=0"`
}

type Syncer struct {
	cfg Config
}

func New(cfg Config) (*Syncer, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}
	return &Syncer{
		cfg: cfg,
	}, nil
}

func (s *Syncer) Start(ctx context.Context) error {
	timer := time.NewTimer(s.cfg.SyncPeriod)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if err := s.askSignatures(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to ask signatures", "error", err)
			}
			timer.Reset(s.cfg.SyncPeriod)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Syncer) askSignatures(ctx context.Context) error {
	// Create context with timeout for the entire sync operation
	syncCtx, cancel := context.WithTimeout(ctx, s.cfg.SyncTimeout)
	defer cancel()

	slog.InfoContext(syncCtx, "Starting signature sync")

	// Collect all pending signature requests across epochs
	wantSignatures, err := s.buildWantSignaturesMap(syncCtx)
	if err != nil {
		return errors.Errorf("failed to build want signatures map: %w", err)
	}

	// If no signatures needed, log and return
	if len(wantSignatures) == 0 {
		slog.InfoContext(syncCtx, "No pending signature requests found")
		return nil
	}

	slog.InfoContext(syncCtx, "Found pending signature requests", "count", len(wantSignatures))

	// Send request to peer
	request := entity.WantSignaturesRequest{
		WantSignatures: wantSignatures,
	}

	response, err := s.cfg.P2PService.SendWantSignaturesRequest(syncCtx, request)
	if err != nil {
		return errors.Errorf("failed to send want signatures request: %w", err)
	}

	slog.InfoContext(syncCtx, "Received signature response", "signatures_count", len(response.Signatures))

	stats := s.processReceivedSignatures(syncCtx, response, wantSignatures)

	slog.InfoContext(syncCtx, "Signature sync completed",
		"processed", stats.ProcessedCount,
		"total_errors", stats.TotalErrors(),
		"unrequested_signatures", stats.UnrequestedSignatureCount,
		"unrequested_hashes", stats.UnrequestedHashCount,
		"signature_request_errors", stats.SignatureRequestErrorCount,
		"public_key_errors", stats.PublicKeyErrorCount,
		"validator_info_errors", stats.ValidatorInfoErrorCount,
		"processing_errors", stats.ProcessingErrorCount,
		"already_exist", stats.AlreadyExistCount,
	)

	return nil
}

func (s *Syncer) buildWantSignaturesMap(ctx context.Context) (map[common.Hash]entity.SignatureBitmap, error) {
	// Get the latest epoch
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest epoch: %w", err)
	}

	// Calculate the starting epoch (go back EpochsToSync epochs)
	var startEpoch uint64
	if latestEpoch >= uint64(s.cfg.EpochsToSync) {
		startEpoch = latestEpoch - uint64(s.cfg.EpochsToSync)
	} else {
		startEpoch = 0
	}

	wantSignatures := make(map[common.Hash]entity.SignatureBitmap)
	totalRequests := 0

	for epoch := latestEpoch; epoch >= startEpoch && totalRequests < s.cfg.MaxSignatureRequestsPerSync; epoch-- {
		var lastHash common.Hash
		remaining := s.cfg.MaxSignatureRequestsPerSync - totalRequests

		for remaining > 0 {
			requests, err := s.cfg.Repo.GetSignatureRequestsByEpochPending(ctx, entity.Epoch(epoch), remaining, lastHash)
			if err != nil {
				return nil, errors.Errorf("failed to get pending signature requests for epoch %d: %w", epoch, err)
			}

			if len(requests) == 0 {
				break
			}

			// Process each request to find missing signatures
			for _, req := range requests {
				reqHash := req.Hash()

				// Get current signature map
				sigMap, err := s.cfg.Repo.GetSignatureMap(ctx, reqHash)
				if err != nil {
					return nil, errors.Errorf("failed to get signature map for request %s: %w", reqHash.Hex(), err)
				}

				// Get missing validators from signature map
				missingValidators := sigMap.GetMissingValidators()
				if !missingValidators.IsEmpty() {
					wantSignatures[reqHash] = missingValidators
				}

				lastHash = reqHash
			}

			totalRequests += len(requests)
			remaining = s.cfg.MaxSignatureRequestsPerSync - totalRequests

			// If we got fewer requests than requested, we've reached the end for this epoch
			if len(requests) < remaining {
				break
			}
		}

		if epoch == 0 {
			break
		}
	}

	return wantSignatures, nil
}

func (s *Syncer) processReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.SignatureBitmap) SignatureProcessingStats {
	var stats SignatureProcessingStats

	for reqHash, signatures := range response.Signatures {
		for _, validatorSig := range signatures {
			// Validate that we actually requested this validator's signature
			requestedBitmap, exists := wantSignatures[reqHash]
			if !exists {
				slog.WarnContext(ctx, "Received signature for unrequested hash", "request_hash", reqHash.Hex())
				stats.UnrequestedHashCount++
				continue
			}

			if !requestedBitmap.Contains(validatorSig.ValidatorIndex) {
				slog.WarnContext(ctx, "Received unrequested signature",
					"request_hash", reqHash.Hex(),
					"validator_index", validatorSig.ValidatorIndex)
				stats.UnrequestedSignatureCount++
				continue
			}

			// Get the original signature request to extract epoch and other details
			sigReq, err := s.cfg.Repo.GetSignatureRequest(ctx, reqHash)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get signature request for processing",
					"request_hash", reqHash.Hex(), "error", err)
				stats.SignatureRequestErrorCount++
				continue
			}

			publicKey, err := crypto.NewPublicKey(sigReq.KeyTag.Type(), validatorSig.Signature.PublicKey)
			if err != nil {
				slog.WarnContext(ctx, "Failed to create public key from signature",
					"request_hash", reqHash.Hex(),
					"validator_index", validatorSig.ValidatorIndex,
					"error", err)
				stats.PublicKeyErrorCount++
				continue
			}

			// Get validator info to extract voting power
			validatorInfo, _, err := s.cfg.Repo.GetValidatorByKey(
				ctx,
				uint64(sigReq.RequiredEpoch),
				sigReq.KeyTag,
				publicKey.OnChain(),
			)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get validator info",
					"request_hash", reqHash.Hex(),
					"validator_index", validatorSig.ValidatorIndex,
					"error", err)
				stats.ValidatorInfoErrorCount++
				continue
			}

			// Process the signature
			param := entity.SaveSignatureParam{
				RequestHash:      reqHash,
				Key:              validatorSig.Signature.PublicKey,
				Signature:        validatorSig.Signature,
				ActiveIndex:      validatorSig.ValidatorIndex,
				VotingPower:      validatorInfo.VotingPower,
				Epoch:            sigReq.RequiredEpoch,
				SignatureRequest: nil,
			}

			if err := s.cfg.SignatureProcessor.ProcessSignature(ctx, param); err != nil {
				if errors.Is(err, entity.ErrEntityAlreadyExist) {
					slog.DebugContext(ctx, "Signature already exists",
						"request_hash", reqHash.Hex(),
						"validator_index", validatorSig.ValidatorIndex)
					stats.AlreadyExistCount++
				} else {
					slog.WarnContext(ctx, "Failed to process received signature",
						"request_hash", reqHash.Hex(),
						"validator_index", validatorSig.ValidatorIndex,
						"error", err)
					stats.ProcessingErrorCount++
				}
				continue
			}

			stats.ProcessedCount++
		}
	}

	return stats
}

func (s *Syncer) HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	slog.InfoContext(ctx, "Handling want signatures request", "request_count", len(request.WantSignatures))

	response := entity.WantSignaturesResponse{
		Signatures: make(map[common.Hash][]entity.ValidatorSignature),
	}

	totalSignatureCount := 0

	for reqHash, requestedIndices := range request.WantSignatures {
		// Check signature count limit before processing each request
		if totalSignatureCount >= s.cfg.MaxResponseSignatureCount {
			return entity.WantSignaturesResponse{}, errors.Errorf("response signature limit exceeded")
		}

		// Get stored signatures for this request
		signatures, err := s.cfg.Repo.GetAllSignatures(ctx, reqHash)
		if err != nil {
			return entity.WantSignaturesResponse{}, errors.Errorf("failed to get signatures for request %s: %w", reqHash.Hex(), err)
		}

		// Get signature request for epoch info
		sigReq, err := s.cfg.Repo.GetSignatureRequest(ctx, reqHash)
		if err != nil {
			return entity.WantSignaturesResponse{}, errors.Errorf("failed to get signature request %s: %w", reqHash.Hex(), err)
		}

		var validatorSigs []entity.ValidatorSignature

		for _, sig := range signatures {
			// Check limit before processing each signature
			if totalSignatureCount >= s.cfg.MaxResponseSignatureCount {
				return entity.WantSignaturesResponse{}, errors.Errorf("response signature limit exceeded")
			}

			publicKey, err := crypto.NewPublicKey(sigReq.KeyTag.Type(), sig.PublicKey)
			if err != nil {
				return entity.WantSignaturesResponse{}, errors.Errorf("failed to get public key: %w", err)
			}

			// Map public key to validator index
			_, activeIndex, err := s.cfg.Repo.GetValidatorByKey(
				ctx,
				uint64(sigReq.RequiredEpoch),
				sigReq.KeyTag,
				publicKey.OnChain(),
			)
			if err != nil {
				return entity.WantSignaturesResponse{}, errors.Errorf("failed to get validator for key: %w", err)
			}

			// Only include if requested
			if requestedIndices.Contains(activeIndex) {
				validatorSigs = append(validatorSigs, entity.ValidatorSignature{
					ValidatorIndex: activeIndex,
					Signature:      sig,
				})
				totalSignatureCount++
			}
		}

		if len(validatorSigs) > 0 {
			response.Signatures[reqHash] = validatorSigs
		}
	}

	slog.InfoContext(ctx, "Want signatures request handled", "response_signatures", totalSignatureCount, "response_requests", len(response.Signatures))

	return response, nil
}
