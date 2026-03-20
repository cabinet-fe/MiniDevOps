//go:build !windows

package service

import (
	"golang.org/x/sys/unix"
)

func readDiskUsage(path string) (uint64, uint64, float64, error) {
	var fs unix.Statfs_t
	if err := unix.Statfs(path, &fs); err != nil {
		return 0, 0, 0, err
	}

	total := uint64(fs.Blocks) * uint64(fs.Bsize)
	free := uint64(fs.Bavail) * uint64(fs.Bsize)
	if total == 0 {
		return 0, 0, 0, nil
	}

	used := total - free
	usage := float64(used) / float64(total) * 100

	return total, free, roundSingleDecimal(usage), nil
}
