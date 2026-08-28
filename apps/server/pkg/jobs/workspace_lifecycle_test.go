package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceInactivityWarningUsesOneUTCClockAndPreservesMailerContract(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	lastAccess := time.Date(2026, time.January, 2, 8, 0, 0, 0, time.UTC)
	workspaceID := uuid.New()
	store := &workspaceWarningStoreStub{pages: [][]workspacedomain.InactivityWarningCandidate{{{
		WorkspaceID:    workspaceID,
		Name:           "Acme",
		Slug:           "acme",
		LastAccessedAt: lastAccess,
		AdminEmails:    []string{"owner@example.com", "admin@example.com"},
	}}}}
	mailerService := &workspaceLifecycleMailerStub{}

	err := processWorkspaceInactivityWarningAt(
		context.Background(),
		store,
		newTestJobLogger(),
		mailerService,
		now,
	)

	require.NoError(t, err)
	require.Equal(t, []workspacedomain.InactivityWarningQuery{{
		InactiveBefore: now.UTC().AddDate(0, -6, 0),
		BatchSize:      workspaceInactivityWarningBatchSize,
	}}, store.queries)
	require.Equal(t, []workspacedomain.InactivityWarningReceipt{{
		WorkspaceID:    workspaceID,
		InactiveBefore: now.UTC().AddDate(0, -6, 0),
		WarningSentAt:  now.UTC(),
	}}, store.receipts)
	require.Len(t, mailerService.templated, 1)
	email := mailerService.templated[0]
	require.Equal(t, []string{"owner@example.com", "admin@example.com"}, email.To)
	require.Equal(t, "workspaces/inactivity_warning", email.Template)
	require.Equal(t, "Acme workspace scheduled for deletion", email.Subject)
	require.Equal(t, map[string]any{
		"WorkspaceName": "Acme",
		"WorkspaceURL":  "https://acme.fortyone.app",
	}, email.Data)
}

func TestWorkspaceInactivityWarningUsesStableCursorAcrossFullPages(t *testing.T) {
	firstPage := make([]workspacedomain.InactivityWarningCandidate, workspaceInactivityWarningBatchSize)
	baseTime := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	for index := range firstPage {
		firstPage[index] = workspacedomain.InactivityWarningCandidate{
			WorkspaceID:    uuid.New(),
			Name:           "Workspace",
			Slug:           "workspace",
			LastAccessedAt: baseTime.Add(time.Duration(index) * time.Minute),
			AdminEmails:    []string{"admin@example.com"},
		}
	}
	lastCandidate := firstPage[len(firstPage)-1]
	secondCandidate := workspacedomain.InactivityWarningCandidate{
		WorkspaceID:    uuid.New(),
		Name:           "Last",
		Slug:           "last",
		LastAccessedAt: lastCandidate.LastAccessedAt.Add(time.Minute),
		AdminEmails:    []string{"admin@example.com"},
	}
	store := &workspaceWarningStoreStub{
		pages: [][]workspacedomain.InactivityWarningCandidate{
			firstPage,
			{secondCandidate},
		},
	}

	err := processWorkspaceInactivityWarningAt(
		context.Background(),
		store,
		newTestJobLogger(),
		&workspaceLifecycleMailerStub{},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Len(t, store.queries, 2)
	require.Equal(t, workspacedomain.InactivityCursor{
		LastAccessedAt: lastCandidate.LastAccessedAt,
		WorkspaceID:    lastCandidate.WorkspaceID,
		Valid:          true,
	}, store.queries[1].Cursor)
	require.Len(t, store.receipts, workspaceInactivityWarningBatchSize+1)
}

func TestWorkspaceInactivityWarningSkipsCandidateWithoutActiveAdmins(t *testing.T) {
	t.Parallel()

	store := &workspaceWarningStoreStub{pages: [][]workspacedomain.InactivityWarningCandidate{{{
		WorkspaceID:    uuid.New(),
		Name:           "Orphaned",
		Slug:           "orphaned",
		LastAccessedAt: time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
	}}}}
	mailerService := &workspaceLifecycleMailerStub{}

	err := processWorkspaceInactivityWarningAt(
		context.Background(),
		store,
		newTestJobLogger(),
		mailerService,
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Empty(t, mailerService.templated)
	require.Empty(t, store.receipts)
}

func TestWorkspaceInactivityWarningDoesNotRecordFailedDelivery(t *testing.T) {
	t.Parallel()

	mailerErr := errors.New("mailer unavailable")
	store := &workspaceWarningStoreStub{pages: [][]workspacedomain.InactivityWarningCandidate{{{
		WorkspaceID:    uuid.New(),
		Name:           "Acme",
		Slug:           "acme",
		LastAccessedAt: time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
		AdminEmails:    []string{"admin@example.com"},
	}}}}

	err := processWorkspaceInactivityWarningAt(
		context.Background(),
		store,
		newTestJobLogger(),
		&workspaceLifecycleMailerStub{templatedErr: mailerErr},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Empty(t, store.receipts)
}

func TestWorkspaceDeletionUsesIndependentInactivityAndWarningCutoffsAcrossPages(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	firstCursor := workspacedomain.InactivityCursor{
		LastAccessedAt: time.Date(2025, time.November, 1, 8, 0, 0, 0, time.UTC),
		WorkspaceID:    uuid.New(),
		Valid:          true,
	}
	secondCursor := workspacedomain.InactivityCursor{
		LastAccessedAt: firstCursor.LastAccessedAt.Add(time.Hour),
		WorkspaceID:    uuid.New(),
		Valid:          true,
	}
	store := &inactiveWorkspaceDeleterStub{results: []workspacedomain.InactivityDeletionResult{
		{
			CandidateCount: maintenancePurgeBatchSize,
			Deleted:        maintenancePurgeBatchSize - 1,
			Blocked:        1,
			Cursor:         firstCursor,
		},
		{
			CandidateCount: 2,
			Deleted:        2,
			Cursor:         secondCursor,
		},
	}}

	err := processWorkspaceDeletionAt(context.Background(), store, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Len(t, store.batches, 2)
	for _, batch := range store.batches {
		require.Equal(t, now.UTC().AddDate(0, -6, 0), batch.InactiveBefore)
		require.Equal(t, now.UTC().Add(-30*24*time.Hour), batch.WarningSentBefore)
		require.Equal(t, now.UTC(), batch.ProcessedAt)
		require.Equal(t, maintenancePurgeBatchSize, batch.BatchSize)
		require.Equal(t, slackrepository.SlackInstallationLifecycleAdvisoryKey, batch.IntegrationLifecycleLockKey)
	}
	require.False(t, store.batches[0].Cursor.Valid)
	require.Equal(t, firstCursor, store.batches[1].Cursor)
}

func TestWorkspaceDeletionRejectsUnaccountedCandidates(t *testing.T) {
	t.Parallel()

	store := &inactiveWorkspaceDeleterStub{results: []workspacedomain.InactivityDeletionResult{{
		CandidateCount: 2,
		Deleted:        1,
		Cursor: workspacedomain.InactivityCursor{
			LastAccessedAt: time.Date(2025, time.November, 1, 8, 0, 0, 0, time.UTC),
			WorkspaceID:    uuid.New(),
			Valid:          true,
		},
	}}}

	err := processWorkspaceDeletionAt(
		context.Background(),
		store,
		newTestJobLogger(),
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.ErrorContains(t, err, "invalid result")
}

type workspaceWarningStoreStub struct {
	pages       [][]workspacedomain.InactivityWarningCandidate
	queries     []workspacedomain.InactivityWarningQuery
	receipts    []workspacedomain.InactivityWarningReceipt
	listErr     error
	recordErr   error
	currentPage int
}

func (store *workspaceWarningStoreStub) ListWorkspaceInactivityWarningCandidates(
	_ context.Context,
	query workspacedomain.InactivityWarningQuery,
) ([]workspacedomain.InactivityWarningCandidate, error) {
	store.queries = append(store.queries, query)
	if store.listErr != nil {
		return nil, store.listErr
	}
	if store.currentPage >= len(store.pages) {
		return nil, nil
	}
	page := store.pages[store.currentPage]
	store.currentPage++
	return page, nil
}

func (store *workspaceWarningStoreStub) RecordWorkspaceInactivityWarning(
	_ context.Context,
	receipt workspacedomain.InactivityWarningReceipt,
) error {
	store.receipts = append(store.receipts, receipt)
	return store.recordErr
}

type workspaceLifecycleMailerStub struct {
	templated    []mailer.TemplatedEmail
	templatedErr error
}

func (*workspaceLifecycleMailerStub) Send(context.Context, mailer.Email) error {
	return nil
}

func (service *workspaceLifecycleMailerStub) SendTemplated(
	_ context.Context,
	email mailer.TemplatedEmail,
) error {
	service.templated = append(service.templated, email)
	return service.templatedErr
}

type inactiveWorkspaceDeleterStub struct {
	results []workspacedomain.InactivityDeletionResult
	batches []workspacedomain.InactivityDeletionBatch
	err     error
}

func (store *inactiveWorkspaceDeleterStub) DeleteInactiveWorkspacesBatch(
	_ context.Context,
	batch workspacedomain.InactivityDeletionBatch,
) (workspacedomain.InactivityDeletionResult, error) {
	store.batches = append(store.batches, batch)
	if store.err != nil {
		return workspacedomain.InactivityDeletionResult{}, store.err
	}
	if len(store.results) == 0 {
		return workspacedomain.InactivityDeletionResult{}, nil
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}
