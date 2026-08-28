//go:build integration

package invitationsrepository

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const repositoryTokenSecret = "repository-invitation-test-key-with-at-least-32-bytes"

type invitationFixture struct {
	pool       *pgxpool.Pool
	repository *repo
	tokens     *invitations.InvitationTokenManager
	inviterID  uuid.UUID
	inviteeID  uuid.UUID
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	invitee    string
}

func newInvitationFixture(t *testing.T) invitationFixture {
	t.Helper()
	database := testkit.NewPostgres(t)
	fixture := invitationFixture{
		pool:       database.Pool,
		repository: New(database.Pool),
		inviterID:  uuid.New(),
		inviteeID:  uuid.New(),
		workspaceA: uuid.New(),
		workspaceB: uuid.New(),
		teamA:      uuid.New(),
		teamB:      uuid.New(),
	}
	suffix := uuid.NewString()
	fixture.invitee = "invitee-" + suffix + "@example.com"
	manager, err := invitations.NewInvitationTokenManager(invitations.InvitationTokenConfig{
		Current: invitations.InvitationTokenKey{ID: "test-v1", Secret: repositoryTokenSecret},
	})
	if err != nil {
		t.Fatalf("create invitation token manager: %v", err)
	}
	fixture.tokens = manager

	ctx := t.Context()
	if _, err := database.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Invitation Admin'),
		       ($4, $5, $6, 'Invitation Recipient')
	`, fixture.inviterID, "admin-"+suffix, "admin-"+suffix+"@example.com",
		fixture.inviteeID, "invitee-"+suffix, fixture.invitee); err != nil {
		t.Fatalf("insert invitation users: %v", err)
	}
	if _, err := database.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Workspace A', $2, $3),
		       ($4, 'Workspace B', $5, $3)
	`, fixture.workspaceA, "workspace-a-"+suffix, fixture.inviterID,
		fixture.workspaceB, "workspace-b-"+suffix); err != nil {
		t.Fatalf("insert invitation workspaces: %v", err)
	}
	if _, err := database.Pool.Exec(ctx, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Team A', $3, '#000000'),
		       ($4, $5, 'Team B', $6, '#ffffff')
	`, fixture.teamA, fixture.workspaceA, "TA-"+suffix,
		fixture.teamB, fixture.workspaceB, "TB-"+suffix); err != nil {
		t.Fatalf("insert invitation teams: %v", err)
	}
	if _, err := database.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $3, 'admin'), ($2, $3, 'admin')
	`, fixture.workspaceA, fixture.workspaceB, fixture.inviterID); err != nil {
		t.Fatalf("insert invitation administrator memberships: %v", err)
	}
	return fixture
}

func (f invitationFixture) newInvitation(t *testing.T, teamIDs ...uuid.UUID) (string, invitations.NewWorkspaceInvitation) {
	t.Helper()
	rawToken, stored, err := f.tokens.Issue()
	if err != nil {
		t.Fatalf("issue invitation token: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	command := invitations.NewWorkspaceInvitation{
		Invitation: invitations.CoreWorkspaceInvitation{
			ID:          uuid.New(),
			WorkspaceID: f.workspaceA,
			InviterID:   f.inviterID,
			Email:       f.invitee,
			Role:        invitations.InvitationRoleMember,
			TeamIDs:     teamIDs,
			ExpiresAt:   now.Add(7 * 24 * time.Hour),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Token: stored,
		EmailOutbox: invitations.InvitationEmailOutboxPayload{
			InviterName:   "Invitation Admin",
			Email:         f.invitee,
			Role:          invitations.InvitationRoleMember,
			ExpiresAt:     now.Add(7 * 24 * time.Hour),
			WorkspaceID:   f.workspaceA,
			WorkspaceName: "Workspace A",
		},
	}
	return rawToken, command
}

func TestCreateInvitationStoresOnlyDigestAndTokenFreeOutbox(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	rawToken, command := fixture.newInvitation(t, fixture.teamA)

	created, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{command})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("created invitation count = %d, want 1", len(created))
	}

	var (
		plaintext *string
		digest    []byte
		nonce     []byte
		keyID     *string
		version   *int16
	)
	if err := fixture.pool.QueryRow(ctx, `
		SELECT token, token_digest, token_nonce, token_key_id, token_version
		FROM workspace_invitations
		WHERE invitation_id = $1
	`, command.Invitation.ID).Scan(&plaintext, &digest, &nonce, &keyID, &version); err != nil {
		t.Fatalf("read stored invitation token: %v", err)
	}
	if plaintext != nil || len(digest) != 32 || len(nonce) != 32 || keyID == nil || *keyID != "test-v1" || version == nil || *version != 1 {
		t.Fatalf("unexpected invitation token storage: plaintext=%v digest=%d nonce=%d keyID=%v version=%v", plaintext != nil, len(digest), len(nonce), keyID, version)
	}
	if bytes.Contains(digest, []byte(rawToken)) || bytes.Contains(nonce, []byte(rawToken)) {
		t.Fatal("raw invitation token was persisted in token metadata")
	}

	var payload string
	if err := fixture.pool.QueryRow(ctx, `
		SELECT event_payload::text
		FROM workspace_invitation_outbox
		WHERE invitation_id = $1 AND event_type = $2
	`, command.Invitation.ID, string(events.InvitationEmail)).Scan(&payload); err != nil {
		t.Fatalf("read invitation email outbox: %v", err)
	}
	if strings.Contains(payload, rawToken) || strings.Contains(payload, `"token"`) {
		t.Fatal("invitation email outbox persisted a raw token")
	}

	lookup, err := fixture.tokens.Lookup(rawToken)
	if err != nil {
		t.Fatalf("derive invitation lookup: %v", err)
	}
	got, err := fixture.repository.GetInvitation(ctx, lookup)
	if err != nil || got.ID != command.Invitation.ID {
		t.Fatalf("lookup digest invitation = (%v, %v), want %s", got.ID, err, command.Invitation.ID)
	}
}

func TestInvitationRepositoryReadsBoundedLegacyTokenShape(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	legacy := base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))
	invitationID := uuid.New()
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO workspace_invitations
			(invitation_id, workspace_id, inviter_id, email, role, token, expires_at)
		VALUES ($1, $2, $3, $4, 'member', $5, NOW() + INTERVAL '1 day')
	`, invitationID, fixture.workspaceA, fixture.inviterID, fixture.invitee, legacy); err != nil {
		t.Fatalf("insert legacy invitation: %v", err)
	}
	lookup, err := fixture.tokens.Lookup(legacy)
	if err != nil {
		t.Fatalf("parse legacy invitation token: %v", err)
	}
	got, err := fixture.repository.GetInvitation(ctx, lookup)
	if err != nil || got.ID != invitationID {
		t.Fatalf("lookup legacy invitation = (%v, %v), want %s", got.ID, err, invitationID)
	}
}

func TestCreateBulkInvitationRollsBackOnCrossWorkspaceTeam(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, first := fixture.newInvitation(t, fixture.teamA)
	_, second := fixture.newInvitation(t, fixture.teamB)
	second.Invitation.Email = "second-" + fixture.invitee
	second.EmailOutbox.Email = second.Invitation.Email

	_, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{first, second})
	if !errors.Is(err, invitations.ErrInvalidInvitationTeam) {
		t.Fatalf("create cross-workspace team invitation error = %v, want ErrInvalidInvitationTeam", err)
	}
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitations", 0)
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitation_outbox", 0)
}

func TestConcurrentInvitationCreationLeavesOneDeliverableCredential(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, first := fixture.newInvitation(t)
	_, second := fixture.newInvitation(t)
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, command := range []invitations.NewWorkspaceInvitation{first, second} {
		command := command
		go func() {
			<-start
			_, createErr := fixture.repository.CreateBulkInvitations(context.Background(), fixture.inviterID, []invitations.NewWorkspaceInvitation{command})
			results <- createErr
		}()
	}
	close(start)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatalf("concurrent invitation creation: %v", err)
		}
	}
	assertCount(t, ctx, fixture.pool, `
		SELECT count(*)
		FROM workspace_invitations
		WHERE workspace_id = $1 AND lower(email) = lower($2) AND used_at IS NULL
	`, 1, fixture.workspaceA, fixture.invitee)
	assertCount(t, ctx, fixture.pool, `
		SELECT count(*)
		FROM workspace_invitation_outbox
		WHERE workspace_id = $1 AND event_type = $2 AND status = 'pending'
	`, 1, fixture.workspaceA, string(events.InvitationEmail))
	assertCount(t, ctx, fixture.pool, `
		SELECT count(*)
		FROM workspace_invitation_outbox
		WHERE workspace_id = $1 AND event_type = $2 AND status = 'cancelled'
	`, 1, fixture.workspaceA, string(events.InvitationEmail))
}

func TestAcceptInvitationIsAtomicSingleUseAndActorScoped(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	rawToken, command := fixture.newInvitation(t, fixture.teamA)
	if _, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{command}); err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	lookup, err := fixture.tokens.Lookup(rawToken)
	if err != nil {
		t.Fatalf("derive invitation lookup: %v", err)
	}

	wrongUserID := uuid.New()
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Wrong Recipient')
	`, wrongUserID, "wrong-"+uuid.NewString(), "wrong-"+fixture.invitee); err != nil {
		t.Fatalf("insert wrong invitee: %v", err)
	}
	_, err = fixture.repository.AcceptInvitation(ctx, invitations.AcceptInvitationCommand{
		Lookup: lookup, UserID: wrongUserID, AcceptedAt: time.Now().UTC(),
	})
	if !errors.Is(err, invitations.ErrInvitationNotFound) {
		t.Fatalf("wrong-recipient acceptance error = %v, want enumeration-safe not found", err)
	}
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_members WHERE user_id = $1", 0, wrongUserID)

	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			_, acceptErr := fixture.repository.AcceptInvitation(context.Background(), invitations.AcceptInvitationCommand{
				Lookup: lookup, UserID: fixture.inviteeID, AcceptedAt: time.Now().UTC(),
			})
			results <- acceptErr
		}()
	}
	ready.Wait()
	close(start)
	firstErr, secondErr := <-results, <-results
	successes, usedErrors := 0, 0
	for _, acceptErr := range []error{firstErr, secondErr} {
		switch {
		case acceptErr == nil:
			successes++
		case errors.Is(acceptErr, invitations.ErrInvitationUsed):
			usedErrors++
		default:
			t.Fatalf("concurrent acceptance error = %v", acceptErr)
		}
	}
	if successes != 1 || usedErrors != 1 {
		t.Fatalf("concurrent outcomes successes=%d used=%d, want 1/1", successes, usedErrors)
	}
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_members WHERE workspace_id = $1 AND user_id = $2", 1, fixture.workspaceA, fixture.inviteeID)
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM team_members WHERE team_id = $1 AND user_id = $2", 1, fixture.teamA, fixture.inviteeID)
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitation_outbox WHERE invitation_id = $1 AND event_type = $2", 1, command.Invitation.ID, string(events.InvitationAccepted))
	assertInvitationUsedAt(t, ctx, fixture.pool, command.Invitation.ID, true)
}

func TestRevokeInvitationEnforcesWorkspaceScope(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, command := fixture.newInvitation(t)
	if _, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{command}); err != nil {
		t.Fatalf("create invitation: %v", err)
	}

	if err := fixture.repository.RevokeInvitation(ctx, fixture.workspaceB, fixture.inviterID, command.Invitation.ID, time.Now().UTC()); !errors.Is(err, invitations.ErrInvitationNotFound) {
		t.Fatalf("cross-workspace revoke error = %v, want ErrInvitationNotFound", err)
	}
	assertInvitationUsedAt(t, ctx, fixture.pool, command.Invitation.ID, false)

	if err := fixture.repository.RevokeInvitation(ctx, fixture.workspaceA, fixture.inviterID, command.Invitation.ID, time.Now().UTC()); err != nil {
		t.Fatalf("same-workspace revoke: %v", err)
	}
	assertInvitationUsedAt(t, ctx, fixture.pool, command.Invitation.ID, true)
}

func TestInvitationOutboxClaimFencing(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, command := fixture.newInvitation(t)
	if _, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{command}); err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	now := time.Now().UTC().Add(time.Second)
	claimed, err := fixture.repository.ClaimInvitationOutboxEvents(ctx, 10, now, now.Add(-10*time.Minute))
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim invitation outbox = (%d, %v), want (1, nil)", len(claimed), err)
	}
	if err := fixture.repository.CompleteInvitationOutboxEvent(ctx, claimed[0].ID, uuid.New(), now); !errors.Is(err, invitations.ErrOutboxClaimLost) {
		t.Fatalf("stale claim completion error = %v, want ErrOutboxClaimLost", err)
	}
	if err := fixture.repository.CompleteInvitationOutboxEvent(ctx, claimed[0].ID, claimed[0].ClaimToken, now); err != nil {
		t.Fatalf("complete current invitation outbox claim: %v", err)
	}
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitation_outbox WHERE outbox_id = $1 AND status = 'completed'", 1, claimed[0].ID)
}

func TestMigration155DownRefusesDigestOnlyInvitation(t *testing.T) {
	fixture := newInvitationFixture(t)
	ctx := t.Context()
	_, command := fixture.newInvitation(t)
	if _, err := fixture.repository.CreateBulkInvitations(ctx, fixture.inviterID, []invitations.NewWorkspaceInvitation{command}); err != nil {
		t.Fatalf("create digest-only invitation: %v", err)
	}
	downMigration, err := migrations.FS.ReadFile("000155_workspace_invitation_token_digests.down.sql")
	if err != nil {
		t.Fatalf("read migration 155 down SQL: %v", err)
	}
	if _, err := fixture.pool.Exec(ctx, string(downMigration)); err == nil || !strings.Contains(err.Error(), "cannot be reversed after digest-only invitations") {
		t.Fatalf("migration 155 down error = %v, want forward-only guard", err)
	}
	assertCount(t, ctx, fixture.pool, "SELECT count(*) FROM workspace_invitations WHERE invitation_id = $1 AND token_digest IS NOT NULL", 1, command.Invitation.ID)
}

func assertInvitationUsedAt(t *testing.T, ctx context.Context, pool *pgxpool.Pool, invitationID uuid.UUID, wantUsed bool) {
	t.Helper()

	var usedAt *time.Time
	if err := pool.QueryRow(ctx, "SELECT used_at FROM workspace_invitations WHERE invitation_id = $1", invitationID).Scan(&usedAt); err != nil {
		t.Fatalf("read invitation used_at: %v", err)
	}
	if (usedAt != nil) != wantUsed {
		t.Fatalf("invitation used = %t, want %t", usedAt != nil, wantUsed)
	}
}

func assertCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := pool.QueryRow(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}
