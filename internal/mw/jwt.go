package mw

import (
	"net/http"
	"strings"

	"github.com/aeilang/urlshortener/pkg/jwt"
	"github.com/labstack/echo/v4"
)

func JWTAuther(jwt *jwt.JWT) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authentication")
			ls := strings.Split(authHeader, " ")
			if len(ls) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			if ls[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			tokenString := ls[1]

			claims, err := jwt.ParseToken(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			c.Set("email", claims.Email)
			c.Set("userID", claims.UserID)

			return next(c)
		}
	}
}
