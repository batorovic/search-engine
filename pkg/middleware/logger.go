package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewLoggerMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.Locals("requestid")

		requestBody := string(c.Body())

		err := c.Next()

		duration := time.Since(start)

		status := c.Response().StatusCode()

		responseBody := string(c.Response().Body())

		fields := []zap.Field{
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("query", c.Context().QueryArgs().String()),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
			zap.String("request_body", requestBody),
			zap.String("response_body", responseBody),
			zap.Int("body_size", len(c.Response().Body())),
		}

		switch {
		case status >= 500:
			logger.Error("request completed", fields...)
		case status >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}

		return err
	}
}
