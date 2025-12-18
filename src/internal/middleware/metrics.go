package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the cache server.
type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	CacheHitsTotal   prometheus.Counter
	CacheMissesTotal prometheus.Counter
	StoredBytesTotal prometheus.Counter
	EntrySize        prometheus.Histogram
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		CacheHitsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
		),
		CacheMissesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
		),
		StoredBytesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "stored_bytes_total",
				Help:      "Total bytes stored in cache",
			},
		),
		EntrySize: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "entry_size_bytes",
				Help:      "Size of cache entries in bytes",
				Buckets:   prometheus.ExponentialBuckets(1024, 2, 20), // 1KB to 1GB
			},
		),
	}
}

// Middleware creates a Gin middleware that records metrics.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		m.RequestsTotal.WithLabelValues(method, status).Inc()
		m.RequestDuration.WithLabelValues(method).Observe(duration)

		// Record cache hit/miss for GET requests
		if method == "GET" && c.Param("key") != "" {
			if c.Writer.Status() == 200 {
				m.CacheHitsTotal.Inc()
			} else if c.Writer.Status() == 404 {
				m.CacheMissesTotal.Inc()
			}
		}

		// Record stored bytes for PUT requests
		if method == "PUT" && c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			size := c.Request.ContentLength
			if size > 0 {
				m.StoredBytesTotal.Add(float64(size))
				m.EntrySize.Observe(float64(size))
			}
		}
	}
}
