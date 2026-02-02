package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"wdpl_back/internal/shared/postgres"
)

type postgresRepository struct {
	db *postgres.DB
}

func NewPostgresRepository(db *postgres.DB) *postgresRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateUser(ctx context.Context, user *User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth.users (
			id, email, password_hash, role, is_active, supabase_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.PasswordHash, user.Role, user.IsActive, user.SupabaseUserID, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, role, is_active, supabase_user_id, created_at, updated_at
		FROM auth.users
		WHERE email = $1
	`, email)

	var u User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.IsActive,
		&u.SupabaseUserID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, role, is_active, supabase_user_id, created_at, updated_at
		FROM auth.users
		WHERE id = $1
	`, userID)

	var u User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.IsActive,
		&u.SupabaseUserID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	if token.ID == "" {
		token.ID = uuid.NewString()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth.refresh_tokens (
			id, user_id, token, expires_at, revoked_at, user_agent, ip, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, token.ID, token.UserID, token.Token, token.ExpiresAt, token.RevokedAt, token.UserAgent, token.IP, token.CreatedAt)
	return err
}

func (r *postgresRepository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, revoked_at, user_agent, ip, created_at
		FROM auth.refresh_tokens
		WHERE token = $1
	`, token)

	var rt RefreshToken
	err := row.Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.RevokedAt,
		&rt.UserAgent,
		&rt.IP,
		&rt.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *postgresRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE auth.refresh_tokens
		SET revoked_at = $1
		WHERE id = $2 AND revoked_at IS NULL
	`, now, tokenID)
	return err
}
