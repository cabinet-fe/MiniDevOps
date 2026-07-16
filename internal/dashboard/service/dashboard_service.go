package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	gomem "github.com/shirou/gopsutil/v4/mem"
	"gorm.io/gorm"

	"bedrock/internal/dashboard/model"
	"bedrock/internal/dashboard/repository"
	"bedrock/internal/rbac"
)

const (
	CardBuildSummary = "build_summary"
	CardSystemInfo   = "system_info"
	CardSystemStatus = "system_status"
)

var ErrUnauthorizedCard = errors.New("仪表盘包含无权限卡片")

type DashboardService struct {
	repo      *repository.DashboardRepository
	version   string
	startTime time.Time
	diskPaths []string
}

func NewDashboardService(repo *repository.DashboardRepository, version string, startTime time.Time, diskPaths []string) *DashboardService {
	paths := make([]string, 0, len(diskPaths))
	for _, path := range diskPaths {
		if path != "" {
			paths = append(paths, path)
		}
	}
	if len(paths) == 0 {
		paths = append(paths, ".")
	}
	return &DashboardService{repo: repo, version: version, startTime: startTime.UTC(), diskPaths: paths}
}

func (s *DashboardService) GetLayout(userID uint, isSuperAdmin bool, permissions []string) (*model.LayoutResponse, error) {
	allowed := allowedCards(isSuperAdmin, permissions)
	layout, err := s.repo.FindLayoutByUserID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.LayoutResponse{Cards: defaultLayout(allowed)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find dashboard layout: %w", err)
	}

	cards, err := decodeCards(layout.CardsJSON)
	if err != nil {
		// A corrupted old preference should never make the dashboard unusable.
		return &model.LayoutResponse{Cards: defaultLayout(allowed)}, nil
	}
	return &model.LayoutResponse{Cards: normalizeCards(cards, allowed)}, nil
}

func (s *DashboardService) PutLayout(userID uint, isSuperAdmin bool, permissions []string, cards []model.CardLayout) (*model.LayoutResponse, error) {
	allowed := allowedCards(isSuperAdmin, permissions)
	if err := validateCards(cards, allowed); err != nil {
		return nil, err
	}
	normalized := normalizeCards(cards, allowed)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("encode dashboard layout: %w", err)
	}
	layout, err := s.repo.FindLayoutByUserID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.repo.CreateLayout(&model.Layout{UserID: userID, CardsJSON: string(raw)}); err != nil {
			return nil, fmt.Errorf("create dashboard layout: %w", err)
		}
		return &model.LayoutResponse{Cards: normalized}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find dashboard layout: %w", err)
	}
	layout.CardsJSON = string(raw)
	if err := s.repo.UpdateLayout(layout); err != nil {
		return nil, fmt.Errorf("update dashboard layout: %w", err)
	}
	return &model.LayoutResponse{Cards: normalized}, nil
}

func (s *DashboardService) BuildSummary() (*model.BuildSummary, error) {
	running, err := s.repo.CountRunsByStatus("running")
	if err != nil {
		return nil, err
	}
	queued, err := s.repo.CountRunsByStatus("queued")
	if err != nil {
		return nil, err
	}
	total, success, err := s.repo.CountFinishedRuns()
	if err != nil {
		return nil, err
	}
	recent, err := s.repo.ListRecentRuns(8)
	if err != nil {
		return nil, err
	}
	rate := float64(0)
	if total > 0 {
		rate = float64(success) * 100 / float64(total)
	}
	return &model.BuildSummary{Running: running, Queued: queued, SuccessRate: rate, Recent: recent}, nil
}

func (s *DashboardService) SystemInfo() (*model.SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	return &model.SystemInfo{
		Version: s.version, OS: runtime.GOOS, Arch: runtime.GOARCH,
		Runtime: runtime.Version(), Hostname: hostname, StartTime: s.startTime,
	}, nil
}

func (s *DashboardService) SystemStatus() (*model.SystemStatus, error) {
	result := &model.SystemStatus{Health: "ok", CollectedAt: time.Now().UTC()}
	if samples, err := cpu.Percent(0, false); err == nil && len(samples) > 0 {
		result.CPUUsagePercent = roundSingleDecimal(samples[0])
	}
	if vm, err := gomem.VirtualMemory(); err == nil {
		result.MemoryUsedBytes = vm.Used
		result.MemoryTotalBytes = vm.Total
		result.MemoryUsagePercent = roundSingleDecimal(vm.UsedPercent)
	}
	for _, path := range s.diskPaths {
		usage, err := disk.Usage(path)
		if err != nil {
			result.Health = "degraded"
			continue
		}
		result.Directories = append(result.Directories, model.DiskStatus{
			Path: path, TotalBytes: usage.Total, FreeBytes: usage.Free, UsedPct: roundSingleDecimal(usage.UsedPercent),
		})
	}
	return result, nil
}

func allowedCards(isSuperAdmin bool, permissions []string) map[string]struct{} {
	allowed := map[string]struct{}{}
	if isSuperAdmin || hasPermission(permissions, "cicd.build_runs:view") {
		allowed[CardBuildSummary] = struct{}{}
	}
	if isSuperAdmin || hasPermission(permissions, "dashboard.system_info:view") {
		allowed[CardSystemInfo] = struct{}{}
	}
	if isSuperAdmin || hasPermission(permissions, "dashboard.system_status:view") {
		allowed[CardSystemStatus] = struct{}{}
	}
	return allowed
}

func hasPermission(codes []string, required string) bool {
	return rbac.HasPermission(rbac.ToSet(codes), required)
}

func defaultLayout(allowed map[string]struct{}) []model.CardLayout {
	all := []string{CardBuildSummary, CardSystemInfo, CardSystemStatus}
	cards := make([]model.CardLayout, 0, len(all))
	for _, id := range all {
		if _, ok := allowed[id]; ok {
			cards = append(cards, model.CardLayout{ID: id, Visible: true, Order: len(cards)})
		}
	}
	return cards
}

func normalizeCards(cards []model.CardLayout, allowed map[string]struct{}) []model.CardLayout {
	byID := map[string]model.CardLayout{}
	for _, card := range cards {
		if _, ok := allowed[card.ID]; ok {
			byID[card.ID] = card
		}
	}
	for _, defaultCard := range defaultLayout(allowed) {
		if _, ok := byID[defaultCard.ID]; !ok {
			byID[defaultCard.ID] = defaultCard
		}
	}
	out := make([]model.CardLayout, 0, len(byID))
	for _, card := range byID {
		out = append(out, card)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].ID < out[j].ID
	})
	for i := range out {
		out[i].Order = i
	}
	return out
}

func validateCards(cards []model.CardLayout, allowed map[string]struct{}) error {
	seen := map[string]struct{}{}
	for _, card := range cards {
		if _, ok := allowed[card.ID]; !ok {
			return ErrUnauthorizedCard
		}
		if _, duplicate := seen[card.ID]; duplicate {
			return fmt.Errorf("重复卡片: %s", card.ID)
		}
		seen[card.ID] = struct{}{}
	}
	return nil
}

func decodeCards(raw string) ([]model.CardLayout, error) {
	var cards []model.CardLayout
	if err := json.Unmarshal([]byte(raw), &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func roundSingleDecimal(value float64) float64 {
	return float64(int(value*10+0.5)) / 10
}
