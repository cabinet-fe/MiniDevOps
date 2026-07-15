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
		{"fixed hit", model.BuildJob{BranchPolicy: "fixed", Branch: "main"}, "main", true},
		{"fixed miss", model.BuildJob{BranchPolicy: "fixed", Branch: "main"}, "develop", false},
		{"empty policy defaults fixed", model.BuildJob{Branch: "main"}, "main", true},
		{"empty policy miss", model.BuildJob{Branch: "main"}, "other", false},
		{"param any branch", model.BuildJob{BranchPolicy: "param", Branch: "main"}, "feature/x", true},
		{"param case insensitive", model.BuildJob{BranchPolicy: "PARAM", Branch: "main"}, "develop", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jobMatchesBranch(tc.job, tc.branch); got != tc.want {
				t.Fatalf("jobMatchesBranch(%+v, %q)=%v want %v", tc.job, tc.branch, got, tc.want)
			}
		})
	}
}
