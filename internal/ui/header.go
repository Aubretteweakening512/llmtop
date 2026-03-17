package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/InfraWhisperer/llmtop/internal/metrics"
)

// RenderHeader renders the fleet summary header bar.
func RenderHeader(summary metrics.FleetSummary, version string, intervalSec int, width int) string {
	title := StyleHeaderTitle.Render("llmtop")
	if version != "" {
		title = StyleHeaderTitle.Render("llmtop " + version)
	}

	workers := StyleHeaderStat.Render(
		fmt.Sprintf("%d workers (%d online)",
			summary.TotalWorkers, summary.OnlineWorkers),
	)

	tokPerSec := StyleHeaderValue.Render(fmt.Sprintf("%.0f tok/s", summary.TotalTokPerSec))

	cacheHit := StyleHeaderValue.Render(fmt.Sprintf("%.0f%%", summary.AvgCacheHit))
	cacheLabel := StyleHeaderStat.Render("cache hit")

	ttft := StyleHeaderValue.Render(fmt.Sprintf("%.0fms", summary.P99TTFT))
	ttftLabel := StyleHeaderStat.Render("P99 TTFT")

	interval := StyleHeaderStat.Render(fmt.Sprintf("↻ %ds", intervalSec))

	dot := StyleHeaderDot.Render("·")

	parts := []string{
		" " + title,
		dot,
		workers,
		dot,
		tokPerSec,
		dot,
		cacheLabel + " " + cacheHit,
		dot,
		ttftLabel + " " + ttft,
		dot,
		interval + " ",
	}

	header := ""
	for _, p := range parts {
		header += p + " "
	}

	return lipgloss.NewStyle().
		Width(width).
		Background(colorDark).
		Foreground(colorWhite).
		Render(header)
}
