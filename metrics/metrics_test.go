package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	assert.Nil(t, SetPusher("47.75.42.25:9091"))
	assert.Nil(t, Start())
	Counter("test_counter", 1, nil)
	Counter("test_counter", 1, nil)
	Counter("test_counter", 1, nil)
	Gauge("test_gauge", 100, map[string]string{"tag": "aaa"})
	Gauge("test_gauge", 100, map[string]string{"tag": "bbb"})
	Timer("test_timer", 100, nil)
	Timer("test_timer", 200, nil)
	Timer("test_timer", 300, nil)
	Stop()
}
