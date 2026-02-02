package auth

import "time"

// User — доменная модель пользователя для фичи авторизации.
type User struct {
	ID             string
	Email          string
	PasswordHash   string
	Role           string
	IsActive       bool
	SupabaseUserID *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// RefreshToken — доменная модель refresh‑токена.
type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	RevokedAt *time.Time
	UserAgent *string
	IP        *string
	CreatedAt time.Time
}
