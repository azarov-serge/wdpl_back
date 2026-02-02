package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEventRepo — in-memory реализация EventRepository.
type mockEventRepo struct {
	events map[string]*Event
}

func (m *mockEventRepo) GetByID(_ context.Context, id string) (*Event, error) {
	if m.events == nil {
		return nil, nil
	}
	return m.events[id], nil
}

func (m *mockEventRepo) List(_ context.Context, limit, offset int) ([]*Event, error) {
	if m.events == nil {
		return nil, nil
	}
	list := make([]*Event, 0, len(m.events))
	for _, e := range m.events {
		list = append(list, e)
	}
	return list, nil
}

func (m *mockEventRepo) Upsert(_ context.Context, event *Event) error {
	if m.events == nil {
		m.events = make(map[string]*Event)
	}
	m.events[event.ID] = event
	return nil
}

// mockDaysRepo — in-memory реализация EventDayRepository для тестов.
type mockDaysRepo struct{}

func (m *mockDaysRepo) ListByEventID(_ context.Context, _ string) ([]*EventDay, error) {
	return nil, nil
}

func (m *mockDaysRepo) Upsert(_ context.Context, _ *EventDay) error {
	return nil
}

// mockDraftRepo — in-memory реализация EventDraftRepository.
type mockDraftRepo struct {
	drafts map[string]*EventDraft
}

func (m *mockDraftRepo) GetByID(_ context.Context, id string) (*EventDraft, error) {
	if m.drafts == nil {
		return nil, nil
	}
	return m.drafts[id], nil
}

func (m *mockDraftRepo) GetByEventID(_ context.Context, eventID string) (*EventDraft, error) {
	if m.drafts == nil {
		return nil, nil
	}
	for _, d := range m.drafts {
		if d.EventID != nil && *d.EventID == eventID {
			return d, nil
		}
	}
	return nil, nil
}

func (m *mockDraftRepo) ListDrafts(_ context.Context) ([]*EventDraft, error) {
	result := make([]*EventDraft, 0, len(m.drafts))
	for _, d := range m.drafts {
		result = append(result, d)
	}
	return result, nil
}

func (m *mockDraftRepo) Upsert(_ context.Context, draft *EventDraft) error {
	if m.drafts == nil {
		m.drafts = make(map[string]*EventDraft)
	}
	m.drafts[draft.ID] = draft
	return nil
}

func (m *mockDraftRepo) DeleteByID(_ context.Context, id string) error {
	if m.drafts == nil {
		return nil
	}
	delete(m.drafts, id)
	return nil
}

// mockDayDraftRepo — in-memory реализация EventDayDraftRepository для тестов.
type mockDayDraftRepo struct {
	dayDrafts map[string]*EventDayDraft
}

func (m *mockDayDraftRepo) GetByID(_ context.Context, id string) (*EventDayDraft, error) {
	if m.dayDrafts == nil {
		return nil, nil
	}
	return m.dayDrafts[id], nil
}

func (m *mockDayDraftRepo) GetByEventIDAndDate(_ context.Context, _ string, _ time.Time) (*EventDayDraft, error) {
	return nil, nil
}

func (m *mockDayDraftRepo) ListByEventID(_ context.Context, _ string) ([]*EventDayDraft, error) {
	return nil, nil
}

func (m *mockDayDraftRepo) Upsert(_ context.Context, draft *EventDayDraft) error {
	if m.dayDrafts == nil {
		m.dayDrafts = make(map[string]*EventDayDraft)
	}
	if draft.ID == "" {
		draft.ID = "day-draft-1"
	}
	m.dayDrafts[draft.ID] = draft
	return nil
}

func (m *mockDayDraftRepo) DeleteByID(_ context.Context, id string) error {
	if m.dayDrafts != nil {
		delete(m.dayDrafts, id)
	}
	return nil
}

func TestSaveDraft_InvalidID(t *testing.T) {
	eventsRepo := &mockEventRepo{}
	daysRepo := &mockDaysRepo{}
	draftsRepo := &mockDraftRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(eventsRepo, daysRepo, draftsRepo, dayDraftsRepo)

	err := svc.SaveDraft(context.Background(), &EventDraft{})
	require.ErrorIs(t, err, ErrInvalidDraftID)
}

func TestSaveDraft_InvalidCreatedBy(t *testing.T) {
	eventsRepo := &mockEventRepo{}
	daysRepo := &mockDaysRepo{}
	draftsRepo := &mockDraftRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(eventsRepo, daysRepo, draftsRepo, dayDraftsRepo)

	err := svc.SaveDraft(context.Background(), &EventDraft{
		ID: "draft-1", Title: "T", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour),
	})
	require.ErrorIs(t, err, ErrInvalidCreatedBy)
}

func TestSaveDraft_CreatesOrUpdates(t *testing.T) {
	eventsRepo := &mockEventRepo{}
	daysRepo := &mockDaysRepo{}
	draftsRepo := &mockDraftRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(eventsRepo, daysRepo, draftsRepo, dayDraftsRepo)

	ctx := context.Background()
	draft := &EventDraft{
		ID:        "draft-1",
		Title:     "Title",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
		CreatedBy: "user-1",
	}

	err := svc.SaveDraft(ctx, draft)
	require.NoError(t, err)

	stored, err := draftsRepo.GetByID(ctx, "draft-1")
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, "Title", stored.Title)
}

func TestPublishDraft_CreatesNewEventAndDeletesDraft(t *testing.T) {
	eventsRepo := &mockEventRepo{}
	draftsRepo := &mockDraftRepo{
		drafts: map[string]*EventDraft{
			"draft-1": {
				ID:        "draft-1",
				Title:     "Event Title",
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
			},
		},
	}
	daysRepo := &mockDaysRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(eventsRepo, daysRepo, draftsRepo, dayDraftsRepo)

	ctx := context.Background()

	event, err := svc.PublishDraft(ctx, "draft-1")
	require.NoError(t, err)
	require.NotNil(t, event)

	// событие создано
	assert.Equal(t, "Event Title", event.Title)

	// черновик удалён
	draft, err := draftsRepo.GetByID(ctx, "draft-1")
	require.NoError(t, err)
	assert.Nil(t, draft)
}

func TestPublishDraft_UpdatesExistingEvent(t *testing.T) {
	eventsRepo := &mockEventRepo{
		events: map[string]*Event{
			"event-1": {
				ID:    "event-1",
				Title: "Old Title",
			},
		},
	}
	eventID := "event-1"
	draftsRepo := &mockDraftRepo{
		drafts: map[string]*EventDraft{
			"draft-1": {
				ID:        "draft-1",
				EventID:   &eventID,
				Title:     "New Title",
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
			},
		},
	}
	daysRepo := &mockDaysRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(eventsRepo, daysRepo, draftsRepo, dayDraftsRepo)

	ctx := context.Background()

	event, err := svc.PublishDraft(ctx, "draft-1")
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.Equal(t, "event-1", event.ID)
	assert.Equal(t, "New Title", event.Title)
}

func TestPublishDraft_NotFound(t *testing.T) {
	eventsRepo := &mockEventRepo{}
	daysRepo := &mockDaysRepo{}
	draftsRepo := &mockDraftRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(eventsRepo, daysRepo, draftsRepo, dayDraftsRepo)

	_, err := svc.PublishDraft(context.Background(), "missing")
	require.ErrorIs(t, err, ErrDraftNotFound)
}

func TestListDrafts(t *testing.T) {
	draftsRepo := &mockDraftRepo{
		drafts: map[string]*EventDraft{
			"d1": {ID: "d1", Title: "Draft 1"},
			"d2": {ID: "d2", Title: "Draft 2"},
		},
	}
	daysRepo := &mockDaysRepo{}
	dayDraftsRepo := &mockDayDraftRepo{}
	svc := NewService(&mockEventRepo{}, daysRepo, draftsRepo, dayDraftsRepo)

	list, err := svc.ListDrafts(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
