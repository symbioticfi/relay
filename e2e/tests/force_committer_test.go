package tests

import (
	"bytes"
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	evmgen "github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

func TestForceCommitterOutsideValidatorSet(t *testing.T) {
	ctx := t.Context()
	deploymentData := loadDeploymentData(t)
	extra := getOperatorData(t, deploymentData.Env.Operators)
	extraIndex := int(deploymentData.Env.Operators)
	extraClient := getGRPCClient(t, extraIndex)
	evmClient := createEVMClient(t, deploymentData)

	fundForceCommitter(t, extra.address)
	require.NoError(t, waitForHealthy(ctx, getHealthEndpoint(extraIndex), time.Minute))

	currentEpoch, err := evmClient.GetCurrentEpoch(ctx)
	require.NoError(t, err)
	networkConfig, err := evmClient.GetConfig(ctx, symbiotic.Timestamp(time.Now().Unix()), currentEpoch)
	require.NoError(t, err)
	require.NoError(t, waitForEpochCommittedOnAllSettlements(ctx, evmClient, networkConfig, currentEpoch, 2*time.Duration(deploymentData.Env.EpochTime)*time.Second))

	scenario := findForceCommitterScenario(t, evmClient, deploymentData, currentEpoch)
	extraKey := extra.blsPrivateKey.PublicKey()

	require.False(t, scenario.signingValset.IsSigner(extraKey.OnChain()), "force committer must not sign the header")
	require.False(t, scenario.signingValset.IsAggregator(extraKey.OnChain()), "force committer must not aggregate the header")
	require.False(t, scenario.targetValset.IsCommitter(extraKey.OnChain()), "force committer must not be a scheduled committer")
	t.Logf("Extra node %s is outside the signing and target validator sets for epoch %d", extra.address, scenario.targetEpoch)
	t.Logf("Scheduled committers at indexes %v can be stopped while signer quorum and an external aggregator remain available", scenario.committerIndexes)

	preTargetEpoch := scenario.targetEpoch - 1
	selectedEpochTimeout := time.Duration(uint64(scenario.targetEpoch-currentEpoch)+1) * time.Duration(deploymentData.Env.EpochTime) * time.Second
	require.NoError(t, waitForEpoch(ctx, evmClient, preTargetEpoch, selectedEpochTimeout))
	require.NoError(t, waitForEpochCommittedOnAllSettlements(
		ctx,
		evmClient,
		scenario.networkConfig,
		preTargetEpoch,
		forceCommitterTimeout(deploymentData.Env.EpochTime),
	))

	committersRestored := false
	t.Cleanup(func() {
		if committersRestored {
			return
		}
		for _, index := range scenario.committerIndexes {
			container := deploymentData.Env.GetSidecarConfigs()[index].ContainerName
			require.NoErrorf(t, startContainer(context.Background(), container), "failed to restart scheduled committer %s", container)
		}
	})
	for _, index := range scenario.committerIndexes {
		container := deploymentData.Env.GetSidecarConfigs()[index].ContainerName
		require.NoErrorf(t, stopContainer(ctx, container), "failed to stop scheduled committer %s", container)
	}

	require.NoError(t, waitForEpoch(ctx, evmClient, scenario.targetEpoch, forceCommitterTimeout(deploymentData.Env.EpochTime)))

	var metadata *apiv1.GetValidatorSetMetadataResponse
	require.NoError(t, waitForErrorIsNil(ctx, forceCommitterTimeout(deploymentData.Env.EpochTime), func() error {
		var err error
		metadata, err = extraClient.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{
			Epoch: new(uint64(scenario.targetEpoch)),
		})
		return err
	}))
	require.NotEmpty(t, metadata.GetRequestId(), "force committer must track the header request metadata")
	t.Logf("Extra node tracked header request %s", metadata.GetRequestId())

	localValset, err := extraClient.GetValidatorSet(ctx, &apiv1.GetValidatorSetRequest{Epoch: new(uint64(scenario.targetEpoch))})
	require.NoError(t, err)
	require.Equal(t, uint64(scenario.targetEpoch), localValset.GetValidatorSet().GetEpoch())
	for _, validator := range localValset.GetValidatorSet().GetValidators() {
		for _, key := range validator.GetKeys() {
			if key.GetTag() == uint32(scenario.targetValset.RequiredKeyTag) {
				require.False(t, bytes.Equal(key.GetPayload(), extraKey.OnChain()), "extra key must be absent from the locally tracked validator set")
			}
		}
	}

	_, err = extraClient.GetLocalValidator(ctx, &apiv1.GetLocalValidatorRequest{})
	require.Equal(t, codes.NotFound, status.Code(err), "extra node must not resolve to a local validator")
	_, err = extraClient.GetSignatureRequest(ctx, &apiv1.GetSignatureRequestRequest{RequestId: metadata.GetRequestId()})
	require.Equal(t, codes.NotFound, status.Code(err), "non-signer must not create a local signing request")
	t.Log("Extra node tracked the valset and request metadata without creating a local signature request")

	var signatures *apiv1.GetSignaturesResponse
	require.NoError(t, waitForErrorIsNil(ctx, forceCommitterTimeout(deploymentData.Env.EpochTime), func() error {
		var err error
		signatures, err = extraClient.GetSignatures(ctx, &apiv1.GetSignaturesRequest{RequestId: metadata.GetRequestId()})
		if err != nil {
			return err
		}
		if len(signatures.GetSignatures()) == 0 {
			return errors.New("force committer has not received validator signatures")
		}
		return nil
	}))

	signedVotingPower := validatorSignatureVotingPower(t, scenario.signingValset, signatures.GetSignatures(), extraKey.Raw())
	require.GreaterOrEqual(t, signedVotingPower.Cmp(scenario.signingValset.QuorumThreshold.Int), 0,
		"force committer must receive signatures representing quorum voting power")
	t.Logf("Extra node stored %d external signatures with voting power %s (quorum %s) and no signature from its own key",
		len(signatures.GetSignatures()), signedVotingPower, scenario.signingValset.QuorumThreshold.String())

	var proof *apiv1.GetAggregationProofResponse
	require.NoError(t, waitForErrorIsNil(ctx, forceCommitterTimeout(deploymentData.Env.EpochTime), func() error {
		var err error
		proof, err = extraClient.GetAggregationProof(ctx, &apiv1.GetAggregationProofRequest{RequestId: metadata.GetRequestId()})
		return err
	}))
	require.NotNil(t, proof.GetAggregationProof())
	require.NotEmpty(t, proof.GetAggregationProof().GetProof(), "force committer must store the proof produced by a validator aggregator")
	t.Logf("Extra node stored the externally aggregated proof (%d bytes)", len(proof.GetAggregationProof().GetProof()))

	require.NoError(t, waitForEpochCommittedOnAllSettlements(
		ctx,
		evmClient,
		scenario.networkConfig,
		scenario.targetEpoch,
		forceCommitterTimeout(deploymentData.Env.EpochTime),
	))

	for _, settlement := range scenario.networkConfig.Settlements {
		chainURL := settlementChainURL(t, settlement.ChainId)
		sender := valsetCommitSender(t, chainURL, settlement.Address, scenario.targetEpoch)
		require.Equal(t, extra.address, sender,
			"header on chain %d must be committed by the out-of-valset force committer", settlement.ChainId)
		t.Logf("Epoch %d committed on chain %d by extra node %s", scenario.targetEpoch, settlement.ChainId, sender)
	}

	restoreScheduledCommitters(t, deploymentData, scenario.committerIndexes, metadata.GetRequestId())
	committersRestored = true
}

func restoreScheduledCommitters(t *testing.T, deploymentData RelayContractsData, indexes []int, requestID string) {
	t.Helper()

	const restoreAttempts = 3
	for _, index := range indexes {
		container := deploymentData.Env.GetSidecarConfigs()[index].ContainerName
		var syncErr error
		for attempt := range restoreAttempts {
			require.NoErrorf(t, startContainer(t.Context(), container), "failed to restart scheduled committer %s", container)
			require.NoErrorf(t, waitForHealthy(t.Context(), getHealthEndpoint(index), 2*time.Minute),
				"scheduled committer %s did not become healthy", container)

			client := getGRPCClient(t, index)
			syncErr = waitForErrorIsNil(t.Context(), 30*time.Second, func() error {
				_, err := client.GetAggregationProof(t.Context(), &apiv1.GetAggregationProofRequest{RequestId: requestID})
				return err
			})
			if syncErr == nil {
				break
			}
			t.Logf("Scheduled committer %s did not reconnect after attempt %d", container, attempt+1)
		}
		require.NoErrorf(t, syncErr, "scheduled committer %s did not catch up the aggregation proof", container)
	}
}

type forceCommitterScenario struct {
	targetEpoch      symbiotic.Epoch
	signingValset    symbiotic.ValidatorSet
	targetValset     symbiotic.ValidatorSet
	networkConfig    symbiotic.NetworkConfig
	committerIndexes []int
}

func findForceCommitterScenario(
	t *testing.T,
	evmClient *evm.Client,
	deploymentData RelayContractsData,
	currentEpoch symbiotic.Epoch,
) forceCommitterScenario {
	t.Helper()

	deriver, err := valsetDeriver.NewDeriver(evmClient, nil)
	require.NoError(t, err)
	configs := deploymentData.Env.GetSidecarConfigs()

	const epochsToCheck = 6
	valsets := make(map[symbiotic.Epoch]symbiotic.ValidatorSet, epochsToCheck+1)
	networkConfigs := make(map[symbiotic.Epoch]symbiotic.NetworkConfig, epochsToCheck+1)
	for offset := range epochsToCheck + 1 {
		epoch := currentEpoch + symbiotic.Epoch(offset)
		captureTimestamp, err := evmClient.GetEpochStart(t.Context(), epoch)
		require.NoError(t, err)
		networkConfig, err := evmClient.GetConfig(t.Context(), captureTimestamp, epoch)
		require.NoError(t, err)
		valset, err := deriver.GetValidatorSet(t.Context(), epoch, networkConfig)
		require.NoError(t, err)
		valsets[epoch] = valset
		networkConfigs[epoch] = networkConfig
	}

	for offset := 2; offset <= epochsToCheck; offset++ {
		targetEpoch := currentEpoch + symbiotic.Epoch(offset)
		targetValset := valsets[targetEpoch]
		committerIndexes := scheduledNodeIndexes(configs, targetValset.IsCommitter)
		if len(committerIndexes) == 0 {
			continue
		}
		if !stoppingNodesPreservesProofFlow(t, valsets[targetEpoch-1], configs, targetEpoch, committerIndexes) {
			continue
		}
		return forceCommitterScenario{
			targetEpoch:      targetEpoch,
			signingValset:    valsets[targetEpoch-1],
			targetValset:     targetValset,
			networkConfig:    networkConfigs[targetEpoch],
			committerIndexes: committerIndexes,
		}
	}

	t.Fatal("could not find an epoch where scheduled committers can be stopped without losing signer quorum or all aggregators")
	return forceCommitterScenario{}
}

func scheduledNodeIndexes(
	configs []RelaySidecarConfig,
	isScheduled func(symbiotic.CompactPublicKey) bool,
) []int {
	indexes := make([]int, 0)
	for index, config := range configs {
		if isScheduled(config.RequiredSymKey.PublicKey().OnChain()) {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func stoppingNodesPreservesProofFlow(
	t *testing.T,
	signingValset symbiotic.ValidatorSet,
	configs []RelaySidecarConfig,
	targetEpoch symbiotic.Epoch,
	stoppedIndexes []int,
) bool {
	t.Helper()

	stopped := make(map[int]struct{}, len(stoppedIndexes))
	for _, index := range stoppedIndexes {
		stopped[index] = struct{}{}
	}

	remainingVotingPower := new(big.Int).Set(signingValset.GetTotalActiveVotingPower().Int)
	hasAggregator := false
	for index, config := range configs {
		key := config.RequiredSymKey.PublicKey().OnChain()
		if _, isStopped := stopped[index]; isStopped {
			if validator, found := signingValset.FindValidatorByKey(signingValset.RequiredKeyTag, key); found && validator.IsActive {
				remainingVotingPower.Sub(remainingVotingPower, validator.VotingPower.Int)
			}
			continue
		}
		if signingValset.IsAggregator(key) {
			hasAggregator = true
		}
	}
	if !hasAggregator || remainingVotingPower.Cmp(signingValset.QuorumThreshold.Int) < 0 {
		t.Logf("Cannot stop candidate committers for epoch %d: signing epoch %d would have aggregator=%t, voting power=%s, quorum=%s",
			targetEpoch, signingValset.Epoch, hasAggregator, remainingVotingPower, signingValset.QuorumThreshold.String())
		return false
	}
	return true
}

func validatorSignatureVotingPower(
	t *testing.T,
	valset symbiotic.ValidatorSet,
	signatures []*apiv1.Signature,
	forceCommitterRawKey []byte,
) *big.Int {
	t.Helper()

	votingPower := new(big.Int)
	seenOperators := make(map[common.Address]struct{}, len(signatures))
	for _, signature := range signatures {
		require.False(t, bytes.Equal(signature.GetPublicKey(), forceCommitterRawKey), "force committer must not sign")

		publicKey, err := blsBn254.FromRaw(blsBn254.RawPublicKey(signature.GetPublicKey()))
		require.NoError(t, err)
		validator, found := valset.FindValidatorByKey(valset.RequiredKeyTag, publicKey.OnChain())
		require.True(t, found, "received signature must belong to the signing validator set")
		_, duplicate := seenOperators[validator.Operator]
		require.False(t, duplicate, "validator signature must be counted once")
		seenOperators[validator.Operator] = struct{}{}
		votingPower.Add(votingPower, validator.VotingPower.Int)
	}
	return votingPower
}

func fundForceCommitter(t *testing.T, address common.Address) {
	t.Helper()

	for chainID, chainURL := range map[uint64]string{
		31337: settlementChains[0],
		31338: settlementChains[1],
	} {
		result, err := fundOperator(t.Context(), getFunderPrivateKey(t), chainURL, symbiotic.CrossChainAddress{
			ChainId: chainID,
			Address: address,
		}, big.NewInt(1e18))
		require.NoErrorf(t, err, "failed to fund force committer on chain %d", chainID)
		t.Logf("Funded force committer on chain %d in transaction %s", chainID, result.TxHash)
	}
}

func waitForEpochCommittedOnAllSettlements(
	ctx context.Context,
	evmClient *evm.Client,
	networkConfig symbiotic.NetworkConfig,
	epoch symbiotic.Epoch,
	timeout time.Duration,
) error {
	return waitForErrorIsNil(ctx, timeout, func() error {
		for _, settlement := range networkConfig.Settlements {
			committed, err := evmClient.IsValsetHeaderCommittedAt(ctx, settlement, epoch)
			if err != nil {
				return err
			}
			if !committed {
				return errors.Errorf("epoch %d is not committed on chain %d", epoch, settlement.ChainId)
			}
		}
		return nil
	})
}

func valsetCommitSender(t *testing.T, chainURL string, settlementAddress common.Address, epoch symbiotic.Epoch) common.Address {
	t.Helper()

	client, err := ethclient.DialContext(t.Context(), chainURL)
	require.NoError(t, err)
	defer client.Close()

	settlement, err := evmgen.NewSettlementFilterer(settlementAddress, client)
	require.NoError(t, err)
	iterator, err := settlement.FilterCommitValSetHeader(&bind.FilterOpts{Start: 0, Context: t.Context()})
	require.NoError(t, err)
	defer iterator.Close()

	var transactionHash common.Hash
	for iterator.Next() {
		if iterator.Event.ValSetHeader.Epoch.Uint64() == uint64(epoch) {
			transactionHash = iterator.Event.Raw.TxHash
			break
		}
	}
	require.NoError(t, iterator.Error())
	require.NotEqual(t, common.Hash{}, transactionHash, "commit event for epoch %d not found", epoch)

	transaction, pending, err := client.TransactionByHash(t.Context(), transactionHash)
	require.NoError(t, err)
	require.False(t, pending)
	chainID, err := client.ChainID(t.Context())
	require.NoError(t, err)
	sender, err := types.Sender(types.LatestSignerForChainID(chainID), transaction)
	require.NoError(t, err)
	return sender
}

func settlementChainURL(t *testing.T, chainID uint64) string {
	t.Helper()

	switch chainID {
	case 31337:
		return settlementChains[0]
	case 31338:
		return settlementChains[1]
	default:
		t.Fatalf("unknown settlement chain %d", chainID)
		return ""
	}
}

func forceCommitterTimeout(epochTime uint64) time.Duration {
	return 3 * time.Duration(epochTime) * time.Second
}
