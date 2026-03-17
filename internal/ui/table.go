package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// Fixed column widths for numeric/compact columns
const (
	colBackend = 8
	colKV      = 6
	colQueue   = 6
	colRun     = 5
	colTTFT    = 10
	colITL     = 10
	colHit     = 6
	colTok     = 7
)

// fixedWidth is the sum of all fixed columns + inter-column spaces + left margin
var fixedWidth = colBackend + colKV + colQueue + colRun + colTTFT + colITL + colHit + colTok + 12 // 12 = spaces + dot + margins

// flexWidths computes dynamic ENDPOINT and MODEL column widths from terminal width.
func flexWidths(termWidth int) (epW, modelW int) {
	avail := termWidth - fixedWidth
	if avail < 30 {
		avail = 30
	}
	epW = avail * 55 / 100
	modelW = avail - epW
	if epW < 15 {
		epW = 15
	}
	if modelW < 10 {
		modelW = 10
	}
	return epW, modelW
}

// SortColumn represents a column that can be sorted.
type SortColumn int

const (
	SortNone    SortColumn = iota
	SortKVCache
	SortQueue
	SortTTFT
	SortHitRate
	SortTokPerSec
)

// SortColumnName returns a human-readable name for the sort column.
func SortColumnName(s SortColumn) string {
	switch s {
	case SortKVCache:
		return "KV%"
	case SortQueue:
		return "Queue"
	case SortTTFT:
		return "TTFT P99"
	case SortHitRate:
		return "Hit%"
	case SortTokPerSec:
		return "Tok/s"
	default:
		return "—"
	}
}

// RenderTable renders the worker metrics table.
func RenderTable(workers []*metrics.WorkerMetrics, selectedIdx int, sortCol SortColumn, filterBackend metrics.Backend, width int) string {
	var sb strings.Builder
	epW, modelW := flexWidths(width)

	header := renderTableHeader(sortCol, epW, modelW)
	sb.WriteString(header)
	sb.WriteString("\n")

	sep := StyleTableSeparator.Render(strings.Repeat("─", max(width-2, 80)))
	sb.WriteString("  " + sep)
	sb.WriteString("\n")

	for i, w := range workers {
		if filterBackend != "" && w.Backend != filterBackend && filterBackend != metrics.BackendUnknown {
			continue
		}
		row := renderWorkerRow(w, i == selectedIdx, epW, modelW)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderTableHeader(sortCol SortColumn, epW, modelW int) string {
	cols := []struct {
		name  string
		width int
		col   SortColumn
	}{
		{"ENDPOINT", epW, SortNone},
		{"BACKEND", colBackend, SortNone},
		{"MODEL", modelW, SortNone},
		{"KV%", colKV, SortKVCache},
		{"QUEUE", colQueue, SortQueue},
		{"RUN", colRun, SortNone},
		{"TTFT P99", colTTFT, SortTTFT},
		{"ITL P99", colITL, SortNone},
		{"HIT%", colHit, SortHitRate},
		{"TOK/S", colTok, SortTokPerSec},
	}

	var parts []string
	for _, c := range cols {
		text := padRight(c.name, c.width)
		if c.col != SortNone && c.col == sortCol {
			parts = append(parts, StyleSortIndicator.Render(text))
		} else {
			parts = append(parts, StyleTableHeader.Render(text))
		}
	}
	return "  " + strings.Join(parts, " ")
}

func renderWorkerRow(w *metrics.WorkerMetrics, selected bool, epW, modelW int) string {
	var dot string
	if w.Online {
		dot = StyleDotOnline.Render("●")
	} else {
		dot = StyleDotOffline.Render("○")
	}

	// Endpoint display: K8s workers show pod name, others show URL
	var ep string
	if strings.HasPrefix(w.Endpoint, "k8s://") {
		ep = w.Label
		if ep == "" {
			ep = strings.TrimPrefix(w.Endpoint, "k8s://")
		}
	} else {
		ep = strings.TrimPrefix(w.Endpoint, "http://")
		ep = strings.TrimPrefix(ep, "https://")
		if w.Label != "" {
			ep = ep + " (" + w.Label + ")"
		}
	}
	epStr := padRight(truncate(ep, epW-2), epW-2)

	backendStr := renderBackend(w.Backend, colBackend)

	model := w.ModelName
	if model == "" {
		model = "—"
	}
	modelStr := padRight(truncate(model, modelW), modelW)

	if !w.Online {
		row := "  " + dot + " " + epStr + " " + backendStr + " " + modelStr + " " +
			padRight("—", colKV) + " " +
			padRight("—", colQueue) + " " +
			padRight("—", colRun) + " " +
			padRight("—", colTTFT) + " " +
			padRight("—", colITL) + " " +
			padRight("—", colHit) + " " +
			padRight("—", colTok)
		if selected {
			return StyleTableRowSelected.Render(row)
		}
		return StyleTableRowOffline.Render(row)
	}

	kvStr := renderKVCache(w.KVCacheUsagePct, colKV)
	queueStr := renderQueue(w.RequestsWaiting, colQueue)
	runStr := padRight(fmt.Sprintf("%d", w.RequestsRunning), colRun)
	ttftStr := renderTTFT(w.TTFT_P99, colTTFT)
	itlStr := renderTTFT(w.ITL_P99, colITL)
	hitStr := renderHitRate(w.CacheHitRatePct, colHit)
	tokStr := renderTokPerSec(w.GenTokPerSec+w.PromptTokPerSec, colTok)

	if selected {
		plain := "  " + dot + " " + stripStyle(epStr) + " " + stripStyle(backendStr) + " " +
			stripStyle(modelStr) + " " + stripStyle(kvStr) + " " + stripStyle(queueStr) + " " +
			stripStyle(runStr) + " " + stripStyle(ttftStr) + " " + stripStyle(itlStr) + " " +
			stripStyle(hitStr) + " " + stripStyle(tokStr)
		return StyleTableRowSelected.Render(plain)
	}

	return "  " + dot + " " + epStr + " " + backendStr + " " + modelStr + " " +
		kvStr + " " + queueStr + " " + runStr + " " + ttftStr + " " + itlStr + " " + hitStr + " " + tokStr
}

func renderBackend(b metrics.Backend, width int) string {
	var style lipgloss.Style
	switch b {
	case metrics.BackendVLLM:
		style = StyleBadgeVLLM
	case metrics.BackendSGLang:
		style = StyleBadgeSGLang
	case metrics.BackendLMCache:
		style = StyleBadgeLMCache
	case metrics.BackendNIM:
		style = StyleBadgeNIM
	default:
		style = StyleBadgeUnknown
	}
	return style.Render(padRight(string(b), width))
}

func renderKVCache(pct float64, width int) string {
	if pct == 0 {
		return StyleMetricNA.Render(padRight("—", width))
	}
	s := fmt.Sprintf("%.0f%%", pct)
	return KVCacheStyle(pct).Render(padRight(s, width))
}

func renderQueue(n int, width int) string {
	if n == 0 {
		return StyleMetricGood.Render(padRight("0", width))
	}
	s := fmt.Sprintf("%d", n)
	return QueueStyle(n).Render(padRight(s, width))
}

func renderTTFT(ms float64, width int) string {
	if ms == 0 {
		return StyleMetricNA.Render(padRight("—", width))
	}
	s := fmt.Sprintf("%.0fms", ms)
	return TTFTStyle(ms).Render(padRight(s, width))
}

func renderHitRate(pct float64, width int) string {
	if pct == 0 {
		return StyleMetricNA.Render(padRight("—", width))
	}
	s := fmt.Sprintf("%.0f%%", pct)
	return StyleMetricGood.Render(padRight(s, width))
}

func renderTokPerSec(tok float64, width int) string {
	if tok == 0 {
		return StyleMetricNA.Render(padRight("—", width))
	}
	s := fmt.Sprintf("%.0f", tok)
	return StyleMetricGood.Render(padRight(s, width))
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n > 3 {
		return s[:n-3] + "..."
	}
	return s[:n]
}

func stripStyle(s string) string {
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
