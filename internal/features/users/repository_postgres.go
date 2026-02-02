package users

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wdpl_back/internal/shared/postgres"
)

type postgresProfileRepo struct {
	db *postgres.DB
}

// NewPostgresProfileRepository возвращает реализацию ProfileRepository для PostgreSQL.
func NewPostgresProfileRepository(db *postgres.DB) ProfileRepository {
	return &postgresProfileRepo{db: db}
}

func (r *postgresProfileRepo) GetByUserID(ctx context.Context, userID string) (*UserProfile, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_id, display_name, avatar_url, bio, locale, timezone, created_at, updated_at
		FROM public.user_profiles
		WHERE user_id = $1
	`, userID)

	var p UserProfile
	var avatarURL, bio sql.NullString
	err := row.Scan(
		&p.UserID,
		&p.DisplayName,
		&avatarURL,
		&bio,
		&p.Locale,
		&p.Timezone,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if avatarURL.Valid {
		p.AvatarURL = &avatarURL.String
	}
	if bio.Valid {
		p.Bio = &bio.String
	}
	return &p, nil
}

func (r *postgresProfileRepo) Create(ctx context.Context, profile *UserProfile) error {
	now := time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	if profile.UpdatedAt.IsZero() {
		profile.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO public.user_profiles (user_id, display_name, avatar_url, bio, locale, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		profile.UserID,
		profile.DisplayName,
		profile.AvatarURL,
		profile.Bio,
		profile.Locale,
		profile.Timezone,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	return err
}

func (r *postgresProfileRepo) Update(ctx context.Context, profile *UserProfile) error {
	profile.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		UPDATE public.user_profiles
		SET display_name = $1, avatar_url = $2, bio = $3, locale = $4, timezone = $5, updated_at = $6
		WHERE user_id = $7
	`,
		profile.DisplayName,
		profile.AvatarURL,
		profile.Bio,
		profile.Locale,
		profile.Timezone,
		profile.UpdatedAt,
		profile.UserID,
	)
	return err
}
