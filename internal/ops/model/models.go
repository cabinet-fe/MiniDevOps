package model

import "time"

const (
	ToolchainBuiltin = "builtin"
	ToolchainCustom  = "custom"
)

const (
	JobQueued      = "queued"
	JobRunning     = "running"
	JobSuccess     = "success"
	JobFailed      = "failed"
	JobInterrupted = "interrupted"
)

// ToolchainDefinition describes an executable and the commands that manage it.
// Custom commands intentionally execute under the Bedrock process UID; this is
// an administrator-only capability, not a sandbox.
type ToolchainDefinition struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Kind              string    `json:"kind" gorm:"size:20;not null;default:builtin"`
	Executable        string    `json:"executable" gorm:"size:200;not null"`
	Description       string    `json:"description" gorm:"size:500"`
	DetectCommand     string    `json:"detect_command" gorm:"type:text"`
	InstallTemplate   string    `json:"install_template" gorm:"type:text"`
	UpgradeTemplate   string    `json:"upgrade_template" gorm:"type:text"`
	UninstallTemplate string    `json:"uninstall_template" gorm:"type:text"`
	VersionsCommand   string    `json:"versions_command" gorm:"type:text"`
	SwitchTemplate    string    `json:"switch_template" gorm:"type:text"`
	DefaultVersion    string    `json:"default_version" gorm:"size:100"`
	CreatedBy         uint      `json:"created_by" gorm:"index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ToolchainDefinition) TableName() string { return "toolchain_definitions" }

// InstallSource is attempted in ascending priority order for install-like jobs.
type InstallSource struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	BaseURL   string    `json:"base_url" gorm:"size:1000;not null"`
	Priority  int       `json:"priority" gorm:"not null;index"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (InstallSource) TableName() string { return "install_sources" }

// ToolchainInstallJob persists async execution, including command snapshots and
// append-only textual logs required for troubleshooting source fallback.
type ToolchainInstallJob struct {
	ID                  uint                `json:"id" gorm:"primaryKey"`
	ToolchainID         uint                `json:"toolchain_id" gorm:"not null;index"`
	Operation           string              `json:"operation" gorm:"size:20;not null"`
	RequestedVersion    string              `json:"requested_version" gorm:"size:100"`
	Status              string              `json:"status" gorm:"size:20;not null;default:queued;index"`
	SourceID            *uint               `json:"source_id" gorm:"index"`
	CommandSnapshot     string              `json:"command_snapshot" gorm:"type:text"`
	LogText             string              `json:"-" gorm:"type:text"`
	ErrorMessage        string              `json:"error_message" gorm:"type:text"`
	CreatedBy           uint                `json:"created_by" gorm:"index"`
	StartedAt           *time.Time          `json:"started_at"`
	FinishedAt          *time.Time          `json:"finished_at"`
	CreatedAt           time.Time           `json:"created_at"`
	ToolchainDefinition ToolchainDefinition `json:"toolchain,omitempty" gorm:"foreignKey:ToolchainID"`
	Source              *InstallSource      `json:"source,omitempty" gorm:"foreignKey:SourceID"`
}

func (ToolchainInstallJob) TableName() string { return "toolchain_install_jobs" }

type ProcessInfo struct {
	PID         int32    `json:"pid"`
	Name        string   `json:"name"`
	CPUPercent  float64  `json:"cpu_percent"`
	MemoryBytes uint64   `json:"memory_bytes"`
	Username    string   `json:"username"`
	StartTime   int64    `json:"start_time"`
	Ports       []uint32 `json:"ports"`
}

type ProcessListOptions struct {
	Keyword string
	PID     *int32
	Port    *uint32
	Sort    string
	Order   string
}
