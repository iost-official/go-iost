package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	assert.Nil(t, SetPusher("47.75.42.25:9091"))
	assert.Nil(t, Start())
	counter := NewCounter("test_counter", nil)
	counter.Add(1, nil)
	counter.Add(1, nil)
	counter.Add(1, nil)

	gauge := NewGauge("test_gauge", []string{"tag"})
	gauge.Set(100, map[string]string{"tag": "aaa"})

	sum := NewSummary("test_summary", nil)
	sum.Observe(100, nil)
	sum.Observe(200, nil)
	sum.Observe(300, nil)

	Stop()
}
