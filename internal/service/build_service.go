package service

import (
	"bufio"
	"buildflow/internal/config"
	"buildflow/internal/model"
	"buildflow/internal/repository"
	"os"
	"sort"
	"strings"
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
	if s.projectRepo != nil {
		if project, err := s.projectRepo.FindByID(build.ProjectID); err == nil {
			resp.ProjectName = project.Name
		}
	}
	if s.envRepo != nil {
		if env, err := s.envRepo.FindByID(build.EnvironmentID); err == nil {
			resp.EnvironmentName = env.Name
			if resp.Branch == "" {
				resp.Branch = env.Branch
			}
		}
	}
	if resp.CurrentStage == "" || ((resp.Status == "failed" || resp.Status == "cancelled") && resp.CurrentStage == "pending") {
		if inferredStage := inferBuildStageFromLog(resp.LogPath); inferredStage != "" {
			resp.CurrentStage = inferredStage
		}
	}
	if resp.CurrentStage == "" {
		if resp.Status == "failed" || resp.Status == "cancelled" {
			resp.CurrentStage = "pending"
		} else {
			resp.CurrentStage = resp.Status
		}
	}
	if build.TriggeredBy > 0 && s.userRepo != nil {
		if user, err := s.userRepo.FindByID(build.TriggeredBy); err == nil {
			resp.TriggeredByName = user.DisplayName
			if resp.TriggeredByName == "" {
				resp.TriggeredByName = user.Username
			}
		}
	}
	return resp, nil
}

func inferBuildStageFromLog(logPath string) string {
	if logPath == "" {
		return ""
	}

	file, err := os.Open(logPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	stage := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "=== Stage: Cloning ==="):
			stage = "cloning"
		case strings.Contains(line, "=== Stage: Building ==="):
			stage = "building"
		case strings.Contains(line, "=== Stage: Deploying ==="):
			stage = "deploying"
		case strings.Contains(line, "=== Build completed successfully ==="):
			stage = "success"
		}
	}

	return stage
}

func (s *BuildService) TriggerBuild(projectID, environmentID, triggeredBy uint, triggerType, branch, commitHash, commitMessage string) (*model.Build, error) {
	if branch == "" {
		env, err := s.envRepo.FindByID(environmentID)
		if err != nil {
			return nil, err
		}
		branch = env.Branch
	}

	num, err := s.repo.GetNextBuildNumber(projectID)
	if err != nil {
		return nil, err
	}
	build := &model.Build{
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		BuildNumber:   num,
		Status:        "pending",
		CurrentStage:  "pending",
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
	TotalProjects   int64                    `json:"total_projects"`
	TodayBuilds     int64                    `json:"today_builds"`
	SuccessRate     float64                  `json:"success_rate"`
	ActiveCount     int                      `json:"active_count"`
	TagSummary      []ProjectTagSummary      `json:"tag_summary"`
	SystemResources DashboardSystemResources `json:"system_resources"`
}

type ProjectTagSummary struct {
	Tag              string `json:"tag"`
	ProjectCount     int    `json:"project_count"`
	EnvironmentCount int    `json:"environment_count"`
}

type DashboardSystemResources struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsedBytes    uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes   uint64  `json:"memory_total_bytes"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	AppMemoryUsedBytes uint64  `json:"app_memory_used_bytes"`
	DiskFreeBytes      uint64  `json:"disk_free_bytes"`
	DiskTotalBytes     uint64  `json:"disk_total_bytes"`
	DiskUsagePercent   float64 `json:"disk_usage_percent"`
}

type DashboardBuildItem struct {
	model.Build
	ProjectName     string `json:"project_name"`
	EnvironmentName string `json:"environment_name"`
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
	activeCount64, err := s.repo.CountActive()
	if err != nil {
		return nil, err
	}
	var successRate float64
	if success, total, err := s.repo.CountSuccessRateInDays(30); err == nil && total > 0 {
		successRate = float64(success) / float64(total) * 100
	}
	projRows, err := s.projectRepo.ListProjectTagsWithEnvCounts(nil)
	if err != nil {
		return nil, err
	}
	tagStats := make(map[string]*ProjectTagSummary)
	for _, row := range projRows {
		tags := strings.Split(row.Tags, ",")
		if row.Tags == "" {
			tags = []string{"未标记"}
		}
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, exists := tagStats[tag]; !exists {
				tagStats[tag] = &ProjectTagSummary{Tag: tag}
			}
			tagStats[tag].ProjectCount++
			tagStats[tag].EnvironmentCount += int(row.EnvCount)
		}
	}
	tagSummary := make([]ProjectTagSummary, 0, len(tagStats))
	for _, item := range tagStats {
		tagSummary = append(tagSummary, *item)
	}
	sort.Slice(tagSummary, func(i, j int) bool {
		return tagSummary[i].Tag < tagSummary[j].Tag
	})

	diskPath := "."
	if config.C != nil && config.C.Build.WorkspaceDir != "" {
		diskPath = config.C.Build.WorkspaceDir
	}

	return &DashboardStats{
		TotalProjects:   totalProjects,
		TodayBuilds:     todayBuilds,
		SuccessRate:     successRate,
		ActiveCount:     int(activeCount64),
		TagSummary:      tagSummary,
		SystemResources: collectDashboardSystemResources(diskPath),
	}, nil
}

func (s *BuildService) GetDashboardSystemResources() DashboardSystemResources {
	diskPath := "."
	if config.C != nil && config.C.Build.WorkspaceDir != "" {
		diskPath = config.C.Build.WorkspaceDir
	}

	return collectDashboardSystemResources(diskPath)
}

func (s *BuildService) GetActiveBuildsList() ([]DashboardBuildItem, error) {
	builds, err := s.repo.FindActiveBuilds()
	if err != nil {
		return nil, err
	}
	return s.decorateDashboardBuilds(builds), nil
}

func (s *BuildService) GetRecentBuilds(limit int) ([]DashboardBuildItem, error) {
	builds, err := s.repo.GetRecentBuilds(limit)
	if err != nil {
		return nil, err
	}
	return s.decorateDashboardBuilds(builds), nil
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

func (s *BuildService) decorateDashboardBuilds(builds []model.Build) []DashboardBuildItem {
	items := make([]DashboardBuildItem, 0, len(builds))
	projectNames := make(map[uint]string, len(builds))
	environmentNames := make(map[uint]string, len(builds))

	for _, build := range builds {
		item := DashboardBuildItem{Build: build}

		if name, ok := projectNames[build.ProjectID]; ok {
			item.ProjectName = name
		} else if s.projectRepo != nil {
			if project, err := s.projectRepo.FindByID(build.ProjectID); err == nil {
				item.ProjectName = project.Name
				projectNames[build.ProjectID] = project.Name
			}
		}

		if name, ok := environmentNames[build.EnvironmentID]; ok {
			item.EnvironmentName = name
		} else if s.envRepo != nil {
			if env, err := s.envRepo.FindByID(build.EnvironmentID); err == nil {
				item.EnvironmentName = env.Name
				environmentNames[build.EnvironmentID] = env.Name
			}
		}

		items = append(items, item)
	}

	return items
}
