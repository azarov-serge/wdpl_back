package users

import "context"

// ProfileRepository описывает операции с таблицей user_profiles.
type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*UserProfile, error)
	Create(ctx context.Context, profile *UserProfile) error
	Update(ctx context.Context, profile *UserProfile) error
}
