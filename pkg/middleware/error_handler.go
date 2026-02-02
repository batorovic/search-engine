package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		msg := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			msg = e.Message
		}

		logger.Error("request error",
			zap.Error(err),
			zap.Int("status", code),
			zap.String("path", c.Path()),
		)

		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": msg,
			},
			"meta": fiber.Map{
				"request_id": c.Locals("requestid"),
			},
		})
	}
}
