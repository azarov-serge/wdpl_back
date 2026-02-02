package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteError(t *testing.T) {
	app := fiber.New()
	app.Get("/err", func(c *fiber.Ctx) error {
		return WriteError(c, fiber.StatusBadRequest, "invalid input")
	})

	res, err := app.Test(httptest.NewRequest("GET", "/err", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, res.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, "invalid input", body["error"])
}

func TestWriteError_StatusAndBody(t *testing.T) {
	app := fiber.New()
	app.Get("/custom", func(c *fiber.Ctx) error {
		return WriteError(c, fiber.StatusForbidden, "forbidden")
	})

	res, err := app.Test(httptest.NewRequest("GET", "/custom", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, res.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, "forbidden", body["error"])
}

func TestWriteInternalError(t *testing.T) {
	app := fiber.New()
	app.Get("/internal", func(c *fiber.Ctx) error {
		return WriteInternalError(c, assert.AnError)
	})

	res, err := app.Test(httptest.NewRequest("GET", "/internal", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, res.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	// В ответ не попадает текст внутренней ошибки (безопасность).
	assert.Equal(t, "internal error", body["error"])
}
