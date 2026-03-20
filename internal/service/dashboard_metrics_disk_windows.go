//go:build windows

package service

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

func readDiskUsage(path string) (uint64, uint64, float64, error) {
	if path == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, 0, 0, err
	}
	p, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return 0, 0, 0, err
	}

	var freeAvailable, totalBytes, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &freeAvailable, &totalBytes, &totalFree); err != nil {
		return 0, 0, 0, err
	}
	if totalBytes == 0 {
		return 0, 0, 0, nil
	}

	used := totalBytes - freeAvailable
	usage := float64(used) / float64(totalBytes) * 100

	return totalBytes, freeAvailable, roundSingleDecimal(usage), nil
}
