package metrics

import (
	"testing"
	"time"
)

func TestComputeFleetSummary(t *testing.T) {
	workers := []*WorkerMetrics{
		{
			Online:          true,
			LastSeen:        time.Now(),
			CacheHitRatePct: 70.0,
			TTFT_P99:        200.0,
			GenTokPerSec:    100.0,
			PromptTokPerSec: 50.0,
		},
		{
			Online:          true,
			LastSeen:        time.Now(),
			CacheHitRatePct: 80.0,
			TTFT_P99:        300.0,
			GenTokPerSec:    200.0,
			PromptTokPerSec: 100.0,
		},
		{
			Online: false, // offline worker
		},
	}

	summary := ComputeFleetSummary(workers)

	if summary.TotalWorkers != 3 {
		t.Errorf("expected TotalWorkers=3, got %d", summary.TotalWorkers)
	}
	if summary.OnlineWorkers != 2 {
		t.Errorf("expected OnlineWorkers=2, got %d", summary.OnlineWorkers)
	}
	if summary.TotalTokPerSec != 450.0 {
		t.Errorf("expected TotalTokPerSec=450.0, got %f", summary.TotalTokPerSec)
	}
	if summary.AvgCacheHit != 75.0 {
		t.Errorf("expected AvgCacheHit=75.0, got %f", summary.AvgCacheHit)
	}
	if summary.P99TTFT != 300.0 {
		t.Errorf("expected P99TTFT=300.0, got %f", summary.P99TTFT)
	}
}

func TestComputeFleetSummary_Empty(t *testing.T) {
	summary := ComputeFleetSummary(nil)
	if summary.TotalWorkers != 0 {
		t.Errorf("expected 0 workers, got %d", summary.TotalWorkers)
	}
}
