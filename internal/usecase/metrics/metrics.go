package metrics

import (
	"math/big"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/symbioticfi/relay/core/entity"
)

type Config struct {
	Registerer prometheus.Registerer
}

type Metrics struct {
	GRPCMetrics

	pkSignDuration             prometheus.Summary
	appSignDuration            prometheus.Summary
	onlyAggregateDuration      prometheus.Summary
	appAggregationDuration     prometheus.Summary
	p2pMessagesSent            *prometheus.CounterVec
	p2pPeerMessagesSent        *prometheus.CounterVec
	evmMethodCall              *prometheus.HistogramVec
	evmCommitGasUsed           *prometheus.HistogramVec
	evmCommitGasPrice          *prometheus.HistogramVec
	appAggProofCompleted       prometheus.Histogram
	appAggProofReceived        prometheus.Histogram
	p2pSyncProcessedSignatures *prometheus.CounterVec
	p2pSyncRequestedHashes     prometheus.Counter
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
	defaultBuckets := []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 7.5, 10, 15, 20, 30, 60}

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

	m.p2pMessagesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_sent_messages_total",
		Help: "Total number of P2P messages sent",
	}, []string{"message_type"})
	all = append(all, m.p2pMessagesSent)

	m.p2pPeerMessagesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_peer_sent_messages_total",
		Help: "Total number of P2P messages sent to peers",
	}, []string{"message_type", "status"})
	all = append(all, m.p2pPeerMessagesSent)

	m.evmMethodCall = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_evm_method_call_duration_seconds",
		Help:    "Duration of EVM method calls in seconds",
		Buckets: defaultBuckets,
	}, []string{"method", "status"})
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

	m.appAggProofCompleted = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_agg_proof_completed_seconds",
		Help:    "Time taken to complete aggregation proof",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2, 3, 5, 7, 8, 9, 10, 12, 15, 17, 20, 25, 30, 35, 40, 45, 50, 55, 60, 90, 120, 150, 180, 240, 300, 600},
	})
	all = append(all, m.appAggProofCompleted)

	m.appAggProofReceived = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "symbiotic_relay_agg_proof_received_seconds",
		Help:    "Time taken to receive aggregation proof",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2, 3, 5, 7, 8, 9, 10, 12, 15, 17, 20, 25, 30, 35, 40, 45, 50, 55, 60, 90, 120, 150, 180, 240, 300, 600},
	})
	all = append(all, m.appAggProofReceived)

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

func (m *Metrics) ObserveP2PMessageSent(messageType string) {
	m.p2pMessagesSent.WithLabelValues(messageType).Add(1)
}

func (m *Metrics) ObserveP2PPeerMessageSent(messageType, status string) {
	m.p2pPeerMessagesSent.WithLabelValues(messageType, status).Add(1)
}

func (m *Metrics) ObserveEVMMethodCall(method string, status string, d time.Duration) {
	m.evmMethodCall.WithLabelValues(method, status).Observe(d.Seconds())
}

func (m *Metrics) ObserveCommitValsetHeaderParams(chainID uint64, gasUsed uint64, effectiveGasPrice *big.Int) {
	m.evmCommitGasUsed.WithLabelValues(strconv.FormatInt(int64(chainID), 10)).Observe(float64(gasUsed))
	gasPrice, _ := effectiveGasPrice.Float64()
	m.evmCommitGasPrice.WithLabelValues(strconv.FormatInt(int64(chainID), 10)).Observe(gasPrice)
}

func (m *Metrics) ObserveAggCompleted(stat entity.SignatureStat) {
	reqReceivedTime, ok := stat.StatMap[entity.SignatureStatStageSignRequestReceived]
	if !ok {
		return
	}
	aggProofCompletedTime, ok := stat.StatMap[entity.SignatureStatStageAggCompleted]
	if !ok {
		return
	}
	m.appAggProofCompleted.Observe(aggProofCompletedTime.Sub(reqReceivedTime).Seconds())
}

func (m *Metrics) ObserveAggReceived(stat entity.SignatureStat) {
	reqReceivedTime, ok := stat.StatMap[entity.SignatureStatStageSignRequestReceived]
	if !ok {
		return
	}
	aggProofReceivedTime, ok := stat.StatMap[entity.SignatureStatStageAggProofReceived]
	if !ok {
		return
	}
	m.appAggProofReceived.Observe(aggProofReceivedTime.Sub(reqReceivedTime).Seconds())
}

func (m *Metrics) ObserveP2PSyncSignaturesProcessed(resultType string, count int) {
	m.p2pSyncProcessedSignatures.WithLabelValues(resultType).Add(float64(count))
}

func (m *Metrics) ObserveP2PSyncRequestedHashes(count int) {
	m.p2pSyncRequestedHashes.Add(float64(count))
}

func (m *Metrics) ObserveP2PSyncAggregationProofsProcessed(resultType string, count int) {
	// Reuse the same metric as signatures for now, but with different labels if needed
	// In future, we might want separate metrics for aggregation proofs
	m.p2pSyncProcessedSignatures.WithLabelValues(resultType).Add(float64(count))
}

func (m *Metrics) ObserveP2PSyncRequestedAggregationProofs(count int) {
	// Reuse the same metric as signature hashes for now
	// In future, we might want separate metrics for aggregation proof requests
	m.p2pSyncRequestedHashes.Add(float64(count))
}
