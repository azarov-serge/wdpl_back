package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/shared/authutils"
	"wdpl_back/internal/shared/config"
)

// LocalsKeyClaims — ключ для хранения JWT‑клеймов в fiber.Ctx.Locals.
const LocalsKeyClaims = "claims"

// RequireAuth возвращает middleware: извлекает Bearer JWT, парсит и кладёт клеймы в c.Locals(LocalsKeyClaims).
// При отсутствии или невалидном токене отвечает 401.
func RequireAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization"})
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization"})
		}
		tokenStr := strings.TrimPrefix(auth, prefix)
		claims, err := authutils.ParseAccessToken(cfg, tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals(LocalsKeyClaims, claims)
		return c.Next()
	}
}

// RequireRole возвращает middleware: проверяет, что роль пользователя из клеймов входит в allowedRoles.
// Ожидает, что RequireAuth уже выполнен и клеймы лежат в c.Locals(LocalsKeyClaims).
func RequireRole(allowedRoles ...string) fiber.Handler {
	set := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		set[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		claimsVal := c.Locals(LocalsKeyClaims)
		if claimsVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		claims, ok := claimsVal.(*authutils.UserClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		if _, ok := set[claims.Role]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}
