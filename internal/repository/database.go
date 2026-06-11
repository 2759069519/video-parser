package repository

import (
	"fmt"
	"log"
	"video-parser/internal/config"
	"video-parser/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.DatabaseConfig) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移
	if err := DB.AutoMigrate(
		&model.Video{},
		&model.Atlas{},
		&model.Profile{},
		&model.ParseRecord{},
	); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	log.Println("数据库连接成功")
}