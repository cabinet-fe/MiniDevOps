package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"buildflow/internal/model"
)

type Scheduler struct {
	maxConcurrent int
	semaphore     chan struct{}
	jobs          chan uint // build IDs
	cancelMap     map[uint]context.CancelFunc
	mu            sync.RWMutex
	pipeline      *Pipeline
	logger        *zap.Logger
	wg            sync.WaitGroup
	done          chan struct{}
	closed        atomic.Bool
}

func NewScheduler(maxConcurrent int, pipeline *Pipeline, logger *zap.Logger) *Scheduler {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Scheduler{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
		jobs:          make(chan uint, 100),
		cancelMap:     make(map[uint]context.CancelFunc),
		pipeline:      pipeline,
		logger:        logger,
		done:          make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go s.run()
}

func (s *Scheduler) run() {
	for buildID := range s.jobs {
		s.semaphore <- struct{}{}
		s.wg.Add(1)
		go func(id uint) {
			defer func() {
				<-s.semaphore
				s.wg.Done()
			}()
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("worker panic recovered, marking build as failed",
						zap.Uint("build_id", id),
						zap.String("panic", fmt.Sprint(r)),
					)
					s.pipeline.failBuild(&model.Build{ID: id}, fmt.Sprintf("internal panic: %v", r))
				}
			}()
			ctx, cancel := context.WithCancel(context.Background())
			s.mu.Lock()
			s.cancelMap[id] = cancel
			s.mu.Unlock()
			defer func() {
				s.mu.Lock()
				delete(s.cancelMap, id)
				s.mu.Unlock()
				cancel()
			}()
			s.pipeline.Execute(ctx, id)
		}(buildID)
	}
}

func (s *Scheduler) Submit(buildID uint) error {
	if s.closed.Load() {
		return fmt.Errorf("scheduler is shut down")
	}
	s.jobs <- buildID
	return nil
}

func (s *Scheduler) Cancel(buildID uint) bool {
	s.mu.RLock()
	cancel, ok := s.cancelMap[buildID]
	s.mu.RUnlock()
	if ok {
		cancel()
		return true
	}
	return false
}

func (s *Scheduler) Shutdown() {
	s.closed.Store(true)
	close(s.jobs)
	s.wg.Wait()
	close(s.done)
}
