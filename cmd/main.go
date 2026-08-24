// @title 短链接服务 API
// @version 1.0
// @description 一个基于 Gin 的短链接生成与访问服务
// @host localhost:9090
// @BasePath /
package main

import (
	"short_url/config"
	"short_url/routers"
)

func main() {
	config.InitConfig()
	r := routers.SetupRouter()
	//router.Run() 会阻塞
	//Gin 的 Run() 方法会启动 HTTP 服务器并阻塞当前 goroutine，直到服务器关闭。
	r.Run(config.Conf.App.Port)
}
