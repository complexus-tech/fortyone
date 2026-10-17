//go:build integration

package workspacesrepository

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type activitySnapshot struct {
	userVersion, workspaceVersion string
	userAt, workspaceAt           time.Time
	userWarning, workspaceWarning *time.Time
}

func readActivitySnapshot(t *testing.T, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) activitySnapshot {
	t.Helper()
	var snapshot activitySnapshot
	require.NoError(t, pool.QueryRow(t.Context(), `
		SELECT CAST(account.xmin AS text), account.last_login_at, account.inactivity_warning_sent_at,
		       CAST(workspace.xmin AS text), workspace.last_accessed_at, workspace.inactivity_warning_sent_at
		FROM users account CROSS JOIN workspaces workspace
		WHERE account.user_id=$1 AND workspace.workspace_id=$2
	`, userID, workspaceID).Scan(&snapshot.userVersion, &snapshot.userAt, &snapshot.userWarning,
		&snapshot.workspaceVersion, &snapshot.workspaceAt, &snapshot.workspaceWarning))
	return snapshot
}

func TestAccessTrackingOnlyWritesAgedOrWarnedRowsIndependently(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspace := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "activity")
	user := insertWorkspaceTestUser(t, ctx, postgres.Pool, "activity", true)
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspace, user, "member")
	repository := New(postgres.Pool)
	for _, test := range []struct {
		name                                             string
		agedUser, agedWorkspace, warnUser, warnWorkspace bool
	}{
		{name: "fresh rows do not write"},
		{name: "only user aged", agedUser: true},
		{name: "only workspace aged", agedWorkspace: true},
		{name: "both aged", agedUser: true, agedWorkspace: true},
		{name: "fresh user warning clears immediately", warnUser: true},
		{name: "fresh workspace warning clears immediately", warnWorkspace: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)
			userAt, workspaceAt := now, now
			if test.agedUser {
				userAt = now.Add(-2 * time.Minute)
			}
			if test.agedWorkspace {
				workspaceAt = now.Add(-2 * time.Minute)
			}
			var userWarning, workspaceWarning *time.Time
			if test.warnUser {
				userWarning = &now
			}
			if test.warnWorkspace {
				workspaceWarning = &now
			}
			_, err := postgres.Pool.Exec(ctx, `UPDATE users SET last_login_at=$2, inactivity_warning_sent_at=$3 WHERE user_id=$1`, user, userAt, userWarning)
			require.NoError(t, err)
			_, err = postgres.Pool.Exec(ctx, `UPDATE workspaces SET last_accessed_at=$2, inactivity_warning_sent_at=$3 WHERE workspace_id=$1`, workspace, workspaceAt, workspaceWarning)
			require.NoError(t, err)
			before := readActivitySnapshot(t, postgres.Pool, workspace, user)
			require.NoError(t, repository.RecordAccess(ctx, workspace, user))
			after := readActivitySnapshot(t, postgres.Pool, workspace, user)
			require.Equal(t, test.agedUser || test.warnUser, before.userVersion != after.userVersion, "user write predicate")
			require.Equal(t, test.agedWorkspace || test.warnWorkspace, before.workspaceVersion != after.workspaceVersion, "workspace write predicate")
			require.Nil(t, after.userWarning)
			require.Nil(t, after.workspaceWarning)
			for range 5 {
				require.NoError(t, repository.RecordAccess(ctx, workspace, user))
			}
			require.Equal(t, after, readActivitySnapshot(t, postgres.Pool, workspace, user), "repeated fresh reads must not rewrite either tuple")
		})
	}
}
