package models

import (
	"time"

	"gorm.io/gorm"
)

//结构体是 GORM 这个 ORM 框架存在的根本意义它是Go 语言的内存对象和MySQL 数据库的表之间的唯一桥梁
//ORM = Object-Relational Mapping（对象关系映射）它的核心目标就是：
//用操作 Go 结构体的方式，来操作数据库表不用写一行原生 SQL，就能完成增删改查
//GORM 依赖反射，只有结构体才能通过反射获取字段名、类型、标签等信息

type ShortUrl struct {
	ID          uint64 `gorm:"primary_key;autoIncrement;column:id;type:bigint unsigned;" json:"-"`
	ShortCode   string `gorm:"column:short_code;type:varchar(8);uniqueIndex;not null" json:"short_code"`
	OriginalURL string `gorm:"column:original_url;type:text;not null" json:"original_url"`
	//为什么结构体里必须定义所有数据库中存在的字段 ——不是为了写，而是为了读。
	CreateAt   time.Time      `gorm:"column:create_at;type:datetime;default:CURRENT_TIMESTAMP" json:"create_at"`
	ExpireAt   *time.Time     `gorm:"column:expire_at;type:datetime" json:"expire_at,omitempty"`
	ClickCount int64          `gorm:"column:click_count;type:bigint;default:0;not null" json:"click_count"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index" json:"-"`
}

type CreateShortUrlRequest struct {
	OriginalUrl string     `json:"original_url"`
	CustomUrl   string     `json:"custom_code" binding:"omitempty,alphanum,min=1,max=8"`
	ExpireIn    string     `json:"expire_time" binding:"omitempty"`
	ExpireAt    *time.Time `json:"expire_at" binding:"omitempty"`
}
