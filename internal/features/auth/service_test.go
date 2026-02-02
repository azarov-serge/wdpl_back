package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wdpl_back/internal/shared/config"
)

// mockUserRepo — простая in-memory реализация UserRepository для тестов.
type mockUserRepo struct {
	usersByEmail map[string]*User
	createErr    error
}

func (m *mockUserRepo) CreateUser(_ context.Context, user *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.usersByEmail == nil {
		m.usersByEmail = make(map[string]*User)
	}
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetUserByEmail(_ context.Context, email string) (*User, error) {
	if m.usersByEmail == nil {
		return nil, nil
	}
	return m.usersByEmail[email], nil
}

func (m *mockUserRepo) GetUserByID(_ context.Context, userID string) (*User, error) {
	if m.usersByEmail == nil {
		return nil, nil
	}
	for _, u := range m.usersByEmail {
		if u.ID == userID {
			return u, nil
		}
	}
	return nil, nil
}

// mockRefreshRepo — in-memory реализация RefreshTokenRepository.
type mockRefreshRepo struct {
	tokensByValue map[string]*RefreshToken
	createErr     error
	revokeErr     error
}

func (m *mockRefreshRepo) CreateRefreshToken(_ context.Context, token *RefreshToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.tokensByValue == nil {
		m.tokensByValue = make(map[string]*RefreshToken)
	}
	m.tokensByValue[token.Token] = token
	return nil
}

func (m *mockRefreshRepo) GetRefreshToken(_ context.Context, token string) (*RefreshToken, error) {
	if m.tokensByValue == nil {
		return nil, nil
	}
	return m.tokensByValue[token], nil
}

func (m *mockRefreshRepo) RevokeRefreshToken(_ context.Context, tokenID string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	for _, rt := range m.tokensByValue {
		if rt.ID == tokenID {
			now := time.Now()
			rt.RevokedAt = &now
			return nil
		}
	}
	return nil
}

// newTestService создаёт Service с замоканными зависимостями.
func newTestService(t *testing.T) (*Service, *mockUserRepo, *mockRefreshRepo) {
	t.Helper()

	cfg := &config.Config{
		JWTSecret:          "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenTTLMin:  15,
		RefreshTokenTTLMin: 30,
	}

	userRepo := &mockUserRepo{}
	refreshRepo := &mockRefreshRepo{}
	s := NewService(userRepo, refreshRepo, cfg)
	return s, userRepo, refreshRepo
}

func TestRegister_Success(t *testing.T) {
	s, userRepo, _ := newTestService(t)

	ctx := context.Background()

	user, tokens, err := s.Register(ctx, "test@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.NotNil(t, tokens)

	// проверяем, что пользователь записан в репозиторий
	stored, err := userRepo.GetUserByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, stored.ID)
	assert.NotEmpty(t, stored.PasswordHash)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	s, userRepo, _ := newTestService(t)

	ctx := context.Background()

	// предварительно создаём пользователя
	_ = userRepo.CreateUser(ctx, &User{
		ID:           "existing-id",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         "user",
		IsActive:     true,
	})

	user, tokens, err := s.Register(ctx, "test@example.com", "password123")
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}

func TestLogin_Success(t *testing.T) {
	s, _, _ := newTestService(t)
	ctx := context.Background()

	// создаём пользователя через Register, чтобы пароль захешировался корректно.
	user, _, err := s.Register(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	loggedInUser, tokens, err := s.Login(ctx, "test@example.com", "password123", "agent", "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, loggedInUser)
	require.NotNil(t, tokens)

	assert.Equal(t, user.ID, loggedInUser.ID)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func TestLogin_InvalidPassword(t *testing.T) {
	s, _, _ := newTestService(t)
	ctx := context.Background()

	_, _, err := s.Register(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	user, tokens, err := s.Login(ctx, "test@example.com", "wrong", "agent", "127.0.0.1")
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}

func TestRefresh_Success(t *testing.T) {
	s, _, refreshRepo := newTestService(t)
	ctx := context.Background()

	// Сначала логинимся, чтобы появился refresh‑токен.
	_, tokens, err := s.Register(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	require.NotEmpty(t, tokens.RefreshToken)

	// Проверяем, что refresh‑токен есть в репозитории.
	storedRT, err := refreshRepo.GetRefreshToken(ctx, tokens.RefreshToken)
	require.NoError(t, err)
	require.NotNil(t, storedRT)

	newTokens, err := s.Refresh(ctx, tokens.RefreshToken, "agent", "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, newTokens)

	assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
}

func TestRefresh_InvalidToken(t *testing.T) {
	s, _, _ := newTestService(t)
	ctx := context.Background()

	tokens, err := s.Refresh(ctx, "non-existent", "agent", "127.0.0.1")
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, tokens)
}

func TestLogout_RevokesToken(t *testing.T) {
	s, _, refreshRepo := newTestService(t)
	ctx := context.Background()

	// создаём refresh‑токен напрямую
	rt := &RefreshToken{
		ID:        "rt-id",
		UserID:    "user-id",
		Token:     "token-value",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	err := refreshRepo.CreateRefreshToken(ctx, rt)
	require.NoError(t, err)

	err = s.Logout(ctx, rt.ID)
	require.NoError(t, err)

	stored, err := refreshRepo.GetRefreshToken(ctx, rt.Token)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.NotNil(t, stored.RevokedAt)
}

func TestRegister_CreateUserError_Propagated(t *testing.T) {
	s, userRepo, _ := newTestService(t)
	ctx := context.Background()

	userRepo.createErr = errors.New("db error")

	user, tokens, err := s.Register(ctx, "test@example.com", "password123")
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}
