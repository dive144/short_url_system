package config

import (
	"fmt"
	"log"
	"short_url/global"
	"short_url/models"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDb() {
	db, err := gorm.Open(mysql.Open(Conf.Database.Dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database, %s", err)
	}
	global.Db = db
	s, err := global.Db.DB()
	if err != nil {
		log.Fatalf("Error connecting to database, %s", err)
	}
	s.SetMaxIdleConns(10)
	s.SetMaxOpenConns(100)
	s.SetConnMaxLifetime(time.Hour)

	//AutoMigrate是初始化操作，应该只在服务启动时执行一次
	err = global.Db.AutoMigrate(&models.ShortUrl{})

	if err != nil {
		panic(err)
	}
	fmt.Println(global.Db) //打印出了mysql了，成功连上了
}
