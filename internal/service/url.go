package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aeilang/urlshortener/internal/model"
	"github.com/aeilang/urlshortener/internal/repo"
)

type ShortCodeGenerator interface {
	GenerateShortCode() string
}

type Cacher interface {
	SetURL(ctx context.Context, url repo.Url) error
	GetURL(ctx context.Context, shortCode string) (*repo.Url, error)
}

type URLService struct {
	querier            repo.Querier
	shortCodeGenerator ShortCodeGenerator
	defaultDuration    time.Duration
	cache              Cacher
	baseURL            string
}

func NewURLService(db *sql.DB, shortCodeGenerator ShortCodeGenerator, duration time.Duration, cache Cacher, baseURL string) *URLService {
	return &URLService{
		querier:            repo.New(db),
		shortCodeGenerator: shortCodeGenerator,
		defaultDuration:    duration,
		cache:              cache,
		baseURL:            baseURL,
	}
}

func (s *URLService) CreateURL(ctx context.Context, req model.CreateURLRequest) (*model.CreateURLResponse, error) {
	var shortCode string
	var isCustom bool
	var expiredAt time.Time

	if req.CustomCode != "" {
		isAvailabel, err := s.querier.IsShortCodeAvailable(ctx, req.CustomCode)
		if err != nil {
			return nil, err
		}
		if !isAvailabel {
			return nil, fmt.Errorf("别名已存在")
		}
		shortCode = req.CustomCode
		isCustom = true
	} else {
		code, err := s.getShortCode(ctx, 0)
		if err != nil {
			return nil, err
		}
		shortCode = code
	}

	if req.Duration == nil {
		expiredAt = time.Now().Add(s.defaultDuration)
	} else {
		expiredAt = time.Now().Add(time.Hour * time.Duration(*req.Duration))
	}

	// 插入数据库
	url, err := s.querier.CreateURL(ctx, repo.CreateURLParams{
		OriginalUrl: req.OriginalURL,
		ShortCode:   shortCode,
		IsCustom:    isCustom,
		ExpiredAt:   expiredAt,
	})
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if err := s.cache.SetURL(ctx, url); err != nil {
		return nil, err
	}

	return &model.CreateURLResponse{
		ShortURL:  s.baseURL + "/" + url.ShortCode,
		ExpiredAt: url.ExpiredAt,
	}, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (string, error) {
	// 先访问cache
	url, err := s.cache.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url != nil {
		return url.OriginalUrl, nil
	}

	// 访问数据库
	url2, err := s.querier.GetUrlByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// 存入缓存
	if err := s.cache.SetURL(ctx, url2); err != nil {
		return "", err
	}

	return url2.OriginalUrl, nil
}

func (s *URLService) getShortCode(ctx context.Context, n int) (string, error) {
	if n > 5 {
		return "", errors.New("重试过多")
	}
	shortCode := s.shortCodeGenerator.GenerateShortCode()

	isAvailable, err := s.querier.IsShortCodeAvailable(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if isAvailable {
		return shortCode, nil
	}

	return s.getShortCode(ctx, n+1)
}
