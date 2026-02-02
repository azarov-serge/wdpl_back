package users

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"wdpl_back/internal/features/auth"
	"wdpl_back/internal/shared/authutils"
	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/http/middleware"
)

// mockAuthUserRepo — минимальный мок auth.UserRepository для теста GetMe (нужен только GetUserByID).
type mockAuthUserRepo struct {
	user *auth.User
}

func (m *mockAuthUserRepo) CreateUser(_ context.Context, _ *auth.User) error { return nil }
func (m *mockAuthUserRepo) GetUserByEmail(_ context.Context, _ string) (*auth.User, error) {
	return nil, nil
}
func (m *mockAuthUserRepo) GetUserByID(_ context.Context, userID string) (*auth.User, error) {
	if m.user != nil && m.user.ID == userID {
		return m.user, nil
	}
	return nil, nil
}

func testUsersHandlerConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		JWTSecret:         "test-jwt-secret-at-least-32-bytes-for-users-me",
		AccessTokenTTLMin: 15,
	}
}

// Интеграционный тест: GET /api/users/me с валидным JWT и мок-репозиториями (без БД).
func TestHandler_GetMe_Success(t *testing.T) {
	cfg := testUsersHandlerConfig(t)
	profileRepo := &mockProfileRepo{}
	userRepo := &mockAuthUserRepo{
		user: &auth.User{
			ID:        "user-me-1",
			Email:     "me@example.com",
			Role:      "user",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	svc := NewService(profileRepo)
	h := NewHandler(svc, userRepo)

	app := fiber.New()
	app.Get("/api/users/me", middleware.RequireAuth(cfg), h.GetMe)

	token, _, err := authutils.GenerateAccessToken(cfg, "user-me-1", "user")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)

	var body UserProfileResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	require.Equal(t, "user-me-1", body.UserID)
	require.Equal(t, "me@example.com", body.Email)
	// При первом запросе профиль создаётся с displayName = email.
	require.Equal(t, "me@example.com", body.DisplayName)
}

func TestHandler_GetMe_Unauthorized(t *testing.T) {
	cfg := testUsersHandlerConfig(t)
	profileRepo := &mockProfileRepo{}
	userRepo := &mockAuthUserRepo{}

	svc := NewService(profileRepo)
	h := NewHandler(svc, userRepo)

	app := fiber.New()
	app.Get("/api/users/me", middleware.RequireAuth(cfg), h.GetMe)

	req := httptest.NewRequest("GET", "/api/users/me", nil)
	// Без заголовка Authorization.

	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, res.StatusCode)
}
