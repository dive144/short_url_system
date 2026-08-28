package shortlink

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrShortCodeConflict = errors.New("short code already exists")

const (
	shortLinkCacheKeyPrefix = "shorturl:v1:link:"
	shortLinkCacheTTL       = 10 * time.Minute
)

type ShortLinkRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewShortLinkRepository(db *gorm.DB, rdb *redis.Client) *ShortLinkRepository {
	return &ShortLinkRepository{db: db, rdb: rdb}
}

//读取流程有问题      → 看Get()
//MySQL查询有问题     → 看getFromMySQL()
//Redis回填有问题     → 看cacheShortLink()

// 根据短码查询整条数据
func (repo *ShortLinkRepository) Get(ctx context.Context, shortCode string) (*ShortUrl, error) {
	shortCode = strings.TrimSpace(shortCode)
	if shortCode == "" {
		return nil, ErrShortLinkNotFound
	}

	key := shortLinkCacheKeyPrefix + shortCode

	if repo.rdb != nil {
		cachedData, err := repo.rdb.Get(ctx, key).Bytes()
		switch {
		case err == nil:
			var shortURL ShortUrl
			if err := json.Unmarshal(cachedData, &shortURL); err == nil {
				return &shortURL, nil
			}
			// 缓存内容损坏时删除该 Key，并继续回源 MySQL。
			log.Printf("Redis缓存数据损坏，key=%s，正在回源MySQL", key)
			if err := repo.rdb.Del(ctx, key).Err(); err != nil && ctx.Err() == nil {
				log.Printf("删除损坏的Redis缓存失败，key=%s，error=%v", key, err)
			}

		case errors.Is(err, redis.Nil):
			// 缓存未命中是正常情况，继续查询 MySQL。

		case ctx.Err() != nil:
			return nil, ctx.Err()

		default:
			// Redis 是缓存；运行时故障不能阻止核心查询链路。
			log.Printf("Redis查询失败，key=%s，error=%v，正在降级查询MySQL", key, err)
		}
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}

	shortURL, err := repo.getFromMySQL(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// 回填缓存属于尽力而为：失败时记录日志，但仍然返回 MySQL 的正确结果。
	repo.cacheShortLink(ctx, key, shortURL)

	return shortURL, nil
}

func (repo *ShortLinkRepository) getFromMySQL(ctx context.Context, shortCode string) (*ShortUrl, error) {
	var shortURL ShortUrl

	err := repo.db.
		WithContext(ctx).
		Where("short_code = ?", shortCode).
		First(&shortURL).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrShortLinkNotFound
	}

	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}

		return nil, fmt.Errorf(
			"查询短链接失败，shortCode=%q: %w",
			shortCode,
			err,
		)
	}

	return &shortURL, nil
}

func (repo *ShortLinkRepository) cacheShortLink(ctx context.Context, key string, shortURL *ShortUrl) {
	if repo.rdb == nil || shortURL == nil || ctx.Err() != nil {
		return
	}

	ttl := shortLinkCacheTTL
	if shortURL.ExpireAt != nil {
		remaining := time.Until(*shortURL.ExpireAt)
		if remaining <= 0 {
			return
		}
		if remaining < ttl {
			ttl = remaining
		}
	}

	data, err := json.Marshal(shortURL)
	if err != nil {
		log.Printf("序列化短链接缓存失败，key=%s，error=%v", key, err)
		return
	}
	if err := repo.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("Redis回填失败，key=%s，error=%v", key, err)
	}
}

// Create
func (repo *ShortLinkRepository) CreateShortLink(ctx context.Context, shorturl *ShortUrl) error {
	create := repo.db.WithContext(ctx).Create(shorturl)
	err := create.Error
	if err != nil {
		// 1062 是 MySQL 唯一键冲突；同时兼容 GORM 开启错误翻译的情况。
		var mysqlErr *mysqldriver.MySQLError
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			(errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) {
			return fmt.Errorf("创建短链接失败，shortCode=%q: %w", shorturl.ShortCode, ErrShortCodeConflict)
		}
		return fmt.Errorf("创建短链接失败，shortCode=%q: %w", shorturl.ShortCode, err)
	}
	return nil
}

func (repo *ShortLinkRepository) IncrementClickCount(ctx context.Context, id uint64) error {
	result := repo.db.
		WithContext(ctx).
		Model(&ShortUrl{}).
		Where("id = ?", id).
		UpdateColumn(
			"click_count",
			gorm.Expr("click_count + ?", 1),
		)

	if result.Error != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		return fmt.Errorf(
			"增加短链接点击次数失败，id=%d: %w",
			id,
			result.Error,
		)
	}

	if result.RowsAffected == 0 {
		return ErrShortLinkNotFound
	}
	return nil
}
