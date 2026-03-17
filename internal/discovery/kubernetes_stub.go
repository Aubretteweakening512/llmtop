//go:build nokubernetes

package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/InfraWhisperer/llmtop/internal/collector"
)

// KubernetesDiscoverer is a stub when built with the nokubernetes tag.
type KubernetesDiscoverer struct{}

// DiscoveredPod is a stub when built with the nokubernetes tag.
type DiscoveredPod struct{}

// NewKubernetesDiscoverer returns an error when Kubernetes support is compiled out.
func NewKubernetesDiscoverer(kubeconfig, namespace, selector string, maxConcurrent int, reqTimeout time.Duration) (*KubernetesDiscoverer, error) {
	return nil, fmt.Errorf("kubernetes discovery disabled in this build (built with -tags nokubernetes)")
}

// ContextName returns an empty string in the stub build.
func (d *KubernetesDiscoverer) ContextName() string { return "" }

// DiscoverPods returns an error in the stub build.
func (d *KubernetesDiscoverer) DiscoverPods(_ context.Context) ([]DiscoveredPod, error) {
	return nil, fmt.Errorf("kubernetes discovery disabled")
}

// ToWorkerConfigs returns nil in the stub build.
func (d *KubernetesDiscoverer) ToWorkerConfigs(_ []DiscoveredPod) []collector.WorkerConfig {
	return nil
}

// DiscoverDCGMPod returns an error in the stub build.
func (d *KubernetesDiscoverer) DiscoverDCGMPod(_ context.Context) (func(context.Context) (string, error), error) {
	return nil, fmt.Errorf("kubernetes discovery disabled")
}
