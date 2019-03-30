// Refer to https://github.com/prometheus/node_exporter

package exporter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/iost-official/go-iost/metrics"
)

const (
	diskSectorSize    = 512
	diskstatsFilename = "/proc/diskstats"
)

type diskstatsCollector struct {
	ignoredDevicesPattern *regexp.Regexp
	item                  []string
	factor                []int
	descs                 []metrics.Gauge
}

func init() {
	registerCollector("diskstats", true, newDiskstatsCollector)
}

// newDiskstatsCollector creates diskstatsCollector
func newDiskstatsCollector() (collector, error) {
	return &diskstatsCollector{
		ignoredDevicesPattern: regexp.MustCompile("^((h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"),
		item: []string{
			// see https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
			//  1 - major number
			//  2 - minor mumber
			//  3 - device name
			"disk_reads_completed_total",  //  4 - reads completed successfully
			"",                            //  5 - reads merged
			"disk_read_bytes_total",       //  6 - sectors read
			"",                            //  7 - time spent reading (ms)
			"disk_writes_completed_total", //  8 - writes completed
			"",                            //  9 - writes merged
			"disk_written_bytes_total",    // 10 - sectors written
			"",                            // 11 - time spent writing (ms)
			// 12 - I/Os currently in progress
			// 13 - time spent doing I/Os (ms)
			// 14 - weighted time spent doing I/Os (ms)
		},
		factor: []int{
			1,
			0,
			diskSectorSize,
			0,
			1,
			0,
			diskSectorSize,
			0,
		},
		descs: make([]metrics.Gauge, 8),
	}, nil
}

// Update implements collector.Update
func (c *diskstatsCollector) Update() error {
	diskStats, err := getDiskStats()
	if err != nil {
		return fmt.Errorf("couldn't get diskstats: %s", err)
	}

	for dev, stats := range diskStats {
		if c.ignoredDevicesPattern.MatchString(dev) {
			continue
		}

		for i, value := range stats {
			// ignore unrecognized additional stats
			if i < len(c.descs) && len(c.item[i]) > 0 {
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("invalid value %s in diskstats: %s", value, err)
				}
				c.setOrCreate(i, v*float64(c.factor[i]), map[string]string{"dev": dev})
			}
		}
	}
	return nil
}

func (c *diskstatsCollector) setOrCreate(index int, value float64, labels map[string]string) {
	if c.descs[index] == nil {
		c.descs[index] = metrics.NewGauge(c.item[index], []string{"dev"})
	}
	c.descs[index].Set(value, labels)
}

func getDiskStats() (map[string][]string, error) {
	file, err := os.Open(diskstatsFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseDiskStats(file)
}

func parseDiskStats(r io.Reader) (map[string][]string, error) {
	var (
		diskStats = map[string][]string{}
		scanner   = bufio.NewScanner(r)
	)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 4 { // we strip major, minor and dev
			return nil, fmt.Errorf("invalid line in %s: %s", diskstatsFilename, scanner.Text())
		}
		dev := parts[2]
		diskStats[dev] = parts[3:]
	}

	return diskStats, scanner.Err()
}
