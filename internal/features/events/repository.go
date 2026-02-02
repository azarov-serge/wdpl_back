package events

import (
	"context"
	"time"
)

// EventRepository описывает операции с публикованными событиями.
type EventRepository interface {
	GetByID(ctx context.Context, id string) (*Event, error)
	List(ctx context.Context, limit, offset int) ([]*Event, error)
	Upsert(ctx context.Context, event *Event) error
}

// EventDayRepository описывает операции с опубликованными днями событий.
type EventDayRepository interface {
	ListByEventID(ctx context.Context, eventID string) ([]*EventDay, error)
	Upsert(ctx context.Context, day *EventDay) error
}

// EventDraftRepository описывает операции с черновиками событий.
type EventDraftRepository interface {
	GetByID(ctx context.Context, id string) (*EventDraft, error)
	GetByEventID(ctx context.Context, eventID string) (*EventDraft, error)
	ListDrafts(ctx context.Context) ([]*EventDraft, error)
	Upsert(ctx context.Context, draft *EventDraft) error
	DeleteByID(ctx context.Context, id string) error
}

// EventDayDraftRepository описывает операции с черновиками дней событий.
type EventDayDraftRepository interface {
	GetByID(ctx context.Context, id string) (*EventDayDraft, error)
	GetByEventIDAndDate(ctx context.Context, eventID string, date time.Time) (*EventDayDraft, error)
	ListByEventID(ctx context.Context, eventID string) ([]*EventDayDraft, error)
	Upsert(ctx context.Context, draft *EventDayDraft) error
	DeleteByID(ctx context.Context, id string) error
}
