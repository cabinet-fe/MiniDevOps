package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000011_dev_environments", upDevEnvironments)
}

func upDevEnvironments(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	for _, table := range []string{"toolchain_install_jobs", "install_sources", "toolchain_definitions"} {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return err
			}
		}
	}
	models := []interface{}{
		&devEnvironmentMigrationModel{},
		&devEnvInstallSourceMigrationModel{},
		&devEnvJobMigrationModel{},
	}
	for _, item := range models {
		if db.Migrator().HasTable(item) {
			continue
		}
		if err := db.Migrator().CreateTable(item); err != nil {
			return err
		}
	}
	return seedBuiltinDevEnvironments(db)
}

type devEnvironmentMigrationModel struct {
	ID              uint      `gorm:"primaryKey"`
	Name            string    `gorm:"size:100;not null;uniqueIndex"`
	Kind            string    `gorm:"size:20;not null;default:builtin"`
	Executable      string    `gorm:"size:200;not null"`
	Description     string    `gorm:"size:500"`
	DetectScript    string    `gorm:"type:text"`
	InstallScript   string    `gorm:"type:text"`
	UpgradeScript   string    `gorm:"type:text"`
	UninstallScript string    `gorm:"type:text"`
	VersionsScript  string    `gorm:"type:text"`
	SwitchScript    string    `gorm:"type:text"`
	DefaultVersion  string    `gorm:"size:100"`
	CreatedBy       uint      `gorm:"index"`
	CreatedAt       time.Time `gorm:""`
	UpdatedAt       time.Time `gorm:""`
}

func (devEnvironmentMigrationModel) TableName() string { return "dev_environments" }

type devEnvInstallSourceMigrationModel struct {
	ID            uint      `gorm:"primaryKey"`
	EnvironmentID uint      `gorm:"not null;uniqueIndex:uidx_dev_env_source_name;index"`
	Name          string    `gorm:"size:100;not null;uniqueIndex:uidx_dev_env_source_name"`
	BaseURL       string    `gorm:"size:1000;not null"`
	Priority      int       `gorm:"not null;index"`
	Enabled       bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:""`
	UpdatedAt     time.Time `gorm:""`
}

func (devEnvInstallSourceMigrationModel) TableName() string { return "dev_env_install_sources" }

type devEnvJobMigrationModel struct {
	ID               uint       `gorm:"primaryKey"`
	EnvironmentID    uint       `gorm:"not null;index"`
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

func (devEnvJobMigrationModel) TableName() string { return "dev_env_jobs" }

type seededDevEnv struct {
	Env     devEnvironmentMigrationModel
	Sources []devEnvInstallSourceMigrationModel
}

func seedBuiltinDevEnvironments(db *gorm.DB) error {
	for _, item := range builtinDevEnvironments() {
		env := item.Env
		if err := db.Where("name = ?", env.Name).FirstOrCreate(&env).Error; err != nil {
			return err
		}
		for _, source := range item.Sources {
			source.EnvironmentID = env.ID
			if err := db.Where("environment_id = ? AND name = ?", env.ID, source.Name).
				FirstOrCreate(&source).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func builtinDevEnvironments() []seededDevEnv {
	return []seededDevEnv{
		{
			Env: devEnvironmentMigrationModel{
				Name: "Go", Kind: "builtin", Executable: "go", Description: "Go 编译环境（需要 asdf >= 0.16）",
				DetectScript: "go version", VersionsScript: "asdf list golang",
				InstallScript:   `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf plugin add golang https://github.com/asdf-community/asdf-golang.git 2>/dev/null || true; asdf install golang "$version" && asdf set -u golang "$version"`,
				UpgradeScript:   `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf plugin update golang || true; asdf install golang "$version" && asdf set -u golang "$version"`,
				UninstallScript: `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || version="$(asdf current golang 2>/dev/null | awk 'NR==1 {print $2}')"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf uninstall golang "$version"`,
				SwitchScript:    `command -v asdf >/dev/null 2>&1 || { echo 'asdf >= 0.16 is required to manage Go versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Go version is required'; exit 2; }; asdf set -u golang "$version"`,
			},
			Sources: []devEnvInstallSourceMigrationModel{
				{Name: "proxy.golang.org", BaseURL: "https://proxy.golang.org", Priority: 10, Enabled: true},
				{Name: "goproxy.cn", BaseURL: "https://goproxy.cn", Priority: 20, Enabled: true},
			},
		},
		{
			Env: devEnvironmentMigrationModel{
				Name: "Node.js", Kind: "builtin", Executable: "node", Description: "Node.js 运行时（需要 fnm）",
				DetectScript: "node --version", VersionsScript: "fnm list",
				InstallScript:   `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; export FNM_NODE_DIST_MIRROR="{{source_url}}"; fnm install "$version" && fnm default "$version"`,
				UpgradeScript:   `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; export FNM_NODE_DIST_MIRROR="{{source_url}}"; fnm install "$version" && fnm default "$version"`,
				UninstallScript: `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; fnm uninstall "$version"`,
				SwitchScript:    `command -v fnm >/dev/null 2>&1 || { echo 'fnm is required to manage Node.js versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Node.js version is required'; exit 2; }; fnm default "$version"`,
			},
			Sources: []devEnvInstallSourceMigrationModel{
				{Name: "nodejs.org", BaseURL: "https://nodejs.org/dist", Priority: 10, Enabled: true},
				{Name: "npmmirror", BaseURL: "https://npmmirror.com/mirrors/node", Priority: 20, Enabled: true},
			},
		},
		{
			Env: devEnvironmentMigrationModel{
				Name: "Java", Kind: "builtin", Executable: "java", Description: "Java 运行时（需要 SDKMAN!）",
				DetectScript: "java -version", VersionsScript: `. "$HOME/.sdkman/bin/sdkman-init.sh" && sdk list java`,
				InstallScript:   `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk install java "$version" && sdk default java "$version"`,
				UpgradeScript:   `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk install java "$version" && sdk default java "$version"`,
				UninstallScript: `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk uninstall java "$version"`,
				SwitchScript:    `[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ] || { echo 'SDKMAN! is required to manage Java versions'; exit 1; }; . "$HOME/.sdkman/bin/sdkman-init.sh"; version="{{version}}"; [ -n "$version" ] || { echo 'a Java identifier is required'; exit 2; }; sdk default java "$version"`,
			},
			Sources: []devEnvInstallSourceMigrationModel{
				{Name: "sdkman candidates", BaseURL: "https://api.sdkman.io/2", Priority: 10, Enabled: true},
			},
		},
		{
			Env: devEnvironmentMigrationModel{
				Name: "Python", Kind: "builtin", Executable: "python3", Description: "Python 运行时（需要 pyenv）",
				DetectScript: "python3 --version", VersionsScript: "pyenv versions",
				InstallScript:   `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; export PYTHON_BUILD_MIRROR_URL="{{source_url}}"; pyenv install --skip-existing "$version" && pyenv global "$version"`,
				UpgradeScript:   `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; export PYTHON_BUILD_MIRROR_URL="{{source_url}}"; pyenv install --skip-existing "$version" && pyenv global "$version"`,
				UninstallScript: `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; pyenv uninstall -f "$version"`,
				SwitchScript:    `command -v pyenv >/dev/null 2>&1 || { echo 'pyenv is required to manage Python versions'; exit 1; }; version="{{version}}"; [ -n "$version" ] || { echo 'a Python version is required'; exit 2; }; pyenv global "$version"`,
			},
			Sources: []devEnvInstallSourceMigrationModel{
				{Name: "python.org", BaseURL: "https://www.python.org/ftp/python", Priority: 10, Enabled: true},
				{Name: "npmmirror python", BaseURL: "https://npmmirror.com/mirrors/python", Priority: 20, Enabled: true},
			},
		},
	}
}
