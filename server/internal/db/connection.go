package db

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接
func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("mini-dev-ops.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		panic("数据库连接失败")
	}

	sqlDB, _ := db.DB()

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 运行数据库迁移
	if err := RunMigrations(db); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化种子数据
	if err := SeedData(db); err != nil {
		log.Fatal("种子数据初始化失败:", err)
	}

	return db, err
}
