package shortlink

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShortLinkHandler struct {
	service *ShortLinkService
}

func NewShortLinkHandler(service *ShortLinkService) *ShortLinkHandler {
	return &ShortLinkHandler{service: service}
}

// 短链接跳转服务
func (handler *ShortLinkHandler) Redirect(c *gin.Context) {
	var shortCode = c.Param("shorturl")
	OriginalUrl, err := handler.service.Resolve(c.Request.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(
			err,
			ErrShortLinkExpired,
		):
			c.JSON(
				http.StatusGone,
				gin.H{
					"code":    410,
					"message": "您访问的短链接已过期",
					"data":    nil,
				},
			)

		case errors.Is(
			err,
			ErrShortLinkNotFound,
		):
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"code":    404,
					"message": "短链接不存在",
				},
			)

		default:
			log.Printf("解析短链接失败，shortCode=%q，error=%v", shortCode, err)
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"code":    500,
					"message": "服务器内部错误",
				},
			)
		}
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, OriginalUrl)
}

//原始网址合法吗？
//过期时间合法吗？
//用户有没有指定短码？
//指定短码是否已被占用？
//没有指定时，系统如何生成短码？
//最后保存什么数据？

// CreateShortUrl 创建短链接
func (handler *ShortLinkHandler) CreateShortLink(c *gin.Context) {
	var re CreateShortUrlRequest
	if err := c.ShouldBindJSON(&re); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	shortLink, err := handler.service.CreateShortLink(c.Request.Context(), &re)
	if err != nil {
		status := http.StatusInternalServerError
		message := "服务器内部错误"
		switch {
		case errors.Is(err, ErrInvalidOriginalURL):
			status = http.StatusBadRequest
			message = "原始网址必须是包含主机名的 http 或 https 地址"
		case errors.Is(err, ErrInvalidExpireTime):
			status = http.StatusBadRequest
			message = "过期时间必须晚于当前时间"
		case errors.Is(err, ErrCustomShortLink):
			status = http.StatusBadRequest
			message = "自定义短码只能包含字母和数字，长度为1到8位"
		case errors.Is(err, ErrCustomShortExists):
			status = http.StatusConflict
			message = "自定义短码已经存在"
		default:
			// 内部错误只写日志，不把数据库信息返回给用户。
			log.Printf("创建短链接失败：%v", err)
		}
		c.JSON(status, gin.H{
			"code":    status,
			"message": message,
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":     201,
		"message":  "短链接创建成功",
		"ShortUrl": shortLink.ShortCode,
	})

}
