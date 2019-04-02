package exporter

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/ilog"
)

const (
	updateInterval = 5 // This value effects irate
)

var (
	factories = make(map[string]func() (collector, error))
)

// collector base type
type collector interface {
	Update() error
}

// Exporter is a set of collectors.
type Exporter struct {
	collectors map[string]collector
	quitCh     chan struct{}
	done       *sync.WaitGroup
}

// New will return a new metrics Exporter.
func New() *Exporter {
	exporter := &Exporter{
		collectors: make(map[string]collector),
		quitCh:     make(chan struct{}),
		done:       new(sync.WaitGroup),
	}

	for key, collectorFactory := range factories {
		ilog.Infof("Add node metrics: %v", key)
		collectorInstance, err := collectorFactory()
		if err != nil {
			ilog.Errorf("Create metrics %v failed: %v", key, err)
		} else {
			exporter.collectors[key] = collectorInstance
		}
	}

	exporter.done.Add(1)
	go exporter.metricsController()

	return exporter
}

func registerCollector(collector string, enabled bool, factory func() (collector, error)) {
	if enabled {
		factories[collector] = factory
	}
}

// Close will close the Exporter.
func (s *Exporter) Close() {
	close(s.quitCh)
	s.done.Wait()
	ilog.Infof("Stopped metrics.")
}

func (s *Exporter) metricsController() {
	ilog.Infof("Node metrics started.")
	setNodeInfoMetrics()
	for {
		select {
		case <-time.After(updateInterval * time.Second):
			for _, collectorInstance := range s.collectors {
				collectorInstance.Update()
			}
		case <-s.quitCh:
			s.done.Done()
			return
		}
	}
}
