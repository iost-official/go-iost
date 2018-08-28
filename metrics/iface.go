package metrics

// Counter defines the API of counter-type metrics.
type Counter interface {
	Add(float64, map[string]string) error
}

// Gauge defines the API of gauge-type metrics.
type Gauge interface {
	Set(float64, map[string]string) error
}

// Summary defines the API of summary-type metrics.
type Summary interface {
	Observe(float64, map[string]string) error
}
