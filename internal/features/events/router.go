package events

import (
	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/http/middleware"
	"wdpl_back/internal/shared/postgres"
)

// Роли, допущенные к редактированию черновиков (STEP3).
var DraftEditorRoles = []string{"admin", "organizer", "editor"}

// RegisterRoutes вешает эндпоинты событий на api (обычно /api).
// Публичные: GET /events, GET /events/:id. Защищённые (редакторы): черновики и POST drafts/:id/publish.
func RegisterRoutes(api fiber.Router, db *postgres.DB, cfg *config.Config) {
	repos := NewPostgresRepository(db)
	svc := NewService(repos.Events, repos.Days, repos.Drafts, repos.DayDrafts)
	h := NewHandler(svc)

	// Сначала защищённые маршруты, чтобы /drafts и т.д. не попали в /:id.
	g := api.Group("/events", middleware.RequireAuth(cfg), middleware.RequireRole(DraftEditorRoles...))
	g.Get("/drafts", h.ListDrafts)
	g.Get("/drafts/:id", h.GetDraft)
	g.Post("/drafts", h.SaveDraft)
	g.Put("/drafts", h.SaveDraft)
	g.Post("/drafts/:id/publish", h.PublishDraft)
	g.Get("/day-drafts/:id", h.GetDayDraft)
	g.Get("/:eventId/day-drafts", h.ListDayDrafts)
	g.Post("/:eventId/day-drafts", h.SaveDayDraft)

	// Публичные: список и детали (без auth). "" — путь группы без суффикса (GET /api/events).
	public := api.Group("/events")
	public.Get("", h.ListEvents)
	public.Get("/:id", h.GetEvent)
}
