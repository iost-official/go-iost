package main

import (
	"runtime"
	"strconv"
	"time"

	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

var (
	nodeInfoGauge = metrics.NewGauge("iost_node_info", []string{"cpu", "mem", "disk", "platform", "git_hash"})
	unknown       = "unknown"
)

func getTotalMem() string {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		ilog.Errorf("Getting memory info failed. err=%v", err)
		return unknown
	}
	return strconv.Itoa(int(vmStat.Total))
}

func getDiskSize() string {
	diskStat, err := disk.Usage(global.GetGlobalConf().DB.LdbPath)
	if err != nil {
		ilog.Errorf("Getting disk info failed. err=%v", err)
		return unknown
	}
	return strconv.Itoa(int(diskStat.Total))
}

func getPlatform() string {
	hostInfo, err := host.Info()
	if err != nil {
		ilog.Errorf("Getting platform info failed. err=%v", err)
		return unknown
	}
	return hostInfo.Platform + "-" + hostInfo.PlatformVersion
}

func setNodeInfoMetrics() {
	nodeInfoGauge.Set(float64(time.Now().Unix()*1e3), map[string]string{
		"cpu":      strconv.Itoa(runtime.NumCPU()),
		"mem":      getTotalMem(),
		"disk":     getDiskSize(),
		"platform": getPlatform(),
		"git_hash": global.GitHash,
	})
}
