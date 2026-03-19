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
}

type EncryptionConfig struct {
	Key string `mapstructure:"key"`
}

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig `mapstructure:"database"`
	JWT       JWTConfig      `mapstructure:"jwt"`
	Build     BuildConfig    `mapstructure:"build"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
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

	C = &cfg
	return &cfg, nil
}
