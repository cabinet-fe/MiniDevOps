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

// cpuUsagePercentFromWindowsGetSystemTimesDeltas computes overall CPU usage from
// deltas of GetSystemTimes counters. Per Windows API docs, kernel time includes
// idle time; elapsed system CPU time is kernel+user (not idle+kernel+user).
func cpuUsagePercentFromWindowsGetSystemTimesDeltas(deltaKernel, deltaUser, deltaIdle int64) float64 {
	total := deltaKernel + deltaUser
	if total <= 0 {
		return 0
	}
	busy := deltaKernel - deltaIdle + deltaUser
	if busy < 0 {
		busy = 0
	}
	if busy > total {
		busy = total
	}
	return roundSingleDecimal(float64(busy) / float64(total) * 100)
}
