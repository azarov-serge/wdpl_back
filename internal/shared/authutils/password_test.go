package authutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	// bcrypt-хэш начинается с $2a$ или $2b$
	assert.Contains(t, hash, "$2")
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	hash1, err := HashPassword("same")
	require.NoError(t, err)
	hash2, err := HashPassword("same")
	require.NoError(t, err)
	// Разные соли — разные хэши.
	assert.NotEqual(t, hash1, hash2)
}

func TestCheckPassword_Valid(t *testing.T) {
	hash, err := HashPassword("mypassword")
	require.NoError(t, err)

	ok := CheckPassword("mypassword", hash)
	assert.True(t, ok)
}

func TestCheckPassword_Invalid(t *testing.T) {
	hash, err := HashPassword("mypassword")
	require.NoError(t, err)

	ok := CheckPassword("wrong", hash)
	assert.False(t, ok)
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("x")
	require.NoError(t, err)

	assert.False(t, CheckPassword("", hash))
}
