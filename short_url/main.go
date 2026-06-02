package main

import (
	"fmt"
	"short_url/config"
	"short_url/routers"
)

func main() {
	fmt.Println("=== main start ===")

	config.InitConfig()
	r := routers.SetupRouter()
	//router.Run() 会阻塞
	//Gin 的 Run() 方法会启动 HTTP 服务器并阻塞当前 goroutine，直到服务器关闭。
	r.Run(config.Conf.App.Port)
}
