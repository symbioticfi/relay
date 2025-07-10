package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	Registerer prometheus.Registerer
}

type Metrics struct {
	pkSignDuration         prometheus.Summary
	appSignDuration        prometheus.Summary
	onlyAggregateDuration  prometheus.Summary
	appAggregationDuration prometheus.Summary
	p2PMessagesSent        *prometheus.CounterVec
	p2pPeerMessagesSent    *prometheus.CounterVec
	evmMethodCall          *prometheus.HistogramVec
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

	m.p2PMessagesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "symbiotic_relay_p2p_sent_messages_total",
		Help: "Total number of P2P messages sent",
	}, []string{"message_type"})
	all = append(all, m.p2PMessagesSent)

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
	m.p2PMessagesSent.WithLabelValues(messageType).Add(1)
}

func (m *Metrics) ObserveP2PPeerMessageSent(messageType, status string) {
	m.p2pPeerMessagesSent.WithLabelValues(messageType, status).Add(1)
}

func (m *Metrics) ObserveEVMMethodCall(method string, status string, d time.Duration) {
	m.evmMethodCall.WithLabelValues(method, status).Observe(d.Seconds())
}
