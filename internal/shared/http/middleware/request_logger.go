package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger возвращает middleware, логирующее метод, путь, статус и длительность запроса.
func RequestLogger(log interface {
	Error(string, ...any)
	Info(string, ...any)
}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		dur := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		if err != nil {
			log.Error("request", "method", method, "path", path, "status", status, "duration_ms", dur.Milliseconds(), "error", err)
			return err
		}
		log.Info("request", "method", method, "path", path, "status", status, "duration_ms", dur.Milliseconds())
		return nil
	}
}
