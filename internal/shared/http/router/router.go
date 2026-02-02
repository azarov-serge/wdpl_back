package router

import (
	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/features/auth"
	"wdpl_back/internal/features/events"
	"wdpl_back/internal/features/users"
	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/http/middleware"
	"wdpl_back/internal/shared/logger"
	"wdpl_back/internal/shared/postgres"
)

// NewFiberApp создаёт и настраивает Fiber‑приложение.
func NewFiberApp(cfg *config.Config, log logger.Logger, db *postgres.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork: false,
	})

	app.Use(middleware.RequestLogger(log))

	// TODO: добавить middleware: request_id, CORS.

	api := app.Group("/api")
	auth.RegisterRoutes(api, db, cfg)
	events.RegisterRoutes(api, db, cfg)
	users.RegisterRoutes(api, db, cfg)

	// healthz — конвенция из Kubernetes (liveness/readiness). Суффикс "z" отличает от путей вроде /health/...
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return app
}
