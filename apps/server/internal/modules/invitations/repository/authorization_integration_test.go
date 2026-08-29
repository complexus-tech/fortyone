//go:build integration

package invitationsrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInvitationAdministrationFailsClosedForUnauthorizedActorStates(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, baseline := fixture.newInvitation(t)
	if _, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{baseline}); err != nil {
		t.Fatalf("create baseline invitation: %v", err)
	}

	states := []struct {
		name                string
		active              bool
		membershipWorkspace uuid.UUID
		role                string
	}{
		{name: "inactive", active: false, membershipWorkspace: fixture.workspaceA, role: "admin"},
		{name: "removed", active: true},
		{name: "guest", active: true, membershipWorkspace: fixture.workspaceA, role: "guest"},
		{name: "cross tenant", active: true, membershipWorkspace: fixture.workspaceB, role: "admin"},
	}
	for _, state := range states {
		state := state
		t.Run(state.name, func(t *testing.T) {
			actorID, actorEmail := insertInvitationAuthorizationActor(
				t,
				fixture.pool,
				state.name,
				state.active,
				state.membershipWorkspace,
				state.role,
			)
			_, createCommand := fixture.newInvitation(t)
			createCommand.Invitation.ID = uuid.New()
			createCommand.Invitation.InviterID = actorID
			createCommand.Invitation.Email = actorEmail
			createCommand.EmailOutbox.Email = actorEmail
			_, err := fixture.repository.CreateBulkInvitations(ctx, actorID, []invitations.NewWorkspaceInvitation{createCommand})
			assertWorkspaceAdminRequired(t, err, "create")
			assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitations WHERE invitation_id = $1", 0, createCommand.Invitation.ID)

			_, err = fixture.repository.ListInvitations(ctx, fixture.workspaceA, actorID, time.Now().UTC())
			assertWorkspaceAdminRequired(t, err, "list")

			err = fixture.repository.RevokeInvitation(ctx, fixture.workspaceA, actorID, baseline.Invitation.ID, time.Now().UTC())
			assertWorkspaceAdminRequired(t, err, "revoke")
			assertInvitationUsedAt(t, ctx, fixture.pool, baseline.Invitation.ID, false)
		})
	}

	listed, err := fixture.repository.ListInvitations(ctx, fixture.workspaceA, fixture.inviterID, time.Now().UTC())
	if err != nil {
		t.Fatalf("list invitations as active admin: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != baseline.Invitation.ID {
		t.Fatalf("authorized invitations = %#v, want baseline %s", listed, baseline.Invitation.ID)
	}
}

func TestInvitationAuthorizationLinearizesConcurrentDemotionAndRemoval(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, baseline := fixture.newInvitation(t)
	if _, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{baseline}); err != nil {
		t.Fatalf("create baseline invitation: %v", err)
	}

	demotionTx := lockInvitationMembership(t, ctx, fixture.pool, fixture.workspaceA, fixture.inviterID)
	_, concurrentCreate := fixture.newInvitation(t)
	concurrentCreate.Invitation.ID = uuid.New()
	concurrentCreate.Invitation.Email = "demotion-" + fixture.invitee
	concurrentCreate.EmailOutbox.Email = concurrentCreate.Invitation.Email
	createResult := make(chan error, 1)
	go func() {
		_, err := fixture.repository.CreateBulkInvitations(
			ctx,
			fixture.inviterID,
			[]invitations.NewWorkspaceInvitation{concurrentCreate},
		)
		createResult <- err
	}()
	waitForBlockedInvitationAuthorization(t, ctx, fixture.pool)
	if _, err := demotionTx.Exec(ctx, `
		UPDATE workspace_members SET role = 'guest'
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.inviterID); err != nil {
		t.Fatalf("demote invitation administrator: %v", err)
	}
	if err := demotionTx.Commit(ctx); err != nil {
		t.Fatalf("commit invitation administrator demotion: %v", err)
	}
	assertWorkspaceAdminRequired(t, <-createResult, "create after concurrent demotion")
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitations WHERE invitation_id = $1", 0, concurrentCreate.Invitation.ID)

	if _, err := fixture.pool.Exec(ctx, `
		UPDATE workspace_members SET role = 'admin'
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.inviterID); err != nil {
		t.Fatalf("restore invitation administrator: %v", err)
	}
	removalTx := lockInvitationMembership(t, ctx, fixture.pool, fixture.workspaceA, fixture.inviterID)
	revokeResult := make(chan error, 1)
	go func() {
		revokeResult <- fixture.repository.RevokeInvitation(
			ctx,
			fixture.workspaceA,
			fixture.inviterID,
			baseline.Invitation.ID,
			time.Now().UTC(),
		)
	}()
	waitForBlockedInvitationAuthorization(t, ctx, fixture.pool)
	if _, err := removalTx.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.inviterID); err != nil {
		t.Fatalf("remove invitation administrator membership: %v", err)
	}
	if err := removalTx.Commit(ctx); err != nil {
		t.Fatalf("commit invitation administrator removal: %v", err)
	}
	assertWorkspaceAdminRequired(t, <-revokeResult, "revoke after concurrent removal")
	assertInvitationUsedAt(t, ctx, fixture.pool, baseline.Invitation.ID, false)
}

func insertInvitationAuthorizationActor(
	t *testing.T,
	pool *pgxpool.Pool,
	label string,
	active bool,
	membershipWorkspace uuid.UUID,
	role string,
) (uuid.UUID, string) {
	t.Helper()
	actorID := uuid.New()
	suffix := uuid.NewString()
	email := "invitation-state-" + suffix + "@example.com"
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, actorID, "invitation-state-"+suffix, email, label, active); err != nil {
		t.Fatalf("insert %s invitation actor: %v", label, err)
	}
	if membershipWorkspace != uuid.Nil {
		if _, err := pool.Exec(t.Context(), `
			INSERT INTO workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, $3)
		`, membershipWorkspace, actorID, role); err != nil {
			t.Fatalf("insert %s invitation actor membership: %v", label, err)
		}
	}
	return actorID, email
}

func lockInvitationMembership(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
) pgx.Tx {
	t.Helper()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin invitation membership lock transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	var role string
	if err := tx.QueryRow(ctx, `
		SELECT CAST(role AS text)
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
		FOR UPDATE
	`, workspaceID, actorID).Scan(&role); err != nil {
		t.Fatalf("lock invitation administrator membership: %v", err)
	}
	return tx
}

func waitForBlockedInvitationAuthorization(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM pg_stat_activity
				WHERE datname = current_database()
				  AND pid <> pg_backend_pid()
				  AND query LIKE '%LockActiveWorkspaceAdmin%'
				  AND wait_event_type = 'Lock'
			)
		`).Scan(&blocked); err != nil {
			t.Fatalf("inspect blocked invitation authorization: %v", err)
		}
		if blocked {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for blocked invitation authorization: %v", ctx.Err())
		case <-deadline.C:
			t.Fatal("invitation authorization did not block on the administrator membership row")
		case <-ticker.C:
		}
	}
}

func assertWorkspaceAdminRequired(t *testing.T, err error, operation string) {
	t.Helper()
	if !errors.Is(err, authorization.ErrWorkspaceAdminRequired) {
		t.Fatalf("%s error = %v, want ErrWorkspaceAdminRequired", operation, err)
	}
}
