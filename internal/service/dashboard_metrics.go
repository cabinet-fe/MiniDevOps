package service

import (
	"math"
	"time"
)

const (
	dashboardCPUBootstrapSleep = 50 * time.Millisecond
	dashboardCPUSampleMinGap   = 10 * time.Millisecond
	dashboardCPUSampleMaxGap   = 30 * time.Second
)

func collectDashboardSystemResources(diskPath string) DashboardSystemResources {
	resources := DashboardSystemResources{}

	fillDashboardCPUUsage(&resources)
	fillDashboardMemoryUsage(&resources)

	if diskPath == "" {
		diskPath = "."
	}
	if total, free, usage, err := readDiskUsage(diskPath); err == nil {
		resources.DiskTotalBytes = total
		resources.DiskFreeBytes = free
		resources.DiskUsagePercent = usage
	} else if diskPath != "." {
		if total, free, usage, fallbackErr := readDiskUsage("."); fallbackErr == nil {
			resources.DiskTotalBytes = total
			resources.DiskFreeBytes = free
			resources.DiskUsagePercent = usage
		}
	}

	return resources
}

func roundSingleDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}
