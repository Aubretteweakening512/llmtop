package collector

import (
	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// parseLMCacheMetrics extracts LMCache-specific metrics from parsed Prometheus data.
func parseLMCacheMetrics(m *metrics.WorkerMetrics, pm *metrics.ParsedMetrics) {
	// Cache hit rate (0.0-1.0 → 0-100%)
	if v, _, ok := pm.GetGaugeAny("lmcache_hit_rate"); ok {
		m.CacheHitRatePct = v * 100
	}

	// Total cache store size in bytes
	if v, _, ok := pm.GetGaugeAny("lmcache_store_size_bytes"); ok {
		m.StoreSizeBytes = v
	}

	// Total evictions
	if v, _, ok := pm.GetGaugeAny("lmcache_eviction_total"); ok {
		m.EvictionTotal = v
	}
}
