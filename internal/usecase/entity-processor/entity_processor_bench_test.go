package entity_processor

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/sync/errgroup"

	badgerrepo "github.com/symbioticfi/relay/internal/client/repository/badger"
	bboltrepo "github.com/symbioticfi/relay/internal/client/repository/bbolt"
	"github.com/symbioticfi/relay/internal/client/repository/cached"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/internal/usecase/entity-processor/mocks"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func BenchmarkProcessSignature(b *testing.B) {
	cases := []struct {
		name       string
		validators int
		requests   int
	}{
		{"10v_1r", 10, 1},
		{"10v_5r", 10, 5},
		{"10v_10r", 10, 10},
	}

	for name, newRepo := range benchBackends() {
		b.Run(name, func(b *testing.B) {
			for _, tc := range cases {
				b.Run(tc.name, func(b *testing.B) {
					repo := newRepo(b)
					epoch := symbiotic.Epoch(100)

					vs, privateKeys := benchCreateValidatorSet(b, epoch, big.NewInt(100000), tc.validators)
					require.NoError(b, repo.SaveNextValsetData(b.Context(), entity.NextValsetData{
						NextValidatorSet:  vs,
						NextNetworkConfig: randomNetworkConfig(),
						PrevValidatorSet:  vs,
						PrevNetworkConfig: randomNetworkConfig(),
					}))

					processor, err := NewEntityProcessor(Config{
						Repo:                     repo,
						Aggregator:               benchMockAggregator(b),
						AggProofSignal:           benchMockAggProofSignal(b),
						SignatureProcessedSignal: signals.New[symbiotic.Signature](signals.DefaultConfig(), "bench", nil),
						Metrics:                  doNothingMetrics{},
					})
					require.NoError(b, err)

					b.ReportAllocs()
					b.ResetTimer()

					for range b.N {
						b.StopTimer()
						sigs := benchGenerateSignatures(b, privateKeys, epoch, tc.requests)
						b.StartTimer()

						eg, ctx := errgroup.WithContext(b.Context())
						for _, sig := range sigs {
							eg.Go(func() error {
								return processor.ProcessSignature(ctx, sig, true)
							})
						}
						require.NoError(b, eg.Wait())
					}
				})
			}
		})
	}
}

func benchGenerateSignatures(
	b *testing.B,
	privateKeys []map[symbiotic.KeyTag]crypto.PrivateKey,
	epoch symbiotic.Epoch,
	numRequests int,
) []symbiotic.Signature {
	b.Helper()

	validators := len(privateKeys)
	sigs := make([]symbiotic.Signature, 0, numRequests*validators)

	for range numRequests {
		req := benchRandomSignatureRequest(b, epoch)
		for v := range validators {
			sigs = append(sigs, benchSignatureForRequest(b, privateKeys[v][req.KeyTag], req))
		}
	}

	return sigs
}

// Benchmark-specific helpers (mirror test helpers but accept *testing.B)

type benchRepoFactory func(b *testing.B) cached.Repository

func benchBackends() map[string]benchRepoFactory {
	return map[string]benchRepoFactory{
		"badger": func(b *testing.B) cached.Repository {
			b.Helper()
			repo, err := badgerrepo.New(badgerrepo.Config{Dir: b.TempDir(), Metrics: repoutil.DoNothingMetrics{}, BlockCacheSize: -1})
			require.NoError(b, err)
			b.Cleanup(func() { repo.Close() })
			return repo
		},
		"bbolt": func(b *testing.B) cached.Repository {
			b.Helper()
			repo, err := bboltrepo.New(bboltrepo.Config{Dir: b.TempDir(), Metrics: repoutil.DoNothingMetrics{}})
			require.NoError(b, err)
			b.Cleanup(func() { repo.Close() })
			return repo
		},
	}
}

func benchRandomBytes(b *testing.B, n int) []byte {
	b.Helper()
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	require.NoError(b, err)
	return buf
}

func benchRandomSignatureRequest(b *testing.B, epoch symbiotic.Epoch) symbiotic.SignatureRequest {
	b.Helper()
	return symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       benchRandomBytes(b, 512),
	}
}

func benchSignatureForRequest(b *testing.B, privateKey crypto.PrivateKey, req symbiotic.SignatureRequest) symbiotic.Signature {
	b.Helper()
	sig, messageHash, err := privateKey.Sign(req.Message)
	require.NoError(b, err)
	return symbiotic.Signature{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: messageHash,
		Signature:   sig,
		PublicKey:   privateKey.PublicKey(),
	}
}

func benchCreateValidatorSet(b *testing.B, epoch symbiotic.Epoch, quorumThreshold *big.Int, count int) (symbiotic.ValidatorSet, []map[symbiotic.KeyTag]crypto.PrivateKey) {
	b.Helper()

	privateKeys := make([]map[symbiotic.KeyTag]crypto.PrivateKey, count)
	validators := make([]symbiotic.Validator, count)

	for i := range count {
		privateKeys[i] = make(map[symbiotic.KeyTag]crypto.PrivateKey)

		privBLS, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
		require.NoError(b, err)
		privateKeys[i][15] = privBLS

		privECDSA, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeEcdsaSecp256k1)
		require.NoError(b, err)
		privateKeys[i][0x10] = privECDSA

		validators[i] = symbiotic.Validator{
			Operator:    common.BytesToAddress(benchRandomBytes(b, 20)),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{
				{Tag: symbiotic.KeyTag(15), Payload: privBLS.PublicKey().OnChain()},
				{Tag: symbiotic.KeyTag(0x10), Payload: privECDSA.PublicKey().OnChain()},
			},
			Vaults: []symbiotic.ValidatorVault{{
				ChainID:     1,
				Vault:       common.BytesToAddress(benchRandomBytes(b, 20)),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
			}},
		}
	}

	list := symbiotic.Validators(validators)
	list.SortByOperatorAddressAsc()

	return symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  symbiotic.ToVotingPower(quorumThreshold),
		Validators:       list,
		Status:           symbiotic.HeaderCommitted,
	}, privateKeys
}

func benchMockAggregator(b *testing.B) *mocks.MockAggregator {
	b.Helper()
	ctrl := gomock.NewController(b)
	m := mocks.NewMockAggregator(ctrl)
	m.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	return m
}

func benchMockAggProofSignal(b *testing.B) *mocks.MockAggProofSignal {
	b.Helper()
	ctrl := gomock.NewController(b)
	m := mocks.NewMockAggProofSignal(ctrl)
	m.EXPECT().Emit(gomock.Any()).Return(nil).AnyTimes()
	return m
}
