package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeObjectiveOverdueStore struct {
	mu            sync.Mutex
	pages         [][]objectivesdomain.OverdueGuidanceRecipient
	page          int
	cursors       []*objectivesdomain.OverdueGuidanceCursor
	recipientAsOf []time.Time
	itemAsOf      []time.Time
	items         map[string][]objectivesdomain.OverdueGuidanceObjective
	itemCalls     map[string]int
	itemError     error
}

func (store *fakeObjectiveOverdueStore) ListOverdueObjectiveGuidanceRecipients(
	_ context.Context,
	asOf time.Time,
	cursor *objectivesdomain.OverdueGuidanceCursor,
	limit int,
) ([]objectivesdomain.OverdueGuidanceRecipient, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.recipientAsOf = append(store.recipientAsOf, asOf)
	if cursor == nil {
		store.cursors = append(store.cursors, nil)
	} else {
		copied := *cursor
		store.cursors = append(store.cursors, &copied)
	}
	if limit != objectiveLeadBatchSize || store.page >= len(store.pages) {
		return []objectivesdomain.OverdueGuidanceRecipient{}, nil
	}
	page := append([]objectivesdomain.OverdueGuidanceRecipient(nil), store.pages[store.page]...)
	store.page++
	return page, nil
}

func (store *fakeObjectiveOverdueStore) ListOverdueObjectiveGuidanceItems(
	_ context.Context,
	asOf time.Time,
	leadUserID uuid.UUID,
	workspaceID uuid.UUID,
) ([]objectivesdomain.OverdueGuidanceObjective, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.itemAsOf = append(store.itemAsOf, asOf)
	key := guidanceRecipientKey(leadUserID, workspaceID)
	store.itemCalls[key]++
	if store.itemError != nil {
		return nil, store.itemError
	}
	return append([]objectivesdomain.OverdueGuidanceObjective(nil), store.items[key]...), nil
}

type fakeOverdueStoryStore struct {
	mu            sync.Mutex
	pages         [][]storydomain.OverdueGuidanceRecipient
	page          int
	cursors       []*storydomain.OverdueGuidanceCursor
	recipientAsOf []time.Time
	itemAsOf      []time.Time
	items         map[string][]storydomain.OverdueGuidanceStory
	itemCalls     map[string]int
}

func (store *fakeOverdueStoryStore) ListOverdueStoryGuidanceRecipients(
	_ context.Context,
	asOf time.Time,
	cursor *storydomain.OverdueGuidanceCursor,
	limit int,
) ([]storydomain.OverdueGuidanceRecipient, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.recipientAsOf = append(store.recipientAsOf, asOf)
	if cursor == nil {
		store.cursors = append(store.cursors, nil)
	} else {
		copied := *cursor
		store.cursors = append(store.cursors, &copied)
	}
	if limit != overdueStoryAssigneeBatchSize || store.page >= len(store.pages) {
		return []storydomain.OverdueGuidanceRecipient{}, nil
	}
	page := append([]storydomain.OverdueGuidanceRecipient(nil), store.pages[store.page]...)
	store.page++
	return page, nil
}

func (store *fakeOverdueStoryStore) ListOverdueStoryGuidanceItems(
	_ context.Context,
	asOf time.Time,
	assigneeID uuid.UUID,
	workspaceID uuid.UUID,
) ([]storydomain.OverdueGuidanceStory, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.itemAsOf = append(store.itemAsOf, asOf)
	key := guidanceRecipientKey(assigneeID, workspaceID)
	store.itemCalls[key]++
	return append([]storydomain.OverdueGuidanceStory(nil), store.items[key]...), nil
}

type lockedGuidanceMailer struct {
	mu     sync.Mutex
	emails []mailer.TemplatedEmail
}

func (service *lockedGuidanceMailer) Send(context.Context, mailer.Email) error { return nil }

func (service *lockedGuidanceMailer) SendTemplated(_ context.Context, email mailer.TemplatedEmail) error {
	service.mu.Lock()
	defer service.mu.Unlock()
	service.emails = append(service.emails, email)
	return nil
}

func (service *lockedGuidanceMailer) count() int {
	service.mu.Lock()
	defer service.mu.Unlock()
	return len(service.emails)
}

func TestProcessObjectiveOverdueUsesCompositeCursorAndDoesNotRedeliver(t *testing.T) {
	t.Parallel()

	workspaceID := guidanceTestUUID(10_000)
	asOfInput := time.Date(2026, time.August, 29, 0, 30, 0, 0, time.FixedZone("test", 2*60*60))
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	firstPage := make([]objectivesdomain.OverdueGuidanceRecipient, objectiveLeadBatchSize)
	for index := range firstPage {
		firstPage[index] = objectiveGuidanceRecipient(guidanceTestUUID(index+1), workspaceID)
	}
	lastRecipient := objectiveGuidanceRecipient(guidanceTestUUID(objectiveLeadBatchSize+1), workspaceID)
	store := &fakeObjectiveOverdueStore{
		pages:     [][]objectivesdomain.OverdueGuidanceRecipient{firstPage, {lastRecipient}},
		items:     make(map[string][]objectivesdomain.OverdueGuidanceObjective),
		itemCalls: make(map[string]int),
	}
	for _, recipient := range []objectivesdomain.OverdueGuidanceRecipient{firstPage[0], lastRecipient} {
		store.items[guidanceRecipientKey(recipient.LeadUserID, workspaceID)] = []objectivesdomain.OverdueGuidanceObjective{
			objectiveGuidanceItem(recipient, workspaceID),
		}
	}
	mailerService := &lockedGuidanceMailer{}

	err := processObjectiveOverdueAt(context.Background(), store, newTestJobLogger(), mailerService, nil, nil, asOfInput)
	require.NoError(t, err)
	require.Equal(t, 2, mailerService.count())
	require.Len(t, store.cursors, 2)
	require.Nil(t, store.cursors[0])
	require.Equal(t, firstPage[len(firstPage)-1].LeadUserID, store.cursors[1].LeadUserID)
	require.Equal(t, workspaceID, store.cursors[1].WorkspaceID)
	requireGuidanceAsOf(t, asOfDate, store.recipientAsOf, store.itemAsOf)
	for key, calls := range store.itemCalls {
		require.Equalf(t, 1, calls, "recipient %s should be loaded once", key)
	}
}

func TestProcessOverdueStoriesEmailUsesCompositeCursorAndDoesNotRedeliver(t *testing.T) {
	t.Parallel()

	workspaceID := guidanceTestUUID(20_000)
	asOfInput := time.Date(2026, time.August, 29, 0, 30, 0, 0, time.FixedZone("test", 2*60*60))
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	firstPage := make([]storydomain.OverdueGuidanceRecipient, overdueStoryAssigneeBatchSize)
	for index := range firstPage {
		firstPage[index] = storyGuidanceRecipient(guidanceTestUUID(index+1), workspaceID)
	}
	lastRecipient := storyGuidanceRecipient(guidanceTestUUID(overdueStoryAssigneeBatchSize+1), workspaceID)
	store := &fakeOverdueStoryStore{
		pages:     [][]storydomain.OverdueGuidanceRecipient{firstPage, {lastRecipient}},
		items:     make(map[string][]storydomain.OverdueGuidanceStory),
		itemCalls: make(map[string]int),
	}
	for _, recipient := range []storydomain.OverdueGuidanceRecipient{firstPage[0], lastRecipient} {
		store.items[guidanceRecipientKey(recipient.AssigneeID, workspaceID)] = []storydomain.OverdueGuidanceStory{
			storyGuidanceItem(recipient, workspaceID),
		}
	}
	mailerService := &lockedGuidanceMailer{}

	err := processOverdueStoriesEmailAt(context.Background(), store, newTestJobLogger(), mailerService, nil, nil, asOfInput)
	require.NoError(t, err)
	require.Equal(t, 2, mailerService.count())
	require.Len(t, store.cursors, 2)
	require.Nil(t, store.cursors[0])
	require.Equal(t, firstPage[len(firstPage)-1].AssigneeID, store.cursors[1].AssigneeID)
	require.Equal(t, workspaceID, store.cursors[1].WorkspaceID)
	requireGuidanceAsOf(t, asOfDate, store.recipientAsOf, store.itemAsOf)
	for key, calls := range store.itemCalls {
		require.Equalf(t, 1, calls, "recipient %s should be loaded once", key)
	}
}

func TestProcessObjectiveOverdueRetriesRecipientReadFailureWithoutFailingSuccessfulBatch(t *testing.T) {
	t.Parallel()

	workspaceID := guidanceTestUUID(30_000)
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	recipient := objectiveGuidanceRecipient(guidanceTestUUID(1), workspaceID)
	store := &fakeObjectiveOverdueStore{
		pages:     [][]objectivesdomain.OverdueGuidanceRecipient{{recipient}},
		items:     make(map[string][]objectivesdomain.OverdueGuidanceObjective),
		itemCalls: make(map[string]int),
		itemError: errors.New("temporary objective read failure"),
	}
	mailerService := &lockedGuidanceMailer{}

	err := processObjectiveOverdueAt(
		context.Background(), store, newTestJobLogger(), mailerService, nil, nil,
		asOfDate,
	)
	require.NoError(t, err)
	require.Zero(t, mailerService.count())
	require.Equal(t, guidanceEmailRecipientAttempts, store.itemCalls[guidanceRecipientKey(recipient.LeadUserID, workspaceID)])
	requireGuidanceAsOf(t, asOfDate, store.recipientAsOf, store.itemAsOf)
}

func requireGuidanceAsOf(t *testing.T, expected time.Time, groups ...[]time.Time) {
	t.Helper()
	for _, values := range groups {
		require.NotEmpty(t, values)
		for _, value := range values {
			require.Equal(t, expected, value)
		}
	}
}

func guidanceTestUUID(value int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", value))
}

func guidanceRecipientKey(userID, workspaceID uuid.UUID) string {
	return userID.String() + "/" + workspaceID.String()
}

func objectiveGuidanceRecipient(userID, workspaceID uuid.UUID) objectivesdomain.OverdueGuidanceRecipient {
	return objectivesdomain.OverdueGuidanceRecipient{
		LeadUserID:    userID,
		LeadEmail:     userID.String() + "@example.com",
		LeadName:      "Objective lead",
		WorkspaceID:   workspaceID,
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
		EmailEnabled:  true,
	}
}

func objectiveGuidanceItem(recipient objectivesdomain.OverdueGuidanceRecipient, workspaceID uuid.UUID) objectivesdomain.OverdueGuidanceObjective {
	return objectivesdomain.OverdueGuidanceObjective{
		ID:             uuid.New(),
		Name:           "Protect launch readiness",
		EndDate:        time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
		LeadUserID:     recipient.LeadUserID,
		LeadEmail:      recipient.LeadEmail,
		LeadName:       recipient.LeadName,
		WorkspaceID:    workspaceID,
		WorkspaceName:  recipient.WorkspaceName,
		WorkspaceSlug:  recipient.WorkspaceSlug,
		TeamID:         uuid.New(),
		DeadlineStatus: "due_today",
		KeyResults:     "[]",
	}
}

func storyGuidanceRecipient(userID, workspaceID uuid.UUID) storydomain.OverdueGuidanceRecipient {
	return storydomain.OverdueGuidanceRecipient{
		AssigneeID:    userID,
		AssigneeEmail: userID.String() + "@example.com",
		AssigneeName:  "Story assignee",
		WorkspaceID:   workspaceID,
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
		EmailEnabled:  true,
	}
}

func storyGuidanceItem(recipient storydomain.OverdueGuidanceRecipient, workspaceID uuid.UUID) storydomain.OverdueGuidanceStory {
	return storydomain.OverdueGuidanceStory{
		ID:             uuid.New(),
		Title:          "Prepare launch metrics",
		EndDate:        time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
		AssigneeID:     recipient.AssigneeID,
		AssigneeEmail:  recipient.AssigneeEmail,
		AssigneeName:   recipient.AssigneeName,
		WorkspaceID:    workspaceID,
		WorkspaceName:  recipient.WorkspaceName,
		WorkspaceSlug:  recipient.WorkspaceSlug,
		TeamID:         uuid.New(),
		TeamName:       "Product",
		TeamCode:       "PRD",
		SequenceID:     42,
		StatusName:     "In progress",
		StatusCategory: "started",
		DeadlineStatus: "due_today",
		EmailEnabled:   true,
	}
}
