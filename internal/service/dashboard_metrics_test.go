package service

import (
	"strings"
	"testing"
)

func TestParseCPUStatLine(t *testing.T) {
	snapshot, err := parseCPUStatLine("cpu  4705 0 4313 1362393 182 0 98 0 0 0")
	if err != nil {
		t.Fatalf("parse cpu stat line: %v", err)
	}

	if snapshot.total != 1371691 {
		t.Fatalf("expected total 1371691, got %d", snapshot.total)
	}
	if snapshot.idle != 1362575 {
		t.Fatalf("expected idle 1362575, got %d", snapshot.idle)
	}
}

func TestCalculateCPUUsage(t *testing.T) {
	previous := cpuSnapshot{idle: 900, total: 1000}
	current := cpuSnapshot{idle: 1050, total: 1300}

	usage := calculateCPUUsage(previous, current)
	if usage != 50 {
		t.Fatalf("expected usage 50, got %.1f", usage)
	}
}

func TestCPUUsagePercentFromWindowsGetSystemTimesDeltas(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		dk   int64
		du   int64
		di   int64
		want float64
	}{
		{
			name: "idle_like_kernel_tracks_idle",
			dk:   100, du: 0, di: 100,
			want: 0,
		},
		{
			name: "half_busy",
			dk:   40, du: 60, di: 50,
			want: 50,
		},
		{
			name: "non_positive_total",
			dk:   -10, du: 5, di: 0,
			want: 0,
		},
		{
			name: "busy_clamped_to_total",
			dk:   100, du: 50, di: -20,
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cpuUsagePercentFromWindowsGetSystemTimesDeltas(tt.dk, tt.du, tt.di)
			if got != tt.want {
				t.Fatalf("expected %.1f, got %.1f", tt.want, got)
			}
		})
	}
}

func TestParseMemoryUsageUsesMemAvailable(t *testing.T) {
	meminfo := strings.NewReader(`
MemTotal:       16384 kB
MemFree:         2048 kB
MemAvailable:    6144 kB
Buffers:         1024 kB
Cached:          2048 kB
`)

	total, used, usage, err := parseMemoryUsage(meminfo)
	if err != nil {
		t.Fatalf("parse memory usage: %v", err)
	}

	if total != 16777216 {
		t.Fatalf("expected total 16777216, got %d", total)
	}
	if used != 10485760 {
		t.Fatalf("expected used 10485760, got %d", used)
	}
	if usage != 62.5 {
		t.Fatalf("expected usage 62.5, got %.1f", usage)
	}
}

func TestParseMemoryUsageFallsBackWhenMemAvailableMissing(t *testing.T) {
	meminfo := strings.NewReader(`
MemTotal:       8192 kB
MemFree:        1024 kB
Buffers:         512 kB
Cached:         1024 kB
`)

	total, used, usage, err := parseMemoryUsage(meminfo)
	if err != nil {
		t.Fatalf("parse memory usage fallback: %v", err)
	}

	if total != 8388608 {
		t.Fatalf("expected total 8388608, got %d", total)
	}
	if used != 5767168 {
		t.Fatalf("expected used 5767168, got %d", used)
	}
	if usage != 68.8 {
		t.Fatalf("expected usage 68.8, got %.1f", usage)
	}
}
