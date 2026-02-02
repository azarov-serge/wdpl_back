package events

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"wdpl_back/internal/shared/postgres"
)

// PostgresRepositories объединяет Postgres-реализации репозиториев событий и черновиков.
type PostgresRepositories struct {
	Events    EventRepository
	Days      EventDayRepository
	Drafts    EventDraftRepository
	DayDrafts EventDayDraftRepository
}

// NewPostgresRepository возвращает реализации репозиториев событий и черновиков.
func NewPostgresRepository(db *postgres.DB) *PostgresRepositories {
	return &PostgresRepositories{
		Events:    &eventRepoImpl{db: db},
		Days:      &eventDayRepoImpl{db: db},
		Drafts:    &eventDraftRepoImpl{db: db},
		DayDrafts: &eventDayDraftRepoImpl{db: db},
	}
}

// — EventRepository

type eventRepoImpl struct{ db *postgres.DB }

func (r *eventRepoImpl) GetByID(ctx context.Context, id string) (*Event, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, start_date, end_date, status, created_at, updated_at
		FROM public.events WHERE id = $1
	`, id)
	var e Event
	err := row.Scan(&e.ID, &e.Title, &e.Description, &e.StartDate, &e.EndDate, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepoImpl) List(ctx context.Context, limit, offset int) ([]*Event, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, start_date, end_date, status, created_at, updated_at
		FROM public.events ORDER BY start_date DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartDate, &e.EndDate, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &e)
	}
	return list, rows.Err()
}

func (r *eventRepoImpl) Upsert(ctx context.Context, event *Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO public.events (id, title, description, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, event.ID, event.Title, event.Description, event.StartDate, event.EndDate, event.Status, event.CreatedAt, event.UpdatedAt)
	return err
}

// — EventDayRepository (опубликованные дни)

type eventDayRepoImpl struct{ db *postgres.DB }

func (r *eventDayRepoImpl) ListByEventID(ctx context.Context, eventID string) ([]*EventDay, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id, date, schedule, session_count, first_session_start, last_session_end, created_at, updated_at
		FROM public.event_days WHERE event_id = $1 ORDER BY date
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*EventDay
	for rows.Next() {
		var d EventDay
		if err := rows.Scan(&d.ID, &d.EventID, &d.Date, &d.Schedule, &d.SessionCount, &d.FirstSessionStart, &d.LastSessionEnd, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *eventDayRepoImpl) Upsert(ctx context.Context, day *EventDay) error {
	if day.ID == "" {
		day.ID = uuid.NewString()
	}
	dateOnly := day.Date.Truncate(24 * time.Hour)
	schedule := day.Schedule
	if schedule == nil {
		schedule = []byte(`{"sessions":[]}`)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO public.event_days (id, event_id, date, schedule, session_count, first_session_start, last_session_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (event_id, date) DO UPDATE SET
			schedule = EXCLUDED.schedule,
			session_count = EXCLUDED.session_count,
			first_session_start = EXCLUDED.first_session_start,
			last_session_end = EXCLUDED.last_session_end,
			updated_at = EXCLUDED.updated_at
	`, day.ID, day.EventID, dateOnly, schedule, day.SessionCount, day.FirstSessionStart, day.LastSessionEnd, day.CreatedAt, day.UpdatedAt)
	return err
}

// — EventDraftRepository

type eventDraftRepoImpl struct{ db *postgres.DB }

func (r *eventDraftRepoImpl) GetByID(ctx context.Context, id string) (*EventDraft, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, event_id, title, description, start_date, end_date, published_at, created_by, created_at, updated_at
		FROM public.event_drafts WHERE id = $1
	`, id)
	var d EventDraft
	err := row.Scan(&d.ID, &d.EventID, &d.Title, &d.Description, &d.StartDate, &d.EndDate, &d.PublishedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *eventDraftRepoImpl) GetByEventID(ctx context.Context, eventID string) (*EventDraft, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, event_id, title, description, start_date, end_date, published_at, created_by, created_at, updated_at
		FROM public.event_drafts WHERE event_id = $1
	`, eventID)
	var d EventDraft
	err := row.Scan(&d.ID, &d.EventID, &d.Title, &d.Description, &d.StartDate, &d.EndDate, &d.PublishedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *eventDraftRepoImpl) ListDrafts(ctx context.Context) ([]*EventDraft, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id, title, description, start_date, end_date, published_at, created_by, created_at, updated_at
		FROM public.event_drafts ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*EventDraft
	for rows.Next() {
		var d EventDraft
		if err := rows.Scan(&d.ID, &d.EventID, &d.Title, &d.Description, &d.StartDate, &d.EndDate, &d.PublishedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *eventDraftRepoImpl) Upsert(ctx context.Context, draft *EventDraft) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO public.event_drafts (id, event_id, title, description, start_date, end_date, published_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			event_id = EXCLUDED.event_id,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			published_at = EXCLUDED.published_at,
			updated_at = EXCLUDED.updated_at
	`, draft.ID, draft.EventID, draft.Title, draft.Description, draft.StartDate, draft.EndDate, draft.PublishedAt, draft.CreatedBy, draft.CreatedAt, draft.UpdatedAt)
	return err
}

func (r *eventDraftRepoImpl) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM public.event_drafts WHERE id = $1`, id)
	return err
}

// — EventDayDraftRepository

type eventDayDraftRepoImpl struct{ db *postgres.DB }

func (r *eventDayDraftRepoImpl) GetByID(ctx context.Context, id string) (*EventDayDraft, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, event_id, date, schedule, published_at, created_by, created_at, updated_at
		FROM public.event_days_drafts WHERE id = $1
	`, id)
	var d EventDayDraft
	err := row.Scan(&d.ID, &d.EventID, &d.Date, &d.Schedule, &d.PublishedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *eventDayDraftRepoImpl) GetByEventIDAndDate(ctx context.Context, eventID string, date time.Time) (*EventDayDraft, error) {
	dateOnly := date.Truncate(24 * time.Hour)
	row := r.db.QueryRowContext(ctx, `
		SELECT id, event_id, date, schedule, published_at, created_by, created_at, updated_at
		FROM public.event_days_drafts WHERE event_id = $1 AND date = $2
	`, eventID, dateOnly)
	var d EventDayDraft
	err := row.Scan(&d.ID, &d.EventID, &d.Date, &d.Schedule, &d.PublishedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *eventDayDraftRepoImpl) ListByEventID(ctx context.Context, eventID string) ([]*EventDayDraft, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id, date, schedule, published_at, created_by, created_at, updated_at
		FROM public.event_days_drafts WHERE event_id = $1 ORDER BY date
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*EventDayDraft
	for rows.Next() {
		var d EventDayDraft
		if err := rows.Scan(&d.ID, &d.EventID, &d.Date, &d.Schedule, &d.PublishedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *eventDayDraftRepoImpl) Upsert(ctx context.Context, draft *EventDayDraft) error {
	if draft.ID == "" {
		draft.ID = uuid.NewString()
	}
	dateOnly := draft.Date.Truncate(24 * time.Hour)
	schedule := draft.Schedule
	if schedule == nil {
		schedule = []byte(`{"sessions":[]}`)
	}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO public.event_days_drafts (id, event_id, date, schedule, published_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (event_id, date) DO UPDATE SET
			schedule = EXCLUDED.schedule,
			published_at = EXCLUDED.published_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`, draft.ID, draft.EventID, dateOnly, schedule, draft.PublishedAt, draft.CreatedBy, draft.CreatedAt, draft.UpdatedAt).Scan(&draft.ID)
	return err
}

func (r *eventDayDraftRepoImpl) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM public.event_days_drafts WHERE id = $1`, id)
	return err
}
