// llmtop - htop for LLM inference clusters
// Real-time terminal dashboard for vLLM, SGLang, and LMCache workers.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/InfraWhisperer/llmtop/internal/collector"
	"github.com/InfraWhisperer/llmtop/internal/discovery"
	"github.com/InfraWhisperer/llmtop/internal/metrics"
	"github.com/InfraWhisperer/llmtop/internal/ui"
	"github.com/InfraWhisperer/llmtop/pkg/config"
)

// k8sFlags holds Kubernetes-related CLI flags.
type k8sFlags struct {
	kubeconfig    string
	namespace     string
	selector      string
	allNamespaces bool
	maxConcurrent int
	noK8s         bool
}

// version is set via -ldflags at build time.
var version = "0.1.0"

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		endpoints    []string
		configFile   string
		intervalSec  int
		once         bool
		outputFmt    string
		dcgmEndpoint string
		kf           k8sFlags
	)

	cmd := &cobra.Command{
		Use:   "llmtop",
		Short: "htop for your LLM inference cluster",
		Long: `llmtop — real-time terminal dashboard for vLLM, SGLang, LMCache, and NIM.

Monitor GPU cache utilization, request queues, TTFT latency, and token
throughput across your entire inference fleet in a single glorious view.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(endpoints, configFile, intervalSec, once, outputFmt, dcgmEndpoint, kf)
		},
	}

	cmd.Flags().StringArrayVarP(&endpoints, "endpoint", "e", nil, "Endpoint URL to monitor (repeatable)")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to config YAML file")
	cmd.Flags().IntVarP(&intervalSec, "interval", "i", 2, "Refresh interval in seconds")
	cmd.Flags().BoolVar(&once, "once", false, "Run once and exit (non-interactive)")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format for --once mode: table, json")
	cmd.Flags().StringVar(&dcgmEndpoint, "dcgm-endpoint", "", "DCGM exporter URL for GPU metrics")

	// Kubernetes discovery flags
	cmd.Flags().StringVar(&kf.kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	cmd.Flags().StringVarP(&kf.namespace, "namespace", "n", "", "Kubernetes namespace for pod discovery")
	cmd.Flags().StringVarP(&kf.selector, "selector", "l", "", "Label selector for pod discovery")
	cmd.Flags().BoolVarP(&kf.allNamespaces, "all-namespaces", "A", false, "Discover across all namespaces")
	cmd.Flags().IntVar(&kf.maxConcurrent, "max-concurrent", 10, "Max concurrent K8s API proxy requests")
	cmd.Flags().BoolVar(&kf.noK8s, "no-k8s", false, "Disable Kubernetes discovery")

	cmd.AddCommand(versionCmd())

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("llmtop %s\n", version)
		},
	}
}

func run(endpoints []string, configFile string, intervalSec int, once bool, outputFmt string, dcgmEndpoint string, kf k8sFlags) error {
	ctx := context.Background()

	// Build worker configs
	var workerConfigs []collector.WorkerConfig
	var cfg *config.Config
	var dcgmFetchFunc func(ctx context.Context) (string, error)

	if configFile != "" {
		var err error
		cfg, err = config.LoadFile(configFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if cfg.Interval > 0 && intervalSec == 2 {
			intervalSec = cfg.Interval
		}
		// Config-level DCGM endpoint; CLI flag overrides.
		if dcgmEndpoint == "" && cfg.DCGMEndpoint != "" {
			dcgmEndpoint = cfg.DCGMEndpoint
		}
		for _, ep := range cfg.Endpoints {
			backend := parseBackend(ep.Backend)
			workerConfigs = append(workerConfigs, collector.WorkerConfig{
				Endpoint:    strings.TrimRight(ep.URL, "/"),
				Backend:     backend,
				Label:       ep.Label,
				MetricsPath: ep.MetricsPath,
			})
		}
	}

	for _, ep := range endpoints {
		workerConfigs = append(workerConfigs, collector.WorkerConfig{
			Endpoint: strings.TrimRight(ep, "/"),
			Backend:  metrics.BackendUnknown,
		})
	}

	// Kubernetes discovery: attempt if no static endpoints and not disabled
	var k8sContext string
	if !kf.noK8s && len(workerConfigs) == 0 {
		kubecfg := kf.kubeconfig
		if kubecfg == "" {
			kubecfg = os.Getenv("KUBECONFIG")
		}

		ns := kf.namespace
		if ns == "" && cfg != nil {
			ns = cfg.Kubernetes.Namespace
		}
		if kf.allNamespaces {
			ns = ""
		}

		sel := kf.selector
		if sel == "" && cfg != nil {
			sel = cfg.Kubernetes.LabelSelector
		}

		mc := kf.maxConcurrent
		if cfg != nil && cfg.Kubernetes.MaxConcurrent > 0 {
			mc = cfg.Kubernetes.MaxConcurrent
		}

		disc, err := discovery.NewKubernetesDiscoverer(kubecfg, ns, sel, mc, 2*time.Second)
		if err == nil {
			k8sContext = disc.ContextName()
			pods, discErr := disc.DiscoverPods(ctx)
			if discErr == nil && len(pods) > 0 {
				fmt.Fprintf(os.Stderr, "llmtop: discovered %d pods via Kubernetes (context: %s)\n", len(pods), k8sContext)
				workerConfigs = append(workerConfigs, disc.ToWorkerConfigs(pods)...)
			}

			// Auto-discover DCGM exporter via K8s if no explicit dcgm endpoint
			if dcgmEndpoint == "" && (cfg == nil || cfg.DCGMEndpoint == "") {
				dcgmFetch, dcgmErr := disc.DiscoverDCGMPod(ctx)
				if dcgmErr == nil {
					fmt.Fprintln(os.Stderr, "llmtop: discovered DCGM exporter via Kubernetes")
					dcgmEndpoint = "k8s://dcgm-exporter" // placeholder for display
					dcgmFetchFunc = dcgmFetch
				}
			}
		}
	}

	// Fall back to localhost discovery if still no endpoints
	if len(workerConfigs) == 0 {
		fmt.Fprintln(os.Stderr, "No endpoints specified, auto-discovering localhost...")
		discovered := discovery.DiscoverLocal(ctx)
		if len(discovered) == 0 {
			fmt.Fprintln(os.Stderr, "No LLM workers found. Use --endpoint, --config, or --kubeconfig to specify endpoints.")
			fmt.Fprintln(os.Stderr, "Checked ports: 8000, 8001, 8002, 8003, 8080, 8081, 8090")
		}
		workerConfigs = append(workerConfigs, discovered...)
	}

	// Create collector
	c := collector.New(workerConfigs, time.Duration(intervalSec)*time.Second)

	// Create DCGM collector if endpoint is configured or auto-discovered
	var dc *collector.DCGMCollector
	if dcgmFetchFunc != nil {
		dc = collector.NewDCGMCollectorWithFetchFunc(
			dcgmEndpoint,
			time.Duration(intervalSec)*time.Second,
			dcgmFetchFunc,
		)
	} else if dcgmEndpoint != "" {
		dc = collector.NewDCGMCollector(
			strings.TrimRight(dcgmEndpoint, "/"),
			time.Duration(intervalSec)*time.Second,
		)
	}

	if once {
		return runOnce(ctx, c, dc, outputFmt)
	}

	return runTUI(ctx, c, dc, intervalSec, k8sContext)
}

func runOnce(ctx context.Context, c *collector.Collector, dc *collector.DCGMCollector, outputFmt string) error {
	c.PollNow()
	workers := c.GetAll()
	summary := metrics.ComputeFleetSummary(workers)

	if dc != nil {
		dc.PollNow()
	}

	switch outputFmt {
	case "json":
		modelGroups := metrics.GroupWorkersByModel(workers)
		envelope := struct {
			Summary     metrics.FleetSummary     `json:"summary"`
			Workers     []*metrics.WorkerMetrics  `json:"workers"`
			ModelGroups []metrics.ModelGroup      `json:"model_groups,omitempty"`
			GPUSummary  *metrics.GPUSummary       `json:"gpu_summary,omitempty"`
			GPUs        []*metrics.GPUInfo        `json:"gpus,omitempty"`
		}{Summary: summary, Workers: workers, ModelGroups: modelGroups}

		if dc != nil {
			gpus := dc.GetAll()
			if len(gpus) > 0 {
				gpuSummary := dc.GetSummary()
				envelope.GPUSummary = &gpuSummary
				envelope.GPUs = gpus
			}
		}

		data, err := json.MarshalIndent(envelope, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default: // table
		printTable(workers, summary)
	}
	return nil
}

func printTable(workers []*metrics.WorkerMetrics, summary metrics.FleetSummary) {
	fmt.Printf("\nllmtop %s  ·  %d workers (%d online)  ·  %.0f tok/s  ·  cache hit %.0f%%  ·  P99 TTFT %.0fms\n\n",
		version, summary.TotalWorkers, summary.OnlineWorkers,
		summary.TotalTokPerSec, summary.AvgCacheHit, summary.P99TTFT)

	w := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ENDPOINT\tBACKEND\tMODEL\tKV%\tQUEUE\tRUN\tTTFT P99\tITL P99\tHIT%\tTOK/S")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 100))
	for _, worker := range workers {
		status := "●"
		if !worker.Online {
			status = "○"
		}
		var ep string
		if strings.HasPrefix(worker.Endpoint, "k8s://") {
			ep = worker.Label
			if ep == "" {
				ep = strings.TrimPrefix(worker.Endpoint, "k8s://")
			}
		} else {
			ep = strings.TrimPrefix(worker.Endpoint, "http://")
			if worker.Label != "" {
				ep = ep + " (" + worker.Label + ")"
			}
		}
		model := worker.ModelName
		if model == "" {
			model = "—"
		}
		_, _ = fmt.Fprintf(w, "%s %s\t%s\t%s\t%.0f%%\t%d\t%d\t%.0fms\t%.0fms\t%.0f%%\t%.0f\n",
			status, ep,
			string(worker.Backend),
			model,
			worker.KVCacheUsagePct,
			worker.RequestsWaiting,
			worker.RequestsRunning,
			worker.TTFT_P99,
			worker.ITL_P99,
			worker.CacheHitRatePct,
			worker.GenTokPerSec+worker.PromptTokPerSec,
		)
	}
	_ = w.Flush() // nolint:errcheck
}

func runTUI(ctx context.Context, c *collector.Collector, dc *collector.DCGMCollector, intervalSec int, k8sContext string) error {
	c.Start(ctx)
	defer c.Stop()

	if dc != nil {
		dc.Start(ctx)
		defer dc.Stop()
	}

	model := ui.NewModel(c, dc, version, intervalSec, k8sContext)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}

func parseBackend(s string) metrics.Backend {
	switch strings.ToLower(s) {
	case "vllm":
		return metrics.BackendVLLM
	case "sglang":
		return metrics.BackendSGLang
	case "lmcache":
		return metrics.BackendLMCache
	case "nim":
		return metrics.BackendNIM
	default:
		return metrics.BackendUnknown
	}
}
