package metrics

import "sort"

// ModelGroup aggregates worker stats for a single model.
type ModelGroup struct {
	ModelName      string
	Backend        Backend
	Workers        int
	OnlineWorkers  int
	TotalTokPerSec float64
	AvgKVCachePct  float64
	TotalQueue     int
	TotalRunning   int
	AvgTTFTP99     float64
	AvgHitRate     float64
}

// GroupWorkersByModel aggregates worker metrics by model name.
func GroupWorkersByModel(workers []*WorkerMetrics) []ModelGroup {
	groups := make(map[string]*modelAccum)

	for _, w := range workers {
		name := w.ModelName
		if name == "" {
			name = "Unknown"
		}

		g, ok := groups[name]
		if !ok {
			g = &modelAccum{name: name, backend: w.Backend}
			groups[name] = g
		}
		g.total++
		if w.Online {
			g.online++
			g.tokPerSec += w.PromptTokPerSec + w.GenTokPerSec
			g.kvSum += w.KVCacheUsagePct
			g.queue += w.RequestsWaiting
			g.running += w.RequestsRunning
			g.ttftSum += w.TTFT_P99
			g.hitSum += w.CacheHitRatePct
			g.onlineCount++
		}
	}

	result := make([]ModelGroup, 0, len(groups))
	for _, g := range groups {
		mg := ModelGroup{
			ModelName:      g.name,
			Backend:        g.backend,
			Workers:        g.total,
			OnlineWorkers:  g.online,
			TotalTokPerSec: g.tokPerSec,
			TotalQueue:     g.queue,
			TotalRunning:   g.running,
		}
		if g.onlineCount > 0 {
			mg.AvgKVCachePct = g.kvSum / float64(g.onlineCount)
			mg.AvgTTFTP99 = g.ttftSum / float64(g.onlineCount)
			mg.AvgHitRate = g.hitSum / float64(g.onlineCount)
		}
		result = append(result, mg)
	}

	// Sort by model name for stable output.
	sort.Slice(result, func(i, j int) bool {
		return result[i].ModelName < result[j].ModelName
	})

	return result
}

type modelAccum struct {
	name        string
	backend     Backend
	total       int
	online      int
	onlineCount int
	tokPerSec   float64
	kvSum       float64
	queue       int
	running     int
	ttftSum     float64
	hitSum      float64
}
