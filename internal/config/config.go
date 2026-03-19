package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	AccessTTL string `mapstructure:"access_ttl"`
	RefreshTTL string `mapstructure:"refresh_ttl"`
}

type BuildConfig struct {
	MaxConcurrent int    `mapstructure:"max_concurrent"`
	WorkspaceDir  string `mapstructure:"workspace_dir"`
	ArtifactDir   string `mapstructure:"artifact_dir"`
	LogDir        string `mapstructure:"log_dir"`
	CacheDir      string `mapstructure:"cache_dir"`
}

type EncryptionConfig struct {
	Key string `mapstructure:"key"`
}

type AdminConfig struct {
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	DisplayName string `mapstructure:"display_name"`
}

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Build      BuildConfig      `mapstructure:"build"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	Admin      AdminConfig      `mapstructure:"admin"`
}

var C *Config

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Join(".", "config"))
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found")
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	v.SetEnvPrefix("BUILDFLOW")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Resolve relative paths based on config file location
	configFile := v.ConfigFileUsed()
	configDir := filepath.Dir(configFile)
	if !filepath.IsAbs(configDir) {
		if abs, err := filepath.Abs(configDir); err == nil {
			configDir = abs
		}
	}

	cfg.Database.Path = resolvePath(configDir, cfg.Database.Path)
	cfg.Build.WorkspaceDir = resolvePath(configDir, cfg.Build.WorkspaceDir)
	cfg.Build.ArtifactDir = resolvePath(configDir, cfg.Build.ArtifactDir)
	cfg.Build.LogDir = resolvePath(configDir, cfg.Build.LogDir)
	cfg.Build.CacheDir = resolvePath(configDir, cfg.Build.CacheDir)

	C = &cfg
	return &cfg, nil
}

func resolvePath(baseDir, targetPath string) string {
	if targetPath == "" || filepath.IsAbs(targetPath) {
		return targetPath
	}
	return filepath.Join(baseDir, targetPath)
}
