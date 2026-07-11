package service

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// ErrKillSelf 禁止终止 BuildFlow 自身进程。
var ErrKillSelf = errors.New("不能终止 BuildFlow 自身")

// ProcessInfo 进程快照，供仪表盘与系统管理页使用。
type ProcessInfo struct {
	PID         int32    `json:"pid"`
	Name        string   `json:"name"`
	MemoryBytes uint64   `json:"memory_bytes"`
	CPUPercent  float64  `json:"cpu_percent"`
	Ports       []uint32 `json:"ports"`
	Username    string   `json:"username"`
	NumThreads  int32    `json:"num_threads"`
	Cmdline     string   `json:"cmdline,omitempty"`
	Status      string   `json:"status,omitempty"`
	CreateTime  int64    `json:"create_time,omitempty"` // unix ms
}

// ProcessListOptions 系统管理页进程列表查询参数。
type ProcessListOptions struct {
	Q        string // 名称/cmdline 子串（不区分大小写）
	PID      *int32
	Port     *uint32
	Sort     string // cpu | memory | name
	Order    string // asc | desc，默认 desc
	Page     int
	PageSize int
	// Detail 为 true 时填充 cmdline/status/create_time（系统管理页）
	Detail bool
	// WithPorts 为 true 时扫描 LISTEN 连接并填充 Ports
	WithPorts bool
}

type cpuSample struct {
	times cpu.TimesStat
	at    time.Time
}

// ProcessService 基于 gopsutil 的本机进程查询与终止（无 DB）。
type ProcessService struct {
	mu       sync.Mutex
	cpuCache map[int32]cpuSample
}

func NewProcessService() *ProcessService {
	return &ProcessService{
		cpuCache: make(map[int32]cpuSample),
	}
}

// ListProcesses 枚举进程，按条件过滤、排序并分页。
func (s *ProcessService) ListProcesses(opts ProcessListOptions) ([]ProcessInfo, int64, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, 0, fmt.Errorf("枚举进程失败: %w", err)
	}

	list := make([]ProcessInfo, 0, len(procs))
	seen := make(map[int32]struct{}, len(procs))
	for _, p := range procs {
		info, ok := s.collectProcess(p, opts.Detail)
		if !ok {
			continue
		}
		seen[info.PID] = struct{}{}
		list = append(list, info)
	}
	s.pruneCPUCache(seen)

	if opts.WithPorts || opts.Port != nil {
		portMap := listenPortMap()
		for i := range list {
			if ports, ok := portMap[list[i].PID]; ok {
				list[i].Ports = ports
			} else if list[i].Ports == nil {
				list[i].Ports = []uint32{}
			}
		}
	} else {
		for i := range list {
			if list[i].Ports == nil {
				list[i].Ports = []uint32{}
			}
		}
	}

	list = filterProcesses(list, opts)
	sortProcesses(list, opts.Sort, opts.Order)

	total := int64(len(list))
	page, pageSize := opts.Page, opts.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	start := (page - 1) * pageSize
	if start >= len(list) {
		return []ProcessInfo{}, total, nil
	}
	end := start + pageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], total, nil
}

// TopProcesses 返回按 CPU/内存排序的前 N 个进程，并为这 N 个 PID 填充监听端口。
func (s *ProcessService) TopProcesses(sortBy string, limit int) ([]ProcessInfo, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	if sortBy != "memory" {
		sortBy = "cpu"
	}

	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("枚举进程失败: %w", err)
	}

	list := make([]ProcessInfo, 0, len(procs))
	seen := make(map[int32]struct{}, len(procs))
	for _, p := range procs {
		info, ok := s.collectProcess(p, false)
		if !ok {
			continue
		}
		seen[info.PID] = struct{}{}
		list = append(list, info)
	}
	s.pruneCPUCache(seen)

	sortProcesses(list, sortBy, "desc")
	if len(list) > limit {
		list = list[:limit]
	}

	portMap := listenPortMap()
	for i := range list {
		if ports, ok := portMap[list[i].PID]; ok {
			list[i].Ports = ports
		} else {
			list[i].Ports = []uint32{}
		}
	}
	return list, nil
}

// KillProcess 向目标进程发送 SIGTERM；禁止终止自身。
func (s *ProcessService) KillProcess(pid int32) (name string, err error) {
	if pid <= 0 {
		return "", errors.New("无效的进程 PID")
	}
	if int(pid) == os.Getpid() {
		return "", ErrKillSelf
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return "", fmt.Errorf("进程不存在: %w", err)
	}
	name, _ = p.Name()
	if name == "" {
		name = fmt.Sprintf("pid-%d", pid)
	}
	if err := p.Terminate(); err != nil {
		return name, fmt.Errorf("终止进程失败: %w", err)
	}
	return name, nil
}

func (s *ProcessService) collectProcess(p *process.Process, detail bool) (ProcessInfo, bool) {
	info := ProcessInfo{PID: p.Pid, Ports: []uint32{}}

	name, err := p.Name()
	if err != nil {
		return info, false
	}
	info.Name = name

	if mem, err := p.MemoryInfo(); err == nil && mem != nil {
		info.MemoryBytes = mem.RSS
	}
	info.CPUPercent = s.cpuPercent(p)
	if u, err := p.Username(); err == nil {
		info.Username = u
	}
	if n, err := p.NumThreads(); err == nil {
		info.NumThreads = n
	}

	if detail {
		if cmd, err := p.Cmdline(); err == nil {
			info.Cmdline = cmd
		}
		if st, err := p.Status(); err == nil && len(st) > 0 {
			info.Status = strings.Join(st, ",")
		}
		if ct, err := p.CreateTime(); err == nil {
			info.CreateTime = ct
		}
	}
	return info, true
}

// cpuPercent 基于上次采样的 CPU times 计算瞬时占用；首次采样返回 0。
func (s *ProcessService) cpuPercent(p *process.Process) float64 {
	times, err := p.Times()
	if err != nil || times == nil {
		return 0
	}
	now := time.Now()
	numCPU := runtime.NumCPU()
	if numCPU < 1 {
		numCPU = 1
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	prev, ok := s.cpuCache[p.Pid]
	s.cpuCache[p.Pid] = cpuSample{times: *times, at: now}
	if !ok {
		return 0
	}

	deltaSec := now.Sub(prev.at).Seconds()
	if deltaSec <= 0 {
		return 0
	}
	// 与 gopsutil Process.Percent(0) 一致：按全部逻辑 CPU 归一化
	delta := deltaSec * float64(numCPU)
	deltaProc := (times.User - prev.times.User) + (times.System - prev.times.System)
	if deltaProc <= 0 {
		return 0
	}
	pct := (deltaProc / delta) * 100
	if pct < 0 {
		return 0
	}
	return roundSingleDecimal(pct)
}

func (s *ProcessService) pruneCPUCache(alive map[int32]struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for pid := range s.cpuCache {
		if _, ok := alive[pid]; !ok {
			delete(s.cpuCache, pid)
		}
	}
}

// listenPortMap 扫描一次 LISTEN 连接，构建 pid → 端口列表。
func listenPortMap() map[int32][]uint32 {
	conns, err := net.Connections("inet")
	if err != nil {
		return map[int32][]uint32{}
	}
	tmp := make(map[int32]map[uint32]struct{})
	for _, c := range conns {
		if c.Status != "LISTEN" || c.Pid <= 0 || c.Laddr.Port == 0 {
			continue
		}
		if tmp[c.Pid] == nil {
			tmp[c.Pid] = make(map[uint32]struct{})
		}
		tmp[c.Pid][c.Laddr.Port] = struct{}{}
	}
	out := make(map[int32][]uint32, len(tmp))
	for pid, set := range tmp {
		ports := make([]uint32, 0, len(set))
		for port := range set {
			ports = append(ports, port)
		}
		sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
		out[pid] = ports
	}
	return out
}

func filterProcesses(list []ProcessInfo, opts ProcessListOptions) []ProcessInfo {
	if opts.Q == "" && opts.PID == nil && opts.Port == nil {
		return list
	}
	q := strings.ToLower(strings.TrimSpace(opts.Q))
	out := make([]ProcessInfo, 0, len(list))
	for _, p := range list {
		if opts.PID != nil && p.PID != *opts.PID {
			continue
		}
		if opts.Port != nil {
			found := false
			for _, port := range p.Ports {
				if port == *opts.Port {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if q != "" {
			name := strings.ToLower(p.Name)
			cmd := strings.ToLower(p.Cmdline)
			if !strings.Contains(name, q) && !strings.Contains(cmd, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func sortProcesses(list []ProcessInfo, sortBy, order string) {
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if sortBy == "" {
		return
	}
	asc := strings.EqualFold(order, "asc")
	switch sortBy {
	case "memory":
		sort.SliceStable(list, func(i, j int) bool {
			if asc {
				return list[i].MemoryBytes < list[j].MemoryBytes
			}
			return list[i].MemoryBytes > list[j].MemoryBytes
		})
	case "name":
		sort.SliceStable(list, func(i, j int) bool {
			if asc {
				return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
			}
			return strings.ToLower(list[i].Name) > strings.ToLower(list[j].Name)
		})
	case "cpu":
		sort.SliceStable(list, func(i, j int) bool {
			if asc {
				return list[i].CPUPercent < list[j].CPUPercent
			}
			return list[i].CPUPercent > list[j].CPUPercent
		})
	}
}
