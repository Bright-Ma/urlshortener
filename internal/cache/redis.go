package cache

import (
	"context"
	"time"

	"github.com/aeilang/urlshortener/config"
	"github.com/aeilang/urlshortener/internal/model"
	"github.com/go-redis/redis/v8"
)

const urlPrifix = "url:"

type RedisCache struct {
	client            *redis.Client
	urlDuration       time.Duration
	emailCodeDuration time.Duration
}

func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client:            client,
		urlDuration:       cfg.UrlDuration,
		emailCodeDuration: cfg.EmailCodeDuration,
	}, nil
}

func (c *RedisCache) SetURL(ctx context.Context, url model.URL) error {
	if err := c.client.Set(ctx, urlPrifix+url.ShortCode, url.OriginalURL, c.urlDuration).Err(); err != nil {
		return err
	}

	return nil
}

func (c *RedisCache) GetURL(ctx context.Context, shortCode string) (originalURL string, err error) {
	originalURL = c.client.Get(ctx, urlPrifix+shortCode).String()

	return originalURL, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
