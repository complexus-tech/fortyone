//go:build integration

package developercredentialsrepository

import (
	"context"
	"fmt"
	"testing"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type credentialFixture struct {
	workspaceA   uuid.UUID
	workspaceB   uuid.UUID
	adminA       uuid.UUID
	adminB       uuid.UUID
	teamA        uuid.UUID
	teamB        uuid.UUID
	adminAccessA developercredentialsdomain.Access
	adminAccessB developercredentialsdomain.Access
}

func newCredentialFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) credentialFixture {
	t.Helper()
	workspaceA := insertCredentialWorkspace(t, ctx, pool, "a")
	workspaceB := insertCredentialWorkspace(t, ctx, pool, "b")
	adminA := insertCredentialUser(t, ctx, pool, "admin-a")
	adminB := insertCredentialUser(t, ctx, pool, "admin-b")
	insertCredentialWorkspaceMember(t, ctx, pool, workspaceA, adminA, "admin")
	insertCredentialWorkspaceMember(t, ctx, pool, workspaceB, adminB, "admin")
	teamA := insertCredentialTeam(t, ctx, pool, workspaceA, "A")
	teamB := insertCredentialTeam(t, ctx, pool, workspaceB, "B")
	insertCredentialTeamMember(t, ctx, pool, teamA, adminA)
	insertCredentialTeamMember(t, ctx, pool, teamB, adminB)
	return credentialFixture{
		workspaceA: workspaceA, workspaceB: workspaceB, adminA: adminA, adminB: adminB,
		teamA: teamA, teamB: teamB,
		adminAccessA: integrationHumanAccess(t, adminA, workspaceA),
		adminAccessB: integrationHumanAccess(t, adminB, workspaceB),
	}
}

func newIntegrationCredentialService(t *testing.T, pool *pgxpool.Pool, clock developercredentials.Clock) *developercredentials.Service {
	t.Helper()
	keyring, err := developercredentials.ParseEncodedTokenKeyring(
		"integration", 1,
		`{"integration@1":"aW50ZWdyYXRpb24tY3JlZGVudGlhbC1rZXktMDAxMjM="}`,
	)
	require.NoError(t, err)
	tokens, err := developercredentials.NewTokenManager(keyring)
	require.NoError(t, err)
	service, err := developercredentials.New(New(pool), tokens, clock, developercredentials.RandomIDGenerator{})
	require.NoError(t, err)
	return service
}

func integrationHumanAccess(t *testing.T, userID uuid.UUID, workspaceID uuid.UUID) developercredentialsdomain.Access {
	t.Helper()
	actor, err := platformauth.NewHumanActor(userID).WithWorkspace(workspaceID)
	require.NoError(t, err)
	return developercredentialsdomain.Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}
}

func insertCredentialWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`,
		id, "Credentials "+label, "credentials-"+label+"-"+uuid.NewString())
	require.NoError(t, err)
	return id
}

func insertCredentialUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, TRUE, FALSE)
	`, id, label+"-"+id.String(), fmt.Sprintf("%s-%s@example.com", label, id), "Credentials "+label)
	require.NoError(t, err)
	return id
}

func insertCredentialWorkspaceMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, role string) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, userID, role)
	require.NoError(t, err)
}

func insertCredentialTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Credentials "+label, workspaceID, "C"+uuid.NewString()[:7])
	require.NoError(t, err)
	return id
}

func insertCredentialTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID)
	require.NoError(t, err)
}

func assertCredentialLastUsedAt(t *testing.T, ctx context.Context, pool *pgxpool.Pool, credentialID uuid.UUID, expected time.Time) {
	t.Helper()
	var actual *time.Time
	require.NoError(t, pool.QueryRow(ctx, `SELECT last_used_at FROM api_credentials WHERE credential_id = $1`, credentialID).Scan(&actual))
	require.NotNil(t, actual)
	require.WithinDuration(t, expected, *actual, time.Microsecond)
}
