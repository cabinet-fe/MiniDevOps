package model

import "time"

const (
	JobQueued      = "queued"
	JobRunning     = "running"
	JobSuccess     = "success"
	JobFailed      = "failed"
	JobInterrupted = "interrupted"
	JobCancelled   = "cancelled"
	JobPending     = "pending"
)

const (
	TriggerManual     = "manual"
	TriggerAPI        = "api"
	TriggerCron       = "cron"
	TriggerBuildEvent = "build_event"
	TriggerDocsGen    = "docs_generate"
)

const (
	SkillPublic  = "public"
	SkillPrivate = "private"
)

const (
	ScopeSkillsRead = "skills:read"
	ScopeAgentsRun  = "agents:run"
)

const (
	EventArtifactReady         = "artifact_ready"
	EventDistributionFinished  = "distribution_finished"
	EventNone                  = "none"
)

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

// AiAgent is a configured agent bound to one CLI, skills, and build-job workspaces.
type AiAgent struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"size:100;not null"`
	Description      string    `json:"description" gorm:"size:500"`
	Enabled          bool      `json:"enabled" gorm:"not null;default:true"`
	CliKey           string    `json:"cli_key" gorm:"size:40;not null;index"`
	SystemPrompt     string    `json:"system_prompt" gorm:"type:text"`
	SkillIDsJSON     string    `json:"-" gorm:"type:text"`
	SkillIDs         []uint    `json:"skill_ids" gorm:"-"`
	BuildJobIDsJSON  string    `json:"-" gorm:"type:text"`
	BuildJobIDs      []uint    `json:"build_job_ids" gorm:"-"`
	OutputDir        string    `json:"output_dir" gorm:"size:200;not null;default:output"`
	ArtifactFormat   string    `json:"artifact_format" gorm:"size:20;not null;default:gzip"`
	MaxArtifacts     int       `json:"max_artifacts" gorm:"not null;default:10"`
	StreamOutput     bool      `json:"stream_output" gorm:"not null;default:false"`
	TimeoutSec       int       `json:"timeout_sec" gorm:"not null;default:600"`
	CreatedBy        uint      `json:"created_by" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (AiAgent) TableName() string { return "ai_agents" }

// AgentTrigger configures how an agent may be started.
type AgentTrigger struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	AgentID           uint      `json:"agent_id" gorm:"not null;index"`
	Type              string    `json:"type" gorm:"size:40;not null"`
	Enabled           bool      `json:"enabled" gorm:"not null;default:true"`
	CronExpression    string    `json:"cron_expression" gorm:"size:100"`
	CronTimezone      string    `json:"cron_timezone" gorm:"size:100;default:UTC"`
	BuildJobID        *uint     `json:"build_job_id" gorm:"index"`
	BuildEvent        string    `json:"build_event" gorm:"size:40"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (AgentTrigger) TableName() string { return "agent_triggers" }

// AgentRun is an independent async execution (DESIGN §5.3).
type AgentRun struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	AgentID         uint       `json:"agent_id" gorm:"not null;index"`
	TriggerType     string     `json:"trigger_type" gorm:"size:40;not null"`
	TriggerID       *uint      `json:"trigger_id" gorm:"index"`
	Status          string     `json:"status" gorm:"size:20;not null;default:queued;index"`
	TriggeredBy     uint       `json:"triggered_by" gorm:"index"`
	BuildRunID      *uint      `json:"build_run_id" gorm:"index"`
	ProjectID       *uint      `json:"project_id" gorm:"index"`
	DocNodeID       *uint      `json:"doc_node_id" gorm:"index"`
	SnapshotJSON    string     `json:"snapshot_json,omitempty" gorm:"type:text"`
	SkillDigestJSON string     `json:"skill_digest_json,omitempty" gorm:"type:text"`
	WorkDir         string     `json:"work_dir" gorm:"size:500"`
	ArtifactPath    string     `json:"artifact_path" gorm:"size:500"`
	LogPath         string     `json:"log_path" gorm:"size:500"`
	OutputText      string     `json:"output_text,omitempty" gorm:"type:text"`
	ErrorMessage    string     `json:"error_message" gorm:"type:text"`
	DurationMs      int64      `json:"duration_ms"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	CreatedAt       time.Time  `json:"created_at"`

	Agent *AiAgent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

func (AgentRun) TableName() string { return "agent_runs" }

// SkillPackage metadata + content-addressed StorageObject reference.
type SkillPackage struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"size:200;not null"`
	Description     string    `json:"description" gorm:"size:1000"`
	Visibility      string    `json:"visibility" gorm:"size:20;not null;default:private;index"`
	StorageObjectID uint      `json:"storage_object_id" gorm:"not null;index"`
	PackageDigest   string    `json:"package_digest" gorm:"size:64;not null;index"`
	SizeBytes       int64     `json:"size_bytes" gorm:"not null"`
	CreatedBy       uint      `json:"created_by" gorm:"index"`
	UpdatedBy       uint      `json:"updated_by" gorm:"index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (SkillPackage) TableName() string { return "skill_packages" }

// PersonalAccessToken stores only the hash; plaintext is returned once on create.
type PersonalAccessToken struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null;index"`
	Name       string     `json:"name" gorm:"size:100;not null"`
	TokenPrefix string    `json:"token_prefix" gorm:"size:16;not null"`
	TokenHash  string     `json:"-" gorm:"size:128;not null;uniqueIndex"`
	ScopesJSON string     `json:"-" gorm:"type:text;not null"`
	Scopes     []string   `json:"scopes" gorm:"-"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (PersonalAccessToken) TableName() string { return "personal_access_tokens" }
