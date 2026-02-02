package events

import (
	"encoding/json"
	"time"
)

// EventResponse — опубликованное событие в ответе API (STEP4).
type EventResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EventDayResponse — опубликованный день события в ответе API.
type EventDayResponse struct {
	ID                string          `json:"id"`
	EventID           string          `json:"eventId"`
	Date              time.Time       `json:"date"`
	Schedule          json.RawMessage `json:"schedule"`
	SessionCount      *int            `json:"sessionCount,omitempty"`
	FirstSessionStart *time.Time      `json:"firstSessionStart,omitempty"`
	LastSessionEnd    *time.Time      `json:"lastSessionEnd,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// EventWithDaysResponse — событие с опубликованными днями (GET /events/:id).
type EventWithDaysResponse struct {
	Event EventResponse      `json:"event"`
	Days  []EventDayResponse `json:"days"`
}

// DraftResponse — черновик события в ответе API.
type DraftResponse struct {
	ID          string     `json:"id"`
	EventID     *string    `json:"eventId,omitempty"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	StartDate   time.Time  `json:"startDate"`
	EndDate     time.Time  `json:"endDate"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedBy   string     `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// SaveDraftRequest — тело запроса на сохранение черновика (created_by задаёт бэкенд из JWT).
type SaveDraftRequest struct {
	ID          string  `json:"id"`
	EventID     *string `json:"eventId,omitempty"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	StartDate   string  `json:"startDate"`
	EndDate     string  `json:"endDate"`
}

// DayDraftResponse — черновик дня события в ответе API.
type DayDraftResponse struct {
	ID          string          `json:"id"`
	EventID     string          `json:"eventId"`
	Date        time.Time       `json:"date"`
	Schedule    json.RawMessage `json:"schedule"`
	PublishedAt *time.Time      `json:"publishedAt,omitempty"`
	CreatedBy   string          `json:"createdBy"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// SaveDayDraftRequest — тело запроса на сохранение черновика дня.
type SaveDayDraftRequest struct {
	EventID  string          `json:"eventId"`
	Date     string          `json:"date"`
	Schedule json.RawMessage `json:"schedule"`
}
