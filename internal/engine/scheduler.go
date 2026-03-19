package engine

import (
	"context"
	"sync"

	"go.uber.org/zap"
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
}

func NewScheduler(maxConcurrent int, pipeline *Pipeline, logger *zap.Logger) *Scheduler {
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

func (s *Scheduler) Submit(buildID uint) {
	s.jobs <- buildID
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
	close(s.jobs)
	s.wg.Wait()
	close(s.done)
}
