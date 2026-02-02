package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// Интеграционный тест: POST /api/auth/sign-in с мок-сервисом (без БД).
func TestHandler_SignIn_Success(t *testing.T) {
	s, _, _ := newTestService(t)
	h := NewHandler(s)
	app := fiber.New()
	app.Post("/api/auth/sign-in", h.SignIn)

	// Регистрируем пользователя через сервис.
	_, _, err := s.Register(ctxBackground(), "signin@example.com", "password123")
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"email":"signin@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/api/auth/sign-in", body)
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)

	var resp AuthResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.NotEmpty(t, resp.AccessToken)
	require.NotEmpty(t, resp.RefreshToken)
	require.Equal(t, "signin@example.com", resp.Email)
}

func TestHandler_SignIn_InvalidBody(t *testing.T) {
	s, _, _ := newTestService(t)
	h := NewHandler(s)
	app := fiber.New()
	app.Post("/api/auth/sign-in", h.SignIn)

	req := httptest.NewRequest("POST", "/api/auth/sign-in", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, res.StatusCode)
}

func TestHandler_SignIn_InvalidCredentials(t *testing.T) {
	s, _, _ := newTestService(t)
	h := NewHandler(s)
	app := fiber.New()
	app.Post("/api/auth/sign-in", h.SignIn)

	_, _, err := s.Register(ctxBackground(), "wrong@example.com", "password123")
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"email":"wrong@example.com","password":"wrong"}`)
	req := httptest.NewRequest("POST", "/api/auth/sign-in", body)
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, res.StatusCode)
}

func ctxBackground() context.Context {
	return context.Background()
}
