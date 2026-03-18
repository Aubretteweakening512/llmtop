package collector

// Compile-time interface satisfaction checks.
var _ MetricsSource = (*Collector)(nil)
var _ GPUSource = (*DCGMCollector)(nil)
