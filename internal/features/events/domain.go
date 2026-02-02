package events

import (
	"encoding/json"
	"time"
)

// Event — публикованное событие (STEP4: домен публичных данных).
type Event struct {
	ID          string
	Title       string
	Description *string
	StartDate   time.Time
	EndDate     time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EventDay — опубликованный день события (расписание по дню в JSONB).
type EventDay struct {
	ID                string
	EventID           string
	Date              time.Time
	Schedule          json.RawMessage
	SessionCount      *int
	FirstSessionStart *time.Time
	LastSessionEnd    *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// EventDraft — черновик события (published_at, created_by). Доступ только у организаторов/админов/редакторов.
type EventDraft struct {
	ID          string
	EventID     *string
	Title       string
	Description *string
	StartDate   time.Time
	EndDate     time.Time
	PublishedAt *time.Time
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EventDayDraft — черновик дня события (расписание по дню в JSONB).
// Допускаются незавершённые данные; доступ только у организаторов/админов/редакторов.
type EventDayDraft struct {
	ID          string
	EventID     string
	Date        time.Time
	Schedule    json.RawMessage
	PublishedAt *time.Time
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
