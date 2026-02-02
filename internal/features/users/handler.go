package users

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/features/auth"
	"wdpl_back/internal/shared/authutils"
	"wdpl_back/internal/shared/http/handler"
	"wdpl_back/internal/shared/http/middleware"
	"wdpl_back/internal/shared/http/response"
)

// Handler реализует HTTP-эндпоинты для профиля пользователя.
type Handler struct {
	service  *Service
	userRepo auth.UserRepository
	validate *validator.Validate
}

// NewHandler создаёт handler профилей.
func NewHandler(service *Service, userRepo auth.UserRepository) *Handler {
	return &Handler{
		service:  service,
		userRepo: userRepo,
		validate: validator.New(),
	}
}

// getClaims извлекает клеймы из контекста (после RequireAuth). При отсутствии возвращает nil, false.
func (h *Handler) getClaims(c *fiber.Ctx) (*authutils.UserClaims, bool) {
	claimsVal := c.Locals(middleware.LocalsKeyClaims)
	if claimsVal == nil {
		return nil, false
	}
	claims, ok := claimsVal.(*authutils.UserClaims)
	return claims, ok
}

// profileToResponse собирает ответ из профиля и email (email из auth.users).
func profileToResponse(profile *UserProfile, email string) UserProfileResponse {
	return UserProfileResponse{
		UserID:      profile.UserID,
		Email:       email,
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Bio:         profile.Bio,
		Locale:      profile.Locale,
		Timezone:    profile.Timezone,
	}
}

// GetMe — GET /api/users/me. Профиль текущего пользователя по JWT.
func (h *Handler) GetMe(c *fiber.Ctx) error {
	claims, ok := h.getClaims(c)
	if !ok || claims == nil {
		return response.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	user, err := h.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	if user == nil {
		return response.WriteError(c, fiber.StatusNotFound, "user not found")
	}

	profile, err := h.service.GetOrCreate(ctx, userID, user.Email)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	return c.JSON(profileToResponse(profile, user.Email))
}

// PutMe — PUT /api/users/me. Обновление профиля текущего пользователя.
func (h *Handler) PutMe(c *fiber.Ctx) error {
	claims, ok := h.getClaims(c)
	if !ok || claims == nil {
		return response.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "validation failed")
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	input := UpdateProfileInput{
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		Bio:         req.Bio,
		Locale:      req.Locale,
		Timezone:    req.Timezone,
	}
	profile, err := h.service.Update(ctx, userID, input)
	if err != nil {
		return response.WriteInternalError(c, err)
	}

	user, err := h.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	email := ""
	if user != nil {
		email = user.Email
	}
	return c.JSON(profileToResponse(profile, email))
}
