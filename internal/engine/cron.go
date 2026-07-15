package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"bedrock/internal/cicd/model"
)

// Clock abstracts time for cron overlap tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// CronScheduler schedules per-BuildJob cron with IANA timezone.
// Overlapping non-terminal runs are skipped; missed triggers during downtime are NOT catch-up
// (robfig/cron does not backfill missed ticks after Start).
type CronScheduler struct {
	cron      *cron.Cron
	entries   map[uint]cron.EntryID
	mu        sync.Mutex
	jobs      JobStore
	runs      RunStore
	enqueuer  RunEnqueuer
	scheduler RunScheduler
	logger    *zap.Logger
	clock     Clock
}

func NewCronScheduler(
	jobs JobStore,
	runs RunStore,
	enqueuer RunEnqueuer,
	scheduler RunScheduler,
	logger *zap.Logger,
) *CronScheduler {
	return &CronScheduler{
		cron:      cron.New(),
		entries:   make(map[uint]cron.EntryID),
		jobs:      jobs,
		runs:      runs,
		enqueuer:  enqueuer,
		scheduler: scheduler,
		logger:    logger,
		clock:     realClock{},
	}
}

// SetClock injects a clock (tests).
func (cs *CronScheduler) SetClock(c Clock) {
	if c != nil {
		cs.clock = c
	}
}

func (cs *CronScheduler) Start() error {
	list, err := cs.jobs.ListCronEnabled()
	if err != nil {
		return fmt.Errorf("load cron jobs: %w", err)
	}
	for _, job := range list {
		if err := cs.addEntry(job); err != nil {
			if cs.logger != nil {
				cs.logger.Warn("skip cron entry", zap.Uint("job_id", job.ID), zap.Error(err))
			}
		}
	}
	cs.cron.Start()
	return nil
}

func (cs *CronScheduler) Stop() {
	cs.cron.Stop()
}

func (cs *CronScheduler) Add(job model.BuildJob) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if entryID, ok := cs.entries[job.ID]; ok {
		cs.cron.Remove(entryID)
		delete(cs.entries, job.ID)
	}
	if !job.TriggerCron || job.CronExpression == "" || !job.Enabled {
		return nil
	}
	return cs.addEntryLocked(job)
}

func (cs *CronScheduler) Remove(jobID uint) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if entryID, ok := cs.entries[jobID]; ok {
		cs.cron.Remove(entryID)
		delete(cs.entries, jobID)
	}
}

func (cs *CronScheduler) Update(job model.BuildJob) error {
	return cs.Add(job)
}

func (cs *CronScheduler) addEntry(job model.BuildJob) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.addEntryLocked(job)
}

func (cs *CronScheduler) addEntryLocked(job model.BuildJob) error {
	jobID := job.ID
	tzName := strings.TrimSpace(job.CronTimezone)
	if tzName == "" {
		tzName = "UTC"
	}
	if _, err := time.LoadLocation(tzName); err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tzName, err)
	}

	// CRON_TZ=IANA embeds per-job timezone; no catch-up of missed ticks.
	spec := "CRON_TZ=" + tzName + " " + job.CronExpression
	entryID, err := cs.cron.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil && cs.logger != nil {
				cs.logger.Error("cron callback panic", zap.Uint("job_id", jobID), zap.Any("panic", r))
			}
		}()
		cs.trigger(jobID)
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression %q (tz=%s): %w", job.CronExpression, tzName, err)
	}

	cs.entries[jobID] = entryID
	return nil
}

func (cs *CronScheduler) trigger(jobID uint) {
	if cs.logger != nil {
		cs.logger.Info("cron triggered", zap.Uint("job_id", jobID), zap.Time("now", cs.clock.Now()))
	}
	active, err := cs.runs.HasNonTerminal(jobID)
	if err != nil {
		if cs.logger != nil {
			cs.logger.Error("cron overlap check failed", zap.Error(err))
		}
		return
	}
	if active {
		if cs.logger != nil {
			cs.logger.Info("cron skipped: non-terminal run exists", zap.Uint("job_id", jobID))
		}
		return
	}
	run, err := cs.enqueuer.EnqueueInternal(jobID, 0, EnqueueParams{TriggerType: "cron"})
	if err != nil {
		if cs.logger != nil {
			cs.logger.Error("cron enqueue failed", zap.Uint("job_id", jobID), zap.Error(err))
		}
		return
	}
	if cs.scheduler != nil {
		_ = cs.scheduler.Submit(run.ID)
	}
}

// TriggerNow is for tests: run the overlap/enqueue path immediately.
func (cs *CronScheduler) TriggerNow(jobID uint) {
	cs.trigger(jobID)
}
