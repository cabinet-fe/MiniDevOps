package service

import (
	"testing"

	"bedrock/internal/cicd/model"
)

func TestJobMatchesBranch(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		job    model.BuildJob
		branch string
		want   bool
	}{
		{"hit", model.BuildJob{Branch: "main"}, "main", true},
		{"miss", model.BuildJob{Branch: "main"}, "develop", false},
		{"empty branch miss", model.BuildJob{Branch: ""}, "main", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jobMatchesBranch(tc.job, tc.branch); got != tc.want {
				t.Fatalf("jobMatchesBranch(%+v, %q)=%v want %v", tc.job, tc.branch, got, tc.want)
			}
		})
	}
}
