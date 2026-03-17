package collector

import (
	"time"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// parseVLLMMetrics extracts vLLM-specific metrics from the parsed Prometheus data.
func parseVLLMMetrics(m *metrics.WorkerMetrics, prev *metrics.WorkerMetrics, pm *metrics.ParsedMetrics) {
	// Running requests
	if v, _, ok := pm.GetGaugeAny("vllm:num_requests_running"); ok {
		m.RequestsRunning = int(v)
	}

	// Waiting requests (queue depth)
	if v, _, ok := pm.GetGaugeAny("vllm:num_requests_waiting"); ok {
		m.RequestsWaiting = int(v)
	}

	// GPU KV cache usage (0.0-1.0 → convert to 0-100%)
	// Newer vLLM versions (and Dynamo) use vllm:kv_cache_usage_perc instead of vllm:gpu_cache_usage_perc
	if v, _, ok := pm.GetGaugeAny("vllm:gpu_cache_usage_perc"); ok {
		m.KVCacheUsagePct = v * 100
	} else if v, _, ok := pm.GetGaugeAny("vllm:kv_cache_usage_perc"); ok {
		m.KVCacheUsagePct = v * 100
	}

	// Prefix cache hit rate (0.0-1.0 → convert to 0-100%)
	// Newer vLLM exports counters (prefix_cache_hits_total / prefix_cache_queries_total) instead of a gauge
	if v, _, ok := pm.GetGaugeAny("vllm:gpu_prefix_cache_hit_rate"); ok {
		m.CacheHitRatePct = v * 100
	} else {
		hits, _, hOk := pm.GetGaugeAny("vllm:prefix_cache_hits_total")
		queries, _, qOk := pm.GetGaugeAny("vllm:prefix_cache_queries_total")
		if hOk && qOk && queries > 0 {
			m.CacheHitRatePct = (hits / queries) * 100
		}
	}

	// Time to first token histogram (in seconds → convert to ms)
	if p50, ok := pm.GetHistogramQuantileAny("vllm:time_to_first_token_seconds", 0.50); ok {
		m.TTFT_P50 = p50 * 1000
	}
	if p99, ok := pm.GetHistogramQuantileAny("vllm:time_to_first_token_seconds", 0.99); ok {
		m.TTFT_P99 = p99 * 1000
	}

	// Inter-token latency histogram (seconds → ms)
	// Newer vLLM uses vllm:inter_token_latency_seconds or vllm:request_time_per_output_token_seconds
	if p50, ok := pm.GetHistogramQuantileAny("vllm:time_per_output_token_seconds", 0.50); ok {
		m.ITL_P50 = p50 * 1000
	} else if p50, ok := pm.GetHistogramQuantileAny("vllm:inter_token_latency_seconds", 0.50); ok {
		m.ITL_P50 = p50 * 1000
	}
	if p99, ok := pm.GetHistogramQuantileAny("vllm:time_per_output_token_seconds", 0.99); ok {
		m.ITL_P99 = p99 * 1000
	} else if p99, ok := pm.GetHistogramQuantileAny("vllm:inter_token_latency_seconds", 0.99); ok {
		m.ITL_P99 = p99 * 1000
	}

	// Token throughput: compute rate from counters
	// We use prev snapshot to compute delta/time
	if prev != nil && prev.Online {
		dt := time.Since(prev.LastSeen).Seconds()
		if dt > 0 {
			if v, _, ok := pm.GetGaugeAny("vllm:prompt_tokens_total"); ok {
				if prev.PromptTokPerSec > 0 || true {
					// prev throughput is stored, recalculate if we have the raw counter
					m.PromptTokPerSec = (v - prevPromptTokensVLLM(prev)) / dt
					if m.PromptTokPerSec < 0 {
						m.PromptTokPerSec = 0
					}
				}
			}
			if v, _, ok := pm.GetGaugeAny("vllm:generation_tokens_total"); ok {
				m.GenTokPerSec = (v - prevGenTokensVLLM(prev)) / dt
				if m.GenTokPerSec < 0 {
					m.GenTokPerSec = 0
				}
			}
		}
	}

	// Store raw counters in a way we can compute rates next time
	// We repurpose EvictionTotal/StoreSizeBytes fields for counter storage
	if v, _, ok := pm.GetGaugeAny("vllm:prompt_tokens_total"); ok {
		m.StoreSizeBytes = v // reuse for prompt counter
	}
	if v, _, ok := pm.GetGaugeAny("vllm:generation_tokens_total"); ok {
		m.EvictionTotal = v // reuse for gen counter
	}
}

func prevPromptTokensVLLM(prev *metrics.WorkerMetrics) float64 {
	return prev.StoreSizeBytes
}

func prevGenTokensVLLM(prev *metrics.WorkerMetrics) float64 {
	return prev.EvictionTotal
}
