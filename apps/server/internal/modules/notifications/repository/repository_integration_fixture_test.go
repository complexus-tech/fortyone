//go:build integration

package notificationsrepository

import (
	"context"
	"fmt"
	"testing"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationIntegrationFixture struct {
	postgres *testkit.Postgres
	repo     *Repository

	workspaceA uuid.UUID
	workspaceB uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	storyA     uuid.UUID
	storyB     uuid.UUID
	actorA     uuid.UUID
	actorB     uuid.UUID
	recipientA uuid.UUID
	recipientB uuid.UUID
	guestA     uuid.UUID
	revocableA uuid.UUID
	inactiveA  uuid.UUID
	outsiderA  uuid.UUID
	system     uuid.UUID
	portalA    uuid.UUID
	portalB    uuid.UUID
	boardA     uuid.UUID
	boardB     uuid.UUID
	feedbackA  uuid.UUID
	feedbackB  uuid.UUID
	slugA      string
	slugB      string
}

func newNotificationIntegrationFixture(t *testing.T, ctx context.Context) notificationIntegrationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := notificationIntegrationFixture{
		postgres:   postgres,
		repo:       New(postgres.Pool),
		workspaceA: uuid.New(), workspaceB: uuid.New(),
		teamA: uuid.New(), teamB: uuid.New(), storyA: uuid.New(), storyB: uuid.New(),
		actorA: uuid.New(), actorB: uuid.New(), recipientA: uuid.New(), recipientB: uuid.New(),
		guestA: uuid.New(), revocableA: uuid.New(), inactiveA: uuid.New(), outsiderA: uuid.New(), system: uuid.New(),
		portalA: uuid.New(), portalB: uuid.New(), boardA: uuid.New(), boardB: uuid.New(),
		feedbackA: uuid.New(), feedbackB: uuid.New(),
		slugA: "notifications-a-" + uuid.NewString(),
		slugB: "notifications-b-" + uuid.NewString(),
	}

	insertNotificationUser(t, ctx, postgres.Pool, fixture.actorA, "actor-a", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.actorB, "actor-b", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.recipientA, "recipient-a", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.recipientB, "recipient-b", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.guestA, "guest-a", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.revocableA, "revocable-a", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.inactiveA, "inactive-a", false, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.outsiderA, "outsider-a", true, false)
	insertNotificationUser(t, ctx, postgres.Pool, fixture.system, "system", true, true)

	insertNotificationWorkspace(t, ctx, postgres.Pool, fixture.workspaceA, fixture.actorA, fixture.slugA, "Notifications A")
	insertNotificationWorkspace(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, fixture.slugB, "Notifications B")
	insertNotificationTeam(t, ctx, postgres.Pool, fixture.teamA, fixture.workspaceA, "NTA")
	insertNotificationTeam(t, ctx, postgres.Pool, fixture.teamB, fixture.workspaceB, "NTB")
	insertNotificationStory(t, ctx, postgres.Pool, fixture.storyA, fixture.teamA, fixture.workspaceA, "Notification story A")
	insertNotificationStory(t, ctx, postgres.Pool, fixture.storyB, fixture.teamB, fixture.workspaceB, "Notification story B")

	for _, userID := range []uuid.UUID{fixture.actorA, fixture.recipientA, fixture.revocableA, fixture.inactiveA} {
		insertNotificationWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, userID, "member")
		insertNotificationTeamMember(t, ctx, postgres.Pool, fixture.teamA, userID)
	}
	insertNotificationWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.guestA, "guest")
	insertNotificationTeamMember(t, ctx, postgres.Pool, fixture.teamA, fixture.guestA)
	for _, userID := range []uuid.UUID{fixture.actorB, fixture.recipientB} {
		insertNotificationWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceB, userID, "member")
		insertNotificationTeamMember(t, ctx, postgres.Pool, fixture.teamB, userID)
	}

	insertNotificationPortal(t, ctx, postgres.Pool, fixture.portalA, fixture.workspaceA)
	insertNotificationPortal(t, ctx, postgres.Pool, fixture.portalB, fixture.workspaceB)
	insertNotificationBoard(t, ctx, postgres.Pool, fixture.boardA, fixture.workspaceA, fixture.portalA, fixture.teamA, "a")
	insertNotificationBoard(t, ctx, postgres.Pool, fixture.boardB, fixture.workspaceB, fixture.portalB, fixture.teamB, "b")
	contributorA := insertNotificationContributor(t, ctx, postgres.Pool, fixture.portalA, fixture.recipientA)
	contributorB := insertNotificationContributor(t, ctx, postgres.Pool, fixture.portalB, fixture.recipientB)
	insertNotificationFeedback(t, ctx, postgres.Pool, fixture.feedbackA, fixture.workspaceA, fixture.portalA, fixture.boardA, contributorA, "feedback-a")
	insertNotificationFeedback(t, ctx, postgres.Pool, fixture.feedbackB, fixture.workspaceB, fixture.portalB, fixture.boardB, contributorB, "feedback-b")

	return fixture
}

func (fixture notificationIntegrationFixture) accessA(actorID uuid.UUID) notificationsdomain.WorkspaceAccess {
	return notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: fixture.workspaceA}
}

func (fixture notificationIntegrationFixture) storyNotification(recipientID uuid.UUID, dedupeKey string) notificationsdomain.NewNotification {
	return notificationsdomain.NewNotification{
		DedupeKey: dedupeKey, RecipientID: recipientID, WorkspaceID: fixture.workspaceA,
		Type: notificationsdomain.NotificationTypeStoryUpdate, EntityType: notificationsdomain.EntityTypeStory,
		EntityID: fixture.storyA, ActorID: fixture.actorA, Title: "Notification story A",
		Message: notificationsdomain.NotificationMessage{
			Template: "{actor} updated the story",
			Variables: map[string]notificationsdomain.Variable{
				"actor": {Value: "Actor A", Type: "actor"},
			},
		},
	}
}

func (fixture notificationIntegrationFixture) feedbackNotification(recipientID uuid.UUID, dedupeKey string) notificationsdomain.NewNotification {
	return notificationsdomain.NewNotification{
		DedupeKey: dedupeKey, RecipientID: recipientID, WorkspaceID: fixture.workspaceA,
		Type: notificationsdomain.NotificationTypeFeedbackComment, EntityType: notificationsdomain.EntityTypeFeedback,
		EntityID: fixture.feedbackA, ActorID: fixture.system, Title: "Feedback A",
		Message: notificationsdomain.NotificationMessage{
			Template:  "Someone replied to your feedback",
			Variables: map[string]notificationsdomain.Variable{},
		},
	}
}

func insertNotificationUser(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id uuid.UUID,
	label string,
	active, system bool,
) {
	t.Helper()
	suffix := uuid.NewString()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Notification "+label, active, system)
}

func insertNotificationWorkspace(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, creatorID uuid.UUID,
	slug, name string,
) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, id, name, slug, creatorID)
}

func insertNotificationTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Notification "+code, workspaceID, code)
}

func insertNotificationWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, userID uuid.UUID,
	role string,
) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, userID, role)
}

func insertNotificationTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID)
}

func insertNotificationStory(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, teamID, workspaceID uuid.UUID,
	title string,
) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO stories (id, team_id, title, workspace_id)
		VALUES ($1, $2, $3, $4)
	`, id, teamID, title, workspaceID)
}

func insertNotificationPortal(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO feedback_portals (id, workspace_id, is_public)
		VALUES ($1, $2, TRUE)
	`, id, workspaceID)
}

func insertNotificationBoard(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, workspaceID, portalID, teamID uuid.UUID,
	suffix string,
) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO feedback_boards (id, workspace_id, portal_id, team_id, name, slug)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, workspaceID, portalID, teamID, "Notification board "+suffix, "notification-board-"+suffix)
}

func insertNotificationContributor(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	portalID, userID uuid.UUID,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO feedback_contributors (id, portal_id, user_id, kind)
		VALUES ($1, $2, $3, 'account')
	`, id, portalID, userID)
	return id
}

func insertNotificationFeedback(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, workspaceID, portalID, boardID, contributorID uuid.UUID,
	slug string,
) {
	t.Helper()
	mustNotificationExec(t, ctx, pool, `
		INSERT INTO feedback_items (
			id, workspace_id, portal_id, board_id, contributor_id, title, slug
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, workspaceID, portalID, boardID, contributorID, "Feedback "+slug, slug)
}

func mustNotificationExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, statement string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, arguments...); err != nil {
		t.Fatalf("execute notification fixture SQL: %v", err)
	}
}

func notificationDedupeKey(prefix string) string {
	return fmt.Sprintf("%s:%s", prefix, uuid.NewString())
}
