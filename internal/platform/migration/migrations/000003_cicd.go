package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000003_cicd", upCICD)
}

func upCICD(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	models := []interface{}{
		&credentialMigrationModel{},
		&repositoryMigrationModel{},
		&serverMigrationModel{},
		&buildJobMigrationModel{},
		&deployTargetMigrationModel{},
		&buildRunMigrationModel{},
		&buildDeployAttemptMigrationModel{},
	}
	for _, m := range models {
		if db.Migrator().HasTable(m) {
			continue
		}
		if err := db.Migrator().CreateTable(m); err != nil {
			return err
		}
	}
	return nil
}

type credentialMigrationModel struct {
	ID               uint      `gorm:"primaryKey"`
	Name             string    `gorm:"size:100;not null;uniqueIndex:idx_cred_name_creator"`
	Type             string    `gorm:"size:20;not null"` // password|token|ssh_key|api_key
	Username         string    `gorm:"size:200"`
	SecretCipher     string    `gorm:"size:4000"` // AES-GCM hex
	PassphraseCipher string    `gorm:"size:2000"` // optional for ssh_key
	Description      string    `gorm:"size:500"`
	CreatedBy        uint      `gorm:"not null;uniqueIndex:idx_cred_name_creator"`
	CreatedAt        time.Time `gorm:""`
	UpdatedAt        time.Time `gorm:""`
}

func (credentialMigrationModel) TableName() string { return "credentials" }

type repositoryMigrationModel struct {
	ID                 uint      `gorm:"primaryKey"`
	Name               string    `gorm:"size:100;not null;uniqueIndex"`
	Description        string    `gorm:"size:500"`
	Tags               string    `gorm:"size:500"`
	RepoURL            string    `gorm:"size:500;not null"`
	DefaultBranch      string    `gorm:"size:200;default:main"`
	AuthType           string    `gorm:"size:20;not null;default:none"` // none|credential
	CredentialID       *uint     `gorm:"index"`
	WebhookSecret      string    `gorm:"size:64"`
	WebhookType        string    `gorm:"size:20;default:auto"`
	WebhookRefPath     string    `gorm:"size:300"`
	WebhookCommitPath  string    `gorm:"size:300"`
	WebhookMessagePath string    `gorm:"size:300"`
	CreatedBy          uint      `gorm:"index"`
	CreatedAt          time.Time `gorm:""`
	UpdatedAt          time.Time `gorm:""`
}

func (repositoryMigrationModel) TableName() string { return "repositories" }

type serverMigrationModel struct {
	ID                uint      `gorm:"primaryKey"`
	Name              string    `gorm:"size:100;not null"`
	Host              string    `gorm:"size:200;not null"`
	Port              int       `gorm:"default:22"`
	OSType            string    `gorm:"size:20;not null;default:linux"`
	Username          string    `gorm:"size:100"`
	AuthType          string    `gorm:"size:20;not null;default:password"` // password|key|ssh_agent|agent
	CredentialID      *uint     `gorm:"index"`
	AgentURL          string    `gorm:"size:500"`
	AgentCredentialID *uint     `gorm:"index"`
	Description       string    `gorm:"size:500"`
	Tags              string    `gorm:"size:500"`
	Status            string    `gorm:"size:20;default:unknown"`
	CreatedBy         uint      `gorm:"index"`
	CreatedAt         time.Time `gorm:""`
	UpdatedAt         time.Time `gorm:""`
}

func (serverMigrationModel) TableName() string { return "servers" }

type buildJobMigrationModel struct {
	ID                uint      `gorm:"primaryKey"`
	RepositoryID      uint      `gorm:"index;not null"`
	Name              string    `gorm:"size:100;not null"`
	Description       string    `gorm:"size:500"`
	Enabled           bool      `gorm:"not null;default:true"`
	BranchPolicy      string    `gorm:"size:20;not null;default:fixed"` // fixed|param
	Branch            string    `gorm:"size:200;default:main"`
	ShallowClone      bool      `gorm:"not null;default:true"`
	BuildScriptType   string    `gorm:"size:20;default:bash"`
	BuildScript       string    `gorm:"type:text"`
	WorkDir           string    `gorm:"size:300"`
	OutputDir         string    `gorm:"size:300"`
	CachePaths        string    `gorm:"type:text"`
	EnvVarNamesJSON   string    `gorm:"type:text"` // JSON array of env var names
	TriggerManual     bool      `gorm:"not null;default:true"`
	TriggerWebhook    bool      `gorm:"not null;default:false"`
	TriggerCron       bool      `gorm:"not null;default:false"`
	CronExpression    string    `gorm:"size:100"`
	CronTimezone      string    `gorm:"size:100;default:UTC"`
	MaxArtifacts      int       `gorm:"default:5"`
	ArtifactFormat    string    `gorm:"size:20;default:gzip"`
	AgentTriggerEvent string    `gorm:"size:40;default:artifact_ready"`
	CreatedBy         uint      `gorm:"index"`
	CreatedAt         time.Time `gorm:""`
	UpdatedAt         time.Time `gorm:""`
}

func (buildJobMigrationModel) TableName() string { return "build_jobs" }

type deployTargetMigrationModel struct {
	ID               uint      `gorm:"primaryKey"`
	BuildJobID       uint      `gorm:"index;not null"`
	ServerID         *uint     `gorm:"index"`
	RemotePath       string    `gorm:"size:500"`
	Method           string    `gorm:"size:20;not null;default:rsync"` // rsync|sftp|scp|agent|local
	PostDeployScript string    `gorm:"type:text"`
	SortOrder        int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:""`
	UpdatedAt        time.Time `gorm:""`
}

func (deployTargetMigrationModel) TableName() string { return "deploy_targets" }

type buildRunMigrationModel struct {
	ID                  uint       `gorm:"primaryKey"`
	BuildJobID          uint       `gorm:"uniqueIndex:idx_job_build_num;not null"`
	BuildNumber         int        `gorm:"uniqueIndex:idx_job_build_num;not null"`
	Status              string     `gorm:"size:20;not null;default:queued"`
	Stage               string     `gorm:"size:20;not null;default:pending"`
	TriggerType         string     `gorm:"size:20"`
	TriggeredBy         uint       `gorm:""`
	Branch              string     `gorm:"size:200"`
	CommitHash          string     `gorm:"size:64"`
	CommitMessage       string     `gorm:"size:500"`
	LogPath             string     `gorm:"size:500"`
	ArtifactPath        string     `gorm:"size:500"`
	DurationMs          int64      `gorm:""`
	ErrorMessage        string     `gorm:"type:text"`
	DistributionSummary string     `gorm:"size:30;default:none"`
	SnapshotJSON        string     `gorm:"type:text"`
	StartedAt           *time.Time `gorm:""`
	FinishedAt          *time.Time `gorm:""`
	CreatedAt           time.Time  `gorm:""`
}

func (buildRunMigrationModel) TableName() string { return "build_runs" }

type buildDeployAttemptMigrationModel struct {
	ID                 uint       `gorm:"primaryKey"`
	BuildRunID         uint       `gorm:"index;not null"`
	BatchNo            int        `gorm:"not null;default:1"`
	DeployTargetID     *uint      `gorm:"index"`
	TargetSnapshotJSON string     `gorm:"type:text"`
	Status             string     `gorm:"size:20;not null;default:pending"`
	LogPath            string     `gorm:"size:500"`
	ErrorMessage       string     `gorm:"type:text"`
	StartedAt          *time.Time `gorm:""`
	FinishedAt         *time.Time `gorm:""`
	CreatedAt          time.Time  `gorm:""`
}

func (buildDeployAttemptMigrationModel) TableName() string { return "build_deploy_attempts" }
