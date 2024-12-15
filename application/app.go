package application

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aeilang/urlshortener/config"
	"github.com/aeilang/urlshortener/database"
	"github.com/aeilang/urlshortener/internal/api"
	"github.com/aeilang/urlshortener/internal/cache"
	"github.com/aeilang/urlshortener/internal/service"
	"github.com/aeilang/urlshortener/pkg/emailsender"
	"github.com/aeilang/urlshortener/pkg/hasher"
	"github.com/aeilang/urlshortener/pkg/jwt"
	"github.com/aeilang/urlshortener/pkg/logger"
	"github.com/aeilang/urlshortener/pkg/randnum"
	"github.com/aeilang/urlshortener/pkg/shortcode"
	"github.com/aeilang/urlshortener/pkg/validator"
	"github.com/labstack/echo/v4"
)

type Application struct {
	e           *echo.Echo
	db          *sql.DB
	redisClient *cache.RedisCache
	urlService  *service.URLService
	urlHandler  *api.URLHandler
	userHandler *api.UserHandler
	cfg         *config.Config
	jwt         *jwt.JWT
}

func InitApp(filePath string) (*Application, error) {
	// 加载配置
	cfg, err := config.NewConfig(filePath)
	if err != nil {
		return nil, err
	}

	// 初始化logger
	logger.InitLogger(cfg.Logger)

	// 初始化数据库
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		return nil, err
	}

	// 初始化redis
	redisClient, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return nil, err
	}

	// 初始化pkg
	emailSender, err := emailsender.NewEmailSend(cfg.Email)
	if err != nil {
		return nil, err
	}

	passwordHash := hasher.NewPasswordHash()

	jwt := jwt.NewJWT(cfg.JWT)

	randNum := randnum.NewRandNum(cfg.RandNum)

	shortCode := shortcode.NewShortCode(cfg.ShortCode)

	customValidator := validator.NewCustomValidator()

	// service
	urlService := service.NewURLService(db, shortCode, redisClient, cfg.App)
	userService := service.NewUserService(db, passwordHash, jwt, redisClient, emailSender, randNum)

	// handler
	urlHandler := api.NewURLHandler(urlService)
	userHandler := api.NewUserHandler(userService)

	// echo
	e := NewEcho(cfg.Server, customValidator)

	a := &Application{
		e:           e,
		db:          db,
		redisClient: redisClient,
		urlService:  urlService,
		urlHandler:  urlHandler,
		userHandler: userHandler,
		cfg:         cfg,
		jwt:         jwt,
	}

	a.initRouter()
	return a, nil
}

func (a *Application) Start() {
	go a.syncViewsToDB()

	go func() {
		if err := a.e.Start(a.cfg.Server.Addr); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	// wait to gracefully shutdown
	a.shutdown()
}

func (a *Application) syncViewsToDB() {
	ticker := time.NewTicker(a.cfg.App.SyncViewDuration)
	defer ticker.Stop()

	for range ticker.C {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := a.urlService.SyncViewsToDB(ctx); err != nil {
				logger.Error(err.Error())
			}
		}()
	}

}

// gracefully shutdown
func (a *Application) shutdown() {
	defer func() {
		if err := a.db.Close(); err != nil {
			logger.Error(err.Error())
		}
	}()
	defer func() {
		if err := a.db.Close(); err != nil {
			logger.Error(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.e.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
}
