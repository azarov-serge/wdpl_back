package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProfileRepo — in-memory реализация ProfileRepository для TDD.
type mockProfileRepo struct {
	byUserID map[string]*UserProfile
}

func (m *mockProfileRepo) GetByUserID(_ context.Context, userID string) (*UserProfile, error) {
	if m.byUserID == nil {
		return nil, nil
	}
	return m.byUserID[userID], nil
}

func (m *mockProfileRepo) Create(_ context.Context, profile *UserProfile) error {
	if m.byUserID == nil {
		m.byUserID = make(map[string]*UserProfile)
	}
	m.byUserID[profile.UserID] = profile
	return nil
}

func (m *mockProfileRepo) Update(_ context.Context, profile *UserProfile) error {
	if m.byUserID == nil {
		m.byUserID = make(map[string]*UserProfile)
	}
	m.byUserID[profile.UserID] = profile
	return nil
}

func TestGetOrCreate_CreatesDefaultIfNotExists(t *testing.T) {
	repo := &mockProfileRepo{}
	svc := NewService(repo)
	ctx := context.Background()

	profile, err := svc.GetOrCreate(ctx, "user-1", "user@example.com")
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, "user-1", profile.UserID)
	assert.Equal(t, "user@example.com", profile.DisplayName)
	assert.Equal(t, "en", profile.Locale)
	assert.Equal(t, "UTC", profile.Timezone)
}

func TestGetOrCreate_ReturnsExisting(t *testing.T) {
	repo := &mockProfileRepo{
		byUserID: map[string]*UserProfile{
			"user-1": {
				UserID:      "user-1",
				DisplayName: "Existing",
				Locale:      "ru",
				Timezone:    "Europe/Moscow",
			},
		},
	}
	svc := NewService(repo)
	ctx := context.Background()

	profile, err := svc.GetOrCreate(ctx, "user-1", "ignored@example.com")
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, "Existing", profile.DisplayName)
	assert.Equal(t, "ru", profile.Locale)
	assert.Equal(t, "Europe/Moscow", profile.Timezone)
}

func TestUpdate_UpdatesOnlyProvidedFields(t *testing.T) {
	repo := &mockProfileRepo{
		byUserID: map[string]*UserProfile{
			"user-1": {
				UserID:      "user-1",
				DisplayName: "Old Name",
				Locale:      "en",
				Timezone:    "UTC",
			},
		},
	}
	svc := NewService(repo)
	ctx := context.Background()

	newName := "New Name"
	newLocale := "ru"

	profile, err := svc.Update(ctx, "user-1", UpdateProfileInput{
		DisplayName: &newName,
		Locale:      &newLocale,
		// avatarURL, bio, timezone оставляем nil — они не должны измениться.
	})
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, "New Name", profile.DisplayName)
	assert.Equal(t, "ru", profile.Locale)
	assert.Equal(t, "UTC", profile.Timezone)
}
