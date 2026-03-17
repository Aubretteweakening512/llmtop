package collector

import (
	"time"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// parseNIMMetrics extracts NIM-specific metrics from parsed Prometheus data.
// NIM exports the same vLLM metrics but without the "vllm:" prefix.
// Counter metrics (prompt_tokens_total, generation_tokens_total) are stored in
// the Prometheus parser's Samples slice despite being typed as counters — the
// parser does not distinguish gauge vs counter storage. This matches vllm.go behavior.
func parseNIMMetrics(m *metrics.WorkerMetrics, prev *metrics.WorkerMetrics, pm *metrics.ParsedMetrics) {
	if v, _, ok := pm.GetGaugeAny("num_requests_running"); ok {
		m.RequestsRunning = int(v)
	}

	if v, _, ok := pm.GetGaugeAny("num_requests_waiting"); ok {
		m.RequestsWaiting = int(v)
	}

	// GPU KV cache usage (0.0-1.0 -> 0-100%)
	if v, _, ok := pm.GetGaugeAny("gpu_cache_usage_perc"); ok {
		m.KVCacheUsagePct = v * 100
	}

	// Prefix cache hit rate (0.0-1.0 -> 0-100%)
	if v, _, ok := pm.GetGaugeAny("gpu_prefix_cache_hit_rate"); ok {
		m.CacheHitRatePct = v * 100
	}

	// TTFT histogram (seconds -> ms)
	if p50, ok := pm.GetHistogramQuantileAny("time_to_first_token_seconds", 0.50); ok {
		m.TTFT_P50 = p50 * 1000
	}
	if p99, ok := pm.GetHistogramQuantileAny("time_to_first_token_seconds", 0.99); ok {
		m.TTFT_P99 = p99 * 1000
	}

	// ITL histogram (seconds -> ms)
	if p50, ok := pm.GetHistogramQuantileAny("time_per_output_token_seconds", 0.50); ok {
		m.ITL_P50 = p50 * 1000
	}
	if p99, ok := pm.GetHistogramQuantileAny("time_per_output_token_seconds", 0.99); ok {
		m.ITL_P99 = p99 * 1000
	}

	// Token throughput: counter-delta rate computation.
	// Reuses StoreSizeBytes for prompt counter, EvictionTotal for gen counter (same pattern as vllm.go).
	if prev != nil && prev.Online {
		dt := time.Since(prev.LastSeen).Seconds()
		if dt > 0 {
			if v, _, ok := pm.GetGaugeAny("prompt_tokens_total"); ok {
				m.PromptTokPerSec = (v - prev.StoreSizeBytes) / dt
				if m.PromptTokPerSec < 0 {
					m.PromptTokPerSec = 0
				}
			}
			if v, _, ok := pm.GetGaugeAny("generation_tokens_total"); ok {
				m.GenTokPerSec = (v - prev.EvictionTotal) / dt
				if m.GenTokPerSec < 0 {
					m.GenTokPerSec = 0
				}
			}
		}
	}

	// Store raw counters for next delta.
	if v, _, ok := pm.GetGaugeAny("prompt_tokens_total"); ok {
		m.StoreSizeBytes = v
	}
	if v, _, ok := pm.GetGaugeAny("generation_tokens_total"); ok {
		m.EvictionTotal = v
	}
}
