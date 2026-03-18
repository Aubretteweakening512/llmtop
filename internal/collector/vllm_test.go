package collector

import (
	"testing"
	"time"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

const sampleVLLMMetrics = `# HELP vllm:num_requests_running Number of requests currently running on GPU.
# TYPE vllm:num_requests_running gauge
vllm:num_requests_running{model_name="meta-llama/Llama-3.1-8B-Instruct"} 5.0
# HELP vllm:num_requests_waiting Number of requests waiting to be processed.
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting{model_name="meta-llama/Llama-3.1-8B-Instruct"} 2.0
# HELP vllm:gpu_cache_usage_perc GPU KV-cache usage.
# TYPE vllm:gpu_cache_usage_perc gauge
vllm:gpu_cache_usage_perc{model_name="meta-llama/Llama-3.1-8B-Instruct"} 0.73
# HELP vllm:gpu_prefix_cache_hit_rate GPU prefix cache hit rate.
# TYPE vllm:gpu_prefix_cache_hit_rate gauge
vllm:gpu_prefix_cache_hit_rate{model_name="meta-llama/Llama-3.1-8B-Instruct"} 0.45
# HELP vllm:time_to_first_token_seconds Histogram of time to first token in seconds.
# TYPE vllm:time_to_first_token_seconds histogram
vllm:time_to_first_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.01"} 5
vllm:time_to_first_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.025"} 20
vllm:time_to_first_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.05"} 40
vllm:time_to_first_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.1"} 48
vllm:time_to_first_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.5"} 50
vllm:time_to_first_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="+Inf"} 50
vllm:time_to_first_token_seconds_sum{model_name="meta-llama/Llama-3.1-8B-Instruct"} 1.25
vllm:time_to_first_token_seconds_count{model_name="meta-llama/Llama-3.1-8B-Instruct"} 50
# HELP vllm:time_per_output_token_seconds Histogram of inter-token latency in seconds.
# TYPE vllm:time_per_output_token_seconds histogram
vllm:time_per_output_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.005"} 10
vllm:time_per_output_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.01"} 30
vllm:time_per_output_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.025"} 80
vllm:time_per_output_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="0.05"} 95
vllm:time_per_output_token_seconds_bucket{model_name="meta-llama/Llama-3.1-8B-Instruct",le="+Inf"} 100
vllm:time_per_output_token_seconds_sum{model_name="meta-llama/Llama-3.1-8B-Instruct"} 1.8
vllm:time_per_output_token_seconds_count{model_name="meta-llama/Llama-3.1-8B-Instruct"} 100
# HELP vllm:prompt_tokens_total Total number of prompt tokens processed.
# TYPE vllm:prompt_tokens_total counter
vllm:prompt_tokens_total{model_name="meta-llama/Llama-3.1-8B-Instruct"} 12000
# HELP vllm:generation_tokens_total Total number of generation tokens produced.
# TYPE vllm:generation_tokens_total counter
vllm:generation_tokens_total{model_name="meta-llama/Llama-3.1-8B-Instruct"} 35000
`

func TestParseVLLMMetrics(t *testing.T) {
	pm := metrics.ParseText(sampleVLLMMetrics)
	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:8000", Online: true}

	parseVLLMMetrics(m, nil, counterState{}, pm)

	if m.RequestsRunning != 5 {
		t.Errorf("RequestsRunning = %d, want 5", m.RequestsRunning)
	}
	if m.RequestsWaiting != 2 {
		t.Errorf("RequestsWaiting = %d, want 2", m.RequestsWaiting)
	}
	if m.KVCacheUsagePct != 73 {
		t.Errorf("KVCacheUsagePct = %f, want 73", m.KVCacheUsagePct)
	}
	if m.CacheHitRatePct != 45 {
		t.Errorf("CacheHitRatePct = %f, want 45", m.CacheHitRatePct)
	}
	if m.TTFT_P50 <= 0 {
		t.Errorf("TTFT_P50 = %f, want > 0", m.TTFT_P50)
	}
	if m.TTFT_P99 <= 0 {
		t.Errorf("TTFT_P99 = %f, want > 0", m.TTFT_P99)
	}
	if m.ITL_P50 <= 0 {
		t.Errorf("ITL_P50 = %f, want > 0", m.ITL_P50)
	}
	if m.ITL_P99 <= 0 {
		t.Errorf("ITL_P99 = %f, want > 0", m.ITL_P99)
	}
}

func TestVLLMRateCalculation(t *testing.T) {
	pm := metrics.ParseText(sampleVLLMMetrics)

	prev := &metrics.WorkerMetrics{
		Online:   true,
		LastSeen: time.Now().Add(-2 * time.Second),
	}
	prevCounters := counterState{
		promptTokensTotal: 10000,
		genTokensTotal:    33000,
	}

	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:8000", Online: true}
	parseVLLMMetrics(m, prev, prevCounters, pm)

	// prompt delta = 12000 - 10000 = 2000 over ~2s = ~1000 tok/s
	if m.PromptTokPerSec < 900 || m.PromptTokPerSec > 1100 {
		t.Errorf("PromptTokPerSec = %f, want ~1000", m.PromptTokPerSec)
	}
	// gen delta = 35000 - 33000 = 2000 over ~2s = ~1000 tok/s
	if m.GenTokPerSec < 900 || m.GenTokPerSec > 1100 {
		t.Errorf("GenTokPerSec = %f, want ~1000", m.GenTokPerSec)
	}
}

func TestDetectVLLM(t *testing.T) {
	pm := metrics.ParseText(sampleVLLMMetrics)
	p := &vllmParser{}
	backend, model := p.Detect(pm)

	if backend != metrics.BackendVLLM {
		t.Errorf("backend = %s, want vLLM", backend)
	}
	if model != "meta-llama/Llama-3.1-8B-Instruct" {
		t.Errorf("model = %s, want meta-llama/Llama-3.1-8B-Instruct", model)
	}
}

func TestVLLMCounterState(t *testing.T) {
	pm := metrics.ParseText(sampleVLLMMetrics)
	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:8000", Online: true}

	counters := parseVLLMMetrics(m, nil, counterState{}, pm)

	// Counter values must NOT leak into WorkerMetrics fields
	if m.StoreSizeBytes != 0 {
		t.Errorf("StoreSizeBytes = %f, want 0 (counters should not leak into exported fields)", m.StoreSizeBytes)
	}
	if m.EvictionTotal != 0 {
		t.Errorf("EvictionTotal = %f, want 0 (counters should not leak into exported fields)", m.EvictionTotal)
	}
	// Counter values should be in the returned counterState
	if counters.promptTokensTotal != 12000 {
		t.Errorf("counters.promptTokensTotal = %f, want 12000", counters.promptTokensTotal)
	}
	if counters.genTokensTotal != 35000 {
		t.Errorf("counters.genTokensTotal = %f, want 35000", counters.genTokensTotal)
	}
}
