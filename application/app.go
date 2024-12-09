package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aeilang/urlshortener/config"
	"github.com/aeilang/urlshortener/database"
	"github.com/aeilang/urlshortener/internal/api"
	"github.com/aeilang/urlshortener/internal/cache"
	"github.com/aeilang/urlshortener/internal/mw"
	"github.com/aeilang/urlshortener/internal/service"
	"github.com/aeilang/urlshortener/pkg/logger"
	"github.com/aeilang/urlshortener/pkg/shortcode"
	"github.com/aeilang/urlshortener/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Application struct {
	e                  *echo.Echo
	db                 *sql.DB
	redisClient        *cache.RedisCache
	urlService         *service.URLService
	urlHandler         *api.URLHandler
	cfg                *config.Config
	shortCodeGenerator *shortcode.ShortCode
}

func (a *Application) Init(filePath string) error {
	cfg, err := config.LoadConfig(filePath)
	if err != nil {
		return fmt.Errorf("加载配置错误: %w", err)
	}
	a.cfg = cfg

	logger.InitLogger(cfg.Logger)

	db, err := database.NewDB(cfg.Database)
	if err != nil {
		return err
	}
	a.db = db

	redisClient, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return err
	}
	a.redisClient = redisClient
	a.shortCodeGenerator = shortcode.NewShortCode(cfg.ShortCode.Length)

	a.urlService = service.NewURLService(db, a.shortCodeGenerator, cfg.App.DefaultDuration, redisClient, cfg.App.BaseURL)

	a.urlHandler = api.NewURLHandler(a.urlService)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Server.WriteTimeout = cfg.Server.WriteTimeout
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Use(mw.Logger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.POST("/api/url", a.urlHandler.CreateURL)
	e.GET("/:code", a.urlHandler.RedirectURL)
	e.Validator = validator.NewCustomValidator()
	a.e = e
	return nil
}

func (a *Application) Run() {
	go a.startServer()
	a.shutdown()
}

func (a *Application) startServer() {
	if err := a.e.Start(a.cfg.Server.Addr); err != nil {
		log.Println(err)
	}
}

func (a *Application) shutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	defer func() {
		if err := a.db.Close(); err != nil {
			log.Println(err)
		}
	}()

	defer func() {
		if err := a.redisClient.Close(); err != nil {
			log.Println(err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.e.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}
