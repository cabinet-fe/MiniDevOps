package db

import (
	"context"
	"log"

	"entgo.io/ent/dialect"
	"minidevops/internal/model"
	"minidevops/internal/model/migrate"
)

// NewClient 创建并初始化ent客户端
func NewClient(databaseURL string) (*model.Client, error) {
	client, err := model.Open(dialect.SQLite, databaseURL)
	if err != nil {
		return nil, err
	}

	// 运行数据库迁移
	if err := migrate.Create(context.Background(), client); err != nil {
		log.Fatalf("创建数据库结构失败: %v", err)
		return nil, err
	}

	return client, nil
}