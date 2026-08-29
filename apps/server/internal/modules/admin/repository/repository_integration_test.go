//go:build integration

package adminrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAdminRepositoryPostgres18SecurityAtomicityAndConcurrency(t *testing.T) {
	fixture := newAdminIntegrationFixture(t)
	ctx := t.Context()

	t.Run("live active internal authorization protects every read", func(t *testing.T) {
		summary, err := fixture.repository.GetDashboardSummary(ctx, admindomain.DashboardSummaryQuery{
			ActorID: fixture.actorA, Now: fixture.now,
		})
		require.NoError(t, err)
		require.Equal(t, 1, summary.SlackInstallations)
		require.Equal(t, 1, summary.GitHubInstallations)

		for _, actorID := range []uuid.UUID{fixture.inactiveAdmin, fixture.externalUser, uuid.New()} {
			_, err := fixture.repository.GetDashboardSummary(ctx, admindomain.DashboardSummaryQuery{
				ActorID: actorID, Now: fixture.now,
			})
			require.ErrorIs(t, err, admindomain.ErrForbidden)
		}

		_, err = fixture.postgres.Pool.Exec(ctx, `
			UPDATE users SET is_active = FALSE WHERE user_id = CAST($1 AS uuid)
		`, fixture.actorA)
		require.NoError(t, err)
		_, err = fixture.repository.GetWorkspaceOverview(ctx, admindomain.GetWorkspaceQuery{
			ActorID: fixture.actorA, WorkspaceID: fixture.workspaceID,
		})
		require.ErrorIs(t, err, admindomain.ErrForbidden)
		_, err = fixture.postgres.Pool.Exec(ctx, `
			UPDATE users SET is_active = TRUE WHERE user_id = CAST($1 AS uuid)
		`, fixture.actorA)
		require.NoError(t, err)
	})

	t.Run("integration availability and typed workspace filters are exact", func(t *testing.T) {
		overview, err := fixture.repository.GetWorkspaceOverview(ctx, admindomain.GetWorkspaceQuery{
			ActorID: fixture.actorA, WorkspaceID: fixture.workspaceID,
		})
		require.NoError(t, err)
		require.True(t, overview.Workspace.SlackInstalled)
		require.True(t, overview.Workspace.GitHubInstalled)
		require.Equal(t, "pro", requireString(t, overview.Workspace.SubscriptionTier))

		result, err := fixture.repository.ListWorkspaces(ctx, admindomain.ListWorkspacesQuery{
			ActorID: fixture.actorA, Page: pagination.OffsetParams{Page: 1, PageSize: 20},
			Status: admindomain.WorkspaceStatusPastDue, Now: fixture.now,
		})
		require.NoError(t, err)
		require.Len(t, result.Items, 1)
		require.Equal(t, fixture.workspaceID, result.Items[0].ID)
	})

	t.Run("workspace trial and immutable audit commit together", func(t *testing.T) {
		trialEndsOn := fixture.now.Add(10 * 24 * time.Hour)
		overview, err := fixture.repository.UpdateWorkspaceTrial(ctx, admindomain.UpdateWorkspaceTrialCommand{
			ActorID: fixture.actorA, WorkspaceID: fixture.workspaceID,
			TrialEndsOn: trialEndsOn, Reason: "approved extension", Now: fixture.now,
		})
		require.NoError(t, err)
		require.WithinDuration(t, trialEndsOn, *overview.Workspace.TrialEndsOn, time.Microsecond)

		var auditCount int
		err = fixture.postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM admin_audit_logs
			WHERE workspace_id = CAST($1 AS uuid) AND action = 'workspace.trial_updated'
		`, fixture.workspaceID).Scan(&auditCount)
		require.NoError(t, err)
		require.Equal(t, 1, auditCount)
	})

	t.Run("audit failures roll back workspace user and note mutations", func(t *testing.T) {
		installFailingAdminAuditTrigger(t, fixture)

		var workspaceUpdatedAt time.Time
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT updated_at FROM workspaces WHERE workspace_id = CAST($1 AS uuid)
		`, fixture.workspaceID).Scan(&workspaceUpdatedAt))
		_, err := fixture.repository.SetWorkspaceDeleted(ctx, admindomain.SetWorkspaceDeletedCommand{
			ActorID: fixture.actorA, WorkspaceID: fixture.workspaceID,
			Deleted: true, Reason: "retention review", Now: fixture.now.Add(time.Minute),
		})
		require.Error(t, err)
		var deletedAt *time.Time
		var updatedAt time.Time
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT deleted_at, updated_at FROM workspaces WHERE workspace_id = CAST($1 AS uuid)
		`, fixture.workspaceID).Scan(&deletedAt, &updatedAt))
		require.Nil(t, deletedAt)
		require.True(t, workspaceUpdatedAt.Equal(updatedAt))

		_, err = fixture.repository.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
			ActorID: fixture.actorA, UserID: fixture.targetUser, Reason: "access review",
			Now:   fixture.now.Add(2 * time.Minute),
			Patch: admindomain.UserStatePatch{IsActive: platformpatch.Set(false)},
		})
		require.Error(t, err)
		var active bool
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT is_active FROM users WHERE user_id = CAST($1 AS uuid)
		`, fixture.targetUser).Scan(&active))
		require.True(t, active)

		_, err = fixture.repository.CreateAdminNote(ctx, admindomain.CreateAdminNoteCommand{
			ActorID: fixture.actorA, TargetType: admindomain.TargetWorkspace,
			TargetID: fixture.workspaceID, WorkspaceID: &fixture.workspaceID, Body: "must roll back",
		})
		require.Error(t, err)
		var noteCount int
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM admin_notes WHERE body = 'must roll back'
		`).Scan(&noteCount))
		require.Zero(t, noteCount)

		dropFailingAdminAuditTrigger(t, fixture)
	})

	t.Run("user note workspace context cannot cross tenant membership", func(t *testing.T) {
		_, err := fixture.repository.CreateAdminNote(ctx, admindomain.CreateAdminNoteCommand{
			ActorID: fixture.actorA, TargetType: admindomain.TargetUser,
			TargetID: fixture.targetUser, WorkspaceID: &fixture.otherWorkspace, Body: "wrong tenant",
		})
		require.ErrorIs(t, err, admindomain.ErrNotFound)

		note, err := fixture.repository.CreateAdminNote(ctx, admindomain.CreateAdminNoteCommand{
			ActorID: fixture.actorA, TargetType: admindomain.TargetUser,
			TargetID: fixture.targetUser, WorkspaceID: &fixture.workspaceID, Body: "verified tenant",
		})
		require.NoError(t, err)
		require.Equal(t, fixture.workspaceID, *note.WorkspaceID)
		var auditCount int
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM admin_audit_logs
			WHERE action = 'admin_note.created' AND metadata ->> 'note_id' = CAST($1 AS text)
		`, note.ID).Scan(&auditCount))
		require.Equal(t, 1, auditCount)
	})

	t.Run("self mutation and cross admin race cannot remove the acting grant", func(t *testing.T) {
		_, err := fixture.repository.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
			ActorID: fixture.actorA, UserID: fixture.actorA, Reason: "self change",
			Now: fixture.now, Patch: admindomain.UserStatePatch{IsInternal: platformpatch.Set(false)},
		})
		require.ErrorIs(t, err, admindomain.ErrSelfMutation)

		start := make(chan struct{})
		results := make(chan error, 2)
		var wait sync.WaitGroup
		for _, pair := range [][2]uuid.UUID{{fixture.actorA, fixture.actorB}, {fixture.actorB, fixture.actorA}} {
			wait.Add(1)
			go func(actorID, targetID uuid.UUID) {
				defer wait.Done()
				<-start
				_, err := fixture.repository.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
					ActorID: actorID, UserID: targetID, Reason: "concurrent review",
					Now:   fixture.now.Add(3 * time.Minute),
					Patch: admindomain.UserStatePatch{IsActive: platformpatch.Set(false)},
				})
				results <- err
			}(pair[0], pair[1])
		}
		close(start)
		wait.Wait()
		close(results)
		successes, forbidden := 0, 0
		for err := range results {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, admindomain.ErrForbidden):
				forbidden++
			default:
				t.Fatalf("unexpected cross-admin mutation error: %v", err)
			}
		}
		require.Equal(t, 1, successes)
		require.Equal(t, 1, forbidden)
		_, err = fixture.postgres.Pool.Exec(ctx, `
			UPDATE users SET is_active = TRUE
			WHERE user_id = ANY(CAST($1 AS uuid[]))
		`, []uuid.UUID{fixture.actorA, fixture.actorB})
		require.NoError(t, err)
	})

	t.Run("concurrent patches serialize without losing unrelated fields", func(t *testing.T) {
		start := make(chan struct{})
		results := make(chan error, 2)
		patches := []admindomain.UserStatePatch{
			{IsActive: platformpatch.Set(false)},
			{IsInternal: platformpatch.Set(true)},
		}
		var wait sync.WaitGroup
		for _, patch := range patches {
			wait.Add(1)
			go func(patch admindomain.UserStatePatch) {
				defer wait.Done()
				<-start
				_, err := fixture.repository.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
					ActorID: fixture.actorA, UserID: fixture.targetUser,
					Patch: patch, Reason: "parallel field review", Now: fixture.now.Add(4 * time.Minute),
				})
				results <- err
			}(patch)
		}
		close(start)
		wait.Wait()
		close(results)
		for err := range results {
			require.NoError(t, err)
		}
		var active, internal bool
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT is_active, is_internal FROM users WHERE user_id = CAST($1 AS uuid)
		`, fixture.targetUser).Scan(&active, &internal))
		require.False(t, active)
		require.True(t, internal)
	})

	t.Run("subscription request and result audits are explicitly two phase", func(t *testing.T) {
		attempt, _, err := fixture.repository.BeginSubscriptionSync(ctx, admindomain.BeginSubscriptionSyncCommand{
			ActorID: fixture.actorA, WorkspaceID: fixture.workspaceID, Reason: "billing reconciliation",
		})
		require.NoError(t, err)
		_, err = fixture.repository.FinishSubscriptionSync(ctx, admindomain.FinishSubscriptionSyncCommand{
			Attempt: attempt, Outcome: admindomain.SubscriptionSyncSucceeded,
		})
		require.NoError(t, err)

		var requested, finished int
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM admin_audit_logs
			WHERE id = CAST($1 AS uuid) AND action = 'subscription.sync_requested'
		`, attempt.AuditID).Scan(&requested))
		require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM admin_audit_logs
			WHERE action = 'subscription.synced'
			  AND metadata ->> 'request_audit_id' = CAST($1 AS text)
		`, attempt.AuditID).Scan(&finished))
		require.Equal(t, 1, requested)
		require.Equal(t, 1, finished)
	})
}

func TestAdminAccountStateOwnsLoginReactivationAndSessionFencing(t *testing.T) {
	fixture := newAdminIntegrationFixture(t)
	ctx := t.Context()
	usersRepository := usersrepository.New(fixture.postgres.Pool)
	warningAt := fixture.now.Add(-48 * time.Hour)
	_, err := fixture.postgres.Pool.Exec(ctx, `
		UPDATE users
		SET inactivity_warning_sent_at = CAST($2 AS timestamptz)
		WHERE user_id = CAST($1 AS uuid)
	`, fixture.targetUser, warningAt)
	require.NoError(t, err)

	var initialVersion int64
	require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
		SELECT auth_session_version
		FROM users
		WHERE user_id = CAST($1 AS uuid)
	`, fixture.targetUser).Scan(&initialVersion))

	disabled, err := fixture.repository.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
		ActorID: fixture.actorA, UserID: fixture.targetUser, Reason: "security incident",
		Now:   fixture.now.Add(time.Minute),
		Patch: admindomain.UserStatePatch{IsActive: platformpatch.Set(false)},
	})
	require.NoError(t, err)
	require.False(t, disabled.User.IsActive)
	require.Equal(t, admindomain.LoginReactivationAdminOnly, disabled.User.LoginReactivationPolicy)

	_, err = usersRepository.ReactivateUserForVerifiedSignIn(ctx, users.VerifiedSignInReactivation{
		UserID: fixture.targetUser, SignedInAt: fixture.now.Add(2 * time.Minute),
	})
	require.ErrorIs(t, err, users.ErrInvalidCredentials)

	enabled, err := fixture.repository.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
		ActorID: fixture.actorA, UserID: fixture.targetUser, Reason: "security review complete",
		Now:   fixture.now.Add(3 * time.Minute),
		Patch: admindomain.UserStatePatch{IsActive: platformpatch.Set(true)},
	})
	require.NoError(t, err)
	require.True(t, enabled.User.IsActive)
	require.Equal(t, admindomain.LoginReactivationVerifiedSignIn, enabled.User.LoginReactivationPolicy)

	var policy string
	var version int64
	var persistedWarning *time.Time
	require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
		SELECT login_reactivation_policy, auth_session_version, inactivity_warning_sent_at
		FROM users
		WHERE user_id = CAST($1 AS uuid)
	`, fixture.targetUser).Scan(&policy, &version, &persistedWarning))
	require.Equal(t, "verified_sign_in", policy)
	require.Equal(t, initialVersion+2, version, "both admin state transitions must fence cached sessions")
	require.Nil(t, persistedWarning, "admin re-enable must restart the inactivity warning grace period")

	var deactivatedAudits, activatedAudits, policyAudits int
	require.NoError(t, fixture.postgres.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE action = 'user.deactivated'),
			COUNT(*) FILTER (WHERE action = 'user.activated'),
			COUNT(*) FILTER (WHERE action = 'user.reactivation_policy_changed')
		FROM admin_audit_logs
		WHERE target_type = 'user'
		  AND target_id = CAST($1 AS uuid)
	`, fixture.targetUser).Scan(&deactivatedAudits, &activatedAudits, &policyAudits))
	require.Equal(t, 1, deactivatedAudits)
	require.Equal(t, 1, activatedAudits)
	require.Equal(t, 2, policyAudits)
}

func requireString(t *testing.T, value *string) string {
	t.Helper()
	require.NotNil(t, value)
	return *value
}

func installFailingAdminAuditTrigger(t *testing.T, fixture adminIntegrationFixture) {
	t.Helper()
	_, err := fixture.postgres.Pool.Exec(t.Context(), `
		CREATE FUNCTION test_reject_admin_audit() RETURNS trigger LANGUAGE plpgsql AS $function$
		BEGIN
			RAISE EXCEPTION 'intentional admin audit failure';
		END
		$function$;
		CREATE TRIGGER test_reject_admin_audit
		BEFORE INSERT ON admin_audit_logs
		FOR EACH ROW EXECUTE FUNCTION test_reject_admin_audit();
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, cleanupErr := fixture.postgres.Pool.Exec(cleanupCtx, `
			DROP TRIGGER IF EXISTS test_reject_admin_audit ON admin_audit_logs;
			DROP FUNCTION IF EXISTS test_reject_admin_audit();
		`)
		if cleanupErr != nil {
			t.Errorf("clean up failing admin audit trigger: %v", cleanupErr)
		}
	})
}

func dropFailingAdminAuditTrigger(t *testing.T, fixture adminIntegrationFixture) {
	t.Helper()
	_, err := fixture.postgres.Pool.Exec(t.Context(), `
		DROP TRIGGER IF EXISTS test_reject_admin_audit ON admin_audit_logs;
		DROP FUNCTION IF EXISTS test_reject_admin_audit();
	`)
	require.NoError(t, err)
}
