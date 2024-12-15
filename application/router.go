// router.go 定义了应用程序的所有HTTP路由

package application

import (
	"github.com/aeilang/urlshortener/internal/mw"
)

// initRouter 初始化所有HTTP路由
func (a *Application) initRouter() {
	// 用户认证相关路由组 /api/auth
	u := a.e.Group("/api/auth")

	// 用户认证相关端点
	u.POST("/login", a.userHandler.Login)                  // 用户登录
	u.POST("/register", a.userHandler.Register)            // 用户注册
	u.POST("/forget", a.userHandler.ForgetPassword)        // 忘记密码
	u.GET("/register/:email", a.userHandler.SendEmailCode) // 发送注册验证码

	// URL缩短服务相关路由
	a.e.GET("/:code", a.urlHandler.RedirectURL) // 短链接重定向

	// 需要JWT认证的URL管理API
	url := a.e.Group("/api", mw.JWTAuther(a.jwt))
	url.POST("/url", a.urlHandler.CreateURL)                // 创建短链接
	url.GET("/urls", a.urlHandler.GetURLs)                  // 获取用户的所有短链接
	url.DELETE("/url/:code", a.urlHandler.DeleteURL)        // 删除短链接
	url.PATCH("/url/:code", a.urlHandler.UpdateURLDuration) // 更新短链接的有效期

}
