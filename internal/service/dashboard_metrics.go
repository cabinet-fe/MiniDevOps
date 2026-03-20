package service

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type cpuSnapshot struct {
	idle  uint64
	total uint64
}

func collectDashboardSystemResources(diskPath string) DashboardSystemResources {
	resources := DashboardSystemResources{}

	if usage, err := readCPUUsage("/proc/stat", 150*time.Millisecond); err == nil {
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

func readCPUUsage(path string, interval time.Duration) (float64, error) {
	first, err := readCPUStatFile(path)
	if err != nil {
		return 0, err
	}

	time.Sleep(interval)

	second, err := readCPUStatFile(path)
	if err != nil {
		return 0, err
	}

	return calculateCPUUsage(first, second), nil
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

func readDiskUsage(path string) (uint64, uint64, float64, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(path, &fs); err != nil {
		return 0, 0, 0, err
	}

	total := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bavail * uint64(fs.Bsize)
	if total == 0 {
		return 0, 0, 0, nil
	}

	used := total - free
	usage := float64(used) / float64(total) * 100

	return total, free, roundSingleDecimal(usage), nil
}

func roundSingleDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}
