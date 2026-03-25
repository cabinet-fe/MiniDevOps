package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buildflow/internal/deployer"
	"buildflow/internal/model"
)

func (p *Pipeline) updateStageKeepSuccess(build *model.Build, stage string) {
	build.CurrentStage = stage
	p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{"current_stage": stage})
}

func (p *Pipeline) deployOneDistribution(ctx context.Context, build *model.Build, project *model.Project, d *model.Distribution, sourceDir string, writeLine func(string)) error {
	method := strings.TrimSpace(strings.ToLower(d.Method))
	if method == "" {
		method = "rsync"
	}
	isLocal := method == "local"
	deployPath := strings.TrimSpace(d.RemotePath)

	if isLocal {
		if deployPath == "" {
			return fmt.Errorf("分发未配置路径")
		}
		if !filepath.IsAbs(deployPath) {
			return fmt.Errorf("本机分发路径须为绝对路径")
		}
	} else {
		if d.ServerID == nil || deployPath == "" {
			return fmt.Errorf("分发未配置服务器或路径")
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	deployOpts := deployer.DeployOptions{
		SourceDir:     sourceDir,
		ArchiveFormat: normalizeArtifactFormat(project.ArtifactFormat),
		RemotePath:    deployPath,
		Logger:        writeLine,
	}
	if !isLocal {
		server, err := p.serverRepo.FindByID(*d.ServerID)
		if err != nil {
			return fmt.Errorf("服务器不存在")
		}
		password, privateKey, agentToken := decryptServerSecrets(server)
		deployOpts.Server = deployer.ServerInfo{
			Host:       server.Host,
			Port:       server.Port,
			OSType:     server.OSType,
			Username:   server.Username,
			AuthType:   server.AuthType,
			Password:   password,
			PrivateKey: privateKey,
			AgentURL:   server.AgentURL,
			AgentToken: agentToken,
		}
	}

	dpl := deployer.NewDeployer(method)
	if err := dpl.Deploy(ctx, deployOpts); err != nil {
		return fmt.Errorf("分发失败: %w", err)
	}
	writeLine("Distribution completed successfully")

	if strings.TrimSpace(d.PostDeployScript) != "" {
		writeLine("=== Executing post-deploy script ===")
		var err error
		if isLocal {
			err = deployer.ExecuteLocalScriptInDir(ctx, deployOpts.RemotePath, d.PostDeployScript, writeLine)
		} else {
			err = deployer.ExecuteRemoteScriptInDir(ctx, deployOpts.Server, deployOpts.RemotePath, d.PostDeployScript, writeLine)
		}
		if err != nil {
			return fmt.Errorf("部署后脚本失败: %w", err)
		}
		writeLine("Post-deploy script completed")
	}
	return nil
}

func parseRedistributeFilterJSON(s string) []uint {
	s = strings.TrimSpace(s)
	if s == "" || s == "[]" || s == "null" {
		return nil
	}
	var ids []uint
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	return ids
}

func filterDistributions(all []model.Distribution, ids []uint) []model.Distribution {
	if len(ids) == 0 {
		return all
	}
	want := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		want[id] = struct{}{}
	}
	out := make([]model.Distribution, 0, len(all))
	for _, d := range all {
		if _, ok := want[d.ID]; ok {
			out = append(out, d)
		}
	}
	return out
}

// runDistributions executes each distribution; does not change Build.Status away from success.
func (p *Pipeline) runDistributions(ctx context.Context, build *model.Build, project *model.Project, env *model.Environment, sourceDir string, writeLine func(string), filterIDs []uint) {
	dists, err := p.distRepo.ListByEnvironmentID(env.ID)
	if err != nil {
		writeLine("ERROR: load distributions: " + err.Error())
		p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"current_stage":          "success",
			"distribution_summary":   "all_failed",
			"redistribute_filter_json": "",
		})
		return
	}
	dists = filterDistributions(dists, filterIDs)
	if len(dists) == 0 {
		writeLine("=== No distribution targets (skipped) ===")
		p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"current_stage":          "success",
			"distribution_summary":   "skipped",
			"redistribute_filter_json": "",
		})
		return
	}

	p.updateStageKeepSuccess(build, "distributing")
	p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
		"distribution_summary": "running",
	})

	writeLine("=== Stage: Distributing ===")
	var nOK, nFail int
	for i := range dists {
		if ctx.Err() != nil {
			for j := i; j < len(dists); j++ {
				p.recordDistributionCancelled(build, dists[j].ID)
			}
			writeLine("ERROR: cancelled")
			break
		}
		d := dists[i]
		row := &model.BuildDistribution{
			BuildID:        build.ID,
			DistributionID: d.ID,
			Status:         "running",
			StartedAt:      ptrTime(time.Now()),
		}
		_ = p.buildDistRepo.Upsert(row)
		writeLine(fmt.Sprintf("--- Distribution #%d (%s → %s) ---", d.ID, d.Method, d.RemotePath))
		err := p.deployOneDistribution(ctx, build, project, &d, sourceDir, writeLine)
		fin := time.Now()
		if err != nil {
			if ctx.Err() != nil {
				row.Status = "cancelled"
				row.ErrorMessage = "cancelled"
			} else {
				row.Status = "failed"
				row.ErrorMessage = err.Error()
			}
			row.FinishedAt = &fin
			_ = p.buildDistRepo.Upsert(row)
			writeLine("ERROR: " + err.Error())
			nFail++
			continue
		}
		row.Status = "success"
		row.ErrorMessage = ""
		row.FinishedAt = &fin
		_ = p.buildDistRepo.Upsert(row)
		nOK++
	}

	summary := "all_success"
	if ctx.Err() != nil {
		summary = "cancelled"
	} else if nFail > 0 && nOK > 0 {
		summary = "partial"
	} else if nFail > 0 && nOK == 0 {
		summary = "all_failed"
	}

	p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
		"current_stage":            "success",
		"distribution_summary":     summary,
		"redistribute_filter_json": "",
	})
	writeLine(fmt.Sprintf("=== Distribution phase finished (%s) ===", summary))
}

func (p *Pipeline) recordDistributionCancelled(build *model.Build, distID uint) {
	row := &model.BuildDistribution{
		BuildID:        build.ID,
		DistributionID: distID,
		Status:         "cancelled",
		ErrorMessage:   "cancelled",
		FinishedAt:     ptrTime(time.Now()),
	}
	_ = p.buildDistRepo.Upsert(row)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func (p *Pipeline) executeRedistributeOnly(ctx context.Context, build *model.Build, project *model.Project, env *model.Environment, writeLine func(string)) {
	artifactPath := strings.TrimSpace(build.ArtifactPath)
	if artifactPath == "" {
		writeLine("ERROR: no artifact_path")
		p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"distribution_summary":     "all_failed",
			"redistribute_filter_json": "",
		})
		return
	}
	if !filepath.IsAbs(artifactPath) {
		artifactPath = filepath.Join(p.artifactDir, artifactPath)
	}
	if _, err := os.Stat(artifactPath); err != nil {
		writeLine("ERROR: " + err.Error())
		p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"distribution_summary":     "all_failed",
			"redistribute_filter_json": "",
		})
		return
	}
	writeLine("=== 重新分发：使用已有产物 ===")
	writeLine("Artifact: " + artifactPath)

	tmpDir, err := os.MkdirTemp("", "buildflow-deploy-*")
	if err != nil {
		writeLine("ERROR: mkdir temp: " + err.Error())
		p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"distribution_summary":     "all_failed",
			"redistribute_filter_json": "",
		})
		return
	}
	defer os.RemoveAll(tmpDir)

	artifactFormat := normalizeArtifactFormat(project.ArtifactFormat)
	if err := extractArtifactArchive(artifactPath, tmpDir, artifactFormat); err != nil {
		writeLine("ERROR: " + err.Error())
		p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"distribution_summary":     "all_failed",
			"redistribute_filter_json": "",
		})
		return
	}

	filter := parseRedistributeFilterJSON(build.RedistributeFilterJSON)
	if ctx.Err() != nil {
		p.cancelBuild(build)
		return
	}
	p.runDistributions(ctx, build, project, env, tmpDir, writeLine, filter)
}

func (p *Pipeline) markBuildArtifactSuccess(build *model.Build, writeLine func(string), hasDistributions bool) {
	finished := time.Now()
	build.FinishedAt = &finished
	if build.StartedAt != nil {
		build.DurationMs = finished.Sub(*build.StartedAt).Milliseconds()
	}
	build.Status = "success"
	build.ErrorMessage = ""
	if hasDistributions {
		build.CurrentStage = "distributing"
		build.DistributionSummary = "running"
	} else {
		build.CurrentStage = "success"
		build.DistributionSummary = "none"
	}
	fields := map[string]interface{}{
		"finished_at":          build.FinishedAt,
		"duration_ms":          build.DurationMs,
		"current_stage":        build.CurrentStage,
		"status":               "success",
		"error_message":        "",
		"distribution_summary": build.DistributionSummary,
	}
	_ = p.buildRepo.UpdateStatus(build.ID, "success", fields)
	writeLine(fmt.Sprintf("=== Build phase succeeded in %dms (artifact ready) ===", build.DurationMs))
	p.notify(build, "success")
}
