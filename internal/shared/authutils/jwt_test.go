package authutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wdpl_back/internal/shared/config"
)

func testJWTConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		JWTSecret:         "test-jwt-secret-at-least-32-bytes-long",
		AccessTokenTTLMin: 15,
	}
}

func TestGenerateAccessToken(t *testing.T) {
	cfg := testJWTConfig(t)

	token, exp, err := GenerateAccessToken(cfg, "user-123", "user")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	assert.True(t, exp.After(time.Now()))
}

func TestParseAccessToken_Valid(t *testing.T) {
	cfg := testJWTConfig(t)

	token, _, err := GenerateAccessToken(cfg, "user-456", "admin")
	require.NoError(t, err)

	claims, err := ParseAccessToken(cfg, token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user-456", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
}

func TestParseAccessToken_InvalidSignature(t *testing.T) {
	cfg := testJWTConfig(t)
	token, _, err := GenerateAccessToken(cfg, "user-1", "user")
	require.NoError(t, err)

	otherCfg := &config.Config{
		JWTSecret:         "other-secret-at-least-32-bytes-long",
		AccessTokenTTLMin: 15,
	}

	claims, err := ParseAccessToken(otherCfg, token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseAccessToken_InvalidToken(t *testing.T) {
	cfg := testJWTConfig(t)

	claims, err := ParseAccessToken(cfg, "not-a-valid-jwt")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseAccessToken_EmptyToken(t *testing.T) {
	cfg := testJWTConfig(t)

	claims, err := ParseAccessToken(cfg, "")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
