package calendarrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type reconciliationBlock struct {
	ID                 uuid.UUID  `db:"block_id"`
	SegmentIndex       int        `db:"segment_index"`
	Title              string     `db:"title"`
	StartAt            time.Time  `db:"start_at"`
	EndAt              time.Time  `db:"end_at"`
	IsLocked           bool       `db:"is_locked"`
	ExternalProvider   *string    `db:"external_provider"`
	ExternalCalendarID *string    `db:"external_calendar_id"`
	ExternalEventID    *string    `db:"external_event_id"`
	ExternalSyncHash   *string    `db:"external_sync_hash"`
	ManualOverrideAt   *time.Time `db:"manual_override_at"`
	ManualOverrideBy   *uuid.UUID `db:"manual_override_by"`
}

type dbScheduleEventOutbox struct {
	ID              uuid.UUID       `db:"outbox_id"`
	WorkspaceID     uuid.UUID       `db:"workspace_id"`
	UserID          uuid.UUID       `db:"user_id"`
	ScheduleBlockID *uuid.UUID      `db:"schedule_block_id"`
	Operation       string          `db:"operation"`
	Provider        string          `db:"provider"`
	CalendarID      string          `db:"calendar_id"`
	ProviderEventID string          `db:"provider_event_id"`
	Payload         json.RawMessage `db:"payload"`
	DedupeKey       string          `db:"dedupe_key"`
	AttemptCount    int             `db:"attempt_count"`
}

type transactionScheduleEventOutboxStore struct {
	tx *sqlx.Tx
}

const scheduleEventUpsertIsCurrentQuery = `
	SELECT EXISTS (
		SELECT 1
		FROM calendar_schedule_blocks block
		INNER JOIN calendar_maya_schedule_ownerships ownership ON
			ownership.workspace_id = block.workspace_id
			AND ownership.story_id = block.story_id
			AND ownership.user_id = block.user_id
		INNER JOIN stories story ON
			story.id = block.story_id
			AND story.workspace_id = block.workspace_id
		INNER JOIN statuses status ON status.status_id = story.status_id
		INNER JOIN users owner_user ON owner_user.user_id = block.user_id AND owner_user.is_active = TRUE
		INNER JOIN workspace_members workspace_member ON
			workspace_member.workspace_id = block.workspace_id
			AND workspace_member.user_id = block.user_id
		INNER JOIN team_members team_member ON
			team_member.team_id = story.team_id
			AND team_member.user_id = block.user_id
		LEFT JOIN team_sprint_settings team_settings ON
			team_settings.team_id = story.team_id
			AND team_settings.workspace_id = story.workspace_id
		LEFT JOIN sprints sprint ON sprint.sprint_id = story.sprint_id
		WHERE block.block_id = $1
			AND block.workspace_id = $2
			AND block.user_id = $3
			AND block.source = 'maya'
			AND block.external_provider = $4
			AND block.external_calendar_id = $5
			AND block.external_event_id = $6
			AND block.title = $7
			AND story.title = $7
			AND block.start_at = $8
			AND block.end_at = $9
			AND story.assignee_id = block.user_id
			AND story.auto_scheduling_enabled = TRUE
			AND ownership.updated_at >= story.updated_at
			AND (team_settings.updated_at IS NULL OR ownership.updated_at >= team_settings.updated_at)
			AND (sprint.updated_at IS NULL OR ownership.updated_at >= sprint.updated_at)
			AND story.deleted_at IS NULL
			AND story.archived_at IS NULL
			AND story.completed_at IS NULL
			AND status.category NOT IN ('completed', 'cancelled')
	)
`

// calendar_schedule_blocks.updated_at is the concurrency token for user-visible
// schedule content. Provider mirror bookkeeping must preserve it.
const markScheduleBlockMirroredQuery = `
	UPDATE calendar_schedule_blocks
	SET external_provider = $4,
		external_calendar_id = $5,
		external_event_id = $3,
		external_sync_hash = $2,
		external_synced_at = CURRENT_TIMESTAMP
	WHERE block_id = $1
`

const mayaScheduleEligibilityQuery = `
	SELECT story.id
	FROM stories story
	INNER JOIN statuses status ON
		status.status_id = story.status_id
	INNER JOIN users selected_user ON
		selected_user.user_id = $3
		AND selected_user.is_active = TRUE
	INNER JOIN workspace_members membership ON
		membership.workspace_id = story.workspace_id
		AND membership.user_id = $3
	INNER JOIN team_members team_membership ON
		team_membership.team_id = story.team_id
		AND team_membership.user_id = $3
	WHERE story.workspace_id = $1
		AND story.id = $2
		AND story.auto_scheduling_enabled = TRUE
		AND story.deleted_at IS NULL
		AND story.archived_at IS NULL
		AND story.completed_at IS NULL
		AND status.category NOT IN ('completed', 'cancelled')
	FOR UPDATE OF story
	FOR SHARE OF status, selected_user, membership, team_membership
`

const mayaScheduleStoryVersionQuery = `
	SELECT story.id
	FROM stories story
	WHERE story.workspace_id = $1
		AND story.id = $2
		AND story.updated_at = $3
	FOR UPDATE OF story
`

func (r *Repo) WithScheduleEventDispatchLock(ctx context.Context, userID uuid.UUID, dispatch func(calendar.ScheduleEventOutboxStore) error) error {
	if userID == uuid.Nil || dispatch == nil {
		return calendar.ErrInvalidScheduleBlock
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin calendar schedule event dispatch: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, uuid.Nil, userID); err != nil {
		return err
	}
	if err := dispatch(transactionScheduleEventOutboxStore{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit calendar schedule event dispatch: %w", err)
	}
	return nil
}

func (r *Repo) ReconcileMayaScheduleBlocks(ctx context.Context, input calendar.MayaScheduleReconcileInput) (calendar.CoreScheduleReconcileResult, error) {
	segments := append([]calendar.MayaScheduleSegmentInput(nil), input.Segments...)
	sort.SliceStable(segments, func(i, j int) bool { return segments[i].SegmentIndex < segments[j].SegmentIndex })
	for index, segment := range segments {
		if segment.SegmentIndex != index || strings.TrimSpace(segment.Title) == "" || !segment.EndAt.After(segment.StartAt) {
			return calendar.CoreScheduleReconcileResult{}, calendar.ErrInvalidScheduleBlock
		}
		if index > 0 && segments[index-1].EndAt.After(segment.StartAt) {
			return calendar.CoreScheduleReconcileResult{}, calendar.ErrCalendarScheduleConflict
		}
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("begin Maya schedule reconciliation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, input.WorkspaceID, input.UserID); err != nil {
		return calendar.CoreScheduleReconcileResult{}, err
	}
	if input.ExpectedStoryUpdatedAt != nil {
		var currentStoryID uuid.UUID
		if err := tx.GetContext(ctx, &currentStoryID, mayaScheduleStoryVersionQuery, input.WorkspaceID, input.StoryID, input.ExpectedStoryUpdatedAt.UTC()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return calendar.CoreScheduleReconcileResult{}, calendar.ErrCalendarScheduleStalePlan
			}
			return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("lock Maya schedule story version: %w", err)
		}
	}
	if len(segments) > 0 {
		var eligibleStoryID uuid.UUID
		if err := tx.GetContext(ctx, &eligibleStoryID, mayaScheduleEligibilityQuery, input.WorkspaceID, input.StoryID, input.UserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return calendar.CoreScheduleReconcileResult{}, calendar.ErrInvalidScheduleBlock
			}
			return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("lock eligible Maya schedule story: %w", err)
		}
	}

	const currentQuery = `
		SELECT block_id, segment_index, title, start_at, end_at, is_locked,
		       external_provider, external_calendar_id, external_event_id, external_sync_hash,
		       manual_override_at, manual_override_by
		FROM calendar_schedule_blocks
		WHERE workspace_id = $1
			AND user_id = $2
			AND story_id = $3
			AND source = 'maya'
		FOR UPDATE
	`
	current := []reconciliationBlock{}
	if err := tx.SelectContext(ctx, &current, currentQuery, input.WorkspaceID, input.UserID, input.StoryID); err != nil {
		return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("list existing Maya schedule segments: %w", err)
	}
	currentByIndex := make(map[int]reconciliationBlock, len(current))
	manualByIndex := make(map[int]reconciliationBlock)
	excludedIDs := make([]string, 0, len(current))
	for _, block := range current {
		if block.ManualOverrideAt != nil {
			manualByIndex[block.SegmentIndex] = block
		} else {
			currentByIndex[block.SegmentIndex] = block
		}
		excludedIDs = append(excludedIDs, block.ID.String())
	}
	for _, blockID := range input.PreemptBlockIDs {
		if blockID != uuid.Nil {
			excludedIDs = append(excludedIDs, blockID.String())
		}
	}
	scheduleProvider, err := preferredScheduleProvider(ctx, tx, input.UserID, current)
	if err != nil {
		return calendar.CoreScheduleReconcileResult{}, err
	}

	const conflictQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM calendar_schedule_blocks block
			WHERE block.user_id = $1
				AND block.start_at < $3
				AND block.end_at > $2
				AND NOT (block.block_id = ANY(CAST($4 AS uuid[])))
			UNION ALL
			SELECT 1
			FROM calendar_busy_windows busy
			INNER JOIN calendar_connections connection ON
				connection.connection_id = busy.connection_id
					AND connection.user_id = busy.user_id
					AND connection.revoked_at IS NULL
					AND connection.cleanup_pending_at IS NULL
			WHERE busy.user_id = $1
				AND busy.start_at < $3
				AND busy.end_at > $2
		)
	`
	for _, segment := range segments {
		if input.AllowConflicts {
			continue
		}
		if _, exists := manualByIndex[segment.SegmentIndex]; exists {
			continue
		}
		if current, exists := currentByIndex[segment.SegmentIndex]; input.Locked && exists && current.StartAt.Equal(segment.StartAt) && current.EndAt.Equal(segment.EndAt) {
			// A locked segment is an explicit instruction to keep this exact
			// time. New busy data can make it conflicting, but must not make the
			// ownership/title watermark impossible to refresh. New or moved
			// locked segments still pass through the normal conflict check.
			continue
		}
		var conflicts bool
		if err := tx.GetContext(ctx, &conflicts, conflictQuery, input.UserID, segment.StartAt, segment.EndAt, pq.Array(excludedIDs)); err != nil {
			return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("validate Maya schedule segment: %w", err)
		}
		if conflicts {
			return calendar.CoreScheduleReconcileResult{}, calendar.ErrCalendarScheduleConflict
		}
	}

	result := calendar.CoreScheduleReconcileResult{
		Blocks:  make([]calendar.CoreScheduleBlock, 0, len(segments)),
		Actions: make([]calendar.ScheduleReconcileAction, 0, len(segments)+len(current)),
	}
	for _, segment := range segments {
		if manual, exists := manualByIndex[segment.SegmentIndex]; exists {
			result.Blocks = append(result.Blocks, toManualOverrideScheduleBlock(input, manual))
			result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUnchanged)
			delete(manualByIndex, segment.SegmentIndex)
			continue
		}
		block, exists := currentByIndex[segment.SegmentIndex]
		blockID := block.ID
		if !exists {
			blockID = uuid.New()
		}
		eventID := calendar.StableGoogleScheduleEventID(blockID)
		if scheduleProvider == calendar.ProviderMicrosoft {
			eventID = "pending:" + blockID.String()
			if exists && block.ExternalProvider != nil && *block.ExternalProvider == string(scheduleProvider) && block.ExternalEventID != nil && strings.TrimSpace(*block.ExternalEventID) != "" {
				eventID = strings.TrimSpace(*block.ExternalEventID)
			}
		}
		event := calendar.ExternalScheduleEventInput{
			CalendarID:  "primary",
			EventID:     eventID,
			BlockID:     blockID,
			StoryID:     input.StoryID,
			WorkspaceID: input.WorkspaceID,
			Title:       segment.Title,
			StartAt:     segment.StartAt.UTC(),
			EndAt:       segment.EndAt.UTC(),
			PrivateProperties: map[string]string{
				"fortyone_source":       "maya_schedule",
				"fortyone_block_id":     blockID.String(),
				"fortyone_story_id":     input.StoryID.String(),
				"fortyone_workspace_id": input.WorkspaceID.String(),
			},
		}
		syncHash := calendar.ScheduleEventSyncHash(event)
		providerChanged := !exists || block.Title != segment.Title || !block.StartAt.Equal(segment.StartAt) || !block.EndAt.Equal(segment.EndAt) ||
			block.ExternalProvider == nil || *block.ExternalProvider != string(scheduleProvider) ||
			block.ExternalCalendarID == nil || *block.ExternalCalendarID != "primary" ||
			block.ExternalEventID == nil || *block.ExternalEventID != eventID
		blockChanged := providerChanged || block.IsLocked != input.Locked

		if exists {
			const updateQuery = `
				UPDATE calendar_schedule_blocks
				SET title = CAST($5 AS text), start_at = $6, end_at = $7, is_locked = $9,
					external_provider = $10, external_calendar_id = 'primary', external_event_id = $8,
					updated_at = CASE WHEN title <> CAST($5 AS text) OR start_at <> $6 OR end_at <> $7 OR is_locked <> $9 THEN CURRENT_TIMESTAMP ELSE updated_at END
				WHERE workspace_id = $1 AND user_id = $2 AND story_id = $3 AND segment_index = $4 AND source = 'maya'
			`
			if _, err := tx.ExecContext(ctx, updateQuery, input.WorkspaceID, input.UserID, input.StoryID, segment.SegmentIndex, segment.Title, segment.StartAt, segment.EndAt, eventID, input.Locked, string(scheduleProvider)); err != nil {
				return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("update Maya schedule segment: %w", err)
			}
		} else {
			const insertQuery = `
				INSERT INTO calendar_schedule_blocks (
					block_id, workspace_id, user_id, story_id, block_type, title, start_at, end_at,
					is_locked, source, segment_index, external_provider, external_calendar_id, external_event_id
				) VALUES ($1, $2, $3, $4, 'work', $5, $6, $7, $10, 'maya', $8, $11, 'primary', $9)
			`
			if _, err := tx.ExecContext(ctx, insertQuery, blockID, input.WorkspaceID, input.UserID, input.StoryID, segment.Title, segment.StartAt, segment.EndAt, segment.SegmentIndex, eventID, input.Locked, string(scheduleProvider)); err != nil {
				return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("create Maya schedule segment: %w", err)
			}
		}
		if scheduleBlockNeedsProviderUpsert(block, exists, scheduleProvider, event, syncHash) {
			if err := enqueueScheduleEventOutbox(ctx, tx, input.WorkspaceID, input.UserID, &blockID, scheduleProvider, calendar.ScheduleEventOperationUpsert, event, syncHash, providerChanged); err != nil {
				return calendar.CoreScheduleReconcileResult{}, err
			}
		}
		if !exists {
			result.Actions = append(result.Actions, calendar.ScheduleReconcileActionCreated)
		} else if blockChanged {
			result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUpdated)
		} else {
			result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUnchanged)
		}
		storyID := input.StoryID
		provider := scheduleProvider
		calendarID := "primary"
		result.Blocks = append(result.Blocks, calendar.CoreScheduleBlock{
			ID:                 blockID,
			WorkspaceID:        input.WorkspaceID,
			UserID:             input.UserID,
			StoryID:            &storyID,
			BlockType:          calendar.ScheduleBlockTypeWork,
			Title:              segment.Title,
			StartAt:            segment.StartAt,
			EndAt:              segment.EndAt,
			IsLocked:           input.Locked,
			Source:             calendar.ScheduleBlockSourceMaya,
			SegmentIndex:       segment.SegmentIndex,
			ExternalProvider:   &provider,
			ExternalCalendarID: &calendarID,
			ExternalEventID:    &eventID,
			ExternalSyncHash:   block.ExternalSyncHash,
		})
		delete(currentByIndex, segment.SegmentIndex)
	}

	for _, block := range currentByIndex {
		provider := scheduleProvider
		if block.ExternalProvider != nil && strings.TrimSpace(*block.ExternalProvider) != "" {
			provider = calendar.Provider(*block.ExternalProvider)
		}
		eventID := calendar.StableGoogleScheduleEventID(block.ID)
		if block.ExternalEventID != nil && strings.TrimSpace(*block.ExternalEventID) != "" {
			eventID = *block.ExternalEventID
		}
		event := calendar.ExternalScheduleEventInput{CalendarID: "primary", EventID: eventID, BlockID: block.ID, StoryID: input.StoryID, WorkspaceID: input.WorkspaceID}
		if err := enqueueScheduleEventOutbox(ctx, tx, input.WorkspaceID, input.UserID, &block.ID, provider, calendar.ScheduleEventOperationDelete, event, "", true); err != nil {
			return calendar.CoreScheduleReconcileResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_schedule_blocks WHERE block_id = $1`, block.ID); err != nil {
			return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("delete obsolete Maya schedule segment: %w", err)
		}
		result.Actions = append(result.Actions, calendar.ScheduleReconcileActionDeleted)
	}
	for _, block := range manualByIndex {
		result.Blocks = append(result.Blocks, toManualOverrideScheduleBlock(input, block))
		result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUnchanged)
	}

	if len(segments) > 0 || input.KeepOwnership {
		const ownershipQuery = `
			INSERT INTO calendar_maya_schedule_ownerships (workspace_id, story_id, user_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (workspace_id, story_id) DO UPDATE SET
				user_id = EXCLUDED.user_id,
				updated_at = CURRENT_TIMESTAMP,
				recovery_attempted_at = NULL
		`
		if _, err := tx.ExecContext(ctx, ownershipQuery, input.WorkspaceID, input.StoryID, input.UserID); err != nil {
			return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("retain Maya schedule ownership: %w", err)
		}
	} else {
		const ownershipQuery = `
			DELETE FROM calendar_maya_schedule_ownerships
			WHERE workspace_id = $1 AND story_id = $2 AND user_id = $3
		`
		if _, err := tx.ExecContext(ctx, ownershipQuery, input.WorkspaceID, input.StoryID, input.UserID); err != nil {
			return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("release Maya schedule ownership: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("commit Maya schedule reconciliation: %w", err)
	}
	return result, nil
}

func toManualOverrideScheduleBlock(input calendar.MayaScheduleReconcileInput, block reconciliationBlock) calendar.CoreScheduleBlock {
	storyID := input.StoryID
	provider := calendar.ProviderGoogle
	if block.ExternalProvider != nil && strings.TrimSpace(*block.ExternalProvider) != "" {
		provider = calendar.Provider(*block.ExternalProvider)
	}
	calendarID := "primary"
	return calendar.CoreScheduleBlock{
		ID:                 block.ID,
		WorkspaceID:        input.WorkspaceID,
		UserID:             input.UserID,
		StoryID:            &storyID,
		BlockType:          calendar.ScheduleBlockTypeWork,
		Title:              block.Title,
		StartAt:            block.StartAt,
		EndAt:              block.EndAt,
		IsLocked:           true,
		Source:             calendar.ScheduleBlockSourceMaya,
		SegmentIndex:       block.SegmentIndex,
		ExternalProvider:   &provider,
		ExternalCalendarID: &calendarID,
		ExternalEventID:    block.ExternalEventID,
		ExternalSyncHash:   block.ExternalSyncHash,
		ManualOverrideAt:   block.ManualOverrideAt,
		ManualOverrideBy:   block.ManualOverrideBy,
	}
}

func preferredScheduleProvider(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, blocks []reconciliationBlock) (calendar.Provider, error) {
	providers := []string{}
	if err := tx.SelectContext(ctx, &providers, `
		SELECT provider
		FROM calendar_connections
		WHERE user_id = $1
			AND revoked_at IS NULL
			AND cleanup_pending_at IS NULL
			AND (
				(provider = 'google' AND $2 = ANY(scopes) AND $3 = ANY(scopes))
				OR (provider = 'microsoft' AND $4 = ANY(scopes))
			)
		ORDER BY CASE provider WHEN 'google' THEN 0 ELSE 1 END
	`, userID, calendar.GoogleCalendarEventsReadonlyScope, calendar.GoogleCalendarEventsOwnedScope, calendar.MicrosoftCalendarReadWriteScope); err != nil {
		return "", fmt.Errorf("list writable calendar providers: %w", err)
	}
	active := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		active[provider] = struct{}{}
	}
	for _, block := range blocks {
		if block.ExternalProvider == nil {
			continue
		}
		if _, ok := active[*block.ExternalProvider]; ok {
			return calendar.Provider(*block.ExternalProvider), nil
		}
	}
	if len(providers) > 0 {
		return calendar.Provider(providers[0]), nil
	}
	return calendar.ProviderGoogle, nil
}

func scheduleBlockNeedsProviderUpsert(block reconciliationBlock, exists bool, provider calendar.Provider, event calendar.ExternalScheduleEventInput, syncHash string) bool {
	if !exists || block.Title != event.Title || !block.StartAt.Equal(event.StartAt) || !block.EndAt.Equal(event.EndAt) {
		return true
	}
	if block.ExternalProvider == nil || *block.ExternalProvider != string(provider) ||
		block.ExternalCalendarID == nil || *block.ExternalCalendarID != event.CalendarID ||
		block.ExternalEventID == nil || *block.ExternalEventID != event.EventID {
		return true
	}
	return block.ExternalSyncHash == nil || *block.ExternalSyncHash != syncHash
}

func enqueueScheduleEventOutbox(ctx context.Context, executor sqlx.ExtContext, workspaceID, userID uuid.UUID, blockID *uuid.UUID, provider calendar.Provider, operation calendar.ScheduleEventOperation, event calendar.ExternalScheduleEventInput, syncHash string, reactivateTerminal bool) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode calendar schedule event outbox payload: %w", err)
	}
	dedupeKey := fmt.Sprintf("%s:%s:%s:%s", provider, operation, event.BlockID, syncHash)
	if blockID != nil {
		if _, err := executor.ExecContext(ctx, `
			UPDATE calendar_schedule_event_outbox
			SET processed_at = CURRENT_TIMESTAMP, last_error = 'Superseded by a newer schedule state.', updated_at = CURRENT_TIMESTAMP
			WHERE schedule_block_id = $1
				AND processed_at IS NULL
				AND dedupe_key <> $2
				AND (provider = $3 OR operation = 'upsert')
		`, *blockID, dedupeKey, string(provider)); err != nil {
			return fmt.Errorf("supersede stale calendar schedule event outbox: %w", err)
		}
	}
	const query = `
		INSERT INTO calendar_schedule_event_outbox (
			workspace_id, user_id, schedule_block_id, operation, provider, calendar_id,
			provider_event_id, payload, dedupe_key
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (dedupe_key) DO UPDATE SET
			payload = EXCLUDED.payload,
			processed_at = CASE
				WHEN $10 OR calendar_schedule_event_outbox.dead_lettered_at IS NULL THEN NULL
				ELSE calendar_schedule_event_outbox.processed_at
			END,
			dead_lettered_at = CASE
				WHEN $10 THEN NULL
				ELSE calendar_schedule_event_outbox.dead_lettered_at
			END,
			attempt_count = CASE
				WHEN $10 OR calendar_schedule_event_outbox.dead_lettered_at IS NULL THEN 0
				ELSE calendar_schedule_event_outbox.attempt_count
			END,
			last_error = CASE
				WHEN $10 OR calendar_schedule_event_outbox.dead_lettered_at IS NULL THEN NULL
				ELSE calendar_schedule_event_outbox.last_error
			END,
			available_at = CASE
				WHEN $10 OR calendar_schedule_event_outbox.dead_lettered_at IS NULL THEN CURRENT_TIMESTAMP
				ELSE calendar_schedule_event_outbox.available_at
			END,
			updated_at = CURRENT_TIMESTAMP
	`
	if _, err := executor.ExecContext(ctx, query, workspaceID, userID, blockID, string(operation), string(provider), event.CalendarID, event.EventID, payload, dedupeKey, reactivateTerminal); err != nil {
		return fmt.Errorf("enqueue calendar schedule event outbox: %w", err)
	}
	return nil
}

func (r *Repo) ListReadyScheduleEventOutboxUsers(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	const query = `
		SELECT outbox.user_id
		FROM calendar_schedule_event_outbox outbox
		INNER JOIN calendar_connections connection ON
			connection.user_id = outbox.user_id
			AND connection.provider = outbox.provider
			AND connection.revoked_at IS NULL
			AND (
				connection.cleanup_pending_at IS NOT NULL
				OR (
					(connection.provider = 'google' AND $2 = ANY(connection.scopes) AND $3 = ANY(connection.scopes))
					OR (connection.provider = 'microsoft' AND $4 = ANY(connection.scopes))
				)
			)
		WHERE outbox.processed_at IS NULL
			AND outbox.dead_lettered_at IS NULL
			AND outbox.available_at <= CURRENT_TIMESTAMP
		GROUP BY outbox.user_id
		ORDER BY MIN(outbox.available_at), outbox.user_id
		LIMIT $1
	`
	userIDs := []uuid.UUID{}
	if err := r.db.SelectContext(ctx, &userIDs, query, limit, calendar.GoogleCalendarEventsOwnedScope, calendar.GoogleCalendarEventsReadonlyScope, calendar.MicrosoftCalendarReadWriteScope); err != nil {
		return nil, fmt.Errorf("list users with ready calendar schedule events: %w", err)
	}
	return userIDs, nil
}

func (r *Repo) ListPendingScheduleEventOutbox(ctx context.Context, userID uuid.UUID, provider calendar.Provider, limit int) ([]calendar.CoreScheduleEventOutbox, error) {
	return listPendingScheduleEventOutbox(ctx, r.db, userID, provider, limit)
}

func (s transactionScheduleEventOutboxStore) ListPendingScheduleEventOutbox(ctx context.Context, userID uuid.UUID, provider calendar.Provider, limit int) ([]calendar.CoreScheduleEventOutbox, error) {
	return listPendingScheduleEventOutbox(ctx, s.tx, userID, provider, limit)
}

func (s transactionScheduleEventOutboxStore) ScheduleEventUpsertIsCurrent(ctx context.Context, item calendar.CoreScheduleEventOutbox, event calendar.ExternalScheduleEventInput) (bool, error) {
	if item.ScheduleBlockID == nil {
		return false, nil
	}
	var current bool
	if err := s.tx.GetContext(
		ctx,
		&current,
		scheduleEventUpsertIsCurrentQuery,
		*item.ScheduleBlockID,
		item.WorkspaceID,
		item.UserID,
		string(item.Provider),
		item.CalendarID,
		item.ProviderEventID,
		event.Title,
		event.StartAt,
		event.EndAt,
	); err != nil {
		return false, fmt.Errorf("validate current calendar schedule upsert: %w", err)
	}
	return current, nil
}

func listPendingScheduleEventOutbox(ctx context.Context, executor sqlx.ExtContext, userID uuid.UUID, provider calendar.Provider, limit int) ([]calendar.CoreScheduleEventOutbox, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	const query = `
		WITH ready AS (
			SELECT outbox_id
			FROM calendar_schedule_event_outbox
			WHERE user_id = $1
				AND provider = $2
				AND processed_at IS NULL
				AND dead_lettered_at IS NULL
				AND available_at <= CURRENT_TIMESTAMP
			ORDER BY created_at, outbox_id
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		UPDATE calendar_schedule_event_outbox outbox
		SET attempt_count = outbox.attempt_count + 1,
			available_at = CURRENT_TIMESTAMP + INTERVAL '5 minutes',
			updated_at = CURRENT_TIMESTAMP
		FROM ready
		WHERE outbox.outbox_id = ready.outbox_id
		RETURNING outbox.outbox_id, outbox.workspace_id, outbox.user_id, outbox.schedule_block_id,
		          outbox.operation, outbox.provider, outbox.calendar_id, outbox.provider_event_id,
		          outbox.payload, outbox.dedupe_key, outbox.attempt_count
	`
	rows := []dbScheduleEventOutbox{}
	if err := sqlx.SelectContext(ctx, executor, &rows, query, userID, string(provider), limit); err != nil {
		return nil, fmt.Errorf("claim calendar schedule event outbox: %w", err)
	}
	items := make([]calendar.CoreScheduleEventOutbox, len(rows))
	for index, row := range rows {
		items[index] = calendar.CoreScheduleEventOutbox{
			ID: row.ID, WorkspaceID: row.WorkspaceID, UserID: row.UserID, ScheduleBlockID: row.ScheduleBlockID,
			Operation: calendar.ScheduleEventOperation(row.Operation), Provider: calendar.Provider(row.Provider),
			CalendarID: row.CalendarID, ProviderEventID: row.ProviderEventID, Payload: row.Payload,
			DedupeKey: row.DedupeKey, AttemptCount: row.AttemptCount,
		}
	}
	return items, nil
}

func (r *Repo) MarkScheduleEventOutboxProcessed(ctx context.Context, item calendar.CoreScheduleEventOutbox, syncHash string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin complete calendar schedule event outbox: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := markScheduleEventOutboxProcessed(ctx, tx, item, syncHash); err != nil {
		return err
	}
	return tx.Commit()
}

func (s transactionScheduleEventOutboxStore) MarkScheduleEventOutboxProcessed(ctx context.Context, item calendar.CoreScheduleEventOutbox, syncHash string) error {
	return markScheduleEventOutboxProcessed(ctx, s.tx, item, syncHash)
}

func markScheduleEventOutboxProcessed(ctx context.Context, executor sqlx.ExtContext, item calendar.CoreScheduleEventOutbox, syncHash string) error {
	if _, err := executor.ExecContext(ctx, `
		UPDATE calendar_schedule_event_outbox
		SET processed_at = CURRENT_TIMESTAMP, last_error = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE outbox_id = $1
	`, item.ID); err != nil {
		return fmt.Errorf("complete calendar schedule event outbox: %w", err)
	}
	if item.Operation == calendar.ScheduleEventOperationUpsert && item.ScheduleBlockID != nil {
		if _, err := executor.ExecContext(ctx, markScheduleBlockMirroredQuery, *item.ScheduleBlockID, syncHash, item.ProviderEventID, string(item.Provider), item.CalendarID); err != nil {
			return fmt.Errorf("mark calendar schedule block mirrored: %w", err)
		}
	}
	return nil
}

func (r *Repo) MarkScheduleEventOutboxFailed(ctx context.Context, item calendar.CoreScheduleEventOutbox, message string, permanent bool) error {
	return markScheduleEventOutboxFailed(ctx, r.db, item, message, permanent)
}

func (s transactionScheduleEventOutboxStore) MarkScheduleEventOutboxFailed(ctx context.Context, item calendar.CoreScheduleEventOutbox, message string, permanent bool) error {
	return markScheduleEventOutboxFailed(ctx, s.tx, item, message, permanent)
}

func markScheduleEventOutboxFailed(ctx context.Context, executor sqlx.ExtContext, item calendar.CoreScheduleEventOutbox, message string, permanent bool) error {
	_, err := executor.ExecContext(ctx, `
		UPDATE calendar_schedule_event_outbox
		SET last_error = $2,
			dead_lettered_at = CASE
				WHEN $3 OR attempt_count >= 8 THEN CURRENT_TIMESTAMP
				ELSE NULL
			END,
			available_at = CASE
				WHEN $3 OR attempt_count >= 8 THEN available_at
				ELSE CURRENT_TIMESTAMP + CASE
					WHEN attempt_count <= 1 THEN INTERVAL '1 minute'
					WHEN attempt_count = 2 THEN INTERVAL '2 minutes'
					WHEN attempt_count = 3 THEN INTERVAL '4 minutes'
					WHEN attempt_count = 4 THEN INTERVAL '8 minutes'
					WHEN attempt_count = 5 THEN INTERVAL '16 minutes'
					WHEN attempt_count = 6 THEN INTERVAL '32 minutes'
					ELSE INTERVAL '1 hour'
				END
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE outbox_id = $1 AND processed_at IS NULL
	`, item.ID, message, permanent)
	if err != nil {
		return fmt.Errorf("mark calendar schedule event outbox failed: %w", err)
	}
	return nil
}

const releaseScheduleEventOutboxQuery = `
	UPDATE calendar_schedule_event_outbox
	SET attempt_count = GREATEST(attempt_count - 1, 0),
		available_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	WHERE outbox_id = ANY(CAST($1 AS uuid[])) AND processed_at IS NULL
`

func (s transactionScheduleEventOutboxStore) ReleaseScheduleEventOutbox(ctx context.Context, outboxIDs []uuid.UUID) error {
	if len(outboxIDs) == 0 {
		return nil
	}
	if _, err := s.tx.ExecContext(ctx, releaseScheduleEventOutboxQuery, pq.Array(outboxIDs)); err != nil {
		return fmt.Errorf("release unprocessed calendar schedule event outbox: %w", err)
	}
	return nil
}

func (s transactionScheduleEventOutboxStore) DeleteCleanupPendingConnectionIfDrained(ctx context.Context, userID uuid.UUID, provider calendar.Provider) error {
	if _, err := s.tx.ExecContext(ctx, `
		DELETE FROM calendar_connections connection
		WHERE connection.user_id = $1
			AND connection.provider = $2
			AND connection.cleanup_pending_at IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM calendar_schedule_event_outbox outbox
				WHERE outbox.user_id = connection.user_id
					AND outbox.provider = connection.provider
					AND outbox.processed_at IS NULL
					AND outbox.dead_lettered_at IS NULL
			)
	`, userID, string(provider)); err != nil {
		return fmt.Errorf("delete drained calendar cleanup connection: %w", err)
	}
	return nil
}
