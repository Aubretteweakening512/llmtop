package collector

import (
	"testing"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

const sampleLMCacheMetrics = `# HELP lmcache_hit_rate Cache hit rate (0.0-1.0).
# TYPE lmcache_hit_rate gauge
lmcache_hit_rate 0.78
# HELP lmcache_store_size_bytes Total cache store size in bytes.
# TYPE lmcache_store_size_bytes gauge
lmcache_store_size_bytes 4294967296
# HELP lmcache_eviction_total Total number of cache evictions.
# TYPE lmcache_eviction_total counter
lmcache_eviction_total 1523
`

func TestParseLMCacheMetrics(t *testing.T) {
	pm := metrics.ParseText(sampleLMCacheMetrics)
	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:9090", Online: true}

	parseLMCacheMetrics(m, pm)

	if m.CacheHitRatePct != 78 {
		t.Errorf("CacheHitRatePct = %f, want 78", m.CacheHitRatePct)
	}
	if m.StoreSizeBytes != 4294967296 {
		t.Errorf("StoreSizeBytes = %f, want 4294967296", m.StoreSizeBytes)
	}
	if m.EvictionTotal != 1523 {
		t.Errorf("EvictionTotal = %f, want 1523", m.EvictionTotal)
	}
}

func TestDetectLMCache(t *testing.T) {
	pm := metrics.ParseText(sampleLMCacheMetrics)
	p := &lmcacheParser{}
	backend, model := p.Detect(pm)

	if backend != metrics.BackendLMCache {
		t.Errorf("backend = %s, want LMCache", backend)
	}
	if model != "" {
		t.Errorf("model = %q, want empty (LMCache has no model)", model)
	}
}
