package model

import "time"

// RiskNoticeSameUID is shown in CLI/agent APIs and UI. AI CLIs run as the
// Bedrock process UID with no OS/container sandbox.
const RiskNoticeSameUID = "AI CLI 与构建脚本均以 Bedrock 进程同一操作系统用户直接执行，无 OS/容器沙箱隔离。"

// CliRuntimeDefinition describes one of the four parallel AI CLIs.
type CliRuntimeDefinition struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Key               string    `json:"key" gorm:"size:40;not null;uniqueIndex"`
	Name              string    `json:"name" gorm:"size:100;not null"`
	BinaryName        string    `json:"binary_name" gorm:"size:100;not null"`
	Description       string    `json:"description" gorm:"size:500"`
	DetectCommand     string    `json:"detect_command" gorm:"type:text"`
	InstallTemplate   string    `json:"install_template" gorm:"type:text"`
	UpgradeTemplate   string    `json:"upgrade_template" gorm:"type:text"`
	UninstallTemplate string    `json:"uninstall_template" gorm:"type:text"`
	DefaultArgs       string    `json:"default_args" gorm:"type:text"`
	EnvTemplateJSON   string    `json:"env_template_json" gorm:"type:text"`
	APIBaseEnv        string    `json:"api_base_env" gorm:"size:100"`
	HealthCommand     string    `json:"health_command" gorm:"type:text"`
	InstallStatus     string    `json:"install_status" gorm:"size:40;default:unknown"`
	InstalledPath     string    `json:"installed_path" gorm:"size:500"`
	InstalledVersion  string    `json:"installed_version" gorm:"size:100"`
	Healthy           bool      `json:"healthy" gorm:"not null;default:false"`
	RiskNotice        string    `json:"risk_notice" gorm:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (CliRuntimeDefinition) TableName() string { return "cli_runtime_definitions" }

// CliInstallSource is tried in ascending priority for install/upgrade.
type CliInstallSource struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CliKey    string    `json:"cli_key" gorm:"size:40;not null;index"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	BaseURL   string    `json:"base_url" gorm:"size:1000;not null"`
	Priority  int       `json:"priority" gorm:"not null;index"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CliInstallSource) TableName() string { return "cli_install_sources" }
