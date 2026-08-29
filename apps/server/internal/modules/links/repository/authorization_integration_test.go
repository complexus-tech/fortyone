//go:build integration

package linksrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	linksdomain "github.com/complexus-tech/projects-api/internal/modules/links/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type linkAuthorizationFixture struct {
	pool       *pgxpool.Pool
	repository *repo
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	storyID    uuid.UUID
	actorID    uuid.UUID
	linkID     uuid.UUID
}

func newLinkAuthorizationFixture(t *testing.T) linkAuthorizationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := linkAuthorizationFixture{
		pool:       postgres.Pool,
		repository: New(logger.NewWithText(io.Discard, slog.LevelError, "links-authorization-test"), postgres.Pool),
		workspaceA: uuid.New(),
		workspaceB: uuid.New(),
		storyID:    uuid.New(),
		actorID:    uuid.New(),
	}
	teamID := uuid.New()
	suffix := uuid.NewString()
	ctx := t.Context()
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Links Actor')
	`, fixture.actorID, "links-auth-"+suffix, "links-auth-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert links actor: %v", err)
	}
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, 'Links Workspace A', $2), ($3, 'Links Workspace B', $4)
	`, fixture.workspaceA, "links-auth-a-"+suffix, fixture.workspaceB, "links-auth-b-"+suffix); err != nil {
		t.Fatalf("insert links workspaces: %v", err)
	}
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, fixture.workspaceA, fixture.actorID); err != nil {
		t.Fatalf("insert links membership: %v", err)
	}
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, 'Links Team', $2, $3, '#000000')
	`, teamID, fixture.workspaceA, "LA-"+suffix[:8]); err != nil {
		t.Fatalf("insert links team: %v", err)
	}
	if _, err := fixture.pool.Exec(ctx, `
		INSERT INTO stories (id, team_id, title, workspace_id)
		VALUES ($1, $2, 'Links Story', $3)
	`, fixture.storyID, teamID, fixture.workspaceA); err != nil {
		t.Fatalf("insert links story: %v", err)
	}
	title := "Original link"
	created, err := fixture.repository.CreateLink(ctx, fixture.actorID, linksdomain.CreateLink{
		Title:       &title,
		URL:         "https://example.com/original",
		StoryID:     fixture.storyID,
		WorkspaceID: fixture.workspaceA,
	})
	if err != nil {
		t.Fatalf("create baseline link: %v", err)
	}
	fixture.linkID = created.LinkID
	return fixture
}

func TestLinkMutationsFailClosedForUnauthorizedActorStates(t *testing.T) {
	fixture := newLinkAuthorizationFixture(t)
	states := []struct {
		name                string
		active              bool
		membershipWorkspace uuid.UUID
		role                string
	}{
		{name: "inactive", active: false, membershipWorkspace: fixture.workspaceA, role: "member"},
		{name: "removed", active: true},
		{name: "guest", active: true, membershipWorkspace: fixture.workspaceA, role: "guest"},
		{name: "cross tenant", active: true, membershipWorkspace: fixture.workspaceB, role: "admin"},
	}

	for _, state := range states {
		state := state
		t.Run(state.name, func(t *testing.T) {
			actorID := insertLinkAuthorizationActor(t, fixture.pool, state.name, state.active, state.membershipWorkspace, state.role)
			_, err := fixture.repository.CreateLink(t.Context(), actorID, linksdomain.CreateLink{
				URL:         "https://example.com/denied-" + state.name,
				StoryID:     fixture.storyID,
				WorkspaceID: fixture.workspaceA,
			})
			assertLinkNotFound(t, err, "create")

			title := "Unauthorized overwrite"
			err = fixture.repository.UpdateLink(t.Context(), actorID, fixture.linkID, fixture.workspaceA, linksdomain.UpdateLink{Title: &title})
			assertLinkNotFound(t, err, "update")

			err = fixture.repository.DeleteLink(t.Context(), actorID, fixture.linkID, fixture.workspaceA)
			assertLinkNotFound(t, err, "delete")
			assertAuthorizedLinkState(t, fixture, "Original link", "https://example.com/original", 1)
		})
	}
}

func TestLinkMutationAuthorizationLinearizesConcurrentDemotionAndRemoval(t *testing.T) {
	fixture := newLinkAuthorizationFixture(t)
	ctx := t.Context()

	demotionTx := lockLinkMembership(t, ctx, fixture.pool, fixture.workspaceA, fixture.actorID)
	createResult := make(chan error, 1)
	go func() {
		_, err := fixture.repository.CreateLink(ctx, fixture.actorID, linksdomain.CreateLink{
			URL:         "https://example.com/concurrent-demotion",
			StoryID:     fixture.storyID,
			WorkspaceID: fixture.workspaceA,
		})
		createResult <- err
	}()
	waitForBlockedLinkQuery(t, ctx, fixture.pool, "CreateLinkForWorkspace")
	if _, err := demotionTx.Exec(ctx, `
		UPDATE workspace_members SET role = 'guest'
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.actorID); err != nil {
		t.Fatalf("demote links actor: %v", err)
	}
	if err := demotionTx.Commit(ctx); err != nil {
		t.Fatalf("commit links actor demotion: %v", err)
	}
	assertLinkNotFound(t, <-createResult, "create after concurrent demotion")
	assertAuthorizedLinkState(t, fixture, "Original link", "https://example.com/original", 1)

	if _, err := fixture.pool.Exec(ctx, `
		UPDATE workspace_members SET role = 'member'
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.actorID); err != nil {
		t.Fatalf("restore links actor role: %v", err)
	}
	removalTx := lockLinkMembership(t, ctx, fixture.pool, fixture.workspaceA, fixture.actorID)
	updateResult := make(chan error, 1)
	updatedTitle := "Concurrent unauthorized update"
	go func() {
		updateResult <- fixture.repository.UpdateLink(
			ctx,
			fixture.actorID,
			fixture.linkID,
			fixture.workspaceA,
			linksdomain.UpdateLink{Title: &updatedTitle},
		)
	}()
	waitForBlockedLinkQuery(t, ctx, fixture.pool, "UpdateLinkForWorkspace")
	if _, err := removalTx.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.actorID); err != nil {
		t.Fatalf("remove links actor membership: %v", err)
	}
	if err := removalTx.Commit(ctx); err != nil {
		t.Fatalf("commit links actor removal: %v", err)
	}
	assertLinkNotFound(t, <-updateResult, "update after concurrent removal")
	assertAuthorizedLinkState(t, fixture, "Original link", "https://example.com/original", 1)
}

func insertLinkAuthorizationActor(
	t *testing.T,
	pool *pgxpool.Pool,
	label string,
	active bool,
	membershipWorkspace uuid.UUID,
	role string,
) uuid.UUID {
	t.Helper()
	actorID := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, actorID, "links-state-"+suffix, "links-state-"+suffix+"@example.com", label, active); err != nil {
		t.Fatalf("insert %s links actor: %v", label, err)
	}
	if membershipWorkspace != uuid.Nil {
		if _, err := pool.Exec(t.Context(), `
			INSERT INTO workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, $3)
		`, membershipWorkspace, actorID, role); err != nil {
			t.Fatalf("insert %s links actor membership: %v", label, err)
		}
	}
	return actorID
}

func lockLinkMembership(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
) pgx.Tx {
	t.Helper()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin membership lock transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	var role string
	if err := tx.QueryRow(ctx, `
		SELECT CAST(role AS text)
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
		FOR UPDATE
	`, workspaceID, actorID).Scan(&role); err != nil {
		t.Fatalf("lock links membership: %v", err)
	}
	return tx
}

func waitForBlockedLinkQuery(t *testing.T, ctx context.Context, pool *pgxpool.Pool, queryName string) {
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
				  AND query LIKE '%' || $1 || '%'
				  AND wait_event_type = 'Lock'
			)
		`, queryName).Scan(&blocked); err != nil {
			t.Fatalf("inspect blocked link mutation: %v", err)
		}
		if blocked {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for blocked link mutation: %v", ctx.Err())
		case <-deadline.C:
			t.Fatalf("link mutation %q did not block on the authorization row", queryName)
		case <-ticker.C:
		}
	}
}

func assertLinkNotFound(t *testing.T, err error, operation string) {
	t.Helper()
	if !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("%s error = %v, want enumeration-safe ErrNotFound", operation, err)
	}
}

func assertAuthorizedLinkState(
	t *testing.T,
	fixture linkAuthorizationFixture,
	wantTitle string,
	wantURL string,
	wantCount int,
) {
	t.Helper()
	var title *string
	var url string
	if err := fixture.pool.QueryRow(t.Context(), `
		SELECT title, url FROM story_links WHERE link_id = $1
	`, fixture.linkID).Scan(&title, &url); err != nil {
		t.Fatalf("read protected link: %v", err)
	}
	if title == nil || *title != wantTitle || url != wantURL {
		t.Fatalf("protected link = %v/%q, want %q/%q", title, url, wantTitle, wantURL)
	}
	var count int
	if err := fixture.pool.QueryRow(t.Context(), "SELECT count(*) FROM story_links").Scan(&count); err != nil {
		t.Fatalf("count protected links: %v", err)
	}
	if count != wantCount {
		t.Fatalf("link count = %d, want %d", count, wantCount)
	}
}
