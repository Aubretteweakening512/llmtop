package collector

import (
	"context"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// MetricsSource provides read access to worker metrics.
// The UI depends on this interface, not on *Collector directly.
type MetricsSource interface {
	Start(ctx context.Context)
	Stop()
	GetAll() []*metrics.WorkerMetrics
	GetHistory(endpoint string) []*metrics.WorkerMetrics
	PollNow(ctx context.Context)
	AddWorker(cfg WorkerConfig)
	RemoveWorker(endpoint string)
	Endpoints() map[string]struct{}
}

// GPUSource provides read access to GPU metrics.
// The UI depends on this interface, not on *DCGMCollector directly.
type GPUSource interface {
	Start(ctx context.Context)
	Stop()
	GetAll() []*metrics.GPUInfo
	GetSummary() metrics.GPUSummary
	PollNow(ctx context.Context)
}
