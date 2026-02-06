package handler

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	CacheHits   metric.Int64Counter
	CacheMisses metric.Int64Counter
	EntrySize   metric.Float64Histogram
}

func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("gradle-cache")

	cacheHits, err := meter.Int64Counter(
		"gradle_cache.cache_hits",
		metric.WithDescription("Total number of cache hits"))
	if err != nil {
		return nil, err
	}

	cacheMisses, err := meter.Int64Counter(
		"gradle_cache.cache_misses",
		metric.WithDescription("Total number of cache misses"))
	if err != nil {
		return nil, err
	}

	entrySize, err := meter.Float64Histogram(
		"gradle_cache.entry_size",
		metric.WithDescription("Current size of items in cache"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		CacheHits:   cacheHits,
		CacheMisses: cacheMisses,
		EntrySize:   entrySize,
	}, nil
}
