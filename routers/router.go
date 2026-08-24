package routers

import (
	_ "short_url/cmd/docs" // 千万不要忘了导入把你上一步生成的docs
	"short_url/handler"

	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
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
	r.GET("/s/:shorturl", handler.VisitShortUrl)
	r.POST("/api/v1/shorten", handler.GenerateShortUrl)
	r.GET("swagger/*any", gs.WrapHandler(files.Handler))
	return r
}
