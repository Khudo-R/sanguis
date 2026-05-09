package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limiter_requests_total",
			Help: "The total number of rate limiter requests",
		},
		[]string{"result"},
	)

	RedisLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "rate_limiter_redis_latency_seconds",
			Help:    "Latency of Redis operations",
			Buckets: prometheus.DefBuckets,
		},
	)
)
