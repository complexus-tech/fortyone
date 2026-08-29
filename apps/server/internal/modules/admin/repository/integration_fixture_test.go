//go:build integration

package adminrepository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type adminIntegrationFixture struct {
	postgres       *testkit.Postgres
	repository     *Repository
	now            time.Time
	actorA         uuid.UUID
	actorB         uuid.UUID
	inactiveAdmin  uuid.UUID
	externalUser   uuid.UUID
	targetUser     uuid.UUID
	workspaceID    uuid.UUID
	otherWorkspace uuid.UUID
}

func newAdminIntegrationFixture(t *testing.T) adminIntegrationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	assertAdminPostgres18(t, ctx, postgres.Pool)

	fixture := adminIntegrationFixture{
		postgres: postgres, repository: New(postgres.Pool),
		now:    time.Now().UTC().Truncate(time.Microsecond),
		actorA: uuid.New(), actorB: uuid.New(), inactiveAdmin: uuid.New(),
		externalUser: uuid.New(), targetUser: uuid.New(),
		workspaceID: uuid.New(), otherWorkspace: uuid.New(),
	}
	insertAdminTestUser(t, postgres.Pool, fixture.actorA, "admin-a", true, true, fixture.now.Add(-5*time.Hour))
	insertAdminTestUser(t, postgres.Pool, fixture.actorB, "admin-b", true, true, fixture.now.Add(-4*time.Hour))
	insertAdminTestUser(t, postgres.Pool, fixture.inactiveAdmin, "inactive", false, true, fixture.now.Add(-3*time.Hour))
	insertAdminTestUser(t, postgres.Pool, fixture.externalUser, "external", true, false, fixture.now.Add(-2*time.Hour))
	insertAdminTestUser(t, postgres.Pool, fixture.targetUser, "target", true, false, fixture.now.Add(-time.Hour))
	insertAdminTestWorkspace(t, postgres.Pool, fixture.workspaceID, fixture.actorA, "admin-main", fixture.now.Add(-2*time.Hour))
	insertAdminTestWorkspace(t, postgres.Pool, fixture.otherWorkspace, fixture.actorA, "admin-other", fixture.now.Add(-time.Hour))

	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES (CAST($1 AS uuid), CAST($2 AS uuid), 'member')
	`, fixture.workspaceID, fixture.targetUser)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO workspace_subscriptions (
			workspace_id, stripe_customer_id, stripe_subscription_id,
			subscription_status, seat_count, subscription_tier, updated_at
		) VALUES (
			CAST($1 AS uuid), CAST($2 AS text), CAST($3 AS text),
			'past_due', 7, 'pro', CAST($4 AS timestamptz)
		)
	`, fixture.workspaceID, "cus_admin", "sub_admin", fixture.now)
	require.NoError(t, err)
	insertAdminTestIntegrations(t, postgres.Pool, fixture)
	return fixture
}

func insertAdminTestUser(
	t *testing.T,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	name string,
	active, internal bool,
	createdAt time.Time,
) {
	t.Helper()
	_, err := pool.Exec(t.Context(), `
		INSERT INTO users (
			user_id, username, email, full_name, is_active, is_internal, created_at, updated_at
		) VALUES (
			CAST($1 AS uuid), CAST($2 AS text), CAST($3 AS text), CAST($4 AS text),
			CAST($5 AS boolean), CAST($6 AS boolean), CAST($7 AS timestamptz), CAST($7 AS timestamptz)
		)
	`, userID, name, name+"@example.test", strings.ToUpper(name), active, internal, createdAt)
	require.NoError(t, err)
}

func insertAdminTestWorkspace(
	t *testing.T,
	pool *pgxpool.Pool,
	workspaceID, creatorID uuid.UUID,
	slug string,
	createdAt time.Time,
) {
	t.Helper()
	_, err := pool.Exec(t.Context(), `
		INSERT INTO workspaces (
			workspace_id, name, slug, created_by, trial_ends_on, created_at, updated_at
		) VALUES (
			CAST($1 AS uuid), CAST($2 AS text), CAST($2 AS text), CAST($3 AS uuid),
			CAST($4 AS timestamptz), CAST($5 AS timestamptz), CAST($5 AS timestamptz)
		)
	`, workspaceID, slug, creatorID, createdAt.Add(7*24*time.Hour), createdAt)
	require.NoError(t, err)
}

func insertAdminTestIntegrations(
	t *testing.T,
	pool *pgxpool.Pool,
	fixture adminIntegrationFixture,
) {
	t.Helper()
	ctx := t.Context()
	_, err := pool.Exec(ctx, `
		INSERT INTO slack_workspaces (
			workspace_id, slack_team_id, slack_team_name, slack_team_domain,
			bot_access_token, installed_by_user_id, is_active
		) VALUES (
			CAST($1 AS uuid), 'T_ADMIN', 'Admin Test', 'admin-test',
			'test-token', CAST($2 AS uuid), TRUE
		)
	`, fixture.workspaceID, fixture.actorA)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO github_installations (
			workspace_id, github_app_id, github_installation_id, account_id,
			account_login, account_type, repository_selection, installed_by_user_id, is_active
		) VALUES (
			CAST($1 AS uuid), 1, 710001, 810001, 'admin-test', 'Organization',
			'all', CAST($2 AS uuid), TRUE
		)
	`, fixture.workspaceID, fixture.actorA)
	require.NoError(t, err)
}

func assertAdminPostgres18(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var version int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT CAST(current_setting('server_version_num') AS integer)",
	).Scan(&version))
	require.GreaterOrEqual(t, version, 180000, "admin persistence contract requires PostgreSQL 18")
}
