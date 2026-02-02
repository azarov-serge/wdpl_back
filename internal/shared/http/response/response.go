package response

import (
	"github.com/gofiber/fiber/v2"
)

// WriteError отправляет JSON {"error": message} с заданным статусом.
// Используется для единообразных ответов об ошибках и чтобы не светить внутренние детали.
func WriteError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

// WriteInternalError отправляет 500 с сообщением "internal error".
// Внутреннюю ошибку err в ответ не включаем (безопасность). Логирование — в middleware или вызывающем коде.
func WriteInternalError(c *fiber.Ctx, err error) error {
	_ = err
	return WriteError(c, fiber.StatusInternalServerError, "internal error")
}
