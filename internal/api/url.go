package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aeilang/urlshortener/internal/model"
	"github.com/labstack/echo/v4"
)

type URLServicer interface {
	CreateURL(ctx context.Context, req model.CreateURLRequest) (shortURL string, err error)
	GetURL(ctx context.Context, shortCode string) (originalURL string, err error)
	IncreViews(ctx context.Context, shortCode string) error
	GetURLs(ctx context.Context, req model.GetURLsRequest) (*model.GetURLsResponse, error)
	DeleteURL(ctx context.Context, shortCode string) error
	UpdateURLDuration(ctx context.Context, req model.UpdateURLDurationReq) error
}

// URLHandler 处理URL相关的HTTP请求
type URLHandler struct {
	urlService URLServicer
}

// NewURLHandler 创建新的URL处理器
func NewURLHandler(urlService URLServicer) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// CreateURL godoc
// @Summary 创建短链接
// @Description 将长URL转换为短URL
// @Tags URL
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param request body model.CreateURLRequest true "创建短链接请求"
// @Success 201 {object} model.CreateURLResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/url [post]
func (h *URLHandler) CreateURL(c echo.Context) error {
	var req model.CreateURLRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, _ := c.Get("userID").(int)
	req.UserID = userID

	shortURL, err := h.urlService.CreateURL(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := model.CreateURLResponse{
		ShortURL: shortURL,
	}

	return c.JSON(http.StatusCreated, resp)
}

// RedirectURL godoc
// @Summary 重定向到原始URL
// @Description 通过短链接代码重定向到原始URL
// @Tags URL
// @Accept json
// @Produce json
// @Param code path string true "短链接代码"
// @Success 301 {string} string "重定向到原始URL"
// @Failure 500 {object} echo.HTTPError
// @Router /{code} [get]
func (h *URLHandler) RedirectURL(c echo.Context) error {
	shortCode := c.Param("code")
	fmt.Println(shortCode)

	originalURL, err := h.urlService.GetURL(c.Request().Context(), shortCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	go func() {
		if err := h.urlService.IncreViews(context.Background(), shortCode); err != nil {
			log.Printf("failed to incre %s's view ", shortCode)
		}
	}()

	return c.Redirect(http.StatusPermanentRedirect, originalURL)
}

// GetURLs godoc
// @Summary 获取用户的所有短链接
// @Description 分页获取当前用户创建的所有短链接
// @Tags URL
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} model.GetURLsResponse
// @Failure 500 {object} echo.HTTPError
// @Router /api/urls [get]
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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteURL godoc
// @Summary 删除短链接
// @Description 删除指定的短链接
// @Tags URL
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param code path string true "短链接代码"
// @Success 204 "No Content"
// @Failure 500 {object} echo.HTTPError
// @Router /api/url/{code} [delete]
func (h *URLHandler) DeleteURL(c echo.Context) error {
	shortCode := c.Param("code")

	if err := h.urlService.DeleteURL(c.Request().Context(), shortCode); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateURLDuration godoc
// @Summary 更新短链接有效期
// @Description 更新指定短链接的有效期
// @Tags URL
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param code path string true "短链接代码"
// @Param request body model.UpdateURLDurationReq true "更新有效期请求"
// @Success 204 "No Content"
// @Failure 500 {object} echo.HTTPError
// @Router /api/url/{code} [patch]
func (h *URLHandler) UpdateURLDuration(c echo.Context) error {
	var req model.UpdateURLDurationReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	req.Code = c.Param("code")

	if err := h.urlService.UpdateURLDuration(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
