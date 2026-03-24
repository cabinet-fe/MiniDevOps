//go:build windows

package service

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type winCPUSnapshot struct {
	idle   uint64
	kernel uint64
	user   uint64
}

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemTimes       = kernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")

	winCPUMu     sync.Mutex
	winCPULast   winCPUSnapshot
	winCPULastAt time.Time
	winCPUValid  bool
)

// memoryStatusEx matches MEMORYSTATUSEX (kernel32).
type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func filetimeTo100ns(ft *windows.Filetime) uint64 {
	return uint64(ft.LowDateTime) | uint64(ft.HighDateTime)<<32
}

func readWindowsCPUSnapshot() (winCPUSnapshot, error) {
	var idle, kernel, user windows.Filetime
	r, _, err := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r == 0 {
		return winCPUSnapshot{}, fmt.Errorf("GetSystemTimes: %w", err)
	}
	i := filetimeTo100ns(&idle)
	k := filetimeTo100ns(&kernel)
	u := filetimeTo100ns(&user)
	return winCPUSnapshot{idle: i, kernel: k, user: u}, nil
}

func calculateWindowsCPUUsage(prev, cur winCPUSnapshot) float64 {
	dk := int64(cur.kernel) - int64(prev.kernel)
	du := int64(cur.user) - int64(prev.user)
	di := int64(cur.idle) - int64(prev.idle)
	return cpuUsagePercentFromWindowsGetSystemTimesDeltas(dk, du, di)
}

func fillDashboardCPUUsage(r *DashboardSystemResources) {
	current, err := readWindowsCPUSnapshot()
	if err != nil {
		return
	}

	winCPUMu.Lock()
	if winCPUValid {
		gap := time.Since(winCPULastAt)
		if gap >= dashboardCPUSampleMinGap && gap <= dashboardCPUSampleMaxGap {
			r.CPUUsagePercent = calculateWindowsCPUUsage(winCPULast, current)
			winCPULast = current
			winCPULastAt = time.Now()
			winCPUMu.Unlock()
			return
		}
	}
	winCPUMu.Unlock()

	time.Sleep(dashboardCPUBootstrapSleep)

	second, err := readWindowsCPUSnapshot()
	if err != nil {
		return
	}

	winCPUMu.Lock()
	defer winCPUMu.Unlock()
	r.CPUUsagePercent = calculateWindowsCPUUsage(current, second)
	winCPULast = second
	winCPULastAt = time.Now()
	winCPUValid = true
}

func fillDashboardMemoryUsage(r *DashboardSystemResources) {
	var ms memoryStatusEx
	ms.dwLength = uint32(unsafe.Sizeof(ms))
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	if ret == 0 {
		return
	}

	total := ms.ullTotalPhys
	if total == 0 {
		return
	}
	avail := ms.ullAvailPhys
	if avail > total {
		avail = total
	}
	used := total - avail
	pct := float64(used) / float64(total) * 100

	r.MemoryTotalBytes = total
	r.MemoryUsedBytes = used
	r.MemoryUsagePercent = roundSingleDecimal(pct)
}
