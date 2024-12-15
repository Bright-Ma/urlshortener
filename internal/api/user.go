package api

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/aeilang/urlshortener/internal/model"
	"github.com/labstack/echo/v4"
)

type UserServicer interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	IsEmailAvaliable(ctx context.Context, email string) error
	Register(ctx context.Context, req model.RegisterReqeust) (*model.LoginResponse, error)
	SendEmailCode(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req model.ForgetPasswordReqeust) (*model.LoginResponse, error)
}

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
	userService UserServicer
}

// NewUserHandler 创建新的用户处理器
func NewUserHandler(userService UserServicer) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Login godoc
// @Summary 用户登录
// @Description 用户通过邮箱和密码登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录请求"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.userService.Login(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, model.ErrUserNameOrPasswordFailed) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

// Register godoc
// @Summary 用户注册
// @Description 新用户注册
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body model.RegisterReqeust true "注册请求"
// @Success 201 {object} model.LoginResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c echo.Context) error {
	var req model.RegisterReqeust
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.userService.IsEmailAvaliable(c.Request().Context(), req.Email); err != nil {
		if errors.Is(err, model.ErrUserNameOrPasswordFailed) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp, err := h.userService.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, model.ErrEmailCodeNotEqual) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

// ForgetPassword godoc
// @Summary 忘记密码
// @Description 通过邮箱验证码重置密码
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body model.ForgetPasswordReqeust true "重置密码请求"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/auth/forget [post]
func (h *UserHandler) ForgetPassword(c echo.Context) error {
	var req model.ForgetPasswordReqeust
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.userService.IsEmailAvaliable(c.Request().Context(), req.Email)
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "邮箱不存在")
	}

	resp, err := h.userService.ResetPassword(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, model.ErrEmailCodeNotEqual) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

// SendEmailCode godoc
// @Summary 发送邮箱验证码
// @Description 向指定邮箱发送验证码
// @Tags 用户
// @Accept json
// @Produce json
// @Param email path string true "邮箱地址"
// @Success 204 "No Content"
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/auth/register/{email} [get]
func (h *UserHandler) SendEmailCode(c echo.Context) error {
	email := c.Param("email")
	log.Printf("email: %s", email)

	if err := h.userService.SendEmailCode(c.Request().Context(), email); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
