package engine

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewScheduler_MinConcurrent_File(t *testing.T) {
	// Covered in pipeline_test.go; keep file for package layout.
	_ = zap.NewNop()
}
