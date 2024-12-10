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
	"github.com/aeilang/urlshortener/pkg/emailsender"
	"github.com/aeilang/urlshortener/pkg/hasher"
	"github.com/aeilang/urlshortener/pkg/jwt"
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
	passwordHash       *hasher.PasswordHash
	emailSender        *emailsender.EmailSend
	urlService         *service.URLService
	userService        *service.UserService
	urlHandler         *api.URLHandler
	userHandler        *api.URLHandler
	cfg                *config.Config
	shortCodeGenerator *shortcode.ShortCode
	jwt                *jwt.JWT
}

func (a *Application) Init(filePath string) error {
	cfg, err := config.LoadConfig(filePath)
	if err != nil {
		return fmt.Errorf("加载配置错误: %w", err)
	}
	a.cfg = cfg

	logger.InitLogger(cfg.Logger)

	a.passwordHash = hasher.NewPasswordHash()
	a.jwt = jwt.NewJWT(cfg.JWT)
	a.emailSender, err = emailsender.NewEmailSend(cfg.Email)
	if err != nil {
		return fmt.Errorf("failed to new EmailSender: %v", err)
	}

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
	a.userService = service.NewUserService(db, a.passwordHash, a.jwt, redisClient)

	a.initEcho()

	a.initRouter()
	return nil
}

func (a *Application) initEcho() {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Server.WriteTimeout = a.cfg.Server.WriteTimeout
	e.Server.ReadTimeout = a.cfg.Server.ReadTimeout

	a.e.Validator = validator.NewCustomValidator()

	e.Use(mw.Logger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	a.e = e
}

func (a *Application) initRouter() {
	// Pubily
	a.e.GET("/:code", a.urlHandler.RedirectURL)

	// auth required
	a.e.POST("/api/url", a.urlHandler.CreateURL)
}

func (a *Application) Run() {
	go a.startServer()
	go a.SyncViewsToDB()
	a.shutdown()
}

func (a *Application) startServer() {
	if err := a.e.Start(a.cfg.Server.Addr); err != nil {
		log.Println(err)
	}
}

func (a *Application) SyncViewsToDB() {
	ticker := time.NewTicker(a.cfg.Redis.SyncViewDuration)
	defer ticker.Stop()

	for range ticker.C {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := a.urlService.SyncViewsToDB(ctx); err != nil {
				log.Printf("sync db failed: %v", err)
			}
		}()
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
