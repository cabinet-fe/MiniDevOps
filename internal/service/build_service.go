package service

import (
	"buildflow/internal/model"
	"buildflow/internal/repository"
)

type BuildService struct {
	repo        *repository.BuildRepository
	projectRepo *repository.ProjectRepository
	envRepo     *repository.EnvironmentRepository
}

func NewBuildService(repo *repository.BuildRepository, projectRepo *repository.ProjectRepository, envRepo *repository.EnvironmentRepository) *BuildService {
	return &BuildService{repo: repo, projectRepo: projectRepo, envRepo: envRepo}
}

func (s *BuildService) TriggerBuild(projectID, environmentID, triggeredBy uint, triggerType, commitHash, commitMessage string) (*model.Build, error) {
	num, err := s.repo.GetNextBuildNumber(projectID)
	if err != nil {
		return nil, err
	}
	build := &model.Build{
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		BuildNumber:   num,
		Status:        "pending",
		TriggerType:   triggerType,
		TriggeredBy:   triggeredBy,
		CommitHash:    commitHash,
		CommitMessage: commitMessage,
	}
	if err := s.repo.Create(build); err != nil {
		return nil, err
	}
	return build, nil
}

func (s *BuildService) GetByID(id uint) (*model.Build, error) {
	return s.repo.FindByID(id)
}

func (s *BuildService) ListByProject(projectID uint, environmentID *uint, page, pageSize int) ([]model.Build, int64, error) {
	return s.repo.List(projectID, environmentID, page, pageSize)
}

func (s *BuildService) Cancel(id uint) error {
	return s.repo.UpdateStatus(id, "cancelled", nil)
}

// DashboardStats holds dashboard statistics.
type DashboardStats struct {
	TotalProjects int64   `json:"total_projects"`
	TodayBuilds   int64   `json:"today_builds"`
	SuccessRate   float64 `json:"success_rate"`
	ActiveCount   int     `json:"active_count"`
}

func (s *BuildService) GetDashboardStats() (*DashboardStats, error) {
	totalProjects, err := s.projectRepo.Count()
	if err != nil {
		return nil, err
	}
	todayBuilds, err := s.repo.CountToday()
	if err != nil {
		return nil, err
	}
	activeBuilds, err := s.repo.FindActiveBuilds()
	if err != nil {
		return nil, err
	}
	// Success rate: total builds with status success / total finished builds
	// Build repo doesn't have CountByStatus - we'd need to add it or compute from trend
	// For simplicity, use 0 if no data
	var successRate float64
	trend, err := s.repo.CountByStatusInDays(30)
	if err == nil && len(trend) > 0 {
		var total, success int64
		for _, r := range trend {
			total += r.Count
			if r.Status == "success" {
				success += r.Count
			}
		}
		if total > 0 {
			successRate = float64(success) / float64(total) * 100
		}
	}
	return &DashboardStats{
		TotalProjects: totalProjects,
		TodayBuilds:   todayBuilds,
		SuccessRate:   successRate,
		ActiveCount:   len(activeBuilds),
	}, nil
}

func (s *BuildService) GetActiveBuildsList() ([]model.Build, error) {
	return s.repo.FindActiveBuilds()
}

func (s *BuildService) GetRecentBuilds(limit int) ([]model.Build, error) {
	return s.repo.GetRecentBuilds(limit)
}

// BuildTrendItem represents build counts by day and status.
type BuildTrendItem struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

func (s *BuildService) GetBuildTrend(days int) ([]BuildTrendItem, error) {
	results, err := s.repo.CountByStatusInDays(days)
	if err != nil {
		return nil, err
	}
	items := make([]BuildTrendItem, 0, len(results))
	for _, r := range results {
		items = append(items, BuildTrendItem{
			Date:   r.Date,
			Status: r.Status,
			Count:  r.Count,
		})
	}
	return items, nil
}
