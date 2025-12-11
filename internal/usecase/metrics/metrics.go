package metrics

import (
	"math/big"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	Registerer prometheus.Registerer
}

type Metrics struct {
	GRPCMetrics

	// signatures and aggregation
	pkSignDuration         prometheus.Summary
	appSignDuration        prometheus.Summary
	onlyAggregateDuration  prometheus.Summary
	appAggregationDuration prometheus.Summary
	aggregationProofSize   *prometheus.HistogramVec

	// p2p
	p2pPeerMessagesSent            *prometheus.CounterVec
	p2pSyncProcessedSignatures     *prometheus.CounterVec
	p2pSyncRequestedHashes         prometheus.Counter
	p2pSyncProcessedAggProofs      *prometheus.CounterVec
	p2pSyncRequestedAggProofHashes prometheus.Counter

	// repo
	repoQueryDuration      *prometheus.HistogramVec
	repoQueryTotalDuration *prometheus.HistogramVec

	// evm
	evmMethodCall     *prometheus.HistogramVec
	evmCommitGasUsed  *prometheus.HistogramVec
	evmCommitGasPrice *prometheus.HistogramVec

	// epoch
	epochsTotal *prometheus.GaugeVec
	epochTime   *prometheus.GaugeVec

	// pruner
	prunedEpochsTotal *prometheus.CounterVec
}

func New(cfg Config) *Metrics {
	registerer := cfg.Registerer
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	return newMetrics(registerer)
}

func newMetrics(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{}

	defaultPercentiles := map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	defaultBuckets := []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 7.5, 10, 15, 20, 30, 60}

	var all []prometheus.Collector

	m.pkSignDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "symbiotic_relay_private_key_sign_duration_seconds",
		Help:       "Duration of private key signing in seconds",
		Objectives: defaultPercentiles,
	})
	all = append(all, m.pkSignDuration)

	m.appSignDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "symbiotic_relay_app_sign_duration_seconds",
		Help:       "Duration of app sign in seconds",
		Objectives: defaultPercentiles,
	})
	all = append(all, m.appSignDuration)

	m.onlyAggregateDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "symbiotic_relay_only_aggregate_duration_seconds",
		Help:       "Duration of just aggregation in seconds",
		Objectives: defaultPercentiles,
	})
	all = append(all, m.onlyAggregateDuration)

	m.appAggregationDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "symbiotic_relay_app_aggregate_duration_seconds",
		Help:       "Duration of app aggregate in seconds",
		Objectives: defaultPercentiles,
	})
	all = append(all, m.appAggregationDuration)

	m.p2pPeerMessagesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_peer_sent_messages_total",
		Help: "Total number of P2P messages sent to peers",
	}, []string{"message_type", "status"})
	all = append(all, m.p2pPeerMessagesSent)

	m.evmMethodCall = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_evm_method_call_duration_seconds",
		Help:    "Duration of EVM method calls in seconds",
		Buckets: defaultBuckets,
	}, []string{"method", "chainId", "status"})
	all = append(all, m.evmMethodCall)

	m.evmCommitGasUsed = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_evm_commit_gas_used",
		Help:    "Gas used for EVM commit operations",
		Buckets: []float64{1e5, 2e5, 3e5, 5e5, 7e5, 1e6, 3e6, 5e6, 7e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12},
	}, []string{"chainId"})
	all = append(all, m.evmCommitGasUsed)

	m.evmCommitGasPrice = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_evm_commit_gas_price",
		Help:    "Gas price for EVM commit operations",
		Buckets: []float64{1e9, 5e9, 1e10, 5e10, 1e11, 5e11, 1e12, 5e12, 1e13, 5e13, 1e14, 5e14, 1e15, 5e15, 1e16},
	}, []string{"chainId"})
	all = append(all, m.evmCommitGasPrice)

	m.p2pSyncProcessedSignatures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_sync_processed_signatures_total",
		Help: "Total number of signatures processed during P2P sync",
	}, []string{"process_result"})
	all = append(all, m.p2pSyncProcessedSignatures)

	m.p2pSyncRequestedHashes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_sync_requested_hashes_total",
		Help: "Total number of requested hashes during P2P sync",
	})
	all = append(all, m.p2pSyncRequestedHashes)

	m.p2pSyncProcessedAggProofs = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_sync_processed_aggregation_proofs_total",
		Help: "Total number of aggregation proofs processed during P2P sync",
	}, []string{"process_result"})
	all = append(all, m.p2pSyncProcessedAggProofs)

	m.p2pSyncRequestedAggProofHashes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_sync_requested_aggregation_proof_hashes_total",
		Help: "Total number of requested aggregation proof hashes during P2P sync",
	})
	all = append(all, m.p2pSyncRequestedAggProofHashes)

	m.repoQueryDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "symbiotic_relay_repo_query_duration_seconds",
		Help: "Duration of repository queries in seconds",
	}, []string{"query_name", "status"})
	all = append(all, m.repoQueryDuration)

	m.repoQueryTotalDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "symbiotic_relay_repo_query_total_duration_seconds",
		Help: " Total duration of repository queries in seconds, including retries",
	}, []string{"query_name", "status"})
	all = append(all, m.repoQueryTotalDuration)

	proofSizeBuckets := []float64{
		256,     // 256B
		512,     // 512B
		1024,    // 1KB
		2048,    // 2KB
		4096,    // 4KB
		8192,    // 8KB
		16384,   // 16KB
		32768,   // 32KB
		65536,   // 64KB
		131072,  // 128KB
		262144,  // 256KB
		393216,  // 384KB
		524288,  // 512KB
		786432,  // 768KB
		1048576, // 1MB
	}
	m.aggregationProofSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_aggregation_proof_size_bytes",
		Help:    "Size of aggregation proofs in bytes, labeled by active validator count",
		Buckets: proofSizeBuckets,
	}, []string{"validator_count"})
	all = append(all, m.aggregationProofSize)

	grpcDurationBuckets := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 5, 10, 15, 20, 40, 45, 60}
	m.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds.",
			Buckets: grpcDurationBuckets,
		},
		[]string{"method", "status_code"},
	)
	all = append(all, m.requestDuration)

	m.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method", "status_code"},
	)
	all = append(all, m.requestsTotal)

	m.requestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_server_requests_in_flight",
			Help: "Current number of gRPC requests being processed.",
		},
		[]string{"method"},
	)
	all = append(all, m.requestsInFlight)

	m.epochsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "symbiotic_relay_epochs_total",
		Help: "Latest number of epochs",
	}, []string{"type"})
	all = append(all, m.epochsTotal)

	m.epochTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "symbiotic_relay_epoch_time_seconds",
		Help: "Time of the latest epoch in seconds",
	}, []string{"type"})
	all = append(all, m.epochTime)

	m.prunedEpochsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_pruned_epochs_total",
		Help: "Total number of epochs pruned from storage",
	}, []string{"entity_type"})
	all = append(all, m.prunedEpochsTotal)

	registerer.MustRegister(all...)
	return m
}

func (m *Metrics) ObservePKSignDuration(d time.Duration) {
	m.pkSignDuration.Observe(d.Seconds())
}

func (m *Metrics) ObserveAppSignDuration(d time.Duration) {
	m.appSignDuration.Observe(d.Seconds())
}

func (m *Metrics) ObserveOnlyAggregateDuration(d time.Duration) {
	m.onlyAggregateDuration.Observe(d.Seconds())
}

func (m *Metrics) ObserveAppAggregateDuration(d time.Duration) {
	m.appAggregationDuration.Observe(d.Seconds())
}

func (m *Metrics) ObserveP2PPeerMessageSent(messageType, status string) {
	m.p2pPeerMessagesSent.WithLabelValues(messageType, status).Add(1)
}

func (m *Metrics) ObserveEVMMethodCall(method string, chainID uint64, status string, d time.Duration) {
	m.evmMethodCall.WithLabelValues(method, strconv.FormatUint(chainID, 10), status).Observe(d.Seconds())
}

func (m *Metrics) ObserveCommitValsetHeaderParams(chainID uint64, gasUsed uint64, effectiveGasPrice *big.Int) {
	m.evmCommitGasUsed.WithLabelValues(strconv.FormatInt(int64(chainID), 10)).Observe(float64(gasUsed))
	gasPrice, _ := effectiveGasPrice.Float64()
	m.evmCommitGasPrice.WithLabelValues(strconv.FormatInt(int64(chainID), 10)).Observe(gasPrice)
}

func (m *Metrics) ObserveP2PSyncSignaturesProcessed(resultType string, count int) {
	m.p2pSyncProcessedSignatures.WithLabelValues(resultType).Add(float64(count))
}

func (m *Metrics) ObserveP2PSyncRequestedHashes(count int) {
	m.p2pSyncRequestedHashes.Add(float64(count))
}

func (m *Metrics) ObserveP2PSyncAggregationProofsProcessed(resultType string, count int) {
	m.p2pSyncProcessedAggProofs.WithLabelValues(resultType).Add(float64(count))
}

func (m *Metrics) ObserveP2PSyncRequestedAggregationProofs(count int) {
	m.p2pSyncRequestedAggProofHashes.Add(float64(count))
}

func (m *Metrics) ObserveAggregationProofSize(proofSizeBytes int, activeValidatorCount int) {
	m.aggregationProofSize.WithLabelValues(strconv.Itoa(activeValidatorCount)).Observe(float64(proofSizeBytes))
}

func (m *Metrics) ObserveRepoQueryDuration(queryName, status string, d time.Duration) {
	m.repoQueryDuration.WithLabelValues(queryName, status).Observe(d.Seconds())
}

func (m *Metrics) ObserveRepoQueryTotalDuration(queryName, status string, d time.Duration) {
	m.repoQueryTotalDuration.WithLabelValues(queryName, status).Observe(d.Seconds())
}

func (m *Metrics) IncPrunedEpochsCount(entityType string) {
	m.prunedEpochsTotal.WithLabelValues(entityType).Inc()
}

func (m *Metrics) ObserveEpoch(epochType string, epochNumber uint64) {
	m.epochsTotal.WithLabelValues(epochType).Set(float64(epochNumber))
	m.epochTime.WithLabelValues(epochType).Set(float64(time.Now().Unix()))
}
