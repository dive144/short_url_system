package shortlink

import (
	"context"
	"errors"
	"log"
	"net/url"
	"short_url/utils"
	"strings"
	"time"
)

var ErrShortLinkExpired = errors.New(
	"short link expired",
)

var ErrShortLinkNotFound = errors.New(
	"short link not found",
)
var ErrCustomShortExists = errors.New(
	"custom short exists")
var ErrCustomShortLink = errors.New(
	"custom short link has an error")
var ErrInvalidOriginalURL = errors.New("invalid original URL")
var ErrInvalidExpireTime = errors.New("invalid expire time")

type ShortLinkService struct {
	repo *ShortLinkRepository
}

func NewShortLinkService(repo *ShortLinkRepository) *ShortLinkService {
	return &ShortLinkService{repo: repo}
}

func (service *ShortLinkService) Resolve(ctx context.Context, shortCode string) (string, error) {
	get, err := service.repo.Get(ctx, shortCode)
	if err != nil {
		return "", err
	}
	//time.Now().After(*existing.ExpireAt) 用来判断是否已经过期了   已过期执行过期操作
	if get.ExpireAt != nil &&
		!time.Now().Before(*get.ExpireAt) {
		//这里已经过期的链接，应该删除； 还差一个逻辑

		return "", ErrShortLinkExpired
	}
	// 链接存在并且没有过期，记录一次点击
	if err := service.repo.IncrementClickCount(
		ctx,
		get.ID,
	); err != nil {
		return "", err
	}
	return get.OriginalURL, nil
}

// Create
func (service *ShortLinkService) CreateShortLink(ctx context.Context, re *CreateShortUrlRequest) (*ShortUrl, error) {
	// 先校验参数，避免非法请求也访问数据库。
	if re == nil {
		return nil, ErrInvalidOriginalURL
	}
	re.OriginalUrl = strings.TrimSpace(re.OriginalUrl)
	target, err := url.Parse(re.OriginalUrl)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" {
		return nil, ErrInvalidOriginalURL
	}
	// nil 表示永久有效；设置了时间时，必须晚于当前时间。
	if re.ExpireAt != nil && !re.ExpireAt.After(time.Now()) {
		return nil, ErrInvalidExpireTime
	}

	var shortCode string
	re.CustomUrl = strings.TrimSpace(re.CustomUrl)
	if re.CustomUrl != "" {
		// 与请求结构体的 max=8、数据库 varchar(8) 保持一致。
		if len(re.CustomUrl) > 8 {
			return nil, ErrCustomShortLink
		}
		for _, char := range re.CustomUrl {
			if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9') {
				return nil, ErrCustomShortLink
			}
		}
		_, err := service.repo.Get(ctx, re.CustomUrl)
		if err != nil {
			if errors.Is(err, ErrShortLinkNotFound) {
				log.Printf("指定url无人使用,正在使用指定url生成短链接......")
				shortCode = re.CustomUrl
			} else {
				// MySQL连接失败、请求超时等未知错误
				return nil, err
			}
		} else {
			//当前定制的url已经存在了
			return nil, ErrCustomShortExists
		}
	}
	//用户没有指定url
	if re.CustomUrl == "" {
		//生成随机的短码
		var err error
		//用户并没有定制url短链接
		shortCode, err = utils.GenerateRandomShortCode(6)
		if err != nil {
			return nil, err
		}
	}
	shortURL := NewShortUrl(shortCode, re)
	//现在就要把shortURL传入repo保存到数据库中去了
	err = service.repo.CreateShortLink(ctx, shortURL)
	if err != nil {
		// 预查询后，其他请求仍可能抢先写入；以数据库唯一索引为最终判定。
		if re.CustomUrl != "" && errors.Is(err, ErrShortCodeConflict) {
			return nil, ErrCustomShortExists
		}
		return nil, err
	}
	return shortURL, nil
}
