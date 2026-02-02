package authutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"wdpl_back/internal/shared/config"
)

// GenerateAccessToken создаёт access JWT для пользователя.
func GenerateAccessToken(cfg *config.Config, userID, role string) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(cfg.AccessTokenTTLMin) * time.Minute)

	claims := &UserClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, exp, nil
}

// ParseAccessToken парсит access‑токен и возвращает клеймы.
func ParseAccessToken(cfg *config.Config, tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
