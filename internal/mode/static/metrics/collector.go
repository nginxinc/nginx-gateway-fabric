package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ManagerMetricsCollector implements ManagerCollector interface and prometheus.Collector interface
type ManagerMetricsCollector struct {
	// Metrics
	reloadsTotal    prometheus.Counter
	reloadsError    prometheus.Counter
	configStale     prometheus.Gauge
	reloadsDuration prometheus.Histogram
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
		configStale: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "nginx_stale_config",
				Namespace:   metricsNamespace,
				Help:        "Indicates if NGINX is not serving the latest configuration.",
				ConstLabels: constLabels,
			},
		),
		reloadsDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:        "nginx_reloads_milliseconds",
				Namespace:   metricsNamespace,
				Help:        "Duration in milliseconds of NGINX reloads",
				ConstLabels: constLabels,
				Buckets:     []float64{100.0, 200.0, 300.0, 400.0, 500.0},
			},
		),
	}
	return nc
}

// IncNginxReloadCount increments the counter of successful NGINX reloads and sets the stale config status to false.
func (mc *ManagerMetricsCollector) IncNginxReloadCount() {
	mc.reloadsTotal.Inc()
	mc.updateConfigStaleStatus(false)
}

// IncNginxReloadErrors increments the counter of NGINX reload errors and sets the stale config status to true.
func (mc *ManagerMetricsCollector) IncNginxReloadErrors() {
	mc.reloadsError.Inc()
	mc.updateConfigStaleStatus(true)
}

// updateConfigStaleStatus updates the last NGINX reload status metric.
func (mc *ManagerMetricsCollector) updateConfigStaleStatus(stale bool) {
	var status float64
	if stale {
		status = 1.0
	}
	mc.configStale.Set(status)
}

// UpdateLastReloadTime updates the last NGINX reload time.
func (mc *ManagerMetricsCollector) UpdateLastReloadTime(duration time.Duration) {
	mc.reloadsDuration.Observe(float64(duration / time.Millisecond))
}

// Describe implements prometheus.Collector interface Describe method.
func (mc *ManagerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	mc.reloadsTotal.Describe(ch)
	mc.reloadsError.Describe(ch)
	mc.configStale.Describe(ch)
	mc.reloadsDuration.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method.
func (mc *ManagerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	mc.reloadsTotal.Collect(ch)
	mc.reloadsError.Collect(ch)
	mc.configStale.Collect(ch)
	mc.reloadsDuration.Collect(ch)
}

// ManagerFakeCollector is a fake collector that will implement ManagerCollector interface.
// Used to initialize the ManagerCollector when metrics are disabled to avoid nil pointer errors.
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
