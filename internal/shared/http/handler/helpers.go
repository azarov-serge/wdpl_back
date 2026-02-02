package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TimeoutContext возвращает контекст с таймаутом из запроса и функцию отмены.
// В хендлерах: ctx, cancel := handler.TimeoutContext(c, 5*time.Second); defer cancel(); ...
func TimeoutContext(c *fiber.Ctx, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), d)
}

// LimitOffset парсит query limit и offset с дефолтами и верхней границей для limit.
// Возвращает (limit, offset). Invalid/negative значения заменяются на дефолты.
func LimitOffset(c *fiber.Ctx, defaultLimit, maxLimit, defaultOffset int) (limit, offset int) {
	limit = c.QueryInt("limit", defaultLimit)
	offset = c.QueryInt("offset", defaultOffset)
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = defaultOffset
	}
	return limit, offset
}
