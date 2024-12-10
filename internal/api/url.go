package api

import (
	"context"
	"log"
	"net/http"

	"github.com/aeilang/urlshortener/internal/model"
	"github.com/labstack/echo/v4"
)

type URLService interface {
	CreateURL(ctx context.Context, req model.CreateURLRequest) (shortURL string, err error)

	GetURL(ctx context.Context, shortCode string) (originalURL string, err error)

	IncreViews(ctx context.Context, shortCode string) error

	GetURLs(ctx context.Context, req model.GetURLsRequest) ([]model.GetURLsResponse, error)
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

	userID, _ := c.Get("userID").(int)
	req.UserID = userID

	// 调用业务函数
	shortURL, err := h.urlService.CreateURL(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := model.CreateURLResponse{
		ShortURL: shortURL,
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

	// 开辟携程增加view
	go func() {
		if err := h.urlService.IncreViews(context.Background(), shortCode); err != nil {
			log.Printf("failed to incre %s's view ", shortCode)
		}
	}()

	return c.Redirect(http.StatusPermanentRedirect, originalURL)
}

// GET /api/urls
func (h *URLHandler) GetURLs(c echo.Context) error {
	userID, _ := c.Get("userID").(int)

	var req model.GetURLsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Size == 0 {
		req.Size = 10
	}
	req.UserID = userID

	resp, err := h.urlService.GetURLs(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, resp)
}
