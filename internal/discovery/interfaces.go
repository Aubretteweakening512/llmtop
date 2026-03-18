package discovery

import "context"

// Discoverer abstracts endpoint discovery so that reconcile loops and tests
// can work with any discovery source (K8s, local, static, mock).
type Discoverer interface {
	Discover(ctx context.Context) ([]Target, error)
	ContextName() string
}
