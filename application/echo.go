// echo.go 负责配置和初始化Echo Web框架

package application

import (
	"github.com/aeilang/urlshortener/config"
	"github.com/aeilang/urlshortener/internal/mw"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// NewEcho 创建并配置一个新的Echo实例
func NewEcho(cfg config.ServerConfig, validator echo.Validator) *echo.Echo {
	e := echo.New()

	// 基本配置
	e.HideBanner = true      // 隐藏Echo的启动Banner
	e.HidePort = true        // 隐藏端口信息
	e.Server.WriteTimeout = cfg.WriteTimeout  // 设置写超时
	e.Server.ReadTimeout = cfg.ReadTimeout    // 设置读超时

	// 设置请求参数验证器
	e.Validator = validator

	// 添加中间件
	e.Use(mw.Logger)                // 日志中间件
	e.Use(middleware.Recover())     // 恢复中间件，用于处理panic
	e.Use(middleware.CORS())        // CORS中间件，处理跨域请求

	// 添加Swagger文档路由
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}
