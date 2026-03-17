// Package collector provides concurrent polling of LLM inference worker metrics.
package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

const (
	// MaxHistory is the number of historical snapshots to retain per worker.
	MaxHistory = 60
	// DefaultTimeout is the HTTP request timeout for metrics polling.
	DefaultTimeout = 3 * time.Second
)

// Collector manages concurrent polling of multiple worker endpoints.
type Collector struct {
	mu       sync.RWMutex
	workers  map[string]*workerState
	client   *http.Client
	interval time.Duration
	cancel   context.CancelFunc
}

// workerState holds the current and historical metrics for a single worker.
type workerState struct {
	mu          sync.Mutex
	current     *metrics.WorkerMetrics
	history     []*metrics.WorkerMetrics
	prev        *metrics.WorkerMetrics // previous snapshot for rate calculation
	metricsPath string
	fetchFunc   func(ctx context.Context) (string, error)
}

// WorkerConfig describes a single worker endpoint to poll.
type WorkerConfig struct {
	Endpoint    string
	Label       string
	Backend     metrics.Backend // hint; auto-detected if Unknown
	MetricsPath string
	FetchFunc   func(ctx context.Context) (string, error) // optional: custom fetcher (e.g., K8s API proxy)
}

// New creates a new Collector with the given worker configs and polling interval.
func New(configs []WorkerConfig, interval time.Duration) *Collector {
	c := &Collector{
		workers: make(map[string]*workerState),
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		interval: interval,
	}
	for _, cfg := range configs {
		m := &metrics.WorkerMetrics{
			Endpoint: cfg.Endpoint,
			Label:    cfg.Label,
			Backend:  cfg.Backend,
			Online:   false,
		}
		mp := cfg.MetricsPath
		if mp == "" {
			mp = "/metrics"
		}
		c.workers[cfg.Endpoint] = &workerState{
			current:     m,
			metricsPath: mp,
			fetchFunc:   cfg.FetchFunc,
		}
	}
	return c
}

// Start begins polling all workers in the background.
func (c *Collector) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	go c.pollLoop(ctx)
}

// Stop halts background polling.
func (c *Collector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

// GetAll returns a snapshot of all current worker metrics.
func (c *Collector) GetAll() []*metrics.WorkerMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*metrics.WorkerMetrics, 0, len(c.workers))
	for _, ws := range c.workers {
		ws.mu.Lock()
		m := copyMetrics(ws.current)
		ws.mu.Unlock()
		result = append(result, m)
	}
	return result
}

// GetHistory returns the history for a specific endpoint.
func (c *Collector) GetHistory(endpoint string) []*metrics.WorkerMetrics {
	c.mu.RLock()
	ws, ok := c.workers[endpoint]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	hist := make([]*metrics.WorkerMetrics, len(ws.history))
	for i, m := range ws.history {
		hist[i] = copyMetrics(m)
	}
	return hist
}

// PollNow triggers an immediate poll of all workers (non-blocking).
func (c *Collector) PollNow() {
	c.mu.RLock()
	endpoints := make([]string, 0, len(c.workers))
	for ep := range c.workers {
		endpoints = append(endpoints, ep)
	}
	c.mu.RUnlock()

	var wg sync.WaitGroup
	for _, ep := range endpoints {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()
			c.pollWorker(endpoint)
		}(ep)
	}
	wg.Wait()
}

// AddWorker adds a new worker endpoint to poll.
func (c *Collector) AddWorker(cfg WorkerConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.workers[cfg.Endpoint]; exists {
		return
	}
	m := &metrics.WorkerMetrics{
		Endpoint: cfg.Endpoint,
		Label:    cfg.Label,
		Backend:  cfg.Backend,
		Online:   false,
	}
	mp := cfg.MetricsPath
	if mp == "" {
		mp = "/metrics"
	}
	c.workers[cfg.Endpoint] = &workerState{
		current:     m,
		metricsPath: mp,
		fetchFunc:   cfg.FetchFunc,
	}
}

func (c *Collector) pollLoop(ctx context.Context) {
	// Do an immediate poll on start
	c.PollNow()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.PollNow()
		}
	}
}

func (c *Collector) pollWorker(endpoint string) {
	c.mu.RLock()
	ws, ok := c.workers[endpoint]
	c.mu.RUnlock()
	if !ok {
		return
	}

	ws.mu.Lock()
	metricsPath := ws.metricsPath
	current := ws.current
	prev := ws.prev
	fetchFunc := ws.fetchFunc
	ws.mu.Unlock()

	var body string
	var err error
	if fetchFunc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
		defer cancel()
		body, err = fetchFunc(ctx)
	} else {
		url := endpoint + metricsPath
		body, err = c.fetchMetrics(url)
		if err != nil && metricsPath == "/metrics" {
			// Try /v1/metrics as fallback (NIM serves metrics there)
			body, err = c.fetchMetrics(endpoint + "/v1/metrics")
			if err == nil {
				ws.mu.Lock()
				ws.metricsPath = "/v1/metrics"
				ws.mu.Unlock()
			}
		}
	}
	if err != nil {
		// Mark offline, preserve last known data
		ws.mu.Lock()
		ws.current = &metrics.WorkerMetrics{
			Endpoint:  current.Endpoint,
			Label:     current.Label,
			Backend:   current.Backend,
			ModelName: current.ModelName,
			Online:    false,
			LastSeen:  current.LastSeen,
		}
		ws.mu.Unlock()
		return
	}

	pm := metrics.ParseText(body)
	updated := parseWorkerMetrics(current, prev, pm)

	ws.mu.Lock()
	ws.prev = copyMetrics(ws.current)
	ws.current = updated
	// Append to history
	ws.history = append(ws.history, copyMetrics(updated))
	if len(ws.history) > MaxHistory {
		ws.history = ws.history[len(ws.history)-MaxHistory:]
	}
	ws.mu.Unlock()
}

func (c *Collector) fetchMetrics(url string) (string, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body %s: %w", url, err)
	}
	return string(body), nil
}

// parseWorkerMetrics extracts WorkerMetrics from parsed prometheus data.
func parseWorkerMetrics(current, prev *metrics.WorkerMetrics, pm *metrics.ParsedMetrics) *metrics.WorkerMetrics {
	m := &metrics.WorkerMetrics{
		Endpoint: current.Endpoint,
		Label:    current.Label,
		Backend:  current.Backend,
		Online:   true,
		LastSeen: time.Now(),
	}

	// Detect backend and model from metrics
	detected, model := detectBackendAndModel(pm)
	if m.Backend == metrics.BackendUnknown || m.Backend == "" {
		m.Backend = detected
	}
	m.ModelName = model

	switch m.Backend {
	case metrics.BackendVLLM:
		parseVLLMMetrics(m, prev, pm)
	case metrics.BackendSGLang:
		parseSGLangMetrics(m, prev, pm)
	case metrics.BackendLMCache:
		parseLMCacheMetrics(m, pm)
	case metrics.BackendNIM:
		parseNIMMetrics(m, prev, pm)
	default:
		// Try all
		switch detected {
		case metrics.BackendVLLM:
			parseVLLMMetrics(m, prev, pm)
		case metrics.BackendSGLang:
			parseSGLangMetrics(m, prev, pm)
		case metrics.BackendLMCache:
			parseLMCacheMetrics(m, pm)
		case metrics.BackendNIM:
			parseNIMMetrics(m, prev, pm)
		}
	}

	// Carry over TTFT history
	if prev != nil {
		m.TTFTHistory = append(prev.TTFTHistory, m.TTFT_P99)
		if len(m.TTFTHistory) > MaxHistory {
			m.TTFTHistory = m.TTFTHistory[len(m.TTFTHistory)-MaxHistory:]
		}
		m.GenTokHistory = append(prev.GenTokHistory, m.GenTokPerSec)
		if len(m.GenTokHistory) > MaxHistory {
			m.GenTokHistory = m.GenTokHistory[len(m.GenTokHistory)-MaxHistory:]
		}
	} else {
		m.TTFTHistory = []float64{m.TTFT_P99}
		m.GenTokHistory = []float64{m.GenTokPerSec}
	}

	return m
}

// detectBackendAndModel tries to identify the backend type and model name from metrics.
func detectBackendAndModel(pm *metrics.ParsedMetrics) (metrics.Backend, string) {
	// Phase 1: check for prefixed backends (existing logic)
	for _, s := range pm.Samples {
		if len(s.Name) >= 5 && s.Name[:5] == "vllm:" {
			return metrics.BackendVLLM, s.Labels["model_name"]
		}
		if len(s.Name) >= 8 && s.Name[:8] == "sglang:n" {
			return metrics.BackendSGLang, s.Labels["model_name"]
		}
		if len(s.Name) >= 7 && s.Name[:7] == "sglang:" {
			return metrics.BackendSGLang, s.Labels["model_name"]
		}
		if len(s.Name) >= 8 && s.Name[:8] == "lmcache_" {
			return metrics.BackendLMCache, ""
		}
	}
	for _, h := range pm.Histograms {
		if len(h.Name) >= 5 && h.Name[:5] == "vllm:" {
			return metrics.BackendVLLM, h.Labels["model_name"]
		}
		if len(h.Name) >= 7 && h.Name[:7] == "sglang:" {
			return metrics.BackendSGLang, h.Labels["model_name"]
		}
	}

	// Phase 2: check for unprefixed NIM metrics (multi-metric conjunction)
	hasRunning := false
	hasCachePerc := false
	hasTTFT := false
	var model string
	for _, s := range pm.Samples {
		switch s.Name {
		case "num_requests_running":
			hasRunning = true
			if m, ok := s.Labels["model_name"]; ok && m != "" {
				model = m
			}
		case "gpu_cache_usage_perc":
			hasCachePerc = true
		}
	}
	for _, h := range pm.Histograms {
		if h.Name == "time_to_first_token_seconds" {
			hasTTFT = true
			if m, ok := h.Labels["model_name"]; ok && m != "" {
				model = m
			}
		}
	}
	if hasRunning && hasCachePerc && hasTTFT {
		return metrics.BackendNIM, model
	}

	return metrics.BackendUnknown, ""
}

func copyMetrics(m *metrics.WorkerMetrics) *metrics.WorkerMetrics {
	if m == nil {
		return nil
	}
	cp := *m
	if m.TTFTHistory != nil {
		cp.TTFTHistory = make([]float64, len(m.TTFTHistory))
		copy(cp.TTFTHistory, m.TTFTHistory)
	}
	if m.GenTokHistory != nil {
		cp.GenTokHistory = make([]float64, len(m.GenTokHistory))
		copy(cp.GenTokHistory, m.GenTokHistory)
	}
	return &cp
}
