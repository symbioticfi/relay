package aggregator_app

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/proof"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type ethClient interface {
	GetQuorumThreshold(ctx context.Context, timestamp *big.Int, keyTag uint8) (*big.Int, error)
}

type valsetDeriver interface {
	GetValidatorSet(ctx context.Context, timestamp *big.Int) (entity.ValidatorSet, error)
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.SignaturesAggregatedMessage) error
	SetSignatureHashMessageHandler(mh func(ctx context.Context, msg entity.P2PSignatureHashMessage) error)
}

type Config struct {
	EthClient     ethClient     `validate:"required"`
	ValsetDeriver valsetDeriver `validate:"required"`
	P2PClient     p2pClient     `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type AggregatorApp struct {
	cfg          Config
	hashStore    *hashStore
	validatorSet entity.ValidatorSet
	inputsBytes  []byte
}

func NewAggregatorApp(ctx context.Context, cfg Config) (*AggregatorApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	nowBig := big.NewInt(time.Now().Unix())

	validatorSet, err := cfg.ValsetDeriver.GetValidatorSet(ctx, nowBig)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator set: %w", err)
	}

	zkInputs, err := initInputs()
	if err != nil {
		return nil, fmt.Errorf("failed to init inputs: %w", err)
	}

	app := &AggregatorApp{
		cfg:          cfg,
		validatorSet: validatorSet,
		hashStore:    newHashStore(),
		inputsBytes:  zkInputs,
	}

	cfg.P2PClient.SetSignatureHashMessageHandler(app.HandleSignatureGeneratedMessage)

	return app, nil
}

func (s *AggregatorApp) HandleSignatureGeneratedMessage(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
	slog.DebugContext(ctx, "received signature hash generated message", "message", msg)

	validator, found := s.validatorSet.FindValidatorByKey(msg.Message.PublicKeyG1)
	if !found {
		return errors.Errorf("validator not found for public key: %x", msg.Message.PublicKeyG1)
	}

	slog.DebugContext(ctx, "found validator", "validator", validator)

	current, err := s.hashStore.PutHash(msg.Message, validator)
	if err != nil {
		return errors.Errorf("failed to put hash: %w", err)
	}

	slog.DebugContext(ctx, "total voting power", "currentVotingPower", current.votingPower.String())

	quorumThreshold, err := s.cfg.EthClient.GetQuorumThreshold(ctx, big.NewInt(time.Now().Unix()), msg.Message.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get quorum threshold: %w", err)
	}

	slog.DebugContext(ctx, "got quorum threshold", "quorumThreshold", quorumThreshold.String())

	coef1e18 := big.NewInt(1e18)

	vpMul1e18 := new(big.Int).Mul(current.votingPower, coef1e18)
	percent1e18 := new(big.Int).Div(vpMul1e18, s.validatorSet.TotalActiveVotingPower)

	thresholdReached := percent1e18.Cmp(quorumThreshold) >= 0
	if !thresholdReached {
		slog.DebugContext(ctx, "quorum not reached yet",
			"percentReached", decimal.NewFromBigInt(percent1e18, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
			"percentQuorumThreshold", decimal.NewFromBigInt(quorumThreshold, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
			"currentVotingPower", current.votingPower,
			"quorumThreshold", quorumThreshold,
			"totalActiveVotingPower", s.validatorSet.TotalActiveVotingPower,
			"aggSignature", current.aggSignature,
			"aggPublicKeyG1", current.aggPublicKeyG1,
			"aggPublicKeyG2", current.aggPublicKeyG2,
		)
		return nil
	}

	slog.DebugContext(ctx, "quorum reached, aggregating signatures",
		"percentReached", decimal.NewFromBigInt(percent1e18, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
		"percentQuorumThreshold", decimal.NewFromBigInt(quorumThreshold, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
		"currentVotingPower", current.votingPower,
		"quorumThreshold", quorumThreshold,
		"totalActiveVotingPower", s.validatorSet.TotalActiveVotingPower,
	)

	// todo ilya, make proof only once when threshold is reached
	start := time.Now()
	proofData, err := proof.DoProve(proof.RawProveInput{
		SignerValidators: current.validators,
		AllValidators:    s.validatorSet.Validators,
		RequiredKeyTag:   msg.Message.KeyTag,
		Message:          msg.Message.MessageHash,
		Signature:        *current.aggSignature,
		SignersAggKeyG2:  *current.aggPublicKeyG2,
	})
	if err != nil {
		return fmt.Errorf("failed to prove: %w", err)
	}

	slog.DebugContext(ctx, "proof created, trying to send aggregated signature message",
		"duration", time.Since(start).String(),
	)
	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, entity.SignaturesAggregatedMessage{
		PublicKeyG1: current.aggPublicKeyG1,
		Proof:       s.makeProofBytes(proofData),
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}

	slog.DebugContext(ctx, "proof sent via p2p", "message", current.aggSignature)

	return nil
}

func (s *AggregatorApp) makeProofBytes(proofData proof.ProofData) []byte {
	var result bytes.Buffer

	result.Write(proofData.Proof)
	result.Write(proofData.Commitments)
	result.Write(proofData.CommitmentPok)
	nonSignersAggVotingPowerBuffer := make([]byte, 32)
	proofData.NonSignersAggVotingPower.FillBytes(nonSignersAggVotingPowerBuffer)
	result.Write(nonSignersAggVotingPowerBuffer)

	return result.Bytes()
}

func initInputs() ([]byte, error) {
	arguments := abi.Arguments{
		{
			Name: "inputs",
			Type: abi.Type{
				T: abi.SliceTy,
				Elem: &abi.Type{
					T: abi.UintTy, Size: 256,
				},
			},
		},
	}

	in := []string{"0", "0", "0", "0", "0", "0", "0", "0", "17452784377140135873242247846499243451530443834097508626974155003329264289405", "0"}
	result := make([]*big.Int, 0, len(in))
	for _, s := range in {
		b, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, errors.Errorf("failed to set string: %s", s)
		}
		result = append(result, b)
	}
	pack, err := arguments.Pack(result)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return pack[64:], nil // remove first 64 bytes of dynamic array prefix, we need only bytes of inputs
}
