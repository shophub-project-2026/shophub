package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shophub_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "shophub_http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	HTTPResponseBytesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shophub_http_response_bytes_total",
			Help: "Total bytes sent in HTTP responses.",
		},
	)
)

var (
	ShopsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shophub_shops_total",
			Help: "Number of shops per user.",
		},
		[]string{"user_id"},
	)

	K8sOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shophub_k8s_operations_total",
			Help: "Total Kubernetes API operations performed by ShopHub.",
		},
		[]string{"operation", "result"},
	)
)
