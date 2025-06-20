package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_hits_total",
		Help: "The total number of hits",
	}, []string{"endpoint", "status"})

	HitTimings = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_hit_timings_seconds",
		Help: "The latency of the hits",
	}, []string{"endpoint"})

	ExternalServiceStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "external_service_status_total",
		Help: "The status of external service responses",
	}, []string{"service", "status"})

	ExternalServiceTimings = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "external_service_timings_seconds",
		Help: "The latency of external service responses",
	}, []string{"service"})
)
