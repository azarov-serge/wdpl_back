package events

import (
	"context"
	"errors"
	"time"
)

// Service инкапсулирует бизнес-логику работы с событиями и черновиками.
type Service struct {
	eventsRepo    EventRepository
	daysRepo      EventDayRepository
	draftsRepo    EventDraftRepository
	dayDraftsRepo EventDayDraftRepository
}

func NewService(eventsRepo EventRepository, daysRepo EventDayRepository, draftsRepo EventDraftRepository, dayDraftsRepo EventDayDraftRepository) *Service {
	return &Service{
		eventsRepo:    eventsRepo,
		daysRepo:      daysRepo,
		draftsRepo:    draftsRepo,
		dayDraftsRepo: dayDraftsRepo,
	}
}

// SaveDraft сохраняет черновик события (создание или обновление). created_by обязателен.
func (s *Service) SaveDraft(ctx context.Context, draft *EventDraft) error {
	now := time.Now()
	if draft.ID == "" {
		return ErrInvalidDraftID
	}
	if draft.CreatedBy == "" {
		return ErrInvalidCreatedBy
	}
	if draft.CreatedAt.IsZero() {
		draft.CreatedAt = now
	}
	draft.UpdatedAt = now
	return s.draftsRepo.Upsert(ctx, draft)
}

// PublishDraft публикует черновик (STEP4): event_draft → events, event_day_drafts → event_days, затем удаляет черновики.
func (s *Service) PublishDraft(ctx context.Context, draftID string) (*Event, error) {
	draft, err := s.draftsRepo.GetByID(ctx, draftID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, ErrDraftNotFound
	}

	now := time.Now()
	eventID := draft.ID
	if draft.EventID != nil {
		eventID = *draft.EventID
	}

	var event *Event
	if draft.EventID != nil {
		event, err = s.eventsRepo.GetByID(ctx, *draft.EventID)
		if err != nil {
			return nil, err
		}
	}
	if event == nil {
		event = &Event{
			ID:        eventID,
			CreatedAt: now,
			Status:    "published",
		}
	}

	event.Title = draft.Title
	event.Description = draft.Description
	event.StartDate = draft.StartDate
	event.EndDate = draft.EndDate
	event.UpdatedAt = now

	if err := s.eventsRepo.Upsert(ctx, event); err != nil {
		return nil, err
	}

	// Публикуем черновики дней в event_days.
	dayDrafts, err := s.dayDraftsRepo.ListByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	for _, dd := range dayDrafts {
		day := &EventDay{
			EventID:   eventID,
			Date:      dd.Date,
			Schedule:  dd.Schedule,
			CreatedAt: dd.CreatedAt,
			UpdatedAt: now,
		}
		if day.Schedule == nil {
			day.Schedule = []byte(`{"sessions":[]}`)
		}
		if err := s.daysRepo.Upsert(ctx, day); err != nil {
			return nil, err
		}
		if err := s.dayDraftsRepo.DeleteByID(ctx, dd.ID); err != nil {
			return nil, err
		}
	}

	if err := s.draftsRepo.DeleteByID(ctx, draft.ID); err != nil {
		return nil, err
	}

	return event, nil
}

// ListPublishedEvents возвращает список опубликованных событий (пагинация).
func (s *Service) ListPublishedEvents(ctx context.Context, limit, offset int) ([]*Event, error) {
	return s.eventsRepo.List(ctx, limit, offset)
}

// GetEventWithDays возвращает событие и его опубликованные дни (для публичного API).
func (s *Service) GetEventWithDays(ctx context.Context, eventID string) (*Event, []*EventDay, error) {
	event, err := s.eventsRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	if event == nil {
		return nil, nil, nil
	}
	days, err := s.daysRepo.ListByEventID(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	return event, days, nil
}

// GetDraft возвращает черновик события по ID.
func (s *Service) GetDraft(ctx context.Context, id string) (*EventDraft, error) {
	return s.draftsRepo.GetByID(ctx, id)
}

// ListDrafts возвращает список черновиков (для админки).
func (s *Service) ListDrafts(ctx context.Context) ([]*EventDraft, error) {
	return s.draftsRepo.ListDrafts(ctx)
}

// SaveDayDraft сохраняет черновик дня события (создание или обновление по event_id + date).
func (s *Service) SaveDayDraft(ctx context.Context, draft *EventDayDraft) error {
	now := time.Now()
	if draft.EventID == "" {
		return ErrInvalidEventID
	}
	if draft.CreatedBy == "" {
		return ErrInvalidCreatedBy
	}
	if draft.CreatedAt.IsZero() {
		draft.CreatedAt = now
	}
	draft.UpdatedAt = now
	return s.dayDraftsRepo.Upsert(ctx, draft)
}

// GetDayDraft возвращает черновик дня по ID.
func (s *Service) GetDayDraft(ctx context.Context, id string) (*EventDayDraft, error) {
	return s.dayDraftsRepo.GetByID(ctx, id)
}

// ListDayDraftsByEventID возвращает черновики дней события для админки.
func (s *Service) ListDayDraftsByEventID(ctx context.Context, eventID string) ([]*EventDayDraft, error) {
	return s.dayDraftsRepo.ListByEventID(ctx, eventID)
}

var (
	ErrInvalidDraftID   = errors.New("invalid draft id")
	ErrDraftNotFound    = errors.New("draft not found")
	ErrInvalidEventID   = errors.New("invalid event id")
	ErrInvalidCreatedBy = errors.New("invalid created_by")
)
