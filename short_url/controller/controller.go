package controller

import (
	"errors"
	"net/http"
	"short_url/global"
	"short_url/models"
	"short_url/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func VisitShortUrl(c *gin.Context) {
	var shortCode = c.Param("shorturl")
	var existing models.ShortUrl
	if err := global.Db.Where("short_code = ?", shortCode).First(&existing).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "您访问的短链接不存在或已被删除",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}
	}
	//time.Now().After(*existing.ExpireAt) 用来判断是否已经过期了   已过期执行过期操作
	if existing.ExpireAt != nil && time.Now().After(*existing.ExpireAt) {
		c.JSON(http.StatusGone, gin.H{
			"code":    410,
			"message": "您访问的短链接已过期",
			"data":    nil,
		})
		return
	}
	// 4. 原子性增加点击次数（并发安全）
	// 用UpdateColumn而不是Save，避免更新其他字段
	go func() {
		global.Db.Model(&models.ShortUrl{}).
			Where("short_code = ?", shortCode).
			UpdateColumn("click_count", gorm.Expr("click_count + 1"))
	}()

	c.Redirect(http.StatusFound, existing.OriginalURL)
}

// gin框架中的api是怎么写的呢；
func GenerateShortUrl(c *gin.Context) {
	print("GenerateShortUrl")
	var re models.CreateShortUrlRequest
	if err := c.ShouldBindJSON(&re); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	if re.CustomUrl != "" {
		var existing models.ShortUrl
		err1 := global.Db.Where("short_code = ?", re.CustomUrl).First(&existing).Error
		if err1 == nil {
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "自定义短码已存在",
				"data":    nil,
			})
			return
		}
		if err1 != gorm.ErrRecordNotFound {
			// 不是"没找到"的错误，说明数据库出错了
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
				"data":    nil,
			})
			return
		}
	}
	// 3. 核心：先插入一条空记录，获取自增ID     之后再修改这条空记录实现获取自增ID的功能
	shortUrl := &models.ShortUrl{
		OriginalURL: re.OriginalUrl,
		ExpireAt:    re.ExpireAt,
	}
	if re.CustomUrl != "" {
		shortUrl.ShortCode = re.CustomUrl
	}
	if err := global.Db.Create(&shortUrl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建短链接失败",
			"data":    nil,
		})
		return
	}

	var shortcode string
	if re.CustomUrl == "" {

		//但是这个时候shortUrl还不知道自己的id到底是多少呢
		//是知道的，Gorm 的Create()方法执行成功后，会自动把自增 ID 赋值给shortUrl.ID
		shortcode = utils.GenerateShortCode(shortUrl.ID)
		err := global.Db.UpdateColumn("short_code", shortcode).Error
		if err != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "生成短码失败",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "短链接创建成功",
		"data": gin.H{
			"short_code":   shortUrl.ShortCode,
			"short_url":    "http://localhost:8080/s/" + shortUrl.ShortCode,
			"original_url": shortUrl.OriginalURL,
			"create_at":    shortUrl.CreateAt,
			"expire_at":    shortUrl.ExpireAt,
			"click_count":  shortUrl.ClickCount,
		},
	})
}
