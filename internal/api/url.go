package api

import (
	"context"
	"net/http"

	"github.com/aeilang/urlshortener/internal/model"
	"github.com/labstack/echo/v4"
)

type URLService interface {
	CreateURL(ctx context.Context, req model.CreateURLRequest) (*model.CreateURLResponse, error)

	GetURL(ctx context.Context, shortCode string) (string, error)
}

type URLHandler struct {
	urlService URLService
}

func NewURLHandler(urlService URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// POST /api/url original_url, custom_code, duration，=> 短URL, 过期时间
func (h *URLHandler) CreateURL(c echo.Context) error {
	// 把数据提取
	var req model.CreateURLRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// 验证数据格式
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// 调用业务函数
	resp, err := h.urlService.CreateURL(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// 返回响应
	return c.JSON(http.StatusCreated, resp)
}

// GET /:code 把短URL重定向到长URL
func (h *URLHandler) RedirectURL(c echo.Context) error {
	// 把code 取出来
	shortCode := c.Param("code")

	// shortcode => url调用业务函数
	originalURL, err := h.urlService.GetURL(c.Request().Context(), shortCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusPermanentRedirect, originalURL)
}
