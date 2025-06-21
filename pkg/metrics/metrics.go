package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	KeyUsageTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "key_usage_total",
			Help: "Number of times a virtual key was used.",
		},
		[]string{"key"},
	)
)

// Register registers the metrics with the provided Prometheus registerer.
func Register(r prometheus.Registerer) {
	r.MustRegister(RequestTotal, RequestDuration, KeyUsageTotal)
}
