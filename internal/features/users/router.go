package users

import (
	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/features/auth"
	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/http/middleware"
	"wdpl_back/internal/shared/postgres"
)

// RegisterRoutes вешает эндпоинты профиля пользователя на api (под /api/users).
// Требуется авторизация (JWT) для GET/PUT /me.
func RegisterRoutes(api fiber.Router, db *postgres.DB, cfg *config.Config) {
	authRepo := auth.NewPostgresRepository(db)
	profileRepo := NewPostgresProfileRepository(db)
	svc := NewService(profileRepo)
	h := NewHandler(svc, authRepo)

	g := api.Group("/users")
	g.Get("/me", middleware.RequireAuth(cfg), h.GetMe)
	g.Put("/me", middleware.RequireAuth(cfg), h.PutMe)
}
