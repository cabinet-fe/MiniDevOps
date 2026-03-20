package service

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cpuSnapshot struct {
	idle  uint64
	total uint64
}

// cpuUsageState caches the last /proc/stat aggregate line so we can compute usage
// from the delta to the current sample without sleeping on every request (the old
// 150ms sleep dominated latency for /dashboard/stats and /dashboard/system-resources).
var (
	cpuUsageMu        sync.Mutex
	cpuLastAgg        cpuSnapshot
	cpuLastAggAt      time.Time
	cpuLastAggValid   bool
	cpuBootstrapSleep = 50 * time.Millisecond
	cpuSampleMinGap   = 10 * time.Millisecond
	cpuSampleMaxGap   = 30 * time.Second
)

func collectDashboardSystemResources(diskPath string) DashboardSystemResources {
	resources := DashboardSystemResources{}

	if usage, err := readCPUUsage("/proc/stat"); err == nil {
		resources.CPUUsagePercent = usage
	}

	if total, used, usage, err := readMemoryUsage("/proc/meminfo"); err == nil {
		resources.MemoryTotalBytes = total
		resources.MemoryUsedBytes = used
		resources.MemoryUsagePercent = usage
	}

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

func readCPUUsage(path string) (float64, error) {
	current, err := readCPUStatFile(path)
	if err != nil {
		return 0, err
	}

	cpuUsageMu.Lock()
	if cpuLastAggValid {
		gap := time.Since(cpuLastAggAt)
		if gap >= cpuSampleMinGap && gap <= cpuSampleMaxGap {
			usage := calculateCPUUsage(cpuLastAgg, current)
			cpuLastAgg = current
			cpuLastAggAt = time.Now()
			cpuUsageMu.Unlock()
			return usage, nil
		}
	}
	cpuUsageMu.Unlock()

	time.Sleep(cpuBootstrapSleep)

	second, err := readCPUStatFile(path)
	if err != nil {
		return 0, err
	}

	cpuUsageMu.Lock()
	defer cpuUsageMu.Unlock()
	usage := calculateCPUUsage(current, second)
	cpuLastAgg = second
	cpuLastAggAt = time.Now()
	cpuLastAggValid = true
	return usage, nil
}

func readCPUStatFile(path string) (cpuSnapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return cpuSnapshot{}, err
	}
	defer file.Close()

	return readCPUStat(file)
}

func readCPUStat(r io.Reader) (cpuSnapshot, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return cpuSnapshot{}, err
		}
		return cpuSnapshot{}, fmt.Errorf("cpu stat is empty")
	}
	return parseCPUStatLine(scanner.Text())
}

func parseCPUStatLine(line string) (cpuSnapshot, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSnapshot{}, fmt.Errorf("invalid cpu stat line")
	}

	var total uint64
	for _, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuSnapshot{}, err
		}
		total += value
	}

	idle, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return cpuSnapshot{}, err
	}
	if len(fields) > 5 {
		iowait, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return cpuSnapshot{}, err
		}
		idle += iowait
	}

	return cpuSnapshot{idle: idle, total: total}, nil
}

func calculateCPUUsage(previous, current cpuSnapshot) float64 {
	if current.total <= previous.total {
		return 0
	}

	totalDelta := float64(current.total - previous.total)
	idleDelta := float64(current.idle - previous.idle)
	usage := (1 - idleDelta/totalDelta) * 100
	if usage < 0 {
		return 0
	}

	return roundSingleDecimal(usage)
}

func readMemoryUsage(path string) (uint64, uint64, float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	return parseMemoryUsage(file)
}

func parseMemoryUsage(r io.Reader) (uint64, uint64, float64, error) {
	scanner := bufio.NewScanner(r)
	values := map[string]uint64{}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		values[strings.TrimSuffix(fields[0], ":")] = value * 1024
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, 0, err
	}

	total := values["MemTotal"]
	if total == 0 {
		return 0, 0, 0, fmt.Errorf("missing MemTotal")
	}

	available := values["MemAvailable"]
	if available == 0 {
		available = values["MemFree"] + values["Buffers"] + values["Cached"]
	}
	if available > total {
		available = total
	}

	used := total - available
	usage := float64(used) / float64(total) * 100

	return total, used, roundSingleDecimal(usage), nil
}

func roundSingleDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}
