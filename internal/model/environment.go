package model

import "time"

type Environment struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ProjectID        uint      `json:"project_id" gorm:"uniqueIndex:idx_proj_env_name;not null"`
	Name             string    `json:"name" gorm:"uniqueIndex:idx_proj_env_name;size:50;not null"`
	Branch           string    `json:"branch" gorm:"size:200;default:main"`
	BuildScript      string    `json:"build_script" gorm:"type:text"`
	BuildScriptType  string    `json:"build_script_type" gorm:"size:20;default:bash"`
	BuildOutputDir   string    `json:"build_output_dir" gorm:"size:300"`
	DeployServerID   *uint     `json:"deploy_server_id"`
	DeployPath       string    `json:"deploy_path" gorm:"size:500"`
	DeployMethod     string    `json:"deploy_method" gorm:"size:20"`
	PostDeployScript string    `json:"post_deploy_script" gorm:"type:text"`
	EnvVars          string    `json:"env_vars" gorm:"type:text"`
	CronExpression   string    `json:"cron_expression" gorm:"size:100"`
	CronEnabled      bool      `json:"cron_enabled" gorm:"default:false"`
	SortOrder        int       `json:"sort_order" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
