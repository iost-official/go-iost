package metrics

import "github.com/prometheus/client_golang/prometheus"

// PromCounter is the implementation of Counter with prometheus's CounterVec.
type PromCounter struct {
	counterVec *prometheus.CounterVec
}

// NewPromCounter returns a instance of PromCounter.
func NewPromCounter(c *prometheus.CounterVec) *PromCounter {
	return &PromCounter{
		counterVec: c,
	}
}

// Add adds the given value to the prometheus Counter.
func (p *PromCounter) Add(value float64, tagkv map[string]string) error {
	counter, err := p.counterVec.GetMetricWith(prometheus.Labels(tagkv))
	if err != nil {
		return err
	}
	counter.Add(value)
	return nil
}

// PromGauge is the implementation of Gauge with prometheus's GaugeVec.
type PromGauge struct {
	gaugeVec *prometheus.GaugeVec
}

// NewPromGauge returns a instance of PromGauge.
func NewPromGauge(g *prometheus.GaugeVec) *PromGauge {
	return &PromGauge{
		gaugeVec: g,
	}
}

// Set sets the given value to the prometheus Gauge.
func (p *PromGauge) Set(value float64, tagkv map[string]string) error {
	gauge, err := p.gaugeVec.GetMetricWith(prometheus.Labels(tagkv))
	if err != nil {
		return err
	}
	gauge.Set(value)
	return nil
}

// PromSummary is the implementation of Summary with prometheus's SummaryVec.
type PromSummary struct {
	summaryVec *prometheus.SummaryVec
}

// NewPromSummary returns a instance of PromSummary.
func NewPromSummary(s *prometheus.SummaryVec) *PromSummary {
	return &PromSummary{
		summaryVec: s,
	}
}

// Observe adds the observations to the prometheus Summary.
func (p *PromSummary) Observe(value float64, tagkv map[string]string) error {
	summary, err := p.summaryVec.GetMetricWith(prometheus.Labels(tagkv))
	if err != nil {
		return err
	}
	summary.Observe(value)
	return nil
}
