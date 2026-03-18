package collector

import (
	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// BackendParser extracts backend-specific metrics from parsed Prometheus data.
// Implementations are registered in the parsers map and dispatched by backend type.
type BackendParser interface {
	// Detect inspects parsed metrics and returns the detected backend and model name.
	// Returns (BackendUnknown, "") if this parser does not recognize the metrics.
	Detect(pm *metrics.ParsedMetrics) (metrics.Backend, string)

	// Parse extracts backend-specific fields into the WorkerMetrics struct.
	// prevCounters carries raw counter values from the previous poll for rate computation.
	// Returns updated counter state.
	Parse(m *metrics.WorkerMetrics, prev *metrics.WorkerMetrics, prevCounters counterState, pm *metrics.ParsedMetrics) counterState
}

// parsers maps backend types to their parser implementation.
var parsers = map[metrics.Backend]BackendParser{}

// detectors is the ordered list of parsers to try during backend detection.
// Order matters: prefixed backends (vLLM, SGLang, LMCache) are checked
// before unprefixed ones (NIM) to avoid false positives.
var detectors []BackendParser

// RegisterParser registers a BackendParser for a backend type.
// Called from init() in each backend file.
func RegisterParser(backend metrics.Backend, parser BackendParser) {
	parsers[backend] = parser
}
