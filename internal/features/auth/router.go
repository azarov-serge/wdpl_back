package auth

import (
	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/postgres"
)

// RegisterRoutes вешает эндпоинты авторизации на api (обычно /api).
func RegisterRoutes(api fiber.Router, db *postgres.DB, cfg *config.Config) {
	repo := NewPostgresRepository(db)
	svc := NewService(repo, repo, cfg)
	h := NewHandler(svc)

	g := api.Group("/auth")
	g.Post("/sign-up", h.SignUp)
	g.Post("/sign-in", h.SignIn)
	g.Post("/sign-out", h.SignOut)
	g.Post("/refresh", h.Refresh)
}
