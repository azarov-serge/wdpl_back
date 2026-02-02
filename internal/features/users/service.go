package users

import (
	"context"
	"time"
)

// UpdateProfileInput — входные данные для обновления профиля (слой сервиса).
type UpdateProfileInput struct {
	DisplayName *string
	AvatarURL   *string
	Bio         *string
	Locale      *string
	Timezone    *string
}

// Service инкапсулирует бизнес-логику работы с профилем пользователя.
type Service struct {
	repo ProfileRepository
}

// NewService создаёт сервис профилей.
func NewService(repo ProfileRepository) *Service {
	return &Service{repo: repo}
}

// GetOrCreate возвращает профиль пользователя.
// Если профиль отсутствует — создаёт с дефолтными значениями (displayName = email, locale = "en", timezone = "UTC").
func (s *Service) GetOrCreate(ctx context.Context, userID, email string) (*UserProfile, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		return profile, nil
	}

	now := time.Now()
	profile = &UserProfile{
		UserID:      userID,
		DisplayName: email,
		Locale:      "en",
		Timezone:    "UTC",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// Update обновляет только переданные поля профиля.
// Если профиля нет — создаёт через GetOrCreate и затем обновляет.
func (s *Service) Update(ctx context.Context, userID string, input UpdateProfileInput) (*UserProfile, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		profile, err = s.GetOrCreate(ctx, userID, userID)
		if err != nil {
			return nil, err
		}
	}

	if input.DisplayName != nil {
		profile.DisplayName = *input.DisplayName
	}
	if input.AvatarURL != nil {
		profile.AvatarURL = input.AvatarURL
	}
	if input.Bio != nil {
		profile.Bio = input.Bio
	}
	if input.Locale != nil {
		profile.Locale = *input.Locale
	}
	if input.Timezone != nil {
		profile.Timezone = *input.Timezone
	}
	profile.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
