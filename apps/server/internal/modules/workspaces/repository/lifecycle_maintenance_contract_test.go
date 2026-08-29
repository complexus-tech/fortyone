package workspacesrepository

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceLifecycleMaintenanceQueriesAreClockedBoundedAndIntegrationSafe(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/lifecycle_maintenance.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: listinactiveworkspacewarningcandidates :many",
		"membership.role = 'admin'",
		"account.is_active = true",
		"workspace.inactivity_warning_sent_at is null",
		"not cast(sqlc.arg(has_cursor) as boolean)",
		"order by workspace.last_accessed_at, workspace.workspace_id",
		"-- name: markworkspaceinactivitywarningsent :execrows",
		"workspace.workspace_id = sqlc.arg(workspace_id)",
		"-- name: lockworkspaceintegrationlifecycle :exec",
		"pg_advisory_xact_lock(sqlc.arg(lock_key))",
		"-- name: listinactiveworkspacedeletioncandidates :many",
		"workspace.inactivity_warning_sent_at <=",
		"sqlc.arg(warning_sent_before)",
		"for update of workspace skip locked",
		"-- name: listdeletedworkspacepurgecandidates :many",
		"workspace.deleted_at is not null",
		"sqlc.arg(deleted_before)",
		"sqlc.arg(after_deleted_at)",
		"order by workspace.deleted_at, workspace.workspace_id",
		"-- name: countworkspacedeletioncandidatesawaitingslackencryption :one",
		"installation.workspace_id = any(cast(sqlc.arg(workspace_ids) as uuid[]))",
		"nullif(installation.bot_access_token, '') is not null",
		"-- name: snapshotworkspaceslackuninstalls :execrows",
		"'workspace_delete'",
		"on conflict (slack_workspace_id, installation_generation, uninstall_kind) do update",
		"where slack_uninstall_outbox.status = 'completed'",
		"-- name: cancelworkspaceslackinboundevents :execrows",
		"event.status in ('pending', 'processing', 'failed')",
		"-- name: cancelworkspaceslackoutbounddeliveries :execrows",
		"delivery.status in ('pending', 'delivering', 'failed')",
		"-- name: deleteinactiveworkspacecandidates :execrows",
		"-- name: deletesoftdeletedworkspacecandidates :execrows",
		"uninstall.installation_generation = installation.installation_generation",
		"nullif(uninstall.credential_payload, '') is not null",
	} {
		require.Contains(t, query, contract)
	}

	require.Equal(t, 3, strings.Count(query, "limit cast(sqlc.arg(batch_size) as integer)"))
	require.NotContains(t, query, " offset ")
	require.NotContains(t, query, "now()")
	require.NotContains(t, query, "current_timestamp")
	require.NotContains(t, query, "interval '")
	require.NotContains(t, query, "::")
}

func TestWorkspaceLifecycleWarningRepositoryMapsUTCAndKeysetCursor(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	inactiveBefore := time.Date(2026, time.February, 28, 10, 15, 0, 0, location)
	afterLastAccess := time.Date(2026, time.January, 4, 9, 0, 0, 0, location)
	workspaceID := uuid.New()
	adminEmails := []string{"admin@example.com"}
	queries := &workspaceLifecycleQueries{
		warningRows: []workspacesql.ListInactiveWorkspaceWarningCandidatesRow{{
			WorkspaceID:    workspaceID,
			Name:           "Acme",
			Slug:           "acme",
			LastAccessedAt: &afterLastAccess,
			AdminEmails:    adminEmails,
		}},
	}
	repository := newWithQueries(queries)

	candidates, err := repository.ListWorkspaceInactivityWarningCandidates(
		context.Background(),
		workspacedomain.InactivityWarningQuery{
			InactiveBefore: inactiveBefore,
			Cursor: workspacedomain.InactivityCursor{
				LastAccessedAt: afterLastAccess,
				WorkspaceID:    workspaceID,
				Valid:          true,
			},
			BatchSize: 100,
		},
	)

	require.NoError(t, err)
	require.Equal(t, workspacesql.ListInactiveWorkspaceWarningCandidatesParams{
		InactiveBefore:      inactiveBefore.UTC(),
		HasCursor:           true,
		AfterLastAccessedAt: afterLastAccess.UTC(),
		AfterWorkspaceID:    workspaceID,
		BatchSize:           100,
	}, queries.warningParams)
	require.Equal(t, []workspacedomain.InactivityWarningCandidate{{
		WorkspaceID:    workspaceID,
		Name:           "Acme",
		Slug:           "acme",
		LastAccessedAt: afterLastAccess.UTC(),
		AdminEmails:    adminEmails,
	}}, candidates)

	adminEmails[0] = "mutated@example.com"
	require.Equal(t, "admin@example.com", candidates[0].AdminEmails[0])
}

func TestWorkspaceLifecycleWarningReceiptAllowsConcurrentStateChange(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	receipt := workspacedomain.InactivityWarningReceipt{
		WorkspaceID:    uuid.New(),
		InactiveBefore: time.Date(2026, time.February, 28, 8, 0, 0, 0, location),
		WarningSentAt:  time.Date(2026, time.August, 28, 8, 0, 0, 0, location),
	}
	queries := &workspaceLifecycleQueries{warningRowsAffected: 0}
	repository := newWithQueries(queries)

	err := repository.RecordWorkspaceInactivityWarning(context.Background(), receipt)

	require.NoError(t, err)
	require.Equal(t, workspacesql.MarkWorkspaceInactivityWarningSentParams{
		WarningSentAt:  receipt.WarningSentAt.UTC(),
		WorkspaceID:    receipt.WorkspaceID,
		InactiveBefore: receipt.InactiveBefore.UTC(),
	}, queries.warningReceiptParams)
}

func TestDeleteInactiveWorkspacesBatchUsesOneOrderedTransaction(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	firstLastAccess := time.Date(2025, time.December, 1, 8, 0, 0, 0, location)
	secondLastAccess := firstLastAccess.Add(time.Hour)
	firstWorkspaceID := uuid.New()
	secondWorkspaceID := uuid.New()
	txQueries := &workspaceLifecycleQueries{
		deletionRows: []workspacesql.ListInactiveWorkspaceDeletionCandidatesRow{
			{WorkspaceID: firstWorkspaceID, LastAccessedAt: &firstLastAccess},
			{WorkspaceID: secondWorkspaceID, LastAccessedAt: &secondLastAccess},
		},
		blocked: 1,
		deleted: 1,
	}
	outerQueries := &workspaceLifecycleQueries{}
	repository := newWithQueries(outerQueries)
	transactionOutcome := ""
	repository.runTransaction = func(
		ctx context.Context,
		operation func(workspacesql.Querier) error,
	) error {
		transactionOutcome = "begun"
		err := operation(txQueries)
		if err != nil {
			transactionOutcome = "rolled_back"
			return err
		}
		transactionOutcome = "committed"
		return nil
	}
	batch := workspacedomain.InactivityDeletionBatch{
		InactiveBefore:              time.Date(2026, time.February, 28, 8, 0, 0, 0, location),
		WarningSentBefore:           time.Date(2026, time.July, 29, 8, 0, 0, 0, location),
		ProcessedAt:                 time.Date(2026, time.August, 28, 8, 0, 0, 0, location),
		Cursor:                      workspacedomain.InactivityCursor{},
		BatchSize:                   500,
		IntegrationLifecycleLockKey: 42,
	}

	result, err := repository.DeleteInactiveWorkspacesBatch(context.Background(), batch)

	require.NoError(t, err)
	require.Equal(t, "committed", transactionOutcome)
	require.Empty(t, outerQueries.calls)
	require.Equal(t, []string{
		"lock",
		"list_candidates",
		"count_blocked",
		"snapshot_uninstalls",
		"cancel_inbound",
		"cancel_outbound",
		"delete_workspaces",
	}, txQueries.calls)
	require.Equal(t, workspacedomain.InactivityDeletionResult{
		CandidateCount: 2,
		Deleted:        1,
		Blocked:        1,
		Cursor: workspacedomain.InactivityCursor{
			LastAccessedAt: secondLastAccess.UTC(),
			WorkspaceID:    secondWorkspaceID,
			Valid:          true,
		},
	}, result)

	txQueries.requireDeletionParams(t, batch, []uuid.UUID{firstWorkspaceID, secondWorkspaceID})
}

func TestDeleteInactiveWorkspacesBatchReturnsNoRolledBackCounts(t *testing.T) {
	t.Parallel()

	databaseErr := errors.New("database unavailable")
	lastAccess := time.Date(2025, time.December, 1, 8, 0, 0, 0, time.UTC)
	txQueries := &workspaceLifecycleQueries{
		deletionRows: []workspacesql.ListInactiveWorkspaceDeletionCandidatesRow{{
			WorkspaceID:    uuid.New(),
			LastAccessedAt: &lastAccess,
		}},
		failAt:  "cancel_inbound",
		failErr: databaseErr,
	}
	repository := newWithQueries(&workspaceLifecycleQueries{})
	transactionOutcome := ""
	repository.runTransaction = func(
		ctx context.Context,
		operation func(workspacesql.Querier) error,
	) error {
		err := operation(txQueries)
		if err != nil {
			transactionOutcome = "rolled_back"
			return err
		}
		transactionOutcome = "committed"
		return nil
	}

	result, err := repository.DeleteInactiveWorkspacesBatch(
		context.Background(),
		validInactivityDeletionBatch(),
	)

	require.ErrorIs(t, err, databaseErr)
	require.Equal(t, "rolled_back", transactionOutcome)
	require.Equal(t, workspacedomain.InactivityDeletionResult{}, result)
	require.Equal(t, []string{
		"lock",
		"list_candidates",
		"count_blocked",
		"snapshot_uninstalls",
		"cancel_inbound",
	}, txQueries.calls)
}

func TestPurgeSoftDeletedWorkspacesBatchUsesSharedOrderedTransactionAndKeysetCursor(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	afterDeletedAt := time.Date(2026, time.August, 20, 8, 0, 0, 0, location)
	firstDeletedAt := afterDeletedAt.Add(time.Hour)
	secondDeletedAt := firstDeletedAt.Add(time.Hour)
	afterWorkspaceID := uuid.New()
	firstWorkspaceID := uuid.New()
	secondWorkspaceID := uuid.New()
	txQueries := &workspaceLifecycleQueries{
		purgeRows: []workspacesql.ListDeletedWorkspacePurgeCandidatesRow{
			{WorkspaceID: firstWorkspaceID, DeletedAt: &firstDeletedAt},
			{WorkspaceID: secondWorkspaceID, DeletedAt: &secondDeletedAt},
		},
		blocked:      1,
		purgeDeleted: 1,
	}
	outerQueries := &workspaceLifecycleQueries{}
	repository := newWithQueries(outerQueries)
	transactionOutcome := ""
	repository.runTransaction = func(
		ctx context.Context,
		operation func(workspacesql.Querier) error,
	) error {
		transactionOutcome = "begun"
		err := operation(txQueries)
		if err != nil {
			transactionOutcome = "rolled_back"
			return err
		}
		transactionOutcome = "committed"
		return nil
	}
	batch := workspacedomain.DeletedWorkspacePurgeBatch{
		DeletedBefore: time.Date(2026, time.August, 26, 8, 0, 0, 0, location),
		ProcessedAt:   time.Date(2026, time.August, 28, 8, 0, 0, 0, location),
		Cursor: workspacedomain.DeletedWorkspacePurgeCursor{
			DeletedAt:   afterDeletedAt,
			WorkspaceID: afterWorkspaceID,
			Valid:       true,
		},
		BatchSize:                   500,
		IntegrationLifecycleLockKey: 42,
	}

	result, err := repository.PurgeSoftDeletedWorkspacesBatch(context.Background(), batch)

	require.NoError(t, err)
	require.Equal(t, "committed", transactionOutcome)
	require.Empty(t, outerQueries.calls)
	require.Equal(t, []string{
		"lock",
		"list_candidates",
		"count_blocked",
		"snapshot_uninstalls",
		"cancel_inbound",
		"cancel_outbound",
		"delete_workspaces",
	}, txQueries.calls)
	require.Equal(t, workspacedomain.DeletedWorkspacePurgeResult{
		CandidateCount: 2,
		Deleted:        1,
		Blocked:        1,
		Cursor: workspacedomain.DeletedWorkspacePurgeCursor{
			DeletedAt:   secondDeletedAt.UTC(),
			WorkspaceID: secondWorkspaceID,
			Valid:       true,
		},
	}, result)

	txQueries.requireSoftDeletedPurgeParams(
		t,
		batch,
		[]uuid.UUID{firstWorkspaceID, secondWorkspaceID},
	)
}

func TestPurgeSoftDeletedWorkspacesBatchReturnsNoRolledBackCounts(t *testing.T) {
	t.Parallel()

	databaseErr := errors.New("database unavailable")
	deletedAt := time.Date(2026, time.August, 20, 8, 0, 0, 0, time.UTC)
	txQueries := &workspaceLifecycleQueries{
		purgeRows: []workspacesql.ListDeletedWorkspacePurgeCandidatesRow{{
			WorkspaceID: uuid.New(),
			DeletedAt:   &deletedAt,
		}},
		failAt:  "cancel_outbound",
		failErr: databaseErr,
	}
	repository := newWithQueries(&workspaceLifecycleQueries{})
	transactionOutcome := ""
	repository.runTransaction = func(
		ctx context.Context,
		operation func(workspacesql.Querier) error,
	) error {
		err := operation(txQueries)
		if err != nil {
			transactionOutcome = "rolled_back"
			return err
		}
		transactionOutcome = "committed"
		return nil
	}

	result, err := repository.PurgeSoftDeletedWorkspacesBatch(
		context.Background(),
		validDeletedWorkspacePurgeBatch(),
	)

	require.ErrorIs(t, err, databaseErr)
	require.Equal(t, "rolled_back", transactionOutcome)
	require.Equal(t, workspacedomain.DeletedWorkspacePurgeResult{}, result)
	require.Equal(t, []string{
		"lock",
		"list_candidates",
		"count_blocked",
		"snapshot_uninstalls",
		"cancel_inbound",
		"cancel_outbound",
	}, txQueries.calls)
}

func TestWorkspaceLifecycleRepositoryRejectsInvalidBatchSizes(t *testing.T) {
	t.Parallel()

	repository := newWithQueries(&workspaceLifecycleQueries{})
	repository.runTransaction = func(context.Context, func(workspacesql.Querier) error) error {
		t.Fatal("transaction must not start for invalid input")
		return nil
	}
	tooLarge := int(math.MaxInt32) + 1

	_, warningErr := repository.ListWorkspaceInactivityWarningCandidates(
		context.Background(),
		workspacedomain.InactivityWarningQuery{
			InactiveBefore: time.Now().UTC(),
			BatchSize:      tooLarge,
		},
	)
	batch := validInactivityDeletionBatch()
	batch.BatchSize = tooLarge
	_, deletionErr := repository.DeleteInactiveWorkspacesBatch(context.Background(), batch)
	purgeBatch := validDeletedWorkspacePurgeBatch()
	purgeBatch.BatchSize = tooLarge
	_, purgeErr := repository.PurgeSoftDeletedWorkspacesBatch(context.Background(), purgeBatch)

	require.ErrorIs(t, warningErr, safecast.ErrOutOfRange)
	require.ErrorIs(t, deletionErr, safecast.ErrOutOfRange)
	require.ErrorIs(t, purgeErr, safecast.ErrOutOfRange)
}

func validInactivityDeletionBatch() workspacedomain.InactivityDeletionBatch {
	return workspacedomain.InactivityDeletionBatch{
		InactiveBefore:              time.Date(2026, time.February, 28, 8, 0, 0, 0, time.UTC),
		WarningSentBefore:           time.Date(2026, time.July, 29, 8, 0, 0, 0, time.UTC),
		ProcessedAt:                 time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
		BatchSize:                   500,
		IntegrationLifecycleLockKey: 42,
	}
}

func validDeletedWorkspacePurgeBatch() workspacedomain.DeletedWorkspacePurgeBatch {
	return workspacedomain.DeletedWorkspacePurgeBatch{
		DeletedBefore:               time.Date(2026, time.August, 26, 8, 0, 0, 0, time.UTC),
		ProcessedAt:                 time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
		BatchSize:                   500,
		IntegrationLifecycleLockKey: 42,
	}
}

type workspaceLifecycleQueries struct {
	workspacesql.Querier
	calls                []string
	warningRows          []workspacesql.ListInactiveWorkspaceWarningCandidatesRow
	warningParams        workspacesql.ListInactiveWorkspaceWarningCandidatesParams
	warningReceiptParams workspacesql.MarkWorkspaceInactivityWarningSentParams
	warningRowsAffected  int64
	deletionRows         []workspacesql.ListInactiveWorkspaceDeletionCandidatesRow
	deletionParams       workspacesql.ListInactiveWorkspaceDeletionCandidatesParams
	purgeRows            []workspacesql.ListDeletedWorkspacePurgeCandidatesRow
	purgeParams          workspacesql.ListDeletedWorkspacePurgeCandidatesParams
	blockedParams        workspacesql.CountWorkspaceDeletionCandidatesAwaitingSlackEncryptionParams
	snapshotParams       workspacesql.SnapshotWorkspaceSlackUninstallsParams
	inboundParams        workspacesql.CancelWorkspaceSlackInboundEventsParams
	outboundParams       workspacesql.CancelWorkspaceSlackOutboundDeliveriesParams
	deleteParams         workspacesql.DeleteInactiveWorkspaceCandidatesParams
	purgeDeleteParams    workspacesql.DeleteSoftDeletedWorkspaceCandidatesParams
	lockParams           workspacesql.LockWorkspaceIntegrationLifecycleParams
	blocked              int64
	deleted              int64
	purgeDeleted         int64
	failAt               string
	failErr              error
}

func (queries *workspaceLifecycleQueries) ListInactiveWorkspaceWarningCandidates(
	_ context.Context,
	params workspacesql.ListInactiveWorkspaceWarningCandidatesParams,
) ([]workspacesql.ListInactiveWorkspaceWarningCandidatesRow, error) {
	queries.warningParams = params
	return queries.warningRows, nil
}

func (queries *workspaceLifecycleQueries) MarkWorkspaceInactivityWarningSent(
	_ context.Context,
	params workspacesql.MarkWorkspaceInactivityWarningSentParams,
) (int64, error) {
	queries.warningReceiptParams = params
	return queries.warningRowsAffected, nil
}

func (queries *workspaceLifecycleQueries) LockWorkspaceIntegrationLifecycle(
	_ context.Context,
	params workspacesql.LockWorkspaceIntegrationLifecycleParams,
) error {
	queries.lockParams = params
	return queries.record("lock")
}

func (queries *workspaceLifecycleQueries) ListInactiveWorkspaceDeletionCandidates(
	_ context.Context,
	params workspacesql.ListInactiveWorkspaceDeletionCandidatesParams,
) ([]workspacesql.ListInactiveWorkspaceDeletionCandidatesRow, error) {
	queries.deletionParams = params
	if err := queries.record("list_candidates"); err != nil {
		return nil, err
	}
	return queries.deletionRows, nil
}

func (queries *workspaceLifecycleQueries) ListDeletedWorkspacePurgeCandidates(
	_ context.Context,
	params workspacesql.ListDeletedWorkspacePurgeCandidatesParams,
) ([]workspacesql.ListDeletedWorkspacePurgeCandidatesRow, error) {
	queries.purgeParams = params
	if err := queries.record("list_candidates"); err != nil {
		return nil, err
	}
	return queries.purgeRows, nil
}

func (queries *workspaceLifecycleQueries) CountWorkspaceDeletionCandidatesAwaitingSlackEncryption(
	_ context.Context,
	params workspacesql.CountWorkspaceDeletionCandidatesAwaitingSlackEncryptionParams,
) (int64, error) {
	queries.blockedParams = params
	if err := queries.record("count_blocked"); err != nil {
		return 0, err
	}
	return queries.blocked, nil
}

func (queries *workspaceLifecycleQueries) SnapshotWorkspaceSlackUninstalls(
	_ context.Context,
	params workspacesql.SnapshotWorkspaceSlackUninstallsParams,
) (int64, error) {
	queries.snapshotParams = params
	return 0, queries.record("snapshot_uninstalls")
}

func (queries *workspaceLifecycleQueries) CancelWorkspaceSlackInboundEvents(
	_ context.Context,
	params workspacesql.CancelWorkspaceSlackInboundEventsParams,
) (int64, error) {
	queries.inboundParams = params
	return 0, queries.record("cancel_inbound")
}

func (queries *workspaceLifecycleQueries) CancelWorkspaceSlackOutboundDeliveries(
	_ context.Context,
	params workspacesql.CancelWorkspaceSlackOutboundDeliveriesParams,
) (int64, error) {
	queries.outboundParams = params
	return 0, queries.record("cancel_outbound")
}

func (queries *workspaceLifecycleQueries) DeleteInactiveWorkspaceCandidates(
	_ context.Context,
	params workspacesql.DeleteInactiveWorkspaceCandidatesParams,
) (int64, error) {
	queries.deleteParams = params
	if err := queries.record("delete_workspaces"); err != nil {
		return 0, err
	}
	return queries.deleted, nil
}

func (queries *workspaceLifecycleQueries) DeleteSoftDeletedWorkspaceCandidates(
	_ context.Context,
	params workspacesql.DeleteSoftDeletedWorkspaceCandidatesParams,
) (int64, error) {
	queries.purgeDeleteParams = params
	if err := queries.record("delete_workspaces"); err != nil {
		return 0, err
	}
	return queries.purgeDeleted, nil
}

func (queries *workspaceLifecycleQueries) record(operation string) error {
	queries.calls = append(queries.calls, operation)
	if queries.failAt == operation {
		return queries.failErr
	}
	return nil
}

func (queries *workspaceLifecycleQueries) requireDeletionParams(
	t *testing.T,
	batch workspacedomain.InactivityDeletionBatch,
	workspaceIDs []uuid.UUID,
) {
	t.Helper()
	require.Equal(t, workspacesql.LockWorkspaceIntegrationLifecycleParams{
		LockKey: batch.IntegrationLifecycleLockKey,
	}, queries.lockParams)
	require.Equal(t, workspacesql.ListInactiveWorkspaceDeletionCandidatesParams{
		InactiveBefore:      batch.InactiveBefore.UTC(),
		WarningSentBefore:   batch.WarningSentBefore.UTC(),
		HasCursor:           false,
		AfterLastAccessedAt: time.Time{}.UTC(),
		AfterWorkspaceID:    uuid.Nil,
		BatchSize:           500,
	}, queries.deletionParams)
	require.Equal(t, workspaceIDs, queries.blockedParams.WorkspaceIds)
	require.Equal(t, int16(credentialvault.CurrentVersion), queries.blockedParams.CredentialKeyVersion)
	require.Equal(t, credentialvault.EnvelopePrefix+"%", queries.blockedParams.CredentialEnvelopePattern)
	require.Equal(t, workspaceIDs, queries.snapshotParams.WorkspaceIds)
	require.NotNil(t, queries.snapshotParams.ProcessedAt)
	require.Equal(t, batch.ProcessedAt.UTC(), *queries.snapshotParams.ProcessedAt)
	require.NotNil(t, queries.inboundParams.ProcessedAt)
	require.Equal(t, batch.ProcessedAt.UTC(), *queries.inboundParams.ProcessedAt)
	require.Equal(t, batch.ProcessedAt.UTC(), queries.outboundParams.ProcessedAt)
	require.Equal(t, batch.InactiveBefore.UTC(), queries.deleteParams.InactiveBefore)
	require.Equal(t, batch.WarningSentBefore.UTC(), queries.deleteParams.WarningSentBefore)
	require.Equal(t, workspaceIDs, queries.deleteParams.WorkspaceIds)
}

func (queries *workspaceLifecycleQueries) requireSoftDeletedPurgeParams(
	t *testing.T,
	batch workspacedomain.DeletedWorkspacePurgeBatch,
	workspaceIDs []uuid.UUID,
) {
	t.Helper()
	require.Equal(t, workspacesql.LockWorkspaceIntegrationLifecycleParams{
		LockKey: batch.IntegrationLifecycleLockKey,
	}, queries.lockParams)
	require.Equal(t, workspacesql.ListDeletedWorkspacePurgeCandidatesParams{
		DeletedBefore:    batch.DeletedBefore.UTC(),
		HasCursor:        true,
		AfterDeletedAt:   batch.Cursor.DeletedAt.UTC(),
		AfterWorkspaceID: batch.Cursor.WorkspaceID,
		BatchSize:        500,
	}, queries.purgeParams)
	require.Equal(t, workspaceIDs, queries.blockedParams.WorkspaceIds)
	require.Equal(t, int16(credentialvault.CurrentVersion), queries.blockedParams.CredentialKeyVersion)
	require.Equal(t, credentialvault.EnvelopePrefix+"%", queries.blockedParams.CredentialEnvelopePattern)
	require.Equal(t, workspaceIDs, queries.snapshotParams.WorkspaceIds)
	require.NotNil(t, queries.snapshotParams.ProcessedAt)
	require.Equal(t, batch.ProcessedAt.UTC(), *queries.snapshotParams.ProcessedAt)
	require.Equal(t, int16(credentialvault.CurrentVersion), queries.snapshotParams.CredentialKeyVersion)
	require.Equal(t, credentialvault.EnvelopePrefix+"%", queries.snapshotParams.CredentialEnvelopePattern)
	require.Equal(t, workspaceIDs, queries.inboundParams.WorkspaceIds)
	require.NotNil(t, queries.inboundParams.ProcessedAt)
	require.Equal(t, batch.ProcessedAt.UTC(), *queries.inboundParams.ProcessedAt)
	require.Equal(t, int16(credentialvault.CurrentVersion), queries.inboundParams.CredentialKeyVersion)
	require.Equal(t, credentialvault.EnvelopePrefix+"%", queries.inboundParams.CredentialEnvelopePattern)
	require.Equal(t, workspaceIDs, queries.outboundParams.WorkspaceIds)
	require.Equal(t, batch.ProcessedAt.UTC(), queries.outboundParams.ProcessedAt)
	require.Equal(t, int16(credentialvault.CurrentVersion), queries.outboundParams.CredentialKeyVersion)
	require.Equal(t, credentialvault.EnvelopePrefix+"%", queries.outboundParams.CredentialEnvelopePattern)
	require.Equal(t, workspacesql.DeleteSoftDeletedWorkspaceCandidatesParams{
		WorkspaceIds:              workspaceIDs,
		DeletedBefore:             batch.DeletedBefore.UTC(),
		CredentialKeyVersion:      int16(credentialvault.CurrentVersion),
		CredentialEnvelopePattern: credentialvault.EnvelopePrefix + "%",
	}, queries.purgeDeleteParams)
}
