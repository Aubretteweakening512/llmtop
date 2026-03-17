package collector

import (
	"testing"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

const sampleNIMMetrics = `# HELP num_requests_running Number of requests in model execution batches.
# TYPE num_requests_running gauge
num_requests_running{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 3.0
# HELP num_requests_waiting Number of requests waiting to be processed.
# TYPE num_requests_waiting gauge
num_requests_waiting{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 1.0
# HELP gpu_cache_usage_perc GPU cache usage as a percentage.
# TYPE gpu_cache_usage_perc gauge
gpu_cache_usage_perc{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 0.42
# HELP gpu_prefix_cache_hit_rate GPU prefix cache hit rate.
# TYPE gpu_prefix_cache_hit_rate gauge
gpu_prefix_cache_hit_rate{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 0.65
# HELP time_to_first_token_seconds Histogram of TTFT.
# TYPE time_to_first_token_seconds histogram
time_to_first_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.01"} 2
time_to_first_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.05"} 8
time_to_first_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.1"} 10
time_to_first_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.5"} 10
time_to_first_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="+Inf"} 10
time_to_first_token_seconds_sum{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 0.32
time_to_first_token_seconds_count{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 10
# HELP time_per_output_token_seconds Histogram of ITL.
# TYPE time_per_output_token_seconds histogram
time_per_output_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.005"} 5
time_per_output_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.01"} 15
time_per_output_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="0.05"} 20
time_per_output_token_seconds_bucket{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b",le="+Inf"} 20
time_per_output_token_seconds_sum{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 0.15
time_per_output_token_seconds_count{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 20
# HELP prompt_tokens_total Total prompt tokens.
# TYPE prompt_tokens_total counter
prompt_tokens_total{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 1500
# HELP generation_tokens_total Total gen tokens.
# TYPE generation_tokens_total counter
generation_tokens_total{model_name="deepseek-ai/deepseek-r1-distill-qwen-7b"} 4200
`

func TestParseNIMMetrics(t *testing.T) {
	pm := metrics.ParseText(sampleNIMMetrics)
	m := &metrics.WorkerMetrics{Endpoint: "http://localhost:8000", Online: true}

	parseNIMMetrics(m, nil, pm)

	if m.RequestsRunning != 3 {
		t.Errorf("RequestsRunning = %d, want 3", m.RequestsRunning)
	}
	if m.RequestsWaiting != 1 {
		t.Errorf("RequestsWaiting = %d, want 1", m.RequestsWaiting)
	}
	if m.KVCacheUsagePct != 42 {
		t.Errorf("KVCacheUsagePct = %f, want 42", m.KVCacheUsagePct)
	}
	if m.CacheHitRatePct != 65 {
		t.Errorf("CacheHitRatePct = %f, want 65", m.CacheHitRatePct)
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
	if m.StoreSizeBytes != 1500 {
		t.Errorf("StoreSizeBytes (prompt counter) = %f, want 1500", m.StoreSizeBytes)
	}
	if m.EvictionTotal != 4200 {
		t.Errorf("EvictionTotal (gen counter) = %f, want 4200", m.EvictionTotal)
	}
}

func TestDetectBackendNIM(t *testing.T) {
	pm := metrics.ParseText(sampleNIMMetrics)
	backend, model := detectBackendAndModel(pm)

	if backend != metrics.BackendNIM {
		t.Errorf("backend = %s, want NIM", backend)
	}
	if model != "deepseek-ai/deepseek-r1-distill-qwen-7b" {
		t.Errorf("model = %s, want deepseek-ai/deepseek-r1-distill-qwen-7b", model)
	}
}

func TestDetectBackendNIM_NoFalsePositive(t *testing.T) {
	partial := `# TYPE num_requests_running gauge
num_requests_running{model_name="test"} 5
`
	pm := metrics.ParseText(partial)
	backend, _ := detectBackendAndModel(pm)

	if backend == metrics.BackendNIM {
		t.Error("should not detect NIM with only one matching metric")
	}
}
