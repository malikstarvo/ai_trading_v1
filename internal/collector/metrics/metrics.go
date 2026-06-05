package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CollectorMetrics struct {
	MessagesTotal         *prometheus.CounterVec
	ReconnectTotal        prometheus.Counter
	ErrorsTotal           *prometheus.CounterVec
	GapEventsTotal        *prometheus.CounterVec
	LastMessageAgeSeconds *prometheus.GaugeVec
	StoredTotal           *prometheus.CounterVec
	ValidationFailed      *prometheus.CounterVec
	RateLimitWaitSeconds  prometheus.Histogram
}

func New(reg *prometheus.Registry) *CollectorMetrics {
	m := &CollectorMetrics{
		MessagesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "collector_messages_total",
			Help: "Total WebSocket messages received by topic.",
		}, []string{"topic"}),

		ReconnectTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "collector_reconnect_total",
			Help: "Total WebSocket reconnection events.",
		}),

		ErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "collector_errors_total",
			Help: "Total errors by source.",
		}, []string{"source"}),

		GapEventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "collector_gap_events_total",
			Help: "Total gap events detected.",
		}, []string{"symbol", "timeframe"}),

		LastMessageAgeSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "collector_last_message_age_seconds",
			Help: "Seconds since last message per topic.",
		}, []string{"topic"}),

		StoredTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "collector_stored_total",
			Help: "Total records persisted to DB.",
		}, []string{"table"}),

		ValidationFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "collector_validation_failed_total",
			Help: "Total validation failures by table and reason.",
		}, []string{"table", "reason"}),

		RateLimitWaitSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "collector_rate_limit_wait_seconds",
			Help:    "Time spent waiting for rate limiter.",
			Buckets: []float64{.001, .005, .01, .05, .1, .5, 1},
		}),
	}

	if reg != nil {
		reg.MustRegister(
			m.MessagesTotal, m.ReconnectTotal, m.ErrorsTotal,
			m.GapEventsTotal, m.LastMessageAgeSeconds,
			m.StoredTotal, m.ValidationFailed, m.RateLimitWaitSeconds,
		)
	}

	return m
}
