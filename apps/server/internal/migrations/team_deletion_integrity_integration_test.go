//go:build integration

package migrations_test

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/internal/migrations"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const teamDeletionIntegrityMigration = "000189_team_deletion_integrity"

func TestTeamDeletionIntegrityRepairsStatusOwnershipAndMigratesDownAndUp(t *testing.T) {
	postgres := testkit.NewPostgresAtMigration(t, 188)
	ctx := t.Context()
	fixture := seedTeamDeletionIntegrityFixture(t, postgres.Pool)
	defaultStatus, equivalentStatus, alternateEquivalentStatus := uuid.New(), uuid.New(), uuid.New()
	for index, status := range []struct {
		id         uuid.UUID
		category   string
		defaultVal bool
	}{
		{defaultStatus, "backlog", true},
		{equivalentStatus, "started", false},
		{alternateEquivalentStatus, "started", false},
	} {
		teamDeletionExec(t, ctx, postgres.Pool, `
			INSERT INTO statuses (status_id, workspace_id, team_id, name, category, is_default, order_index)
			VALUES ($1, $2, $3, 'Owning status', $4, $5, $6)
		`, status.id, fixture.workspaceID, fixture.survivingTeamID, status.category, status.defaultVal, index)
	}
	foreignFallbackStatus := uuid.New()
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO statuses (status_id, workspace_id, team_id, name, category)
		VALUES ($1, $2, $3, 'Other category', 'unstarted')
	`, foreignFallbackStatus, fixture.workspaceID, fixture.teamID)
	emptyTeam := uuid.New()
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'No statuses', 'EMPTY', 'blue')
	`, emptyTeam, fixture.workspaceID)
	equivalentStory, fallbackStory, noStatusesStory, nullStatusStory, validStory := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, story := range []struct {
		id     uuid.UUID
		team   uuid.UUID
		status *uuid.UUID
	}{
		{equivalentStory, fixture.survivingTeamID, &fixture.statusID},
		{fallbackStory, fixture.survivingTeamID, &foreignFallbackStatus},
		{noStatusesStory, emptyTeam, &fixture.statusID},
		{nullStatusStory, fixture.survivingTeamID, nil},
		{validStory, fixture.survivingTeamID, &defaultStatus},
	} {
		teamDeletionExec(t, ctx, postgres.Pool, `
			INSERT INTO stories (id, workspace_id, team_id, title, status_id)
			VALUES ($1, $2, $3, 'Repair fixture', $4)
		`, story.id, fixture.workspaceID, story.team, story.status)
	}
	teamDeletionExec(t, ctx, postgres.Pool, "UPDATE stories SET deleted_at = now() WHERE id = $1", equivalentStory)
	teamDeletionExec(t, ctx, postgres.Pool, "UPDATE stories SET archived_at = now() WHERE id = $1", fallbackStory)

	applyTeamDeletionIntegrityMigration(t, postgres.Pool, ".up.sql")
	for _, expected := range []struct {
		story  uuid.UUID
		status *uuid.UUID
	}{
		{equivalentStory, &equivalentStatus},
		{fallbackStory, &defaultStatus},
		{noStatusesStory, nil},
		{nullStatusStory, nil},
		{validStory, &defaultStatus},
	} {
		var status *uuid.UUID
		require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT status_id FROM stories WHERE id = $1", expected.story).Scan(&status))
		require.Equal(t, expected.status, status)
	}

	_, err := postgres.Pool.Exec(ctx, "UPDATE stories SET status_id = $1 WHERE id = $2", fixture.statusID, validStory)
	var constraintErr *pgconn.PgError
	require.ErrorAs(t, err, &constraintErr)
	require.Equal(t, "stories_team_status_fkey", constraintErr.ConstraintName)
	_, err = postgres.Pool.Exec(ctx, "UPDATE stories SET team_id = $1 WHERE id = $2", fixture.teamID, validStory)
	require.ErrorAs(t, err, &constraintErr)
	require.Equal(t, "stories_team_status_fkey", constraintErr.ConstraintName)

	// A corrected source duplicates safely; team transfers must change the
	// status in the same statement, while an explicit NULL remains supported.
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO stories (id, workspace_id, team_id, title, status_id)
		SELECT $1, workspace_id, team_id, 'Duplicate', status_id FROM stories WHERE id = $2
	`, uuid.New(), equivalentStory)
	teamDeletionExec(t, ctx, postgres.Pool, "UPDATE stories SET team_id = $1, status_id = $2 WHERE id = $3", fixture.teamID, fixture.statusID, validStory)
	teamDeletionExec(t, ctx, postgres.Pool, "UPDATE stories SET status_id = NULL WHERE id = $1", validStory)

	applyTeamDeletionIntegrityMigration(t, postgres.Pool, ".down.sql")
	var repairedStatus uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT status_id FROM stories WHERE id = $1", equivalentStory).Scan(&repairedStatus))
	require.Equal(t, equivalentStatus, repairedStatus, "rollback must not restore corrupted status ownership")
	applyTeamDeletionIntegrityMigration(t, postgres.Pool, ".up.sql")
	teamDeletionExec(t, ctx, postgres.Pool, "DELETE FROM teams WHERE team_id = $1", fixture.teamID)
	assertTeamDeletionRowExists(t, postgres.Pool, "SELECT EXISTS (SELECT 1 FROM stories WHERE id = $1)", equivalentStory, true)
}

func TestTeamDeletionIntegrityCascadesPreserveOtherTeamsAndProviderCleanup(t *testing.T) {
	postgres := testkit.NewPostgresAtMigration(t, 189)
	ctx := t.Context()
	fixture := seedTeamDeletionIntegrityFixture(t, postgres.Pool)
	objectiveID, labelID, storyID, survivingStoryID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO objectives (objective_id, workspace_id, team_id, name, sequence_id)
		VALUES ($1, $2, $3, 'Delete with team', 1)
	`, objectiveID, fixture.workspaceID, fixture.teamID)
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO labels (label_id, workspace_id, team_id, name)
		VALUES ($1, $2, $3, 'Delete with team')
	`, labelID, fixture.workspaceID, fixture.teamID)
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO stories (id, workspace_id, team_id, title, status_id, objective_id)
		VALUES ($1, $2, $3, 'Deleted story', $4, $5), ($6, $2, $7, 'Surviving story', NULL, $5)
	`, storyID, fixture.workspaceID, fixture.teamID, fixture.statusID, objectiveID, survivingStoryID, fixture.survivingTeamID)
	teamDeletionExec(t, ctx, postgres.Pool, "INSERT INTO story_labels (story_id, label_id) VALUES ($1, $3), ($2, $3)", storyID, survivingStoryID, labelID)

	portalID, boardID, survivingBoardID := uuid.New(), uuid.New(), uuid.New()
	contributorID := uuid.New()
	targetItemID, sourceItemID, sameTeamSourceID := uuid.New(), uuid.New(), uuid.New()
	teamDeletionExec(t, ctx, postgres.Pool, "INSERT INTO feedback_portals (id, workspace_id) VALUES ($1, $2)", portalID, fixture.workspaceID)
	teamDeletionExec(t, ctx, postgres.Pool, "INSERT INTO feedback_contributors (id, portal_id, kind) VALUES ($1, $2, 'anonymous')", contributorID, portalID)
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO feedback_boards (id, workspace_id, portal_id, team_id, name, slug)
		VALUES ($1, $2, $3, $4, 'Deleted board', 'deleted'), ($5, $2, $3, $6, 'Surviving board', 'surviving')
	`, boardID, fixture.workspaceID, portalID, fixture.teamID, survivingBoardID, fixture.survivingTeamID)
	for _, item := range []struct{ id, board uuid.UUID }{
		{targetItemID, boardID}, {sourceItemID, survivingBoardID}, {sameTeamSourceID, boardID},
	} {
		teamDeletionExec(t, ctx, postgres.Pool, `
			INSERT INTO feedback_items (id, workspace_id, portal_id, board_id, title, description, slug, contributor_id)
			VALUES ($1, $2, $3, $4, 'Feedback', 'Preserve source content', $5, $6)
		`, item.id, fixture.workspaceID, portalID, item.board, item.id.String(), contributorID)
	}
	teamDeletionExec(t, ctx, postgres.Pool, `
		UPDATE feedback_items SET merged_into_item_id = $1, merged_at = now(), merged_by_user_id = $4
		WHERE id IN ($2, $3)
	`, targetItemID, sourceItemID, sameTeamSourceID, fixture.userID)
	mergeEventID := uuid.New()
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO feedback_item_merge_outbox (
			merge_event_id, source_item_id, target_item_id, workspace_id, portal_id,
			merged_by_user_id, merged_at, event_payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, now(), '{"original":"merge evidence"}')
	`, mergeEventID, sourceItemID, targetItemID, fixture.workspaceID, portalID, fixture.userID)

	googleBlockID, microsoftBlockID := uuid.New(), uuid.New()
	for index, block := range []struct {
		id       uuid.UUID
		provider string
	}{
		{googleBlockID, "google"}, {microsoftBlockID, "microsoft"},
	} {
		teamDeletionExec(t, ctx, postgres.Pool, `
			INSERT INTO calendar_schedule_blocks (
				block_id, workspace_id, user_id, story_id, block_type, title,
				start_at, end_at, source, segment_index,
				external_provider, external_calendar_id, external_event_id
			)
			VALUES ($1, $2, $3, $4, 'work', 'Scheduled story', now(), now() + interval '1 hour', 'maya', $5, $6, 'calendar', 'same-provider-event-id')
		`, block.id, fixture.workspaceID, fixture.userID, storyID, index, block.provider)
	}
	// An existing stale outbox row exercises the conflict path, including its
	// provider correction, delivery retry state and provider-scoped dedupe key.
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO calendar_schedule_event_outbox (
			workspace_id, user_id, schedule_block_id, operation, provider,
			calendar_id, provider_event_id, dedupe_key, processed_at, dead_lettered_at, attempt_count
		)
		VALUES ($1, $2, $3, 'delete', 'google', 'old-calendar', 'old-event', $4, now(), now(), 2)
	`, fixture.workspaceID, fixture.userID, microsoftBlockID, "microsoft:delete:"+microsoftBlockID.String()+":")

	// Every cascade, feedback detach and provider-outbox mutation must roll
	// back together when the surrounding team-deletion transaction rolls back.
	tx, err := postgres.Pool.Begin(ctx)
	require.NoError(t, err)
	teamDeletionExec(t, ctx, tx, "DELETE FROM teams WHERE team_id = $1", fixture.teamID)
	require.NoError(t, tx.Rollback(ctx))
	assertTeamDeletionRowExists(t, postgres.Pool, "SELECT EXISTS (SELECT 1 FROM teams WHERE team_id = $1)", fixture.teamID, true)
	var originalTarget uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT merged_into_item_id FROM feedback_items WHERE id = $1", sourceItemID).Scan(&originalTarget))
	require.Equal(t, targetItemID, originalTarget)
	var outboxCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT count(*) FROM calendar_schedule_event_outbox WHERE workspace_id = $1", fixture.workspaceID).Scan(&outboxCount))
	require.Equal(t, 1, outboxCount)

	teamDeletionExec(t, ctx, postgres.Pool, "DELETE FROM teams WHERE team_id = $1", fixture.teamID)
	for _, deleted := range []struct {
		query string
		id    uuid.UUID
	}{
		{"SELECT EXISTS (SELECT 1 FROM teams WHERE team_id = $1)", fixture.teamID},
		{"SELECT EXISTS (SELECT 1 FROM objectives WHERE objective_id = $1)", objectiveID},
		{"SELECT EXISTS (SELECT 1 FROM labels WHERE label_id = $1)", labelID},
		{"SELECT EXISTS (SELECT 1 FROM story_labels WHERE label_id = $1)", labelID},
		{"SELECT EXISTS (SELECT 1 FROM stories WHERE id = $1)", storyID},
		{"SELECT EXISTS (SELECT 1 FROM statuses WHERE status_id = $1)", fixture.statusID},
		{"SELECT EXISTS (SELECT 1 FROM feedback_boards WHERE id = $1)", boardID},
		{"SELECT EXISTS (SELECT 1 FROM feedback_items WHERE id = $1)", targetItemID},
		{"SELECT EXISTS (SELECT 1 FROM feedback_items WHERE id = $1)", sameTeamSourceID},
		{"SELECT EXISTS (SELECT 1 FROM calendar_schedule_blocks WHERE block_id = $1)", googleBlockID},
		{"SELECT EXISTS (SELECT 1 FROM calendar_schedule_blocks WHERE block_id = $1)", microsoftBlockID},
	} {
		assertTeamDeletionRowExists(t, postgres.Pool, deleted.query, deleted.id, false)
	}
	assertTeamDeletionRowExists(t, postgres.Pool, "SELECT EXISTS (SELECT 1 FROM stories WHERE id = $1 AND objective_id IS NULL)", survivingStoryID, true)
	assertTeamDeletionRowExists(t, postgres.Pool, "SELECT EXISTS (SELECT 1 FROM feedback_boards WHERE id = $1)", survivingBoardID, true)
	assertTeamDeletionRowExists(t, postgres.Pool, `
		SELECT EXISTS (SELECT 1 FROM feedback_items WHERE id = $1 AND status = 'closed'
			AND merged_into_item_id IS NULL AND merged_at IS NULL AND merged_by_user_id IS NULL
			AND description = 'Preserve source content')
	`, sourceItemID, true)
	assertTeamDeletionRowExists(t, postgres.Pool, `
		SELECT EXISTS (SELECT 1 FROM feedback_item_merge_outbox WHERE merge_event_id = $1
			AND event_payload = '{"original":"merge evidence"}' AND status = 'pending')
	`, mergeEventID, true)

	rows, err := postgres.Pool.Query(ctx, `
		SELECT provider, calendar_id, provider_event_id, dedupe_key, attempt_count,
			processed_at IS NULL AND dead_lettered_at IS NULL AS ready
		FROM calendar_schedule_event_outbox WHERE workspace_id = $1 ORDER BY provider
	`, fixture.workspaceID)
	require.NoError(t, err)
	defer rows.Close()
	for _, expected := range []struct {
		provider string
		blockID  uuid.UUID
	}{
		{"google", googleBlockID}, {"microsoft", microsoftBlockID},
	} {
		require.True(t, rows.Next())
		var provider, calendarID, eventID, dedupeKey string
		var attempts int
		var ready bool
		require.NoError(t, rows.Scan(&provider, &calendarID, &eventID, &dedupeKey, &attempts, &ready))
		require.Equal(t, expected.provider, provider)
		require.Equal(t, "calendar", calendarID)
		require.Equal(t, "same-provider-event-id", eventID)
		require.Equal(t, provider+":delete:"+expected.blockID.String()+":", dedupeKey)
		require.Zero(t, attempts)
		require.True(t, ready)
	}
	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func TestTeamDeletionIntegrityAllowsBulkFeedbackDeletion(t *testing.T) {
	postgres := testkit.NewPostgresAtMigration(t, 189)
	ctx := t.Context()
	fixture := seedTeamDeletionIntegrityFixture(t, postgres.Pool)
	portalID, boardID, targetID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contributorID := uuid.New()
	teamDeletionExec(t, ctx, postgres.Pool, "INSERT INTO feedback_portals (id, workspace_id) VALUES ($1, $2)", portalID, fixture.workspaceID)
	teamDeletionExec(t, ctx, postgres.Pool, "INSERT INTO feedback_contributors (id, portal_id, kind) VALUES ($1, $2, 'anonymous')", contributorID, portalID)
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO feedback_boards (id, workspace_id, portal_id, team_id, name, slug)
		VALUES ($1, $2, $3, $4, 'Board', 'board')
	`, boardID, fixture.workspaceID, portalID, fixture.teamID)
	teamDeletionExec(t, ctx, postgres.Pool, `
		INSERT INTO feedback_items (id, workspace_id, portal_id, board_id, title, slug, contributor_id)
		VALUES ($1, $2, $3, $4, 'Target', 'target', $6), ($5, $2, $3, $4, 'Source', 'source', $6)
	`, targetID, fixture.workspaceID, portalID, boardID, sourceID, contributorID)
	teamDeletionExec(t, ctx, postgres.Pool, "UPDATE feedback_items SET merged_into_item_id = $1, merged_at = now() WHERE id = $2", targetID, sourceID)
	teamDeletionExec(t, ctx, postgres.Pool, "DELETE FROM feedback_items WHERE board_id = $1", boardID)
	assertTeamDeletionRowExists(t, postgres.Pool, "SELECT EXISTS (SELECT 1 FROM feedback_items WHERE board_id = $1)", boardID, false)
}

type teamDeletionIntegrityFixture struct {
	workspaceID, teamID, survivingTeamID, statusID, userID uuid.UUID
}

func seedTeamDeletionIntegrityFixture(t *testing.T, pool *pgxpool.Pool) teamDeletionIntegrityFixture {
	t.Helper()
	ctx := t.Context()
	fixture := teamDeletionIntegrityFixture{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	teamDeletionExec(t, ctx, pool, "INSERT INTO users (user_id, username, email) VALUES ($1, $2, $3)", fixture.userID, fixture.userID.String(), fixture.userID.String()+"@example.com")
	teamDeletionExec(t, ctx, pool, "INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, 'Deletion integrity', $2)", fixture.workspaceID, fixture.workspaceID.String())
	teamDeletionExec(t, ctx, pool, "INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'admin')", fixture.workspaceID, fixture.userID)
	teamDeletionExec(t, ctx, pool, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Delete team', 'DEL', 'blue'), ($3, $2, 'Surviving team', 'SUR', 'green')
	`, fixture.teamID, fixture.workspaceID, fixture.survivingTeamID)
	teamDeletionExec(t, ctx, pool, `
		INSERT INTO statuses (status_id, workspace_id, team_id, name, category, is_default)
		VALUES ($1, $2, $3, 'In progress', 'started', true)
	`, fixture.statusID, fixture.workspaceID, fixture.teamID)
	return fixture
}

func applyTeamDeletionIntegrityMigration(t *testing.T, pool *pgxpool.Pool, direction string) {
	t.Helper()
	script, err := migrations.FS.ReadFile(teamDeletionIntegrityMigration + direction)
	require.NoError(t, err)
	teamDeletionExec(t, t.Context(), pool, string(script))
}

func teamDeletionExec(t *testing.T, ctx context.Context, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, query string, args ...any) {
	t.Helper()
	_, err := db.Exec(ctx, query, args...)
	require.NoError(t, err)
}

func assertTeamDeletionRowExists(t *testing.T, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, query string, id uuid.UUID, expected bool) {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRow(t.Context(), query, id).Scan(&exists))
	require.Equal(t, expected, exists, query)
}
