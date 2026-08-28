// @title 短链接服务 API
// @version 1.0
// @description 一个基于 Gin 的短链接生成与访问服务
// @host localhost:9090
// @BasePath /
package main

import (
	"log"
	"short_url/config"
	"short_url/routers"
)

func main() {
	//加载本地配置文件
	cfg := config.SetUpConfig()

	//连接mysql数据库
	db, err := config.NewDb(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	err = config.AutoMigrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	defer config.CloseDb(db) //项目结束时关闭数据库

	// 连接 Redis
	rdb, err := config.NewRedisClient(cfg) //现在获取到了redis的客户端了，接下来应该做些什么呢
	if err != nil {
		log.Printf("Redis unavailable, falling back to MySQL: %v", err)
	} else {
		log.Printf("Redis connected (cache enabled)")
		defer config.CloseRedisClient(rdb) //项目结束时关闭Redis
	}

	//设置路由
	r := routers.SetupRouter(db, rdb)
	log.Printf("Server is running on port %s", cfg.App.Port)
	//router.Run() 会阻塞
	//Gin 的 Run() 方法会启动 HTTP 服务器并阻塞当前 goroutine，直到服务器关闭。
	r.Run(cfg.App.Port)
}
