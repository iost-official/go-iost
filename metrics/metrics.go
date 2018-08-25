package metrics

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// errors
var (
	ErrDuplMetricsType = errors.New("duplicated metrics name")
)

type metricsType int

const (
	metricsTypeCounter metricsType = iota
	metricsTypeGauge
	metricsTypeTimer
)

// Client is the struct responsible for sending different type of metrics.
type Client struct {
	allMetrics *sync.Map // map[string]metricsType

	counter *sync.Map // map[string]prometheus.CounterVec
	gauge   *sync.Map // map[string]prometheus.GaugeVec
	timer   *sync.Map // map[string]prometheus.SummaryVec

	pusher *push.Pusher
}

// NewClient returns a new Client.
func NewClient() *Client {
	// pusher := push.New("127.0.0.1:9091", "iost").Collector(cpuTemp).Collector(hdFailures)
	return &Client{
		allMetrics: new(sync.Map),
		counter:    new(sync.Map),
		gauge:      new(sync.Map),
		timer:      new(sync.Map),
	}
}

// Counter sends a counter type metrics.
func (c *Client) Counter(name string, value float64, tagkv map[string]string) error {
	err := c.checkAndSet(name, metricsTypeCounter)
	if err != nil {
		return err
	}
	counter, _ := c.counter.LoadOrStore(name, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
	}, nil))
	counter.(*prometheus.CounterVec).With(prometheus.Labels(tagkv)).Inc()
	return nil
}

// Gauge sends a gauge type metrics.
func (c *Client) Gauge(name string, value float64, tagkv map[string]string) error {
	return nil
}

// Timer sends a summary type metrics.
func (c *Client) Timer(name string, value float64, tagkv map[string]string) error {
	return nil
}

func (c *Client) checkAndSet(name string, mt metricsType) error {
	metrTyp, exist := c.allMetrics.Load(name)
	if !exist {
		c.allMetrics.Store(name, mt)
		return nil
	}
	if metrTyp != mt {
		return ErrDuplMetricsType
	}
	return nil
}
