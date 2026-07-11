package service

import (
	"errors"
	"os"
	"testing"
)

func TestKillProcessRejectsSelfPID(t *testing.T) {
	svc := NewProcessService()
	self := int32(os.Getpid())

	name, err := svc.KillProcess(self)
	if !errors.Is(err, ErrKillSelf) {
		t.Fatalf("expected ErrKillSelf, got name=%q err=%v", name, err)
	}
	if err == nil || err.Error() != "不能终止 BuildFlow 自身" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestKillProcessRejectsInvalidPID(t *testing.T) {
	svc := NewProcessService()
	if _, err := svc.KillProcess(0); err == nil {
		t.Fatal("expected error for pid 0")
	}
	if _, err := svc.KillProcess(-1); err == nil {
		t.Fatal("expected error for negative pid")
	}
}

func TestSortProcessesByCPUAndMemory(t *testing.T) {
	list := []ProcessInfo{
		{PID: 1, Name: "a", CPUPercent: 10, MemoryBytes: 300},
		{PID: 2, Name: "b", CPUPercent: 50, MemoryBytes: 100},
		{PID: 3, Name: "c", CPUPercent: 20, MemoryBytes: 200},
	}

	sortProcesses(list, "cpu", "desc")
	if list[0].PID != 2 || list[1].PID != 3 || list[2].PID != 1 {
		t.Fatalf("cpu desc order wrong: %+v", pidsOf(list))
	}

	sortProcesses(list, "memory", "asc")
	if list[0].PID != 2 || list[1].PID != 3 || list[2].PID != 1 {
		t.Fatalf("memory asc order wrong: %+v", pidsOf(list))
	}

	sortProcesses(list, "name", "asc")
	if list[0].Name != "a" || list[1].Name != "b" || list[2].Name != "c" {
		t.Fatalf("name asc order wrong: %+v", namesOf(list))
	}

	unchanged := []ProcessInfo{
		{PID: 1, Name: "a", CPUPercent: 10},
		{PID: 2, Name: "b", CPUPercent: 50},
		{PID: 3, Name: "c", CPUPercent: 20},
	}
	sortProcesses(unchanged, "", "desc")
	if unchanged[0].PID != 1 || unchanged[1].PID != 2 || unchanged[2].PID != 3 {
		t.Fatalf("empty sort should keep original order: %+v", pidsOf(unchanged))
	}
}

func TestFilterProcessesByQPIDPort(t *testing.T) {
	list := []ProcessInfo{
		{PID: 10, Name: "nginx", Cmdline: "/usr/sbin/nginx", Ports: []uint32{80, 443}},
		{PID: 20, Name: "redis-server", Cmdline: "redis-server *:6379", Ports: []uint32{6379}},
		{PID: 30, Name: "buildflow", Cmdline: "./buildflow", Ports: []uint32{8080}},
	}

	pid := int32(20)
	filtered := filterProcesses(list, ProcessListOptions{PID: &pid})
	if len(filtered) != 1 || filtered[0].PID != 20 {
		t.Fatalf("pid filter: %+v", filtered)
	}

	port := uint32(80)
	filtered = filterProcesses(list, ProcessListOptions{Port: &port})
	if len(filtered) != 1 || filtered[0].Name != "nginx" {
		t.Fatalf("port filter: %+v", filtered)
	}

	filtered = filterProcesses(list, ProcessListOptions{Q: "REDIS"})
	if len(filtered) != 1 || filtered[0].PID != 20 {
		t.Fatalf("q filter name: %+v", filtered)
	}

	filtered = filterProcesses(list, ProcessListOptions{Q: "buildflow"})
	if len(filtered) != 1 || filtered[0].PID != 30 {
		t.Fatalf("q filter cmdline/name: %+v", filtered)
	}

	filtered = filterProcesses(list, ProcessListOptions{Q: "no-such-process"})
	if len(filtered) != 0 {
		t.Fatalf("expected empty, got %+v", filtered)
	}
}

func TestTopProcessesReturnsLimitedSorted(t *testing.T) {
	svc := NewProcessService()
	// 首次采样 CPU 多为 0，按内存排序更稳定
	list, err := svc.TopProcesses("memory", 5)
	if err != nil {
		t.Fatalf("TopProcesses: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one process")
	}
	if len(list) > 5 {
		t.Fatalf("expected at most 5, got %d", len(list))
	}
	for i := 1; i < len(list); i++ {
		if list[i].MemoryBytes > list[i-1].MemoryBytes {
			t.Fatalf("not sorted by memory desc at %d: %d > %d", i, list[i].MemoryBytes, list[i-1].MemoryBytes)
		}
		if list[i].Ports == nil {
			t.Fatalf("ports should be non-nil slice for pid %d", list[i].PID)
		}
	}
}

func pidsOf(list []ProcessInfo) []int32 {
	out := make([]int32, len(list))
	for i, p := range list {
		out[i] = p.PID
	}
	return out
}

func namesOf(list []ProcessInfo) []string {
	out := make([]string, len(list))
	for i, p := range list {
		out[i] = p.Name
	}
	return out
}
