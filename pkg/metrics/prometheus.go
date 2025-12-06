// Package metrics provides Prometheus metrics collection for TFDrift-Falco.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for TFDrift-Falco
type Metrics struct {
	// Total number of drift alerts by severity and resource type
	DriftAlertsTotal *prometheus.CounterVec

	// Current number of unresolved drift alerts
	UnresolvedAlerts *prometheus.GaugeVec

	// Histogram of drift detection latency
	DetectionLatency prometheus.Histogram

	// Total number of events processed
	EventsProcessed *prometheus.CounterVec

	// Current status of TFDrift-Falco components
	ComponentStatus *prometheus.GaugeVec
}

var (
	// Global metrics instance
	DefaultMetrics *Metrics
)

// NewMetrics creates a new metrics instance with all collectors registered
func NewMetrics(namespace string) *Metrics {
	if namespace == "" {
		namespace = "tfdrift"
	}

	m := &Metrics{
		DriftAlertsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "drift_alerts_total",
				Help:      "Total number of drift alerts detected",
			},
			[]string{"severity", "resource_type", "provider"},
		),

		UnresolvedAlerts: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "unresolved_alerts",
				Help:      "Current number of unresolved drift alerts",
			},
			[]string{"severity", "resource_type"},
		),

		DetectionLatency: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "detection_latency_seconds",
				Help:      "Histogram of drift detection latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
		),

		EventsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "events_processed_total",
				Help:      "Total number of events processed",
			},
			[]string{"event_type", "source", "status"},
		),

		ComponentStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "component_status",
				Help:      "Status of TFDrift-Falco components (1=healthy, 0=unhealthy)",
			},
			[]string{"component"},
		),
	}

	DefaultMetrics = m
	return m
}

// RecordDriftAlert records a new drift alert
func (m *Metrics) RecordDriftAlert(severity, resourceType, provider string) {
	m.DriftAlertsTotal.WithLabelValues(severity, resourceType, provider).Inc()
	m.UnresolvedAlerts.WithLabelValues(severity, resourceType).Inc()
}

// ResolveAlert decrements the unresolved alerts counter
func (m *Metrics) ResolveAlert(severity, resourceType string) {
	m.UnresolvedAlerts.WithLabelValues(severity, resourceType).Dec()
}

// RecordEvent records a processed event
func (m *Metrics) RecordEvent(eventType, source, status string) {
	m.EventsProcessed.WithLabelValues(eventType, source, status).Inc()
}

// RecordDetectionLatency records drift detection latency
func (m *Metrics) RecordDetectionLatency(seconds float64) {
	m.DetectionLatency.Observe(seconds)
}

// SetComponentStatus sets the health status of a component
func (m *Metrics) SetComponentStatus(component string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	m.ComponentStatus.WithLabelValues(component).Set(status)
}

// StartMetricsServer starts an HTTP server to expose Prometheus metrics
func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}
