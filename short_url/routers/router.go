package routers

import (
	"short_url/controller"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	// ✅ 加在所有路由的最前面！
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "✅ 路由系统完全正常！")
	})
	r.POST("/test", func(c *gin.Context) {
		c.String(200, "✅ POST路由也完全正常！")
	})
	//参数占位符 = 留一个空位  这个位置不写死，用户填什么我接收什么，把内容存到名叫 code 的参数里
	r.GET("/s/:shorturl", controller.VisitShortUrl)
	r.POST("/api/v1/shorten", controller.GenerateShortUrl)

	return r
}
