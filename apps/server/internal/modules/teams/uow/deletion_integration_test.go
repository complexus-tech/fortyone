//go:build integration

package teamuow

import (
	"context"
	"fmt"
	"testing"
	"time"

	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestTeamDeletionCommitsCompleteGraphAndDurableCleanup(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	fixture := seedTeamDeletionFixture(t, ctx, postgres.Pool, 505)
	manager := newTeamDeletionManager(t, postgres.Pool)
	command := fixture.command(t)
	require.NoError(t, manager.Delete(ctx, command))

	for _, table := range []string{"teams", "team_members", "statuses", "objectives", "key_results", "labels", "stories", "feedback_boards"} {
		assertTeamDeletionCount(t, ctx, postgres.Pool, "SELECT COUNT(*) FROM "+pgx.Identifier{table}.Sanitize()+" WHERE team_id = $1", 0, fixture.teamID)
	}
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM story_comments WHERE story_id = ANY(CAST($1 AS uuid[]))`, 0, fixture.storyIDs)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM feedback_items WHERE id = $1`, 0, fixture.feedbackID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM documents WHERE document_id = $1`, 1, fixture.documentID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM document_relationships WHERE document_id = $1`, 1, fixture.documentID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM document_relationships WHERE document_id = $1 AND entity_id = $2`, 1, fixture.documentID, fixture.otherStoryID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM notifications WHERE workspace_id = $1`, 2, fixture.workspaceID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM notifications WHERE entity_id = $1`, 1, fixture.otherStoryID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM notifications WHERE entity_id = $1`, 1, fixture.otherFeedbackID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM teams WHERE team_id = $1`, 1, fixture.otherTeamID)
	for _, table := range []string{"objectives", "labels", "statuses"} {
		assertTeamDeletionCount(t, ctx, postgres.Pool, "SELECT COUNT(*) FROM "+pgx.Identifier{table}.Sanitize()+" WHERE team_id = $1", 1, fixture.otherTeamID)
	}
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM stories WHERE id = $1`, 1, fixture.otherStoryID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM feedback_items WHERE id = $1`, 1, fixture.otherFeedbackID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM labels WHERE label_id = $1`, 1, fixture.workspaceLabelID)

	for _, attachmentID := range fixture.orphanIDs {
		assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM attachments WHERE attachment_id = $1`, 0, attachmentID)
		assertTeamDeletionCount(t, ctx, postgres.Pool, `
			SELECT COUNT(*) FROM attachment_object_deletion_outbox
			WHERE attachment_id = $1 AND workspace_id = $2
			  AND storage_provider = 'aws' AND container_name = 'team-deletion-test'
			  AND status = 'pending'
		`, 1, attachmentID, fixture.workspaceID)
	}
	for _, attachmentID := range fixture.sharedIDs {
		assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM attachments WHERE attachment_id = $1`, 1, attachmentID)
		assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM attachment_object_deletion_outbox WHERE attachment_id = $1`, 0, attachmentID)
	}
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM attachment_object_deletion_outbox`, len(fixture.orphanIDs))
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM story_mutation_events`, len(fixture.storyIDs))
	for _, storyID := range fixture.storyIDs {
		assertTeamDeletionCount(t, ctx, postgres.Pool, `
			SELECT COUNT(*) FROM story_mutation_events
			WHERE story_id = $1 AND workspace_id = $2 AND event_type = 'story.deleted'
			  AND actor_kind = 'human_user' AND actor_id = $3
			  AND status = 'pending' AND occurred_at = $4
		`, 1, storyID, fixture.workspaceID, fixture.adminID, command.DeletedAt)
	}
}

func TestTeamDeletionRollsBackGraphEventsAndOutboxOnLateFailure(t *testing.T) {
	for _, failure := range []string{"team_delete", "object_deletion_enqueue"} {
		t.Run(failure, func(t *testing.T) {
			postgres := testkit.NewPostgres(t)
			ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
			defer cancel()
			fixture := seedTeamDeletionFixture(t, ctx, postgres.Pool, 3)
			manager := newTeamDeletionManager(t, postgres.Pool)
			before := snapshotTeamDeletionGraph(t, ctx, postgres.Pool)
			mustTeamDeletionExec(t, ctx, postgres.Pool, `
				CREATE FUNCTION reject_team_deletion_test() RETURNS trigger LANGUAGE plpgsql AS $$
				BEGIN
				    IF NOT EXISTS (SELECT 1 FROM story_mutation_events WHERE event_type = 'story.deleted') THEN
				        RAISE EXCEPTION 'failure injection ran before durable story deletion';
				    END IF;
				    RAISE EXCEPTION 'forced late team deletion failure';
				END;
				$$;
			`)
			if failure == "team_delete" {
				mustTeamDeletionExec(t, ctx, postgres.Pool, `
					CREATE TRIGGER zz_reject_team_deletion_test AFTER DELETE ON teams
					FOR EACH ROW EXECUTE FUNCTION reject_team_deletion_test();
				`)
			} else {
				mustTeamDeletionExec(t, ctx, postgres.Pool, `
					CREATE TRIGGER reject_team_deletion_test AFTER INSERT ON attachment_object_deletion_outbox
					FOR EACH ROW EXECUTE FUNCTION reject_team_deletion_test();
				`)
			}
			err := manager.Delete(ctx, fixture.command(t))
			require.ErrorContains(t, err, "forced late team deletion failure")
			require.Equal(t, before, snapshotTeamDeletionGraph(t, ctx, postgres.Pool), "every database side effect must roll back")
		})
	}
}

func TestTeamDeletionRevalidatesAccessAndDeletesEmptyTeam(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := seedTeamDeletionFixture(t, ctx, postgres.Pool, 0)
	manager := newTeamDeletionManager(t, postgres.Pool)
	before := snapshotTeamDeletionGraph(t, ctx, postgres.Pool)
	command := fixture.command(t)
	for _, kind := range []platformauth.PrincipalKind{platformauth.PrincipalHumanUser, platformauth.PrincipalPersonalToken} {
		credentialID := uuid.Nil
		if kind == platformauth.PrincipalPersonalToken {
			credentialID = uuid.New()
		}
		readOnlyActor, err := platformauth.NewActor(fixture.adminID, kind, credentialID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), platformauth.UnrestrictedTeamAccess())
		require.NoError(t, err)
		readOnlyActor, err = readOnlyActor.WithWorkspace(fixture.workspaceID)
		require.NoError(t, err)
		command.Actor = readOnlyActor
		require.ErrorIs(t, manager.Delete(ctx, command), teamsdomain.ErrDeletionForbidden)
	}
	command = fixture.command(t)
	otherWorkspaceID := uuid.New()
	mustTeamDeletionExec(t, ctx, postgres.Pool, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, 'Foreign', $2)`, otherWorkspaceID, "foreign-"+uuid.NewString())
	mustTeamDeletionExec(t, ctx, postgres.Pool, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'admin')`, otherWorkspaceID, fixture.adminID)
	command.WorkspaceID = otherWorkspaceID
	actor, err := platformauth.NewHumanActor(fixture.adminID).WithWorkspace(otherWorkspaceID)
	require.NoError(t, err)
	command.Actor = actor
	require.ErrorIs(t, manager.Delete(ctx, command), teamsdomain.ErrNotFound)
	require.Equal(t, before, snapshotTeamDeletionGraph(t, ctx, postgres.Pool))

	command = fixture.command(t)
	mustTeamDeletionExec(t, ctx, postgres.Pool, `UPDATE workspace_members SET role = 'member' WHERE workspace_id = $1 AND user_id = $2`, fixture.workspaceID, fixture.adminID)
	require.ErrorIs(t, manager.Delete(ctx, command), teamsdomain.ErrNotFound)
	mustTeamDeletionExec(t, ctx, postgres.Pool, `UPDATE workspace_members SET role = 'admin' WHERE workspace_id = $1 AND user_id = $2`, fixture.workspaceID, fixture.adminID)
	mustTeamDeletionExec(t, ctx, postgres.Pool, `UPDATE users SET is_active = FALSE WHERE user_id = $1`, fixture.adminID)
	require.ErrorIs(t, manager.Delete(ctx, command), teamsdomain.ErrNotFound)
	mustTeamDeletionExec(t, ctx, postgres.Pool, `UPDATE users SET is_active = TRUE WHERE user_id = $1`, fixture.adminID)
	require.Equal(t, before, snapshotTeamDeletionGraph(t, ctx, postgres.Pool))

	require.NoError(t, manager.Delete(ctx, command))
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM teams WHERE team_id = $1`, 0, fixture.teamID)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM story_mutation_events`, 0)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM attachment_object_deletion_outbox`, 0)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT COUNT(*) FROM teams WHERE team_id = $1`, 1, fixture.otherTeamID)
}

type teamDeletionFixture struct {
	workspaceID, adminID, teamID, otherTeamID uuid.UUID
	storyIDs                                  []uuid.UUID
	otherStoryID                              uuid.UUID
	documentID                                uuid.UUID
	feedbackID, otherFeedbackID               uuid.UUID
	workspaceLabelID                          uuid.UUID
	orphanIDs, sharedIDs                      []uuid.UUID
}

func (fixture teamDeletionFixture) command(t *testing.T) teamsdomain.Deletion {
	t.Helper()
	actor, err := platformauth.NewHumanActor(fixture.adminID).WithWorkspace(fixture.workspaceID)
	require.NoError(t, err)
	return teamsdomain.Deletion{TeamID: fixture.teamID, WorkspaceID: fixture.workspaceID, Actor: actor, DeletedAt: time.Now().UTC().Truncate(time.Microsecond)}
}

func newTeamDeletionManager(t *testing.T, pool *pgxpool.Pool) *Manager {
	t.Helper()
	manager, err := New(pool, teamsrepository.New(pool), storiesrepository.NewMutationRepository(nil, pool), attachmentsrepository.New(pool), "aws", "team-deletion-test")
	require.NoError(t, err)
	return manager
}

func seedTeamDeletionFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, storyCount int) teamDeletionFixture {
	t.Helper()
	fixture := teamDeletionFixture{workspaceID: uuid.New(), adminID: uuid.New(), teamID: uuid.New(), otherTeamID: uuid.New()}
	mustTeamDeletionExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, 'Team deletion admin', TRUE)
	`, fixture.adminID, "deletion-"+fixture.adminID.String(), fixture.adminID.String()+"@example.test")
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, 'Team deletion workspace', $2)`, fixture.workspaceID, "deletion-"+fixture.workspaceID.String())
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'admin')`, fixture.workspaceID, fixture.adminID)
	for _, teamID := range []uuid.UUID{fixture.teamID, fixture.otherTeamID} {
		mustTeamDeletionExec(t, ctx, pool, `INSERT INTO teams (team_id, workspace_id, name, code, color) VALUES ($1, $2, $3, $4, '#123456')`, teamID, fixture.workspaceID, teamID.String(), teamID.String()[:8])
		mustTeamDeletionExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, fixture.adminID)
	}
	if storyCount == 0 {
		return fixture
	}

	statusID, objectiveID, keyResultID, labelID, commentID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	otherObjectiveID, otherStatusID := uuid.New(), uuid.New()
	fixture.workspaceLabelID = uuid.New()
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO statuses (status_id, workspace_id, team_id, name, category, is_default) VALUES ($1, $2, $3, 'Backlog', 'backlog', TRUE)`, statusID, fixture.workspaceID, fixture.teamID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO statuses (status_id, workspace_id, team_id, name, category, is_default) VALUES ($1, $2, $3, 'Backlog', 'backlog', TRUE)`, otherStatusID, fixture.workspaceID, fixture.otherTeamID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO objectives (objective_id, workspace_id, team_id, sequence_id, name) VALUES ($1, $2, $3, 1, 'Team objective'), ($4, $2, $5, 1, 'Retained objective')`, objectiveID, fixture.workspaceID, fixture.teamID, otherObjectiveID, fixture.otherTeamID)
	mustTeamDeletionExec(t, ctx, pool, `
		INSERT INTO key_results (id, objective_id, team_id, sequence_id, name, measurement_type, start_date, end_date)
		VALUES ($1, $2, $3, 1, 'Team result', 'number', CURRENT_DATE, CURRENT_DATE + 10)
	`, keyResultID, objectiveID, fixture.teamID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO labels (label_id, workspace_id, team_id, name) VALUES ($1, $2, $3, 'Team label'), ($4, $2, NULL, 'Workspace label')`, labelID, fixture.workspaceID, fixture.teamID, fixture.workspaceLabelID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO labels (workspace_id, team_id, name) VALUES ($1, $2, 'Retained label')`, fixture.workspaceID, fixture.otherTeamID)
	rows, err := pool.Query(ctx, `
		INSERT INTO stories (workspace_id, team_id, status_id, objective_id, key_result_id, title, sequence_id, deleted_at, archived_at)
		SELECT $1, $2, $3, $4, $5, 'Team story ' || CAST(n AS text), n,
		       CASE WHEN n = 1 THEN CURRENT_TIMESTAMP END,
		       CASE WHEN n = 2 THEN CURRENT_TIMESTAMP END
		FROM generate_series(1, CAST($6 AS integer)) AS n
		RETURNING id
	`, fixture.workspaceID, fixture.teamID, statusID, objectiveID, keyResultID, storyCount)
	require.NoError(t, err)
	fixture.storyIDs, err = pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	require.NoError(t, err)
	storyID := fixture.storyIDs[len(fixture.storyIDs)-1]
	fixture.otherStoryID = uuid.New()
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO stories (id, workspace_id, team_id, objective_id, status_id, title) VALUES ($1, $2, $3, $4, $5, 'Preserved story')`, fixture.otherStoryID, fixture.workspaceID, fixture.otherTeamID, otherObjectiveID, otherStatusID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO story_labels (story_id, label_id) VALUES ($1, $2), ($1, $3)`, storyID, labelID, fixture.workspaceLabelID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO story_comments (comment_id, story_id, commenter_id, content) VALUES ($1, $2, $3, 'Comment')`, commentID, storyID, fixture.adminID)

	fixture.documentID = uuid.New()
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO documents (document_id, workspace_id, title, created_by, updated_by) VALUES ($1, $2, 'Retained workspace document', $3, $3)`, fixture.documentID, fixture.workspaceID, fixture.adminID)
	for _, entity := range []struct {
		kind string
		id   uuid.UUID
	}{{"story", storyID}, {"objective", objectiveID}, {"story", fixture.otherStoryID}} {
		mustTeamDeletionExec(t, ctx, pool, `INSERT INTO document_relationships (document_id, workspace_id, entity_type, entity_id, created_by) VALUES ($1, $2, $3, $4, $5)`, fixture.documentID, fixture.workspaceID, entity.kind, entity.id, fixture.adminID)
	}
	for _, entity := range []struct {
		kind string
		id   uuid.UUID
	}{{"story", storyID}, {"comment", commentID}, {"objective", objectiveID}, {"key_result", keyResultID}, {"story", fixture.otherStoryID}} {
		mustTeamDeletionExec(t, ctx, pool, `
			INSERT INTO notifications (recipient_id, workspace_id, type, entity_type, entity_id, actor_id, title, message)
			VALUES ($1, $2, 'mention', CAST($3 AS entity_type), $4, $1, 'Entity notification', '{}')
		`, fixture.adminID, fixture.workspaceID, entity.kind, entity.id)
	}
	portalID, contributorID := uuid.New(), uuid.New()
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO feedback_portals (id, workspace_id) VALUES ($1, $2)`, portalID, fixture.workspaceID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO feedback_contributors (id, portal_id, kind) VALUES ($1, $2, 'anonymous')`, contributorID, portalID)
	for _, target := range []struct {
		teamID uuid.UUID
		itemID *uuid.UUID
	}{{fixture.teamID, &fixture.feedbackID}, {fixture.otherTeamID, &fixture.otherFeedbackID}} {
		boardID := uuid.New()
		*target.itemID = uuid.New()
		mustTeamDeletionExec(t, ctx, pool, `INSERT INTO feedback_boards (id, workspace_id, portal_id, team_id, name, slug) VALUES ($1, $2, $3, $4, 'Feedback board', $5)`, boardID, fixture.workspaceID, portalID, target.teamID, boardID.String())
		mustTeamDeletionExec(t, ctx, pool, `INSERT INTO feedback_items (id, workspace_id, portal_id, board_id, title, slug, contributor_id) VALUES ($1, $2, $3, $4, 'Feedback item', $5, $6)`, *target.itemID, fixture.workspaceID, portalID, boardID, target.itemID.String(), contributorID)
		mustTeamDeletionExec(t, ctx, pool, `
			INSERT INTO notifications (recipient_id, workspace_id, type, entity_type, entity_id, actor_id, title, message)
			VALUES ($1, $2, 'mention', 'feedback', $3, $1, 'Feedback notification', '{}')
		`, fixture.adminID, fixture.workspaceID, *target.itemID)
	}
	for range 3 {
		fixture.orphanIDs = append(fixture.orphanIDs, insertTeamDeletionAttachment(t, ctx, pool, fixture))
		fixture.sharedIDs = append(fixture.sharedIDs, insertTeamDeletionAttachment(t, ctx, pool, fixture))
	}
	// Generic, inline, and feedback-only orphans all need durable deletion.
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2), ($1, $3), ($1, $4)`, storyID, fixture.orphanIDs[0], fixture.sharedIDs[0], fixture.sharedIDs[1])
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO story_inline_attachments (story_id, attachment_id, created_by) VALUES ($1, $2, $3), ($1, $4, $3)`, storyID, fixture.orphanIDs[1], fixture.adminID, fixture.sharedIDs[2])
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO feedback_item_attachments (item_id, attachment_id) VALUES ($1, $2), ($1, $3)`, fixture.feedbackID, fixture.orphanIDs[2], fixture.sharedIDs[2])
	// Other-team stories, workspace documents, and other-team feedback retain media.
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)`, fixture.otherStoryID, fixture.sharedIDs[0])
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO document_attachments (document_id, attachment_id, created_by) VALUES ($1, $2, $3)`, fixture.documentID, fixture.sharedIDs[1], fixture.adminID)
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO feedback_item_attachments (item_id, attachment_id) VALUES ($1, $2)`, fixture.otherFeedbackID, fixture.sharedIDs[2])
	return fixture
}

func insertTeamDeletionAttachment(t *testing.T, ctx context.Context, pool *pgxpool.Pool, fixture teamDeletionFixture) uuid.UUID {
	t.Helper()
	id := uuid.New()
	mustTeamDeletionExec(t, ctx, pool, `INSERT INTO attachments (attachment_id, workspace_id, uploaded_by, filename, blob_name, size, mime_type) VALUES ($1, $2, $3, 'attachment.png', $4, 128, 'image/png')`, id, fixture.workspaceID, fixture.adminID, "team-deletion/"+id.String())
	return id
}

func mustTeamDeletionExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	_, err := pool.Exec(ctx, query, args...)
	require.NoError(t, err)
}

func assertTeamDeletionCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, expected int, args ...any) {
	t.Helper()
	var count int
	require.NoError(t, pool.QueryRow(ctx, query, args...).Scan(&count))
	require.Equal(t, expected, count, query)
}

func snapshotTeamDeletionGraph(t *testing.T, ctx context.Context, pool *pgxpool.Pool) map[string]string {
	t.Helper()
	result := make(map[string]string)
	for _, table := range []string{
		"teams", "team_members", "statuses", "objectives", "key_results", "labels", "story_labels", "stories", "story_comments",
		"attachments", "story_attachments", "story_inline_attachments", "feedback_boards", "feedback_items", "feedback_item_attachments",
		"documents", "document_relationships", "document_attachments", "notifications", "story_mutation_events", "attachment_object_deletion_outbox",
	} {
		query := fmt.Sprintf(`SELECT CAST(COALESCE(jsonb_agg(to_jsonb(entry) ORDER BY CAST(to_jsonb(entry) AS text)), '[]') AS text) FROM %s AS entry`, pgx.Identifier{table}.Sanitize())
		var state string
		require.NoError(t, pool.QueryRow(ctx, query).Scan(&state))
		result[table] = state
	}
	return result
}
