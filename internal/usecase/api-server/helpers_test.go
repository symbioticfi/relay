package api_server

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/core/entity"
	deriverMocks "github.com/symbioticfi/relay/core/usecase/valset-deriver/mocks"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
)

// testSetup contains all the mocks and test helper functions
// This unified setup can be used for all API server endpoint tests
type testSetup struct {
	ctrl          *gomock.Controller
	mockEvmClient *deriverMocks.MockEvmClient
	mockRepo      *mocks.Mockrepo
	mockDeriver   *mocks.Mockderiver
	handler       *grpcHandler
}

// newTestSetup creates a new test setup with mocked dependencies
// This reuses the same construction pattern as production code
func newTestSetup(t *testing.T) *testSetup {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockEvmClient := deriverMocks.NewMockEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)

	// Create Config using the same structure as production
	cfg := Config{
		Address:           ":8080", // not used in unit tests
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,

		// Inject mocked dependencies
		EvmClient:    mockEvmClient,
		Repo:         mockRepo,
		Deriver:      mockDeriver,
		ServeMetrics: false,
	}

	// Create grpcHandler using the same pattern as production
	handler := &grpcHandler{
		cfg: cfg,
	}

	return &testSetup{
		ctrl:          ctrl,
		mockEvmClient: mockEvmClient,
		mockRepo:      mockRepo,
		mockDeriver:   mockDeriver,
		handler:       handler,
	}
}

// createTestValidatorSet creates a sample validator set for testing
// This is the simpler version used for GetValidatorSetHeader tests
func createTestValidatorSet(epoch uint64) entity.ValidatorSet {
	return entity.ValidatorSet{
		Version:            1,
		RequiredKeyTag:     entity.KeyTag(15),
		Epoch:              epoch,
		CaptureTimestamp:   1640995200, // 2022-01-01 00:00:00 UTC
		QuorumThreshold:    entity.ToVotingPower(big.NewInt(670)),
		PreviousHeaderHash: common.HexToHash("0xdef456"),
		Validators: []entity.Validator{
			{
				Operator:    common.HexToAddress("0x123"),
				VotingPower: entity.ToVotingPower(big.NewInt(1000)),
				IsActive:    true,
				Keys: []entity.ValidatorKey{
					{
						Tag:     entity.KeyTag(15),
						Payload: entity.CompactPublicKey("test-key"),
					},
				},
			},
		},
	}
}

// createTestValidatorSetWithMultipleValidators creates a sample validator set with multiple validators for testing
// This is the richer version used for GetValidatorByAddress tests
func createTestValidatorSetWithMultipleValidators(epoch uint64) entity.ValidatorSet {
	return entity.ValidatorSet{
		Version:            1,
		RequiredKeyTag:     entity.KeyTag(15),
		Epoch:              epoch,
		CaptureTimestamp:   1640995200, // 2022-01-01 00:00:00 UTC
		QuorumThreshold:    entity.ToVotingPower(big.NewInt(670)),
		PreviousHeaderHash: common.HexToHash("0xdef456"),
		Validators: []entity.Validator{
			{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000123"),
				VotingPower: entity.ToVotingPower(big.NewInt(1000)),
				IsActive:    true,
				Keys: []entity.ValidatorKey{
					{
						Tag:     entity.KeyTag(15),
						Payload: entity.CompactPublicKey("test-key-1"),
					},
				},
				Vaults: []entity.ValidatorVault{
					{
						ChainID:     1,
						Vault:       common.HexToAddress("0x456"),
						VotingPower: entity.ToVotingPower(big.NewInt(1000)),
					},
				},
			},
			{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000abc"),
				VotingPower: entity.ToVotingPower(big.NewInt(2000)),
				IsActive:    true,
				Keys: []entity.ValidatorKey{
					{
						Tag:     entity.KeyTag(15),
						Payload: entity.CompactPublicKey("test-key-2"),
					},
				},
				Vaults: []entity.ValidatorVault{
					{
						ChainID:     1,
						Vault:       common.HexToAddress("0xdef"),
						VotingPower: entity.ToVotingPower(big.NewInt(2000)),
					},
				},
			},
			{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000789"),
				VotingPower: entity.ToVotingPower(big.NewInt(500)),
				IsActive:    false,
				Keys: []entity.ValidatorKey{
					{
						Tag:     entity.KeyTag(15),
						Payload: entity.CompactPublicKey("test-key-3"),
					},
				},
			},
		},
	}
}
