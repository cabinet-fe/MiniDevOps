package engine

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewScheduler_MinConcurrent(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop()
	s := NewScheduler(0, &Pipeline{}, logger)
	if s.maxConcurrent != 1 {
		t.Fatalf("maxConcurrent: got %d want 1", s.maxConcurrent)
	}
}
