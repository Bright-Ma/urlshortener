package api

import (
	"context"
	"errors"
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

type UserHandler struct {
	userService UserServicer
}

// POST /api/auth/login
// {email， password} -> {access_token, username, email}
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

// Post /api/auth/register
// {username, email, password, email_code}
// {access_token, username, password}
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

// POST /api/auth/forget
// {email, password, email_code} -> {access_token, username, email}
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
		return echo.NewHTTPError(http.StatusBadRequest, "该邮箱未注册")
	}

	if err != nil && !errors.Is(err, model.ErrEmailAleadyExist) {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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

// GET /api/auth/register/:email
func (h *UserHandler) SendEmailCode(c echo.Context) error {
	email := c.Param("email")
	req := model.SendCodeRequest{
		Email: email,
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.userService.SendEmailCode(c.Request().Context(), email); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}
