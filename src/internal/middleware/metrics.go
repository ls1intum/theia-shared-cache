package middleware

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds all metrics for the cache server.
type Metrics struct {
	RequestsTotal   metric.Int64Counter
	RequestDuration metric.Float64Histogram
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("gradle-cache")

	requestsTotal, err := meter.Int64Counter(
		"gradle_cache.requests_total",
		metric.WithDescription("Total number of cache requests"))
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"gradle_cache.request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
	}, nil

}

// Middleware creates a Gin middleware that records metrics.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		attr := metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("status", status),
		)
		m.RequestsTotal.Add(c.Request.Context(), 1, attr)

		m.RequestDuration.Record(c.Request.Context(), duration,
			metric.WithAttributes(
				attribute.String("method", method)))

	}
}
