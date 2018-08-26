package metrics

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// errors
var (
	ErrDuplMetricsType   = errors.New("duplicated metrics name")
	ErrNilPusher         = errors.New("pusher is nil")
	ErrPusherUnavailable = errors.New("pusher addr unavailable")
	ErrMetricsStopped    = errors.New("metrics has been stopped")
)

type metricsType int

const (
	metricsTypeCounter metricsType = iota
	metricsTypeGauge
	metricsTypeTimer
)

var (
	pushInterval = time.Millisecond * 500
)

// Client is the struct responsible for sending different type of metrics.
type Client struct {
	isRunning uint32

	metricsTypeMap *sync.Map // map[string]metricsType

	allMetrics map[metricsType]*sync.Map

	pusher *push.Pusher
	exitCh chan struct{}
}

// NewClient returns a new Client.
func NewClient() *Client {
	c := &Client{
		metricsTypeMap: new(sync.Map),
		exitCh:         make(chan struct{}),
	}
	c.allMetrics = map[metricsType]*sync.Map{
		metricsTypeCounter: new(sync.Map), // map[string]*prometheus.CounterVec
		metricsTypeGauge:   new(sync.Map), // map[string]*prometheus.GaugeVec
		metricsTypeTimer:   new(sync.Map), // map[string]*prometheus.SummaryVec
	}
	return c
}

// SetPusher sets the pusher with the given addr.
func (c *Client) SetPusher(addr string) error {
	if !isAddrAvailable(addr) {
		return ErrPusherUnavailable
	}
	c.pusher = push.New(addr, "iost")
	return nil
}

// Start starts the pusher loop.
func (c *Client) Start() error {
	if c.pusher == nil {
		return ErrNilPusher
	}
	if !atomic.CompareAndSwapUint32(&c.isRunning, 0, 1) {
		return nil
	}
	go c.startPush()
	return nil
}

// Stop stops the pusher loop.
func (c *Client) Stop() {
	if c.pusher == nil {
		return
	}
	if !atomic.CompareAndSwapUint32(&c.isRunning, 1, 0) {
		return
	}
	c.exitCh <- struct{}{}
	<-c.exitCh
	c.pusher.Add()
}

// Counter sends a counter type metrics.
func (c *Client) Counter(name string, value float64, tagkv map[string]string) error {
	return c.doMetrics(metricsTypeCounter, name, value, tagkv)
}

// Gauge sends a gauge type metrics.
func (c *Client) Gauge(name string, value float64, tagkv map[string]string) error {
	return c.doMetrics(metricsTypeGauge, name, value, tagkv)
}

// Timer sends a summary type metrics.
func (c *Client) Timer(name string, value float64, tagkv map[string]string) error {
	return c.doMetrics(metricsTypeTimer, name, value, tagkv)
}

func (c *Client) doMetrics(mt metricsType, name string, value float64, tagkv map[string]string) error {
	err := c.checkAndSet(name, mt)
	if err != nil {
		return err
	}
	metr, exist := c.allMetrics[mt].LoadOrStore(name, c.getMetricsInstance(mt, name, tagkv))
	if !exist {
		c.pusher.Collector(metr.(prometheus.Collector))
	}
	switch mt {
	case metricsTypeCounter:
		counter, err := metr.(*prometheus.CounterVec).GetMetricWith(prometheus.Labels(tagkv))
		if err != nil {
			return err
		}
		counter.Add(value)
	case metricsTypeGauge:
		gauge, err := metr.(*prometheus.GaugeVec).GetMetricWith(prometheus.Labels(tagkv))
		if err != nil {
			return err
		}
		gauge.Set(value)
	case metricsTypeTimer:
		sum, err := metr.(*prometheus.SummaryVec).GetMetricWith(prometheus.Labels(tagkv))
		if err != nil {
			return err
		}
		sum.Observe(value)
	default:
		return nil
	}
	return nil
}

func (c *Client) checkAndSet(name string, mt metricsType) error {
	if c.pusher == nil {
		return ErrNilPusher
	}
	if atomic.LoadUint32(&c.isRunning) == 0 {
		return ErrMetricsStopped
	}
	metrTyp, exist := c.metricsTypeMap.Load(name)
	if !exist {
		c.metricsTypeMap.Store(name, mt)
		return nil
	}
	if metrTyp != mt {
		return ErrDuplMetricsType
	}
	return nil
}

func (c *Client) startPush() {
	timer := time.NewTimer(pushInterval)
	for {
		select {
		case <-timer.C:
			c.pusher.Add()
			timer.Reset(pushInterval)
		case <-c.exitCh:
			timer.Stop()
			c.exitCh <- struct{}{}
			return
		}
	}
}

func (c *Client) getMetricsInstance(mt metricsType, name string, tagkv map[string]string) interface{} {
	labelName := make([]string, 0, len(tagkv))
	for k := range tagkv {
		labelName = append(labelName, k)
	}
	switch mt {
	case metricsTypeCounter:
		return prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: "-",
		}, labelName)
	case metricsTypeGauge:
		return prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "-",
		}, labelName)
	case metricsTypeTimer:
		return prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: name,
			Help: "-",
		}, labelName)
	default:
		return nil
	}
}
