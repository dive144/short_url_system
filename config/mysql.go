package config

import (
	"fmt"
	"short_url/internal/shortlink"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDb(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.Database.Dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	s, err := db.DB()
	if err != nil {
		return nil, err
	}
	s.SetMaxIdleConns(10)
	s.SetMaxOpenConns(100)
	s.SetConnMaxLifetime(time.Hour)

	fmt.Println(db) //打印出了mysql了，成功连上了
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&shortlink.ShortUrl{})
}

func CloseDb(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
