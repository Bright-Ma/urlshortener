package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aeilang/urlshortener/config"
	"github.com/aeilang/urlshortener/internal/model"
	"github.com/aeilang/urlshortener/internal/repo"
)

type ShortCodeGenerator interface {
	GenerateShortCode() string
}

type URLCacher interface {
	SetURL(ctx context.Context, url model.URL) error
	GetURL(ctx context.Context, shortCode string) (string, error)
	DelURL(ctx context.Context, shortCode string) error
	IncreViews(ctx context.Context, shortCode string) error
	ScanViews(ctx context.Context, cursor uint64, batchSize int64) (keys []string, nextCursor uint64, err error)
	GetViews(ctx context.Context, shortCode string) (int, error)
	DelViews(ctx context.Context, shortCode string) error
}

type URLService struct {
	querier            repo.Querier
	shortCodeGenerator ShortCodeGenerator
	cache              URLCacher
	urlDefaultDuration time.Duration
	baseURL            string
}

func NewURLService(db *sql.DB, shortCodeGenerator ShortCodeGenerator, cache URLCacher, cfg config.AppConfig) *URLService {
	return &URLService{
		querier:            repo.New(db),
		shortCodeGenerator: shortCodeGenerator,
		cache:              cache,
		urlDefaultDuration: cfg.DefaultDuration,
		baseURL:            cfg.BaseURL,
	}
}

func (s *URLService) CreateURL(ctx context.Context, req model.CreateURLRequest) (shortURL string, err error) {
	var shortCode string
	var isCustom bool
	var expiredAt time.Time

	if req.CustomCode != "" {
		isAvailabel, err := s.querier.IsShortCodeAvailable(ctx, req.CustomCode)
		if err != nil {
			return "", err
		}
		if !isAvailabel {
			return "", fmt.Errorf("别名已存在")
		}
		shortCode = req.CustomCode
		isCustom = true
	} else {
		code, err := s.getShortCode(ctx, 0)
		if err != nil {
			return "", err
		}
		shortCode = code
	}

	if req.Duration == nil {
		expiredAt = time.Now().Add(s.urlDefaultDuration)
	} else {
		expiredAt = time.Now().Add(time.Hour * time.Duration(*req.Duration))
	}

	// 插入数据库
	if err := s.querier.CreateURL(ctx, repo.CreateURLParams{
		OriginalUrl: req.OriginalURL,
		ShortCode:   shortCode,
		IsCustom:    isCustom,
		ExpiredAt:   expiredAt,
		UserID:      int32(req.UserID),
	}); err != nil {
		return "", err
	}

	url := model.URL{
		OriginalURL: req.OriginalURL,
		ShortCode:   shortCode,
	}

	// 存入缓存
	if err := s.cache.SetURL(ctx, url); err != nil {
		return "", err
	}

	ShortURL := s.baseURL + "/" + url.ShortCode

	return ShortURL, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (originalURL string, err error) {
	// 先访问cache
	originalURL, err = s.cache.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if originalURL != "" {
		return originalURL, nil
	}

	// 访问数据库
	row, err := s.querier.GetUrlByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	url := model.URL{
		OriginalURL: row.OriginalUrl,
		ShortCode:   row.ShortCode,
	}

	// 存入缓存
	if err := s.cache.SetURL(ctx, url); err != nil {
		return "", err
	}

	return url.OriginalURL, nil
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

func (s *URLService) GetURLs(ctx context.Context, req model.GetURLsRequest) (*model.GetURLsResponse, error) {
	rows, err := s.querier.GetURLsByUserID(ctx, repo.GetURLsByUserIDParams{
		UserID: int32(req.UserID),
		Limit:  int32(req.Size),
		Offset: int32((req.Page - 1) * req.Size),
	})
	if err != nil {
		return nil, err
	}

	items := make([]model.FullURL, len(rows))
	total := 0

	for i := range rows {
		row := &rows[i]
		views, err := s.cache.GetViews(ctx, row.ShortCode)
		if err != nil {
			return nil, err
		}

		row.Views += int32(views)

		items[i] = model.FullURL{
			OriginalURL: row.OriginalUrl,
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, row.ShortCode),
			ExpiredAt:   row.ExpiredAt,
			IsCustom:    row.IsCustom,
			Views:       uint(row.Views),
			ID:          int(row.ID),
		}
		total = int(row.Total)
	}

	resp := model.GetURLsResponse{
		Items: items,
		Total: total,
	}

	return &resp, nil
}

func (s *URLService) DeleteURL(ctx context.Context, code string) error {
	if err := s.querier.DeleteURLByShortCode(ctx, code); err != nil {
		return err
	}

	if err := s.cache.DelURL(ctx, code); err != nil {
		return err
	}

	if err := s.cache.DelViews(ctx, code); err != nil {
		return err
	}

	return nil
}

func (s *URLService) UpdateURLDuration(ctx context.Context, req model.UpdateURLDurationReq) error {
	return s.querier.UpdateURLExpiredByShortCode(ctx, repo.UpdateURLExpiredByShortCodeParams{
		ExpiredAt: req.ExpiredAt,
		ShortCode: req.Code,
	})
}

func (s *URLService) IncreViews(ctx context.Context, shortCode string) error {
	return s.cache.IncreViews(ctx, shortCode)
}

func (s *URLService) SyncViewsToDB(ctx context.Context) error {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = s.cache.ScanViews(ctx, cursor, 100)
		if err != nil {
			return err
		}

		for _, key := range keys {
			views, err := s.cache.GetViews(ctx, key)
			if err != nil {
				return err
			}

			if views == 0 {
				continue
			}

			if err := s.cache.DelViews(ctx, key); err != nil {
				return err
			}

			shortCode := strings.Split(key, ":")[1]

			if err := s.querier.UpdateViewsByShortCode(ctx, repo.UpdateViewsByShortCodeParams{
				Views:     int32(views),
				ShortCode: shortCode,
			}); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
