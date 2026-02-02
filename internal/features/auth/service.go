package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"wdpl_back/internal/shared/authutils"
	"wdpl_back/internal/shared/config"
)

// Service инкапсулирует бизнес-логику авторизации.
type Service struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	cfg              *config.Config
}

// NewService создаёт сервис с зависимостями от интерфейсов (DIP).
// Одна реализация *postgresRepository реализует оба интерфейса — в роутере передают repo дважды.
func NewService(userRepo UserRepository, refreshTokenRepo RefreshTokenRepository, cfg *config.Config) *Service {
	return &Service{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrEmailExists        = errors.New("user with this email already exists")
)

type AuthTokens struct {
	AccessToken   string
	AccessExpiry  time.Time
	RefreshToken  string
	RefreshExpiry time.Time
}

func (s *Service) Register(ctx context.Context, email, password string) (*User, *AuthTokens, error) {
	existing, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, ErrEmailExists
	}

	hash, err := authutils.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	user := &User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		Role:         "user",
		IsActive:     true,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(ctx, user.ID, user.Role, "", "")
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *Service) Login(ctx context.Context, email, password, userAgent, ip string) (*User, *AuthTokens, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || !authutils.CheckPassword(password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, nil, ErrUserInactive
	}

	tokens, err := s.issueTokens(ctx, user.ID, user.Role, userAgent, ip)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshTokenStr string, userAgent, ip string) (*AuthTokens, error) {
	rt, err := s.refreshTokenRepo.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}
	if rt == nil || rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		// Для простоты считаем все такие случаи 401 на уровне handler.
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, rt.UserID, "", userAgent, ip)
	if err != nil {
		return nil, err
	}

	// Опционально можно отозвать старый refresh и создать новый.
	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, refreshTokenID string) error {
	return s.refreshTokenRepo.RevokeRefreshToken(ctx, refreshTokenID)
}

// RevokeByToken отзывает refresh‑токен по его строковому значению (для SignOut).
// Если токен не найден — возвращает nil (идемпотентность).
func (s *Service) RevokeByToken(ctx context.Context, tokenStr string) error {
	rt, err := s.refreshTokenRepo.GetRefreshToken(ctx, tokenStr)
	if err != nil || rt == nil {
		return nil
	}
	return s.refreshTokenRepo.RevokeRefreshToken(ctx, rt.ID)
}

func (s *Service) issueTokens(ctx context.Context, userID, role, userAgent, ip string) (*AuthTokens, error) {
	accessToken, accessExp, err := authutils.GenerateAccessToken(s.cfg, userID, role)
	if err != nil {
		return nil, err
	}

	refreshExp := time.Now().Add(time.Duration(s.cfg.RefreshTokenTTLMin) * time.Minute)
	refreshToken := uuid.NewString()

	rt := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: refreshExp,
	}
	if userAgent != "" {
		rt.UserAgent = &userAgent
	}
	if ip != "" {
		rt.IP = &ip
	}

	if err := s.refreshTokenRepo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:   accessToken,
		AccessExpiry:  accessExp,
		RefreshToken:  refreshToken,
		RefreshExpiry: refreshExp,
	}, nil
}
