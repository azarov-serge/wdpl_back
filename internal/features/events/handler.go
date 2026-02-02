package events

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"wdpl_back/internal/shared/authutils"
	"wdpl_back/internal/shared/http/handler"
	"wdpl_back/internal/shared/http/middleware"
	"wdpl_back/internal/shared/http/response"
)

// Handler реализует HTTP‑эндпоинты для событий (публичные) и черновиков (защищённые).
type Handler struct {
	service *Service
}

// NewHandler создаёт handler событий.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListEvents — GET /api/events (публичный, пагинация: limit, offset).
func (h *Handler) ListEvents(c *fiber.Ctx) error {
	limit, offset := handler.LimitOffset(c, 20, 100, 0)
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	list, err := h.service.ListPublishedEvents(ctx, limit, offset)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	resp := make([]EventResponse, 0, len(list))
	for _, e := range list {
		resp = append(resp, eventToResponse(e))
	}
	return c.JSON(resp)
}

// GetEvent — GET /api/events/:id (публичный, событие + дни).
func (h *Handler) GetEvent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "missing id")
	}
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	event, days, err := h.service.GetEventWithDays(ctx, id)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	if event == nil {
		return response.WriteError(c, fiber.StatusNotFound, "event not found")
	}
	daysResp := make([]EventDayResponse, 0, len(days))
	for _, d := range days {
		daysResp = append(daysResp, eventDayToResponse(d))
	}
	return c.JSON(EventWithDaysResponse{Event: eventToResponse(event), Days: daysResp})
}

// PublishDraft — POST /api/events/drafts/:id/publish (черновик → events + event_days).
func (h *Handler) PublishDraft(c *fiber.Ctx) error {
	draftID := c.Params("id")
	if draftID == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "missing id")
	}
	ctx, cancel := handler.TimeoutContext(c, 10*time.Second)
	defer cancel()

	event, err := h.service.PublishDraft(ctx, draftID)
	if err != nil {
		if errors.Is(err, ErrDraftNotFound) {
			return response.WriteError(c, fiber.StatusNotFound, "draft not found")
		}
		return response.WriteInternalError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(eventToResponse(event))
}

// ListDrafts — GET /api/events/drafts
func (h *Handler) ListDrafts(c *fiber.Ctx) error {
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	list, err := h.service.ListDrafts(ctx)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	resp := make([]DraftResponse, 0, len(list))
	for _, d := range list {
		resp = append(resp, draftToResponse(d))
	}
	return c.JSON(resp)
}

// GetDraft — GET /api/events/drafts/:id
func (h *Handler) GetDraft(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "missing id")
	}
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	draft, err := h.service.GetDraft(ctx, id)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	if draft == nil {
		return response.WriteError(c, fiber.StatusNotFound, "draft not found")
	}
	return c.JSON(draftToResponse(draft))
}

// SaveDraft — POST /api/events/drafts или PUT /api/events/drafts
// createdBy берётся из JWT (только организаторы/админы/редакторы проходят middleware).
func (h *Handler) SaveDraft(c *fiber.Ctx) error {
	claimsVal := c.Locals(middleware.LocalsKeyClaims)
	if claimsVal == nil {
		return response.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	claims, ok := claimsVal.(*authutils.UserClaims)
	if !ok {
		return response.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var req SaveDraftRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if req.ID == "" || req.Title == "" || req.StartDate == "" || req.EndDate == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "id, title, startDate, endDate required")
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid startDate")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid endDate")
	}

	draft := &EventDraft{
		ID:          req.ID,
		EventID:     req.EventID,
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedBy:   claims.UserID,
	}
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	if err := h.service.SaveDraft(ctx, draft); err != nil {
		if errors.Is(err, ErrInvalidDraftID) || errors.Is(err, ErrInvalidCreatedBy) {
			return response.WriteError(c, fiber.StatusBadRequest, err.Error())
		}
		return response.WriteInternalError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(draftToResponse(draft))
}

// ListDayDrafts — GET /api/events/:eventId/day-drafts
func (h *Handler) ListDayDrafts(c *fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "missing eventId")
	}
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	list, err := h.service.ListDayDraftsByEventID(ctx, eventID)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	resp := make([]DayDraftResponse, 0, len(list))
	for _, d := range list {
		resp = append(resp, dayDraftToResponse(d))
	}
	return c.JSON(resp)
}

// GetDayDraft — GET /api/events/day-drafts/:id
func (h *Handler) GetDayDraft(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "missing id")
	}
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	draft, err := h.service.GetDayDraft(ctx, id)
	if err != nil {
		return response.WriteInternalError(c, err)
	}
	if draft == nil {
		return response.WriteError(c, fiber.StatusNotFound, "day draft not found")
	}
	return c.JSON(dayDraftToResponse(draft))
}

// SaveDayDraft — POST /api/events/:eventId/day-drafts
func (h *Handler) SaveDayDraft(c *fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "missing eventId")
	}
	claimsVal := c.Locals(middleware.LocalsKeyClaims)
	if claimsVal == nil {
		return response.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	claims, ok := claimsVal.(*authutils.UserClaims)
	if !ok {
		return response.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	createdBy := claims.UserID

	var req SaveDayDraftRequest
	if err := c.BodyParser(&req); err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid body")
	}
	if req.EventID == "" {
		req.EventID = eventID
	}
	if req.EventID != eventID {
		return response.WriteError(c, fiber.StatusBadRequest, "eventId mismatch")
	}
	if req.Date == "" {
		return response.WriteError(c, fiber.StatusBadRequest, "date required")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return response.WriteError(c, fiber.StatusBadRequest, "invalid date")
	}
	schedule := req.Schedule
	if schedule == nil {
		schedule = []byte(`{"sessions":[]}`)
	}

	draft := &EventDayDraft{
		EventID:   req.EventID,
		Date:      date,
		Schedule:  schedule,
		CreatedBy: createdBy,
	}
	ctx, cancel := handler.TimeoutContext(c, 5*time.Second)
	defer cancel()

	if err := h.service.SaveDayDraft(ctx, draft); err != nil {
		if errors.Is(err, ErrInvalidEventID) || errors.Is(err, ErrInvalidCreatedBy) {
			return response.WriteError(c, fiber.StatusBadRequest, err.Error())
		}
		return response.WriteInternalError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(dayDraftToResponse(draft))
}

func draftToResponse(d *EventDraft) DraftResponse {
	return DraftResponse{
		ID:          d.ID,
		EventID:     d.EventID,
		Title:       d.Title,
		Description: d.Description,
		StartDate:   d.StartDate,
		EndDate:     d.EndDate,
		PublishedAt: d.PublishedAt,
		CreatedBy:   d.CreatedBy,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func eventToResponse(e *Event) EventResponse {
	return EventResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		StartDate:   e.StartDate,
		EndDate:     e.EndDate,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func eventDayToResponse(d *EventDay) EventDayResponse {
	return EventDayResponse{
		ID:                d.ID,
		EventID:           d.EventID,
		Date:              d.Date,
		Schedule:          d.Schedule,
		SessionCount:      d.SessionCount,
		FirstSessionStart: d.FirstSessionStart,
		LastSessionEnd:    d.LastSessionEnd,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func dayDraftToResponse(d *EventDayDraft) DayDraftResponse {
	return DayDraftResponse{
		ID:          d.ID,
		EventID:     d.EventID,
		Date:        d.Date,
		Schedule:    d.Schedule,
		PublishedAt: d.PublishedAt,
		CreatedBy:   d.CreatedBy,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
