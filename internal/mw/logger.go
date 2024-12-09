package mw

import (
	"time"

	"github.com/aeilang/urlshortener/pkg/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)
		if err != nil {
			c.Error(err)
		}

		req := c.Request()
		res := c.Response()

		fields := []zap.Field{
			zap.String("remote_ip", c.RealIP()),
			zap.String("latency", time.Since(start).String()),
			zap.String("host", req.Host),
			zap.String("request", req.Method+" "+req.RequestURI),
			zap.Int("status", res.Status),
			zap.Int64("size", res.Size),
			zap.String("user_agent", req.UserAgent()),
		}

		id := req.Header.Get(echo.HeaderXRequestID)
		if id != "" {
			fields = append(fields, zap.String("request_id", id))
		}

		n := res.Status
		switch {
		case n >= 500:
			logger.Error("Server error", fields...)
		case n >= 400:
			logger.Warn("Client error", fields...)
		case n >= 300:
			logger.Info("Redirection", fields...)
		default:
			logger.Info("Success", fields...)
		}

		return nil
	}
}
