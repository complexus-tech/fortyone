package workspacesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

var errWorkspaceLifecycleMaintenanceUnavailable = errors.New("workspace lifecycle maintenance repository is not configured")

// ListWorkspaceInactivityWarningCandidates returns one stable, bounded page of
// active workspaces that crossed the inactivity threshold and have not yet
// received a warning.
func (r *repo) ListWorkspaceInactivityWarningCandidates(
	ctx context.Context,
	query workspacedomain.InactivityWarningQuery,
) ([]workspacedomain.InactivityWarningCandidate, error) {
	if r == nil || r.queries == nil {
		return nil, errWorkspaceLifecycleMaintenanceUnavailable
	}
	if err := validateInactivityWarningQuery(query); err != nil {
		return nil, err
	}
	databaseBatchSize, err := safecast.Int32(query.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("validate workspace warning batch size: %w", err)
	}

	rows, err := r.queries.ListInactiveWorkspaceWarningCandidates(
		ctx,
		workspacesql.ListInactiveWorkspaceWarningCandidatesParams{
			InactiveBefore:      query.InactiveBefore.UTC(),
			HasCursor:           query.Cursor.Valid,
			AfterLastAccessedAt: query.Cursor.LastAccessedAt.UTC(),
			AfterWorkspaceID:    query.Cursor.WorkspaceID,
			BatchSize:           databaseBatchSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list workspace inactivity warning candidates: %w", err)
	}

	candidates := make([]workspacedomain.InactivityWarningCandidate, len(rows))
	for index, row := range rows {
		if row.LastAccessedAt == nil {
			return nil, fmt.Errorf("map workspace inactivity warning candidate %s: last access time is missing", row.WorkspaceID)
		}
		candidates[index] = workspacedomain.InactivityWarningCandidate{
			WorkspaceID:    row.WorkspaceID,
			Name:           row.Name,
			Slug:           row.Slug,
			LastAccessedAt: row.LastAccessedAt.UTC(),
			AdminEmails:    append([]string(nil), row.AdminEmails...),
		}
	}
	return candidates, nil
}

// RecordWorkspaceInactivityWarning stores the warning time only while the
// workspace still satisfies the candidate predicate. A zero-row update is a
// benign concurrent state change: the accepted email must not rewind a newer
// access or warning record.
func (r *repo) RecordWorkspaceInactivityWarning(
	ctx context.Context,
	receipt workspacedomain.InactivityWarningReceipt,
) error {
	if r == nil || r.queries == nil {
		return errWorkspaceLifecycleMaintenanceUnavailable
	}
	if receipt.WorkspaceID == uuid.Nil {
		return errors.New("workspace warning receipt requires a workspace")
	}
	if receipt.InactiveBefore.IsZero() || receipt.WarningSentAt.IsZero() {
		return errors.New("workspace warning receipt requires cutoff and sent times")
	}
	if !receipt.InactiveBefore.Before(receipt.WarningSentAt) {
		return errors.New("workspace warning inactivity cutoff must precede the sent time")
	}

	rows, err := r.queries.MarkWorkspaceInactivityWarningSent(
		ctx,
		workspacesql.MarkWorkspaceInactivityWarningSentParams{
			WarningSentAt:  receipt.WarningSentAt.UTC(),
			WorkspaceID:    receipt.WorkspaceID,
			InactiveBefore: receipt.InactiveBefore.UTC(),
		},
	)
	if err != nil {
		return fmt.Errorf("record workspace inactivity warning: %w", err)
	}
	if rows > 1 {
		return fmt.Errorf("record workspace inactivity warning: affected %d rows, want at most 1", rows)
	}
	return nil
}

// DeleteInactiveWorkspacesBatch atomically snapshots integration uninstall
// authority, cancels in-flight Slack work, and hard-deletes one locked page of
// inactivity-policy candidates. Workspaces with legacy credentials are
// counted but left untouched for a later retry after credential migration.
func (r *repo) DeleteInactiveWorkspacesBatch(
	ctx context.Context,
	batch workspacedomain.InactivityDeletionBatch,
) (workspacedomain.InactivityDeletionResult, error) {
	if r == nil || r.queries == nil || r.runTransaction == nil {
		return workspacedomain.InactivityDeletionResult{}, errWorkspaceLifecycleMaintenanceUnavailable
	}
	if err := validateInactivityDeletionBatch(batch); err != nil {
		return workspacedomain.InactivityDeletionResult{}, err
	}
	databaseBatchSize, err := safecast.Int32(batch.BatchSize)
	if err != nil {
		return workspacedomain.InactivityDeletionResult{}, fmt.Errorf("validate inactive workspace deletion batch size: %w", err)
	}

	batch.InactiveBefore = batch.InactiveBefore.UTC()
	batch.WarningSentBefore = batch.WarningSentBefore.UTC()
	batch.ProcessedAt = batch.ProcessedAt.UTC()
	batch.Cursor.LastAccessedAt = batch.Cursor.LastAccessedAt.UTC()

	var cursor workspacedomain.InactivityCursor
	result, err := r.deleteWorkspaceCandidatesBatch(ctx, workspaceDeletionCoreRequest{
		processedAt:                 batch.ProcessedAt,
		integrationLifecycleLockKey: batch.IntegrationLifecycleLockKey,
		listCandidates: func(queries workspacesql.Querier) ([]uuid.UUID, error) {
			candidates, listErr := queries.ListInactiveWorkspaceDeletionCandidates(
				ctx,
				workspacesql.ListInactiveWorkspaceDeletionCandidatesParams{
					InactiveBefore:      batch.InactiveBefore,
					WarningSentBefore:   batch.WarningSentBefore,
					HasCursor:           batch.Cursor.Valid,
					AfterLastAccessedAt: batch.Cursor.LastAccessedAt,
					AfterWorkspaceID:    batch.Cursor.WorkspaceID,
					BatchSize:           databaseBatchSize,
				},
			)
			if listErr != nil {
				return nil, fmt.Errorf("list inactive workspace deletion candidates: %w", listErr)
			}
			workspaceIDs := make([]uuid.UUID, len(candidates))
			for index, candidate := range candidates {
				if candidate.WorkspaceID == uuid.Nil || candidate.LastAccessedAt == nil {
					return nil, fmt.Errorf("map inactive workspace deletion candidate at index %d: identity and last access time are required", index)
				}
				workspaceIDs[index] = candidate.WorkspaceID
			}
			if len(candidates) > 0 {
				lastCandidate := candidates[len(candidates)-1]
				cursor = workspacedomain.InactivityCursor{
					LastAccessedAt: lastCandidate.LastAccessedAt.UTC(),
					WorkspaceID:    lastCandidate.WorkspaceID,
					Valid:          true,
				}
			}
			return workspaceIDs, nil
		},
		deleteCandidates: func(
			queries workspacesql.Querier,
			scope workspaceDeletionCredentialScope,
		) (int64, error) {
			return queries.DeleteInactiveWorkspaceCandidates(
				ctx,
				workspacesql.DeleteInactiveWorkspaceCandidatesParams{
					WorkspaceIds:              scope.workspaceIDs,
					InactiveBefore:            batch.InactiveBefore,
					WarningSentBefore:         batch.WarningSentBefore,
					CredentialKeyVersion:      scope.credentialVersion,
					CredentialEnvelopePattern: scope.credentialPattern,
				},
			)
		},
	})
	if err != nil {
		return workspacedomain.InactivityDeletionResult{}, fmt.Errorf("delete inactive workspace batch: %w", err)
	}
	return workspacedomain.InactivityDeletionResult{
		CandidateCount: result.candidateCount,
		Deleted:        result.deleted,
		Blocked:        result.blocked,
		Cursor:         cursor,
	}, nil
}

// PurgeSoftDeletedWorkspacesBatch atomically retires integrations and
// hard-deletes one locked page of workspaces whose trash-retention cutoff has
// elapsed. The destructive integration sequence is shared with inactivity
// deletion; only candidate selection and the final policy predicate differ.
func (r *repo) PurgeSoftDeletedWorkspacesBatch(
	ctx context.Context,
	batch workspacedomain.DeletedWorkspacePurgeBatch,
) (workspacedomain.DeletedWorkspacePurgeResult, error) {
	if r == nil || r.queries == nil || r.runTransaction == nil {
		return workspacedomain.DeletedWorkspacePurgeResult{}, errWorkspaceLifecycleMaintenanceUnavailable
	}
	if err := validateDeletedWorkspacePurgeBatch(batch); err != nil {
		return workspacedomain.DeletedWorkspacePurgeResult{}, err
	}
	databaseBatchSize, err := safecast.Int32(batch.BatchSize)
	if err != nil {
		return workspacedomain.DeletedWorkspacePurgeResult{}, fmt.Errorf("validate soft-deleted workspace purge batch size: %w", err)
	}

	batch.DeletedBefore = batch.DeletedBefore.UTC()
	batch.ProcessedAt = batch.ProcessedAt.UTC()
	batch.Cursor.DeletedAt = batch.Cursor.DeletedAt.UTC()

	var cursor workspacedomain.DeletedWorkspacePurgeCursor
	result, err := r.deleteWorkspaceCandidatesBatch(ctx, workspaceDeletionCoreRequest{
		processedAt:                 batch.ProcessedAt,
		integrationLifecycleLockKey: batch.IntegrationLifecycleLockKey,
		listCandidates: func(queries workspacesql.Querier) ([]uuid.UUID, error) {
			candidates, listErr := queries.ListDeletedWorkspacePurgeCandidates(
				ctx,
				workspacesql.ListDeletedWorkspacePurgeCandidatesParams{
					DeletedBefore:    batch.DeletedBefore,
					HasCursor:        batch.Cursor.Valid,
					AfterDeletedAt:   batch.Cursor.DeletedAt,
					AfterWorkspaceID: batch.Cursor.WorkspaceID,
					BatchSize:        databaseBatchSize,
				},
			)
			if listErr != nil {
				return nil, fmt.Errorf("list soft-deleted workspace purge candidates: %w", listErr)
			}
			workspaceIDs := make([]uuid.UUID, len(candidates))
			for index, candidate := range candidates {
				if candidate.WorkspaceID == uuid.Nil || candidate.DeletedAt == nil {
					return nil, fmt.Errorf("map soft-deleted workspace purge candidate at index %d: identity and deletion time are required", index)
				}
				workspaceIDs[index] = candidate.WorkspaceID
			}
			if len(candidates) > 0 {
				lastCandidate := candidates[len(candidates)-1]
				cursor = workspacedomain.DeletedWorkspacePurgeCursor{
					DeletedAt:   lastCandidate.DeletedAt.UTC(),
					WorkspaceID: lastCandidate.WorkspaceID,
					Valid:       true,
				}
			}
			return workspaceIDs, nil
		},
		deleteCandidates: func(
			queries workspacesql.Querier,
			scope workspaceDeletionCredentialScope,
		) (int64, error) {
			return queries.DeleteSoftDeletedWorkspaceCandidates(
				ctx,
				workspacesql.DeleteSoftDeletedWorkspaceCandidatesParams{
					WorkspaceIds:              scope.workspaceIDs,
					DeletedBefore:             batch.DeletedBefore,
					CredentialKeyVersion:      scope.credentialVersion,
					CredentialEnvelopePattern: scope.credentialPattern,
				},
			)
		},
	})
	if err != nil {
		return workspacedomain.DeletedWorkspacePurgeResult{}, fmt.Errorf("purge soft-deleted workspace batch: %w", err)
	}
	return workspacedomain.DeletedWorkspacePurgeResult{
		CandidateCount: result.candidateCount,
		Deleted:        result.deleted,
		Blocked:        result.blocked,
		Cursor:         cursor,
	}, nil
}

type workspaceDeletionCoreRequest struct {
	processedAt                 time.Time
	integrationLifecycleLockKey int64
	listCandidates              func(workspacesql.Querier) ([]uuid.UUID, error)
	deleteCandidates            func(workspacesql.Querier, workspaceDeletionCredentialScope) (int64, error)
}

type workspaceDeletionCredentialScope struct {
	workspaceIDs      []uuid.UUID
	credentialVersion int16
	credentialPattern string
}

type workspaceDeletionCoreResult struct {
	candidateCount int
	deleted        int64
	blocked        int64
}

// deleteWorkspaceCandidatesBatch is the single integration-safe destructive
// transaction used by every workspace deletion policy. Callers own only the
// policy-specific locked candidate read and final guarded DELETE.
func (r *repo) deleteWorkspaceCandidatesBatch(
	ctx context.Context,
	request workspaceDeletionCoreRequest,
) (workspaceDeletionCoreResult, error) {
	credentialVersion, err := safecast.Int16(credentialvault.CurrentVersion)
	if err != nil {
		return workspaceDeletionCoreResult{}, fmt.Errorf("validate credential vault version: %w", err)
	}
	if request.listCandidates == nil || request.deleteCandidates == nil {
		return workspaceDeletionCoreResult{}, errors.New("workspace deletion candidate and delete operations are required")
	}

	credentialScope := workspaceDeletionCredentialScope{
		credentialVersion: credentialVersion,
		credentialPattern: credentialvault.EnvelopePrefix + "%",
	}
	var committed workspaceDeletionCoreResult
	err = r.withinTransaction(ctx, func(queries workspacesql.Querier) error {
		if lockErr := queries.LockWorkspaceIntegrationLifecycle(
			ctx,
			workspacesql.LockWorkspaceIntegrationLifecycleParams{
				LockKey: request.integrationLifecycleLockKey,
			},
		); lockErr != nil {
			return fmt.Errorf("lock integration lifecycle: %w", lockErr)
		}

		workspaceIDs, listErr := request.listCandidates(queries)
		if listErr != nil {
			return listErr
		}
		if len(workspaceIDs) == 0 {
			return nil
		}
		seen := make(map[uuid.UUID]struct{}, len(workspaceIDs))
		for index, workspaceID := range workspaceIDs {
			if workspaceID == uuid.Nil {
				return fmt.Errorf("workspace deletion candidate at index %d has no identity", index)
			}
			if _, duplicate := seen[workspaceID]; duplicate {
				return fmt.Errorf("workspace deletion candidate %s is duplicated", workspaceID)
			}
			seen[workspaceID] = struct{}{}
		}
		credentialScope.workspaceIDs = workspaceIDs

		blocked, countErr := queries.CountWorkspaceDeletionCandidatesAwaitingSlackEncryption(
			ctx,
			workspacesql.CountWorkspaceDeletionCandidatesAwaitingSlackEncryptionParams{
				WorkspaceIds:              workspaceIDs,
				CredentialKeyVersion:      credentialScope.credentialVersion,
				CredentialEnvelopePattern: credentialScope.credentialPattern,
			},
		)
		if countErr != nil {
			return fmt.Errorf("count workspaces awaiting Slack credential encryption: %w", countErr)
		}
		if blocked < 0 || blocked > int64(len(workspaceIDs)) {
			return fmt.Errorf(
				"count workspaces awaiting Slack credential encryption: got %d, want 0..%d",
				blocked,
				len(workspaceIDs),
			)
		}

		processedAt := request.processedAt
		if _, snapshotErr := queries.SnapshotWorkspaceSlackUninstalls(
			ctx,
			workspacesql.SnapshotWorkspaceSlackUninstallsParams{
				ProcessedAt:               &processedAt,
				WorkspaceIds:              workspaceIDs,
				CredentialKeyVersion:      credentialScope.credentialVersion,
				CredentialEnvelopePattern: credentialScope.credentialPattern,
			},
		); snapshotErr != nil {
			return fmt.Errorf("snapshot Slack uninstalls: %w", snapshotErr)
		}
		if _, cancelErr := queries.CancelWorkspaceSlackInboundEvents(
			ctx,
			workspacesql.CancelWorkspaceSlackInboundEventsParams{
				ProcessedAt:               &processedAt,
				WorkspaceIds:              workspaceIDs,
				CredentialKeyVersion:      credentialScope.credentialVersion,
				CredentialEnvelopePattern: credentialScope.credentialPattern,
			},
		); cancelErr != nil {
			return fmt.Errorf("cancel Slack inbox work: %w", cancelErr)
		}
		if _, cancelErr := queries.CancelWorkspaceSlackOutboundDeliveries(
			ctx,
			workspacesql.CancelWorkspaceSlackOutboundDeliveriesParams{
				ProcessedAt:               processedAt,
				WorkspaceIds:              workspaceIDs,
				CredentialKeyVersion:      credentialScope.credentialVersion,
				CredentialEnvelopePattern: credentialScope.credentialPattern,
			},
		); cancelErr != nil {
			return fmt.Errorf("cancel Slack outbound work: %w", cancelErr)
		}

		deleted, deleteErr := request.deleteCandidates(queries, credentialScope)
		if deleteErr != nil {
			return fmt.Errorf("delete workspace candidates: %w", deleteErr)
		}
		expectedDeleted := int64(len(workspaceIDs)) - blocked
		if deleted != expectedDeleted {
			return fmt.Errorf(
				"delete workspace candidates: affected %d rows, want %d",
				deleted,
				expectedDeleted,
			)
		}
		committed = workspaceDeletionCoreResult{
			candidateCount: len(workspaceIDs),
			deleted:        deleted,
			blocked:        blocked,
		}
		return nil
	})
	if err != nil {
		return workspaceDeletionCoreResult{}, err
	}
	return committed, nil
}

func validateInactivityWarningQuery(query workspacedomain.InactivityWarningQuery) error {
	if query.InactiveBefore.IsZero() {
		return errors.New("workspace warning inactivity cutoff is required")
	}
	if query.BatchSize <= 0 {
		return errors.New("workspace warning batch size must be positive")
	}
	return validateInactivityCursor(query.Cursor)
}

func validateInactivityDeletionBatch(batch workspacedomain.InactivityDeletionBatch) error {
	if batch.InactiveBefore.IsZero() || batch.WarningSentBefore.IsZero() || batch.ProcessedAt.IsZero() {
		return errors.New("inactive workspace deletion cutoffs and processing time are required")
	}
	if !batch.InactiveBefore.Before(batch.ProcessedAt) || !batch.WarningSentBefore.Before(batch.ProcessedAt) {
		return errors.New("inactive workspace deletion cutoffs must precede the processing time")
	}
	if batch.BatchSize <= 0 {
		return errors.New("inactive workspace deletion batch size must be positive")
	}
	if batch.IntegrationLifecycleLockKey == 0 {
		return errors.New("inactive workspace deletion requires an integration lifecycle lock")
	}
	return validateInactivityCursor(batch.Cursor)
}

func validateInactivityCursor(cursor workspacedomain.InactivityCursor) error {
	if !cursor.Valid {
		return nil
	}
	if cursor.LastAccessedAt.IsZero() || cursor.WorkspaceID == uuid.Nil {
		return errors.New("workspace inactivity cursor requires last access time and workspace ID")
	}
	return nil
}

func validateDeletedWorkspacePurgeBatch(batch workspacedomain.DeletedWorkspacePurgeBatch) error {
	if batch.DeletedBefore.IsZero() || batch.ProcessedAt.IsZero() {
		return errors.New("soft-deleted workspace purge cutoff and processing time are required")
	}
	if !batch.DeletedBefore.Before(batch.ProcessedAt) {
		return errors.New("soft-deleted workspace purge cutoff must precede the processing time")
	}
	if batch.BatchSize <= 0 {
		return errors.New("soft-deleted workspace purge batch size must be positive")
	}
	if batch.IntegrationLifecycleLockKey == 0 {
		return errors.New("soft-deleted workspace purge requires an integration lifecycle lock")
	}
	return validateDeletedWorkspacePurgeCursor(batch.Cursor)
}

func validateDeletedWorkspacePurgeCursor(cursor workspacedomain.DeletedWorkspacePurgeCursor) error {
	if !cursor.Valid {
		return nil
	}
	if cursor.DeletedAt.IsZero() || cursor.WorkspaceID == uuid.Nil {
		return errors.New("soft-deleted workspace purge cursor requires deletion time and workspace ID")
	}
	return nil
}
