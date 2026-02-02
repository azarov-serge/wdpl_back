package authutils

import "github.com/golang-jwt/jwt/v5"

// UserClaims описывает содержимое JWT‑токена.
// Важно: сюда не кладём чувствительные данные (пароли и т.п.).
type UserClaims struct {
	UserID string `json:"userID"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
