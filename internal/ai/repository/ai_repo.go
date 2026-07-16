package repository

import (
	"time"

	"gorm.io/gorm"

	"bedrock/internal/ai/model"
)

type AIRepository struct {
	db *gorm.DB
}

func NewAIRepository(db *gorm.DB) *AIRepository {
	return &AIRepository{db: db}
}

func (r *AIRepository) ListCLIs() ([]model.CliRuntimeDefinition, error) {
	var items []model.CliRuntimeDefinition
	err := r.db.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *AIRepository) FindCLIByKey(key string) (*model.CliRuntimeDefinition, error) {
	var item model.CliRuntimeDefinition
	if err := r.db.Where("key = ?", key).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AIRepository) UpdateCLI(item *model.CliRuntimeDefinition) error {
	return r.db.Save(item).Error
}

func (r *AIRepository) ListSources(cliKey string) ([]model.CliInstallSource, error) {
	var items []model.CliInstallSource
	q := r.db.Order("priority ASC, id ASC")
	if cliKey != "" {
		q = q.Where("cli_key = ?", cliKey)
	}
	err := q.Find(&items).Error
	return items, err
}

func (r *AIRepository) ListEnabledSources(cliKey string) ([]model.CliInstallSource, error) {
	var items []model.CliInstallSource
	err := r.db.Where("cli_key = ? AND enabled = ?", cliKey, true).
		Order("priority ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *AIRepository) FindSource(id uint) (*model.CliInstallSource, error) {
	var item model.CliInstallSource
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AIRepository) CreateSource(item *model.CliInstallSource) error {
	return r.db.Create(item).Error
}

func (r *AIRepository) UpdateSource(item *model.CliInstallSource) error {
	return r.db.Save(item).Error
}

func (r *AIRepository) DeleteSource(id uint) error {
	return r.db.Delete(&model.CliInstallSource{}, id).Error
}

func (r *AIRepository) CreateAgent(agent *model.AiAgent) error {
	return r.db.Create(agent).Error
}

func (r *AIRepository) UpdateAgent(agent *model.AiAgent) error {
	return r.db.Save(agent).Error
}

func (r *AIRepository) DeleteAgent(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("agent_id = ?", id).Delete(&model.AgentTrigger{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.AiAgent{}, id).Error
	})
}

func (r *AIRepository) FindAgent(id uint) (*model.AiAgent, error) {
	var agent model.AiAgent
	if err := r.db.First(&agent, id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AIRepository) ListAgents(page, pageSize int) ([]model.AiAgent, int64, error) {
	q := r.db.Model(&model.AiAgent{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var items []model.AiAgent
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *AIRepository) CreateTrigger(t *model.AgentTrigger) error {
	return r.db.Create(t).Error
}

func (r *AIRepository) UpdateTrigger(t *model.AgentTrigger) error {
	return r.db.Save(t).Error
}

func (r *AIRepository) DeleteTrigger(id uint) error {
	return r.db.Delete(&model.AgentTrigger{}, id).Error
}

func (r *AIRepository) FindTrigger(id uint) (*model.AgentTrigger, error) {
	var t model.AgentTrigger
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *AIRepository) ListTriggers(agentID uint) ([]model.AgentTrigger, error) {
	var items []model.AgentTrigger
	err := r.db.Where("agent_id = ?", agentID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *AIRepository) ListCronTriggers() ([]model.AgentTrigger, error) {
	var items []model.AgentTrigger
	err := r.db.Where("type = ? AND enabled = ?", model.TriggerCron, true).Find(&items).Error
	return items, err
}

func (r *AIRepository) ListBuildEventTriggers(buildJobID uint, event string) ([]model.AgentTrigger, error) {
	var items []model.AgentTrigger
	err := r.db.Where(
		"type = ? AND enabled = ? AND build_job_id = ? AND (build_event = ? OR build_event = '' OR build_event IS NULL)",
		model.TriggerBuildEvent, true, buildJobID, event,
	).Find(&items).Error
	return items, err
}

func (r *AIRepository) CreateRun(run *model.AgentRun) error {
	return r.db.Create(run).Error
}

func (r *AIRepository) UpdateRun(run *model.AgentRun) error {
	return r.db.Save(run).Error
}

func (r *AIRepository) UpdateRunFields(id uint, fields map[string]any) error {
	return r.db.Model(&model.AgentRun{}).Where("id = ?", id).Updates(fields).Error
}

func (r *AIRepository) FindRun(id uint) (*model.AgentRun, error) {
	var run model.AgentRun
	if err := r.db.Preload("Agent").First(&run, id).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *AIRepository) ListRuns(page, pageSize int, agentID uint, status string) ([]model.AgentRun, int64, error) {
	q := r.db.Model(&model.AgentRun{})
	if agentID > 0 {
		q = q.Where("agent_id = ?", agentID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var items []model.AgentRun
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *AIRepository) ListRunsByStatuses(statuses ...string) ([]model.AgentRun, error) {
	var items []model.AgentRun
	err := r.db.Where("status IN ?", statuses).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *AIRepository) CountActiveRuns(agentID uint) (int64, error) {
	var n int64
	err := r.db.Model(&model.AgentRun{}).
		Where("agent_id = ? AND status IN ?", agentID, []string{model.JobQueued, model.JobRunning, model.JobPending}).
		Count(&n).Error
	return n, err
}

func (r *AIRepository) MarkRunningRunsInterrupted() (int64, error) {
	now := time.Now().UTC()
	res := r.db.Model(&model.AgentRun{}).
		Where("status = ?", model.JobRunning).
		Updates(map[string]any{
			"status":        model.JobInterrupted,
			"error_message": "interrupted by server restart",
			"finished_at":   now,
		})
	return res.RowsAffected, res.Error
}

func (r *AIRepository) CreateSkill(skill *model.SkillPackage) error {
	return r.db.Create(skill).Error
}

func (r *AIRepository) UpdateSkill(skill *model.SkillPackage) error {
	return r.db.Save(skill).Error
}

func (r *AIRepository) DeleteSkill(id uint) error {
	return r.db.Delete(&model.SkillPackage{}, id).Error
}

func (r *AIRepository) FindSkill(id uint) (*model.SkillPackage, error) {
	var skill model.SkillPackage
	if err := r.db.First(&skill, id).Error; err != nil {
		return nil, err
	}
	return &skill, nil
}

func (r *AIRepository) ListSkills(page, pageSize int, userID uint, isSuperAdmin bool) ([]model.SkillPackage, int64, error) {
	q := r.db.Model(&model.SkillPackage{})
	if !isSuperAdmin {
		q = q.Where("visibility = ? OR created_by = ?", model.SkillPublic, userID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var items []model.SkillPackage
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *AIRepository) CreatePAT(token *model.PersonalAccessToken) error {
	return r.db.Create(token).Error
}

func (r *AIRepository) FindPATByHash(hash string) (*model.PersonalAccessToken, error) {
	var token model.PersonalAccessToken
	if err := r.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *AIRepository) FindPAT(id uint) (*model.PersonalAccessToken, error) {
	var token model.PersonalAccessToken
	if err := r.db.First(&token, id).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *AIRepository) ListPATs(userID uint) ([]model.PersonalAccessToken, error) {
	var items []model.PersonalAccessToken
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *AIRepository) UpdatePAT(token *model.PersonalAccessToken) error {
	return r.db.Save(token).Error
}

func (r *AIRepository) DeletePAT(id uint) error {
	return r.db.Delete(&model.PersonalAccessToken{}, id).Error
}
