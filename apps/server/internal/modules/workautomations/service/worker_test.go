package workautomations

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	"github.com/google/uuid"
)

type workerStore struct {
	WorkerRepository
	claim      domain.Claimed
	claimed    bool
	authorized bool
	duplicate  bool
	events     []domain.Event
	advances   int
	completed  []string
	errors     []string
	released   string
	next       *time.Time
}

func (s *workerStore) Claim(context.Context) (domain.Claimed, bool, error) {
	if s.claimed {
		return domain.Claimed{}, false, nil
	}
	s.claimed = true
	return s.claim, true, nil
}
func (s *workerStore) Authorized(context.Context, domain.Claimed) (bool, error) {
	return s.authorized, nil
}
func (s *workerStore) Events(context.Context, domain.Claimed) ([]domain.Event, error) {
	return s.events, nil
}
func (s *workerStore) ClaimRun(context.Context, domain.Claimed, string) (uuid.UUID, bool, error) {
	return uuid.New(), !s.duplicate, nil
}
func (s *workerStore) CompleteRun(_ context.Context, _ domain.Claimed, _ uuid.UUID, status string, _ *uuid.UUID, errorText string) error {
	s.completed = append(s.completed, status)
	s.errors = append(s.errors, errorText)
	return nil
}
func (s *workerStore) Advance(context.Context, domain.Claimed, domain.Event, string) error {
	s.advances++
	return nil
}
func (s *workerStore) Release(_ context.Context, _ domain.Claimed, next *time.Time, errorText string) error {
	s.next, s.released = next, errorText
	return nil
}

type storyActions struct {
	StoryActions
	story   domain.Story
	updates int
	creates int
	key     string
	err     error
}

func (a *storyActions) Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.Story, error) {
	return a.story, nil
}
func (a *storyActions) Update(context.Context, uuid.UUID, uuid.UUID, domain.Story, domain.Actions, string) error {
	a.updates++
	return a.err
}
func (a *storyActions) Create(_ context.Context, _, _, _ uuid.UUID, _ domain.Draft, key string) (uuid.UUID, error) {
	a.creates++
	a.key = key
	return uuid.New(), a.err
}

func ruleFixture() (*workerStore, *storyActions) {
	team := uuid.New()
	priority := "High"
	configuration, _ := json.Marshal(domain.RuleConfiguration{Version: 1, Trigger: "story.created", Actions: domain.Actions{Priority: &priority}})
	store := &workerStore{authorized: true, claim: domain.Claimed{Automation: domain.Automation{ID: uuid.New(), OwnerID: uuid.New(), TeamID: team, Kind: "rule", Configuration: configuration}, WorkspaceID: uuid.New(), LeaseToken: uuid.New()}, events: []domain.Event{{ID: uuid.New(), StoryID: uuid.New(), Kind: "story.created", CreatedAt: time.Now()}}}
	actions := &storyActions{story: domain.Story{TeamID: team, Priority: "Low", UpdatedAt: time.Now()}}
	return store, actions
}
func TestRulesClaimEachEventAndDoNotReplayCompletedRuns(t *testing.T) {
	store, actions := ruleFixture()
	store.duplicate = true
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if actions.updates != 0 || store.advances != 1 || len(store.completed) != 0 {
		t.Fatal("duplicate event mutated the task")
	}
	store, actions = ruleFixture()
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if actions.updates != 1 || store.completed[0] != "succeeded" || store.advances != 1 {
		t.Fatal("matching event was not applied exactly once")
	}
}
func TestRulesRespectOwnerRevocationAndSkipUnchangedValues(t *testing.T) {
	store, actions := ruleFixture()
	store.authorized = false
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if actions.updates != 0 || store.released == "" {
		t.Fatal("revoked owner performed a mutation")
	}
	store, actions = ruleFixture()
	actions.story.Priority = "High"
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if actions.updates != 0 || store.completed[0] != "skipped" {
		t.Fatal("unchanged assignment repeated a mutation")
	}
}
func TestRulesPersistSafeFailuresInsteadOfReflectingTaskPayloads(t *testing.T) {
	store, actions := ruleFixture()
	actions.err = errors.New("private-value customer description")
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.completed[0] != "failed" || store.released == "" || strings.Contains(store.released, "private-value") {
		t.Fatal("failure was hidden or reflected private payload")
	}
}
func TestRecurringCreationUsesStableOccurrenceKeyAndFutureCalendarDate(t *testing.T) {
	store, actions := ruleFixture()
	store.claim.Kind = "recurrence"
	due := time.Date(2026, 10, 1, 7, 0, 0, 0, time.UTC)
	store.claim.DueAt = &due
	raw, _ := json.Marshal(domain.RecurrenceConfiguration{Version: 1, Schedule: domain.Schedule{Frequency: "daily", Timezone: "Africa/Harare", StartsOn: "2026-10-01", LocalTime: "09:00", MonthDay: 1}, Draft: domain.Draft{Title: "Review", Priority: "High"}})
	store.claim.Configuration = raw
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	want := "team-automation:" + store.claim.ID.String() + ":2026-10-01T07:00:00Z"
	if actions.creates != 1 || actions.key != want || store.next == nil || !store.next.After(time.Now()) {
		t.Fatalf("unstable recurrence key or next date: %+v %+v", actions, store.next)
	}
	store.claimed = false
	if _, err := NewWorker(store, actions).DispatchBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if actions.key != want {
		t.Fatal("recovery changed the idempotency key")
	}
}
