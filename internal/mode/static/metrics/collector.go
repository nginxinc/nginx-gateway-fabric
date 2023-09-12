package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ManagerCollector is an interface for the metrics of the Nginx runtime manager
type ManagerCollector interface {
	IncNginxReloadCount()
	IncNginxReloadErrors()
	UpdateLastReloadTime(ms time.Duration)
}

// ManagerMetricsCollector implements ManagerCollector interface and prometheus.Collector interface
type ManagerMetricsCollector struct {
	// Metrics
	reloadsTotal     prometheus.Counter
	reloadsError     prometheus.Counter
	lastReloadStatus prometheus.Gauge
	lastReloadTime   prometheus.Gauge
}

// NewManagerMetricsCollector creates a new ManagerMetricsCollector
func NewManagerMetricsCollector(constLabels map[string]string) *ManagerMetricsCollector {
	nc := &ManagerMetricsCollector{
		reloadsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:        "nginx_reloads_total",
				Namespace:   metricsNamespace,
				Help:        "Number of successful NGINX reloads",
				ConstLabels: constLabels,
			}),
		reloadsError: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:        "nginx_reload_errors_total",
				Namespace:   metricsNamespace,
				Help:        "Number of unsuccessful NGINX reloads",
				ConstLabels: constLabels,
			},
		),
		lastReloadStatus: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "nginx_last_reload_status",
				Namespace:   metricsNamespace,
				Help:        "Status of the last NGINX reload",
				ConstLabels: constLabels,
			},
		),
		lastReloadTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "nginx_last_reload_milliseconds",
				Namespace:   metricsNamespace,
				Help:        "Duration in milliseconds of the last NGINX reload",
				ConstLabels: constLabels,
			},
		),
	}
	return nc
}

// IncNginxReloadCount increments the counter of successful NGINX reloads and sets the last reload status to true.
func (mc *ManagerMetricsCollector) IncNginxReloadCount() {
	mc.reloadsTotal.Inc()
	mc.updateLastReloadStatus(true)
}

// IncNginxReloadErrors increments the counter of NGINX reload errors and sets the last reload status to false.
func (mc *ManagerMetricsCollector) IncNginxReloadErrors() {
	mc.reloadsError.Inc()
	mc.updateLastReloadStatus(false)
}

// updateLastReloadStatus updates the last NGINX reload status metric.
func (mc *ManagerMetricsCollector) updateLastReloadStatus(up bool) {
	var status float64
	if up {
		status = 1.0
	}
	mc.lastReloadStatus.Set(status)
}

// UpdateLastReloadTime updates the last NGINX reload time.
func (mc *ManagerMetricsCollector) UpdateLastReloadTime(duration time.Duration) {
	mc.lastReloadTime.Set(float64(duration / time.Millisecond))
}

// Describe implements prometheus.Collector interface Describe method.
func (mc *ManagerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	mc.reloadsTotal.Describe(ch)
	mc.reloadsError.Describe(ch)
	mc.lastReloadStatus.Describe(ch)
	mc.lastReloadTime.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method.
func (mc *ManagerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	mc.reloadsTotal.Collect(ch)
	mc.reloadsError.Collect(ch)
	mc.lastReloadStatus.Collect(ch)
	mc.lastReloadTime.Collect(ch)
}

// ManagerFakeCollector is a fake collector that will implement ManagerCollector interface.
// Used to initilise the ManagerCollector when metrics are disabled to avoid nil pointer errors.
type ManagerFakeCollector struct{}

// NewManagerFakeCollector creates a fake collector that implements ManagerCollector interface.
func NewManagerFakeCollector() *ManagerFakeCollector {
	return &ManagerFakeCollector{}
}

// IncNginxReloadCount implements a fake IncNginxReloadCount.
func (mc *ManagerFakeCollector) IncNginxReloadCount() {}

// IncNginxReloadErrors implements a fake IncNginxReloadErrors.
func (mc *ManagerFakeCollector) IncNginxReloadErrors() {}

// UpdateLastReloadTime implements a fake UpdateLastReloadTime.
func (mc *ManagerFakeCollector) UpdateLastReloadTime(_ time.Duration) {}
