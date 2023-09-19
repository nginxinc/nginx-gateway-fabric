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
				Buckets:     []float64{500, 1000, 5000, 10000, 30000},
			},
		),
	}
	return nc
}

// IncNginxReloadCount increments the counter of successful NGINX reloads and sets the stale config status to false.
func (mc *ManagerMetricsCollector) IncReloadCount() {
	mc.reloadsTotal.Inc()
	mc.updateConfigStaleStatus(false)
}

// IncNginxReloadErrors increments the counter of NGINX reload errors and sets the stale config status to true.
func (mc *ManagerMetricsCollector) IncReloadErrors() {
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

// ObserveLastReloadTime adds the last NGINX reload time to the histogram.
func (mc *ManagerMetricsCollector) ObserveLastReloadTime(duration time.Duration) {
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

// ManagerNoopCollector is a no-op collector that will implement ManagerCollector interface.
// Used to initialize the ManagerCollector when metrics are disabled to avoid nil pointer errors.
type ManagerNoopCollector struct{}

// NewManagerNoopCollector creates a no-op collector that implements ManagerCollector interface.
func NewManagerNoopCollector() *ManagerNoopCollector {
	return &ManagerNoopCollector{}
}

// IncReloadCount implements a no-op IncReloadCount.
func (mc *ManagerNoopCollector) IncReloadCount() {}

// IncReloadErrors implements a no-op IncReloadErrors.
func (mc *ManagerNoopCollector) IncReloadErrors() {}

// ObserveLastReloadTime implements a no-op ObserveLastReloadTime.
func (mc *ManagerNoopCollector) ObserveLastReloadTime(_ time.Duration) {}
