package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000006_dashboard_ops", upDashboardOps)
}

func upDashboardOps(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	models := []interface{}{
		&dashboardLayoutMigrationModel{},
		&toolchainDefinitionMigrationModel{},
		&installSourceMigrationModel{},
		&toolchainInstallJobMigrationModel{},
	}
	for _, item := range models {
		if db.Migrator().HasTable(item) {
			continue
		}
		if err := db.Migrator().CreateTable(item); err != nil {
			return err
		}
	}
	return seedBuiltinToolchains(db)
}

type dashboardLayoutMigrationModel struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;uniqueIndex"`
	CardsJSON string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:""`
	UpdatedAt time.Time `gorm:""`
}

func (dashboardLayoutMigrationModel) TableName() string { return "dashboard_layouts" }

type toolchainDefinitionMigrationModel struct {
	ID                uint      `gorm:"primaryKey"`
	Name              string    `gorm:"size:100;not null;uniqueIndex"`
	Kind              string    `gorm:"size:20;not null;default:builtin"`
	Executable        string    `gorm:"size:200;not null"`
	Description       string    `gorm:"size:500"`
	DetectCommand     string    `gorm:"type:text"`
	InstallTemplate   string    `gorm:"type:text"`
	UpgradeTemplate   string    `gorm:"type:text"`
	UninstallTemplate string    `gorm:"type:text"`
	VersionsCommand   string    `gorm:"type:text"`
	SwitchTemplate    string    `gorm:"type:text"`
	DefaultVersion    string    `gorm:"size:100"`
	CreatedBy         uint      `gorm:"index"`
	CreatedAt         time.Time `gorm:""`
	UpdatedAt         time.Time `gorm:""`
}

func (toolchainDefinitionMigrationModel) TableName() string { return "toolchain_definitions" }

type installSourceMigrationModel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;not null;uniqueIndex"`
	BaseURL   string    `gorm:"size:1000;not null"`
	Priority  int       `gorm:"not null;index"`
	Enabled   bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:""`
	UpdatedAt time.Time `gorm:""`
}

func (installSourceMigrationModel) TableName() string { return "install_sources" }

type toolchainInstallJobMigrationModel struct {
	ID               uint       `gorm:"primaryKey"`
	ToolchainID      uint       `gorm:"not null;index"`
	Operation        string     `gorm:"size:20;not null"`
	RequestedVersion string     `gorm:"size:100"`
	Status           string     `gorm:"size:20;not null;default:queued;index"`
	SourceID         *uint      `gorm:"index"`
	CommandSnapshot  string     `gorm:"type:text"`
	LogText          string     `gorm:"type:text"`
	ErrorMessage     string     `gorm:"type:text"`
	CreatedBy        uint       `gorm:"index"`
	StartedAt        *time.Time `gorm:""`
	FinishedAt       *time.Time `gorm:""`
	CreatedAt        time.Time  `gorm:""`
}

func (toolchainInstallJobMigrationModel) TableName() string { return "toolchain_install_jobs" }

func seedBuiltinToolchains(db *gorm.DB) error {
	for _, builtin := range builtinToolchains() {
		if err := db.Where("name = ?", builtin.Name).FirstOrCreate(&builtin).Error; err != nil {
			return err
		}
	}
	sources := []installSourceMigrationModel{
		{Name: "Primary", BaseURL: "https://pypi.org/simple", Priority: 10, Enabled: true},
		{Name: "Secondary", BaseURL: "https://mirrors.aliyun.com/pypi/simple", Priority: 20, Enabled: true},
	}
	for _, source := range sources {
		if err := db.Where("name = ?", source.Name).FirstOrCreate(&source).Error; err != nil {
			return err
		}
	}
	return nil
}

func builtinToolchains() []toolchainDefinitionMigrationModel {
	return []toolchainDefinitionMigrationModel{
		{
			Name: "Go", Kind: "builtin", Executable: "go", Description: "Go 编译工具链（需要 asdf >= 0.16）",
			DetectCommand: "go version", VersionsCommand: "asdf list golang",
			InstallTemplate:   `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf plugin add golang https://github.com/asdf-community/asdf-golang.git 2>/dev/null || true; asdf install golang "$version" && asdf set -u golang "$version"`,
			UpgradeTemplate:   `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf plugin update golang || true; asdf install golang "$version" && asdf set -u golang "$version"`,
			UninstallTemplate: `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || version="$(asdf current golang 2>/dev/null | awk 'NR==1 {print $2}')"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf uninstall golang "$version"`,
			SwitchTemplate:    `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf set -u golang "$version"`,
		},
		{
			Name: "Node.js", Kind: "builtin", Executable: "node", Description: "Node.js 运行时（需要 fnm）",
			DetectCommand: "node --version", VersionsCommand: "fnm list",
			InstallTemplate:   `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; fnm install "$version" && fnm default "$version"`,
			UpgradeTemplate:   `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; fnm install "$version" && fnm default "$version"`,
			UninstallTemplate: `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; fnm uninstall "$version"`,
			SwitchTemplate:    `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; fnm default "$version"`,
		},
		{
			Name: "Java", Kind: "builtin", Executable: "java", Description: "Java 运行时（需要 SDKMAN!）",
			DetectCommand: "java -version", VersionsCommand: `. "$HOME/.sdkman/bin/sdkman-init.sh" && sdk list java`,
			InstallTemplate:   `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk install java "$version" && sdk default java "$version"`,
			UpgradeTemplate:   `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk install java "$version" && sdk default java "$version"`,
			UninstallTemplate: `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk uninstall java "$version"`,
			SwitchTemplate:    `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk default java "$version"`,
		},
		{
			Name: "Python", Kind: "builtin", Executable: "python3", Description: "Python 运行时（需要 pyenv）",
			DetectCommand: "python3 --version", VersionsCommand: "pyenv versions",
			InstallTemplate:   `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; pyenv install --skip-existing "$version" && pyenv global "$version"`,
			UpgradeTemplate:   `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; pyenv install --skip-existing "$version" && pyenv global "$version"`,
			UninstallTemplate: `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; pyenv uninstall -f "$version"`,
			SwitchTemplate:    `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; pyenv global "$version"`,
		},
	}
}
