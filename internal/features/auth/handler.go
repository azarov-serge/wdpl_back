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

// refreshCookieName — имя HTTP‑only cookie для refresh‑токена.
// Важно: фронт не имеет доступа к значению, только браузер отправляет его на бекенд.
const refreshCookieName = "refreshToken"

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

	setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshExpiry)

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

	setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshExpiry)

	return c.JSON(AuthResponse{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	// Пытаемся взять refresh‑токен из HTTP‑only cookie.
	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken == "" {
		// Для обратной совместимости поддерживаем вариант с токеном в теле.
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
		refreshToken = req.RefreshToken
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	userAgent := c.Get("User-Agent")
	ip := c.IP()

	tokens, err := h.service.Refresh(ctx, refreshToken, userAgent, ip)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.WriteError(c, fiber.StatusUnauthorized, "invalid refresh token")
		}
		return response.WriteInternalError(c, err)
	}

	setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshExpiry)

	// Для простого сценария checkAuth фронт может вызывать /refresh и по 200 понимать,
	// что пользователь авторизован.
	return c.JSON(fiber.Map{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

func (h *Handler) SignOut(c *fiber.Ctx) error {
	// Сначала пробуем взять токен из cookie, чтобы фронт не держал его в JS.
	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken == "" {
		// Для обратной совместимости: можно прислать токен в теле.
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
		refreshToken = req.RefreshToken
	}

	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	if err := h.service.RevokeByToken(ctx, refreshToken); err != nil {
		return response.WriteInternalError(c, err)
	}

	clearRefreshCookie(c)

	return c.SendStatus(fiber.StatusNoContent)
}

// setRefreshCookie выставляет HTTP‑only cookie с refresh‑токеном.
// Secure=true подразумевает использование HTTPS в продакшене.
func setRefreshCookie(c *fiber.Ctx, token string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   true,
	})
}

// clearRefreshCookie очищает refresh‑cookie на клиенте.
func clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
	})
}
