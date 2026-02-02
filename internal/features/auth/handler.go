package auth

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/shared/http/handler"
	"wdpl_back/internal/shared/http/response"
)

// Handler реализует HTTP‑эндпоинты для авторизации.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) SignUp(c *fiber.Ctx) error {
	var req SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "validation failed")
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	user, tokens, err := h.service.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.WriteError(c, fiber.StatusUnauthorized, "invalid credentials")
		}
		if errors.Is(err, ErrUserInactive) {
			return response.WriteError(c, fiber.StatusForbidden, "user is inactive")
		}
		if errors.Is(err, ErrEmailExists) {
			return response.WriteError(c, fiber.StatusBadRequest, "user with this email already exists")
		}
		return response.WriteInternalError(c, err)
	}

	return c.JSON(AuthResponse{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *Handler) SignIn(c *fiber.Ctx) error {
	var req SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "validation failed")
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	userAgent := c.Get("User-Agent")
	ip := c.IP()

	user, tokens, err := h.service.Login(ctx, req.Email, req.Password, userAgent, ip)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.WriteError(c, fiber.StatusUnauthorized, "invalid credentials")
		}
		if errors.Is(err, ErrUserInactive) {
			return response.WriteError(c, fiber.StatusForbidden, "user is inactive")
		}
		return response.WriteInternalError(c, err)
	}

	return c.JSON(AuthResponse{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "validation failed")
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	userAgent := c.Get("User-Agent")
	ip := c.IP()

	tokens, err := h.service.Refresh(ctx, req.RefreshToken, userAgent, ip)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.WriteError(c, fiber.StatusUnauthorized, "invalid refresh token")
		}
		return response.WriteInternalError(c, err)
	}

	// Для простого сценария checkAuth фронт может вызывать /refresh и по 200 понимать,
	// что пользователь авторизован.
	return c.JSON(fiber.Map{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

func (h *Handler) SignOut(c *fiber.Ctx) error {
	type SignOutRequest struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}

	var req SignOutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "validation failed")
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	if err := h.service.RevokeByToken(ctx, req.RefreshToken); err != nil {
		return response.WriteInternalError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
