package configs

import (
	"os"
)

// Config 应用配置
type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	MountPath string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "minidevops.db"),
		JWTSecret: getEnv("JWT_SECRET", "minidevops-secret-key-2024"),
		MountPath: getEnv("MOUNT_PATH", "~/dev-ops"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
