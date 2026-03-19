package service

import (
	"buildflow/internal/model"
	"buildflow/internal/repository"
	"sort"
)

type BuildService struct {
	repo        *repository.BuildRepository
	projectRepo *repository.ProjectRepository
	envRepo     *repository.EnvironmentRepository
	userRepo    *repository.UserRepository
}

func NewBuildService(repo *repository.BuildRepository, projectRepo *repository.ProjectRepository, envRepo *repository.EnvironmentRepository, userRepo *repository.UserRepository) *BuildService {
	return &BuildService{repo: repo, projectRepo: projectRepo, envRepo: envRepo, userRepo: userRepo}
}

// BuildDetailResponse extends Build with associated names.
type BuildDetailResponse struct {
	model.Build
	ProjectName     string `json:"project_name"`
	EnvironmentName string `json:"environment_name"`
	TriggeredByName string `json:"triggered_by_name"`
}

func (s *BuildService) GetBuildDetail(id uint) (*BuildDetailResponse, error) {
	build, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := &BuildDetailResponse{Build: *build}
	if project, err := s.projectRepo.FindByID(build.ProjectID); err == nil {
		resp.ProjectName = project.Name
	}
	if env, err := s.envRepo.FindByID(build.EnvironmentID); err == nil {
		resp.EnvironmentName = env.Name
	}
	if build.TriggeredBy > 0 {
		if user, err := s.userRepo.FindByID(build.TriggeredBy); err == nil {
			resp.TriggeredByName = user.DisplayName
			if resp.TriggeredByName == "" {
				resp.TriggeredByName = user.Username
			}
		}
	}
	return resp, nil
}

func (s *BuildService) TriggerBuild(projectID, environmentID, triggeredBy uint, triggerType, branch, commitHash, commitMessage string) (*model.Build, error) {
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
		Branch:        branch,
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

// BuildListItem extends Build with project and environment names for list display.
type BuildListItem struct {
	model.Build
	ProjectName     string `json:"project_name"`
	EnvironmentName string `json:"environment_name"`
}

func (s *BuildService) ListAll(page, pageSize int) ([]BuildListItem, int64, error) {
	builds, total, err := s.repo.ListAll(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	items := make([]BuildListItem, 0, len(builds))
	for _, b := range builds {
		item := BuildListItem{Build: b}
		if project, err := s.projectRepo.FindByID(b.ProjectID); err == nil {
			item.ProjectName = project.Name
		}
		if env, err := s.envRepo.FindByID(b.EnvironmentID); err == nil {
			item.EnvironmentName = env.Name
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (s *BuildService) Cancel(id uint) error {
	return s.repo.UpdateStatus(id, "cancelled", nil)
}

// DashboardStats holds dashboard statistics.
type DashboardStats struct {
	TotalProjects int64                 `json:"total_projects"`
	TodayBuilds   int64                 `json:"today_builds"`
	SuccessRate   float64               `json:"success_rate"`
	ActiveCount   int                   `json:"active_count"`
	GroupSummary  []ProjectGroupSummary `json:"group_summary"`
}

type ProjectGroupSummary struct {
	GroupName        string `json:"group_name"`
	ProjectCount     int    `json:"project_count"`
	EnvironmentCount int    `json:"environment_count"`
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
	projects, err := s.projectRepo.ListAll(nil)
	if err != nil {
		return nil, err
	}
	groupStats := make(map[string]*ProjectGroupSummary)
	for _, project := range projects {
		groupName := project.GroupName
		if groupName == "" {
			groupName = "未分组"
		}
		if _, exists := groupStats[groupName]; !exists {
			groupStats[groupName] = &ProjectGroupSummary{GroupName: groupName}
		}
		groupStats[groupName].ProjectCount++
		groupStats[groupName].EnvironmentCount += len(project.Environments)
	}
	groupSummary := make([]ProjectGroupSummary, 0, len(groupStats))
	for _, item := range groupStats {
		groupSummary = append(groupSummary, *item)
	}
	sort.Slice(groupSummary, func(i, j int) bool {
		return groupSummary[i].GroupName < groupSummary[j].GroupName
	})
	return &DashboardStats{
		TotalProjects: totalProjects,
		TodayBuilds:   todayBuilds,
		SuccessRate:   successRate,
		ActiveCount:   len(activeBuilds),
		GroupSummary:  groupSummary,
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
