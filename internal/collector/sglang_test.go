package collector

import (
	"math"
	"testing"
	"time"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

const sampleSGLangMetrics = `# HELP sglang:num_running_reqs Number of requests currently running.
# TYPE sglang:num_running_reqs gauge
sglang:num_running_reqs{model_name="Qwen/Qwen2.5-72B-Instruct"} 8.0
# HELP sglang:num_waiting_reqs Number of requests waiting in the queue.
# TYPE sglang:num_waiting_reqs gauge
sglang:num_waiting_reqs{model_name="Qwen/Qwen2.5-72B-Instruct"} 3.0
# HELP sglang:token_usage Token usage (KV cache utilization).
# TYPE sglang:token_usage gauge
sglang:token_usage{model_name="Qwen/Qwen2.5-72B-Instruct"} 0.58
# HELP sglang:cache_hit_rate Prefix cache hit rate.
# TYPE sglang:cache_hit_rate gauge
sglang:cache_hit_rate{model_name="Qwen/Qwen2.5-72B-Instruct"} 0.32
# HELP sglang:time_to_first_token_seconds Histogram of time to first token.
# TYPE sglang:time_to_first_token_seconds histogram
sglang:time_to_first_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.01"} 3
sglang:time_to_first_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.05"} 12
sglang:time_to_first_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.1"} 18
sglang:time_to_first_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.5"} 20
sglang:time_to_first_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="+Inf"} 20
sglang:time_to_first_token_seconds_sum{model_name="Qwen/Qwen2.5-72B-Instruct"} 0.95
sglang:time_to_first_token_seconds_count{model_name="Qwen/Qwen2.5-72B-Instruct"} 20
# HELP sglang:time_per_output_token_seconds Histogram of inter-token latency.
# TYPE sglang:time_per_output_token_seconds histogram
sglang:time_per_output_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.005"} 8
sglang:time_per_output_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.01"} 25
sglang:time_per_output_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="0.05"} 38
sglang:time_per_output_token_seconds_bucket{model_name="Qwen/Qwen2.5-72B-Instruct",le="+Inf"} 40
sglang:time_per_output_token_seconds_sum{model_name="Qwen/Qwen2.5-72B-Instruct"} 0.52
sglang:time_per_output_token_seconds_count{model_name="Qwen/Qwen2.5-72B-Instruct"} 40
# HELP sglang:prompt_tokens_total Total prompt tokens processed.
# TYPE sglang:prompt_tokens_total counter
sglang:prompt_tokens_total{model_name="Qwen/Qwen2.5-72B-Instruct"} 8500
# HELP sglang:generation_tokens_total Total generation tokens produced.
# TYPE sglang:generation_tokens_total counter
sglang:generation_tokens_total{model_name="Qwen/Qwen2.5-72B-Instruct"} 22000
`

func TestParseSGLangMetrics(t *testing.T) {
	pm := metrics.ParseText(sampleSGLangMetrics)
	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:30000", Online: true}

	parseSGLangMetrics(m, nil, counterState{}, pm)

	if m.RequestsRunning != 8 {
		t.Errorf("RequestsRunning = %d, want 8", m.RequestsRunning)
	}
	if m.RequestsWaiting != 3 {
		t.Errorf("RequestsWaiting = %d, want 3", m.RequestsWaiting)
	}
	if !approxEqual(m.KVCacheUsagePct, 58, 0.01) {
		t.Errorf("KVCacheUsagePct = %f, want 58", m.KVCacheUsagePct)
	}
	if !approxEqual(m.CacheHitRatePct, 32, 0.01) {
		t.Errorf("CacheHitRatePct = %f, want 32", m.CacheHitRatePct)
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

func TestDetectSGLang(t *testing.T) {
	pm := metrics.ParseText(sampleSGLangMetrics)
	p := &sglangParser{}
	backend, model := p.Detect(pm)

	if backend != metrics.BackendSGLang {
		t.Errorf("backend = %s, want SGLang", backend)
	}
	if model != "Qwen/Qwen2.5-72B-Instruct" {
		t.Errorf("model = %s, want Qwen/Qwen2.5-72B-Instruct", model)
	}
}

func TestSGLangCounterState(t *testing.T) {
	pm := metrics.ParseText(sampleSGLangMetrics)
	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:30000", Online: true}

	counters := parseSGLangMetrics(m, nil, counterState{}, pm)

	// Counter values must NOT leak into WorkerMetrics fields
	if m.StoreSizeBytes != 0 {
		t.Errorf("StoreSizeBytes = %f, want 0 (counters should not leak)", m.StoreSizeBytes)
	}
	if m.EvictionTotal != 0 {
		t.Errorf("EvictionTotal = %f, want 0 (counters should not leak)", m.EvictionTotal)
	}
	// Counter values should be in the returned counterState
	if counters.promptTokensTotal != 8500 {
		t.Errorf("counters.promptTokensTotal = %f, want 8500", counters.promptTokensTotal)
	}
	if counters.genTokensTotal != 22000 {
		t.Errorf("counters.genTokensTotal = %f, want 22000", counters.genTokensTotal)
	}
}

func TestSGLangRateCalculation(t *testing.T) {
	pm := metrics.ParseText(sampleSGLangMetrics)

	prev := &metrics.WorkerMetrics{
		Online:   true,
		LastSeen: time.Now().Add(-2 * time.Second),
	}
	prevCounters := counterState{
		promptTokensTotal: 7500,
		genTokensTotal:    20000,
	}

	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:30000", Online: true}
	parseSGLangMetrics(m, prev, prevCounters, pm)

	// prompt delta = 8500 - 7500 = 1000 over ~2s = ~500 tok/s
	if m.PromptTokPerSec < 450 || m.PromptTokPerSec > 550 {
		t.Errorf("PromptTokPerSec = %f, want ~500", m.PromptTokPerSec)
	}
	// gen delta = 22000 - 20000 = 2000 over ~2s = ~1000 tok/s
	if m.GenTokPerSec < 900 || m.GenTokPerSec > 1100 {
		t.Errorf("GenTokPerSec = %f, want ~1000", m.GenTokPerSec)
	}
}
