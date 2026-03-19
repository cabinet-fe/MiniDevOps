package engine

import (
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"buildflow/internal/model"
	"buildflow/internal/repository"
)

// CronScheduler manages cron-based timed builds for environments.
type CronScheduler struct {
	cron        *cron.Cron
	entries     map[uint]cron.EntryID // envID -> cron entry ID
	mu          sync.Mutex
	envRepo     *repository.EnvironmentRepository
	buildRepo   *repository.BuildRepository
	scheduler   *Scheduler
	logger      *zap.Logger
}

// NewCronScheduler creates a new CronScheduler.
func NewCronScheduler(
	envRepo *repository.EnvironmentRepository,
	buildRepo *repository.BuildRepository,
	scheduler *Scheduler,
	logger *zap.Logger,
) *CronScheduler {
	return &CronScheduler{
		cron:      cron.New(),
		entries:   make(map[uint]cron.EntryID),
		envRepo:   envRepo,
		buildRepo: buildRepo,
		scheduler: scheduler,
		logger:    logger,
	}
}

// Start loads all enabled cron environments and starts the cron scheduler.
func (cs *CronScheduler) Start() error {
	envs, err := cs.envRepo.ListCronEnabled()
	if err != nil {
		return fmt.Errorf("load cron environments: %w", err)
	}
	for _, env := range envs {
		if err := cs.addEntry(env); err != nil {
			cs.logger.Warn("skip cron entry", zap.Uint("env_id", env.ID), zap.Error(err))
		}
	}
	cs.cron.Start()
	return nil
}

// Stop stops the cron scheduler.
func (cs *CronScheduler) Stop() {
	cs.cron.Stop()
}

// Add registers a cron entry for the given environment.
func (cs *CronScheduler) Add(env model.Environment) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	// Remove existing entry if any
	if entryID, ok := cs.entries[env.ID]; ok {
		cs.cron.Remove(entryID)
		delete(cs.entries, env.ID)
	}
	if !env.CronEnabled || env.CronExpression == "" {
		return nil
	}
	return cs.addEntry(env)
}

// Remove removes the cron entry for the given environment ID.
func (cs *CronScheduler) Remove(envID uint) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if entryID, ok := cs.entries[envID]; ok {
		cs.cron.Remove(entryID)
		delete(cs.entries, envID)
	}
}

// Update updates the cron entry for the given environment.
func (cs *CronScheduler) Update(env model.Environment) error {
	return cs.Add(env)
}

func (cs *CronScheduler) addEntry(env model.Environment) error {
	envID := env.ID
	projectID := env.ProjectID
	expression := env.CronExpression

	entryID, err := cs.cron.AddFunc(expression, func() {
		cs.logger.Info("cron triggered build",
			zap.Uint("env_id", envID),
			zap.Uint("project_id", projectID),
		)
		num, err := cs.buildRepo.GetNextBuildNumber(projectID)
		if err != nil {
			cs.logger.Error("cron: get next build number", zap.Error(err))
			return
		}
		build := &model.Build{
			ProjectID:     projectID,
			EnvironmentID: envID,
			BuildNumber:   num,
			Status:        "pending",
			TriggerType:   "cron",
			TriggeredBy:   0, // system trigger
		}
		if err := cs.buildRepo.Create(build); err != nil {
			cs.logger.Error("cron: create build", zap.Error(err))
			return
		}
		cs.scheduler.Submit(build.ID)
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", expression, err)
	}

	cs.mu.Lock()
	cs.entries[envID] = entryID
	cs.mu.Unlock()

	return nil
}
