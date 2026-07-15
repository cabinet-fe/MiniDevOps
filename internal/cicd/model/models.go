package model

import "time"

// Credential is the shared secret store (AES-GCM ciphertext; never echoed by API).
type Credential struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"size:100;not null;uniqueIndex:idx_cred_name_creator"`
	Type             string    `json:"type" gorm:"size:20;not null"`
	Username         string    `json:"username" gorm:"size:200"`
	SecretCipher     string    `json:"-" gorm:"size:4000"`
	PassphraseCipher string    `json:"-" gorm:"size:2000"`
	Description      string    `json:"description" gorm:"size:500"`
	CreatedBy        uint      `json:"created_by" gorm:"not null;uniqueIndex:idx_cred_name_creator"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	HasSecret     bool `json:"has_secret" gorm:"-"`
	HasPassphrase bool `json:"has_passphrase" gorm:"-"`
}

func (Credential) TableName() string { return "credentials" }

// Repository is a Git source configuration.
type Repository struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	Name               string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description        string    `json:"description" gorm:"size:500"`
	Tags               string    `json:"tags" gorm:"size:500"`
	RepoURL            string    `json:"repo_url" gorm:"size:500;not null"`
	DefaultBranch      string    `json:"default_branch" gorm:"size:200;default:main"`
	AuthType           string    `json:"auth_type" gorm:"size:20;not null;default:none"`
	CredentialID       *uint     `json:"credential_id" gorm:"index"`
	WebhookSecret      string    `json:"webhook_secret,omitempty" gorm:"size:64"`
	WebhookType        string    `json:"webhook_type" gorm:"size:20;default:auto"`
	WebhookRefPath     string    `json:"webhook_ref_path" gorm:"size:300"`
	WebhookCommitPath  string    `json:"webhook_commit_path" gorm:"size:300"`
	WebhookMessagePath string    `json:"webhook_message_path" gorm:"size:300"`
	CreatedBy          uint      `json:"created_by" gorm:"index"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Repository) TableName() string { return "repositories" }

// Server is a deploy host. Secrets live in Credential; bind requires cicd.credentials:use.
type Server struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" gorm:"size:100;not null"`
	Host              string    `json:"host" gorm:"size:200;not null"`
	Port              int       `json:"port" gorm:"default:22"`
	OSType            string    `json:"os_type" gorm:"size:20;not null;default:linux"`
	Username          string    `json:"username" gorm:"size:100"`
	AuthType          string    `json:"auth_type" gorm:"size:20;not null;default:password"`
	CredentialID      *uint     `json:"credential_id" gorm:"index"`
	AgentURL          string    `json:"agent_url" gorm:"size:500"`
	AgentCredentialID *uint     `json:"agent_credential_id" gorm:"index"`
	Description       string    `json:"description" gorm:"size:500"`
	Tags              string    `json:"tags" gorm:"size:500"`
	Status            string    `json:"status" gorm:"size:20;default:unknown"`
	CreatedBy         uint      `json:"created_by" gorm:"index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (Server) TableName() string { return "servers" }

// BuildJob belongs to a Repository (1:N).
type BuildJob struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	RepositoryID      uint      `json:"repository_id" gorm:"index;not null"`
	Name              string    `json:"name" gorm:"size:100;not null"`
	Description       string    `json:"description" gorm:"size:500"`
	Enabled           bool      `json:"enabled" gorm:"not null;default:true"`
	BranchPolicy      string    `json:"branch_policy" gorm:"size:20;not null;default:fixed"`
	Branch            string    `json:"branch" gorm:"size:200;default:main"`
	ShallowClone      bool      `json:"shallow_clone" gorm:"not null;default:true"`
	BuildScriptType   string    `json:"build_script_type" gorm:"size:20;default:bash"`
	BuildScript       string    `json:"build_script" gorm:"type:text"`
	WorkDir           string    `json:"work_dir" gorm:"size:300"`
	OutputDir         string    `json:"output_dir" gorm:"size:300"`
	CachePaths        string    `json:"cache_paths" gorm:"type:text"`
	EnvVarNamesJSON   string    `json:"-" gorm:"type:text"`
	EnvVarNames       []string  `json:"env_var_names" gorm:"-"`
	TriggerManual     bool      `json:"trigger_manual" gorm:"not null;default:true"`
	TriggerWebhook    bool      `json:"trigger_webhook" gorm:"not null;default:false"`
	TriggerCron       bool      `json:"trigger_cron" gorm:"not null;default:false"`
	CronExpression    string    `json:"cron_expression" gorm:"size:100"`
	CronTimezone      string    `json:"cron_timezone" gorm:"size:100;default:UTC"`
	MaxArtifacts      int       `json:"max_artifacts" gorm:"default:5"`
	ArtifactFormat    string    `json:"artifact_format" gorm:"size:20;default:gzip"`
	AgentTriggerEvent string    `json:"agent_trigger_event" gorm:"size:40;default:artifact_ready"`
	CreatedBy         uint      `json:"created_by" gorm:"index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	DeployTargets []DeployTarget `json:"deploy_targets,omitempty" gorm:"foreignKey:BuildJobID"`
}

func (BuildJob) TableName() string { return "build_jobs" }

// DeployTarget is private to a BuildJob (1:N); not shared across jobs.
type DeployTarget struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	BuildJobID       uint      `json:"build_job_id" gorm:"index;not null"`
	ServerID         *uint     `json:"server_id" gorm:"index"`
	RemotePath       string    `json:"remote_path" gorm:"size:500"`
	Method           string    `json:"method" gorm:"size:20;not null;default:rsync"`
	PostDeployScript string    `json:"post_deploy_script" gorm:"type:text"`
	SortOrder        int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (DeployTarget) TableName() string { return "deploy_targets" }

// BuildRun status (result) vs stage (activity) — see DESIGN §5.2.
// status: queued|running|success|failed|cancelled|interrupted
// stage: pending|cloning|building|archiving|distributing|idle
// distribution_summary: none|running|all_success|partial|all_failed|cancelled
type BuildRun struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	BuildJobID          uint       `json:"build_job_id" gorm:"uniqueIndex:idx_job_build_num;not null"`
	BuildNumber         int        `json:"build_number" gorm:"uniqueIndex:idx_job_build_num;not null"`
	Status              string     `json:"status" gorm:"size:20;not null;default:queued"`
	Stage               string     `json:"stage" gorm:"size:20;not null;default:pending"`
	TriggerType         string     `json:"trigger_type" gorm:"size:20"`
	TriggeredBy         uint       `json:"triggered_by"`
	Branch              string     `json:"branch" gorm:"size:200"`
	CommitHash          string     `json:"commit_hash" gorm:"size:64"`
	CommitMessage       string     `json:"commit_message" gorm:"size:500"`
	LogPath             string     `json:"log_path" gorm:"size:500"`
	ArtifactPath        string     `json:"artifact_path" gorm:"size:500"`
	DurationMs          int64      `json:"duration_ms"`
	ErrorMessage        string     `json:"error_message" gorm:"type:text"`
	DistributionSummary string     `json:"distribution_summary" gorm:"size:30;default:none"`
	SnapshotJSON        string     `json:"snapshot_json,omitempty" gorm:"type:text"`
	StartedAt           *time.Time `json:"started_at"`
	FinishedAt          *time.Time `json:"finished_at"`
	CreatedAt           time.Time  `json:"created_at"`

	DeployAttempts []BuildDeployAttempt `json:"deploy_attempts,omitempty" gorm:"foreignKey:BuildRunID"`
}

func (BuildRun) TableName() string { return "build_runs" }

// BuildDeployAttempt is one target row in a distribute/redeploy batch (append-only in Wave 4).
type BuildDeployAttempt struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	BuildRunID         uint       `json:"build_run_id" gorm:"index;not null"`
	BatchNo            int        `json:"batch_no" gorm:"not null;default:1"`
	DeployTargetID     *uint      `json:"deploy_target_id" gorm:"index"`
	TargetSnapshotJSON string     `json:"target_snapshot_json,omitempty" gorm:"type:text"`
	Status             string     `json:"status" gorm:"size:20;not null;default:pending"`
	LogPath            string     `json:"log_path" gorm:"size:500"`
	ErrorMessage       string     `json:"error_message" gorm:"type:text"`
	StartedAt          *time.Time `json:"started_at"`
	FinishedAt         *time.Time `json:"finished_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (BuildDeployAttempt) TableName() string { return "build_deploy_attempts" }
