// Refer to https://github.com/prometheus/node_exporter

package exporter

import (
	"fmt"

	"github.com/iost-official/go-iost/metrics"
	"github.com/lufia/iostat"
)

type diskstatsCollector struct {
	item   []string
	descs  []metrics.Gauge
	update []func(stat *iostat.DriveStats) float64
}

func init() {
	registerCollector("diskstats", true, newDiskstatsCollector)
}

// newDiskstatsCollector creates diskstatsCollector
func newDiskstatsCollector() (collector, error) {
	return &diskstatsCollector{
		item: []string{
			"disk_reads_completed_total",
			"disk_read_bytes_total",
			"disk_read_time_seconds_total",
			"disk_writes_completed_total",
			"disk_written_bytes_total",
			"disk_write_time_seconds_total",
		},
		descs: make([]metrics.Gauge, 6),
		update: []func(stat *iostat.DriveStats) float64{
			func(stat *iostat.DriveStats) float64 {
				return float64(stat.NumRead)
			},
			func(stat *iostat.DriveStats) float64 {
				return float64(stat.BytesRead)
			},
			func(stat *iostat.DriveStats) float64 {
				return stat.TotalReadTime.Seconds()
			},
			func(stat *iostat.DriveStats) float64 {
				return float64(stat.NumWrite)
			},
			func(stat *iostat.DriveStats) float64 {
				return float64(stat.BytesWritten)
			},
			func(stat *iostat.DriveStats) float64 {
				return stat.TotalWriteTime.Seconds()
			},
		},
	}, nil
}

// Update implements collector.Update
func (c *diskstatsCollector) Update() error {
	diskStats, err := iostat.ReadDriveStats()
	if err != nil {
		return fmt.Errorf("couldn't get diskstats: %s", err)
	}

	for _, stats := range diskStats {

		for i, desc := range c.descs {
			v := c.update[i](stats)
			if desc == nil {
				c.descs[i] = metrics.NewGauge(c.item[i], []string{"dev"})
				desc = c.descs[i]
			}
			desc.Set(v, map[string]string{"dev": stats.Name})
		}
	}
	return nil
}
