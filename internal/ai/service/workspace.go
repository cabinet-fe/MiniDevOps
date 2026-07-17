package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"bedrock/internal/ai/model"
	cicdmodel "bedrock/internal/cicd/model"
)

// BuildJobFinder loads build jobs for agent workspace softlinks.
type BuildJobFinder interface {
	FindByID(id uint) (*cicdmodel.BuildJob, error)
}

func (s *AgentService) SetBuildJobFinder(f BuildJobFinder) { s.jobs = f }

func (s *AgentService) agentRoot(agentID uint) string {
	return filepath.Join(s.workDir, "agents", fmt.Sprintf("agent-%d", agentID))
}

// SyncAgentWorkspace ensures the persistent agent directory layout:
// skills under .agents/skills, job-* softlinks to build-job workspaces, SYSTEM_PROMPT.md.
// jobDirs are absolute paths of successfully linked build-job workspaces (for run logs).
func (s *AgentService) SyncAgentWorkspace(agent *model.AiAgent, userID uint, isSuperAdmin bool) (digests map[uint]string, jobDirs []string, err error) {
	if agent == nil {
		return nil, nil, fmt.Errorf("agent is nil")
	}
	decodeSkillIDs(agent)
	decodeBuildJobIDs(agent)

	root := s.agentRoot(agent.ID)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, nil, err
	}

	skillsRoot := filepath.Join(root, ".agents", "skills")
	_ = os.RemoveAll(skillsRoot)
	digests = map[uint]string{}
	if s.skills != nil {
		digests, err = s.skills.InjectSkills(root, agent.SkillIDs, userID, isSuperAdmin)
		if err != nil {
			return nil, nil, err
		}
	}

	jobDirs, err = s.rebuildJobLinks(root, agent.BuildJobIDs)
	if err != nil {
		return nil, nil, err
	}
	// Drop legacy per-job OpenCode external_directory configs from the prior approach.
	_ = os.Remove(filepath.Join(root, "opencode.json"))

	promptPath := filepath.Join(root, "SYSTEM_PROMPT.md")
	if err := os.WriteFile(promptPath, []byte(agent.SystemPrompt), 0o644); err != nil {
		return nil, nil, err
	}
	return digests, jobDirs, nil
}

func (s *AgentService) rebuildJobLinks(agentRoot string, jobIDs []uint) ([]string, error) {
	entries, err := os.ReadDir(agentRoot)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "job-") {
			continue
		}
		_ = os.Remove(filepath.Join(agentRoot, name))
	}
	if len(jobIDs) == 0 {
		return nil, nil
	}
	if s.jobs == nil {
		if s.logger != nil {
			s.logger.Warn("build job finder not configured; skipping agent job softlinks")
		}
		return nil, nil
	}
	var linked []string
	for _, id := range jobIDs {
		job, err := s.jobs.FindByID(id)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("skip agent job softlink: job not found",
					zap.Uint("build_job_id", id), zap.Error(err))
			}
			continue
		}
		target := filepath.Join(s.workDir, fmt.Sprintf("repo-%d", job.RepositoryID), fmt.Sprintf("job-%d", job.ID))
		absTarget, err := filepath.Abs(target)
		if err != nil {
			absTarget = target
		}
		if _, err := os.Stat(absTarget); err != nil {
			if s.logger != nil {
				s.logger.Warn("skip agent job softlink: workspace missing",
					zap.Uint("build_job_id", id), zap.String("target", absTarget), zap.Error(err))
			}
			continue
		}
		link := filepath.Join(agentRoot, fmt.Sprintf("job-%d", id))
		if err := os.Symlink(absTarget, link); err != nil {
			if s.logger != nil {
				s.logger.Warn("failed to create agent job softlink",
					zap.Uint("build_job_id", id), zap.String("link", link), zap.Error(err))
			}
			continue
		}
		linked = append(linked, absTarget)
	}
	return linked, nil
}

// appendFullPermissionArgs enables each CLI's broad / bypass-sandbox mode so
// job-* softlinks under the agent workspace can be followed. Scope is enforced
// via prompt splicing (agentWorkspaceScopeHint), not per-directory allow lists.
func appendFullPermissionArgs(cliKey string, args []string) []string {
	switch cliKey {
	case "claude_code", "opencode", "reasonix":
		return append(args, "--dangerously-skip-permissions")
	case "codex":
		return append(args, "--dangerously-bypass-approvals-and-sandbox")
	default:
		return args
	}
}

// agentWorkspaceScopeHint asks the CLI to stay inside BEDROCK_AGENT_WORKDIR
// (softlinks like ./job-{id} are in scope) and write deliverables to OUTPUT.
func agentWorkspaceScopeHint() string {
	return "你的工作目录是 $BEDROCK_AGENT_WORKDIR（agents 下本智能体目录）。" +
		"只能在该目录内读写；通过 ./job-{id} 软链访问构建任务代码。" +
		"禁止访问该目录之外的任意路径。" +
		" Your working directory is $BEDROCK_AGENT_WORKDIR (this agent under agents/)." +
		" Read/write only inside it; access bound build-job code via ./job-{id} softlinks." +
		" Do not access any path outside this directory." +
		" Write deliverable files into $BEDROCK_AGENT_OUTPUT."
}

func (s *AgentService) removeAgentWorkspace(agentID uint) {
	_ = os.RemoveAll(s.agentRoot(agentID))
}

func dirHasRegularFiles(dir string) (bool, error) {
	has := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info != nil && info.Mode().IsRegular() {
			has = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return has, nil
}

func encodeBuildJobIDs(agent *model.AiAgent, ids []uint) error {
	if ids == nil {
		ids = []uint{}
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	agent.BuildJobIDsJSON = string(b)
	agent.BuildJobIDs = ids
	return nil
}

func decodeBuildJobIDs(agent *model.AiAgent) {
	if agent.BuildJobIDsJSON == "" {
		agent.BuildJobIDs = []uint{}
		return
	}
	_ = json.Unmarshal([]byte(agent.BuildJobIDsJSON), &agent.BuildJobIDs)
	if agent.BuildJobIDs == nil {
		agent.BuildJobIDs = []uint{}
	}
}
