package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	ServerAddr  string
	DatabaseURL string
	JWTSecret   string
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "file:./data/minidevops.db?cache=shared&_fk=1"),
		JWTSecret:   getEnv("JWT_SECRET", "minidevops_secret_key"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}