//go:build integration

package linksrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	linksdomain "github.com/complexus-tech/projects-api/internal/modules/links/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestLinkMutationsEnforceWorkspaceOwnership(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tx, err := postgres.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin test transaction: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	workspaceA, storyA, actorA := insertLinkTestStory(t, ctx, tx, "a")
	workspaceB, _, actorB := insertLinkTestStory(t, ctx, tx, "b")
	repository := New(
		logger.NewWithText(io.Discard, slog.LevelError, "links-integration-test"),
		tx,
	)

	if _, err := repository.CreateLink(ctx, actorB, linksdomain.CreateLink{
		URL:         "https://example.com/rejected",
		StoryID:     storyA,
		WorkspaceID: workspaceB,
	}); !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("cross-workspace create error = %v, want linksdomain.ErrNotFound", err)
	}

	oversizedValue := strings.Repeat("x", 256)
	t.Run("create rejects oversized title", func(t *testing.T) {
		withRollbackRepository(t, ctx, tx, func(repository *repo) {
			_, err := repository.CreateLink(ctx, actorA, linksdomain.CreateLink{
				Title:       &oversizedValue,
				URL:         "https://example.com/valid",
				StoryID:     storyA,
				WorkspaceID: workspaceA,
			})
			assertStringDataRightTruncation(t, err)
		})
	})
	t.Run("create rejects oversized URL", func(t *testing.T) {
		withRollbackRepository(t, ctx, tx, func(repository *repo) {
			_, err := repository.CreateLink(ctx, actorA, linksdomain.CreateLink{
				URL:         oversizedValue,
				StoryID:     storyA,
				WorkspaceID: workspaceA,
			})
			assertStringDataRightTruncation(t, err)
		})
	})

	title := "Architecture notes"
	created, err := repository.CreateLink(ctx, actorA, linksdomain.CreateLink{
		Title:       &title,
		URL:         "https://example.com/original",
		StoryID:     storyA,
		WorkspaceID: workspaceA,
	})
	if err != nil {
		t.Fatalf("same-workspace create: %v", err)
	}

	t.Run("update rejects oversized title", func(t *testing.T) {
		withRollbackRepository(t, ctx, tx, func(repository *repo) {
			err := repository.UpdateLink(
				ctx,
				actorA,
				created.LinkID,
				workspaceA,
				linksdomain.UpdateLink{Title: &oversizedValue},
			)
			assertStringDataRightTruncation(t, err)
		})
	})
	t.Run("update rejects oversized URL", func(t *testing.T) {
		withRollbackRepository(t, ctx, tx, func(repository *repo) {
			err := repository.UpdateLink(
				ctx,
				actorA,
				created.LinkID,
				workspaceA,
				linksdomain.UpdateLink{URL: &oversizedValue},
			)
			assertStringDataRightTruncation(t, err)
		})
	})
	assertStoredLink(t, ctx, tx, created.LinkID, title, "https://example.com/original")

	rejectedTitle := "Cross-tenant overwrite"
	if err := repository.UpdateLink(
		ctx,
		actorB,
		created.LinkID,
		workspaceB,
		linksdomain.UpdateLink{Title: &rejectedTitle},
	); !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("cross-workspace update error = %v, want linksdomain.ErrNotFound", err)
	}
	assertStoredLink(t, ctx, tx, created.LinkID, title, "https://example.com/original")

	updatedURL := "https://example.com/updated"
	if err := repository.UpdateLink(
		ctx,
		actorA,
		created.LinkID,
		workspaceA,
		linksdomain.UpdateLink{URL: &updatedURL},
	); err != nil {
		t.Fatalf("same-workspace update: %v", err)
	}
	assertStoredLink(t, ctx, tx, created.LinkID, title, updatedURL)

	if err := repository.DeleteLink(ctx, actorB, created.LinkID, workspaceB); !errors.Is(err, linksdomain.ErrNotFound) {
		t.Fatalf("cross-workspace delete error = %v, want linksdomain.ErrNotFound", err)
	}
	assertStoredLink(t, ctx, tx, created.LinkID, title, updatedURL)

	if err := repository.DeleteLink(ctx, actorA, created.LinkID, workspaceA); err != nil {
		t.Fatalf("same-workspace delete: %v", err)
	}
	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM story_links WHERE link_id = $1)", created.LinkID).Scan(&exists); err != nil {
		t.Fatalf("check deleted link: %v", err)
	}
	if exists {
		t.Fatal("same-workspace delete left the link behind")
	}
}

func withRollbackRepository(
	t *testing.T,
	ctx context.Context,
	parent pgx.Tx,
	test func(*repo),
) {
	t.Helper()

	tx, err := parent.Begin(ctx)
	if err != nil {
		t.Fatalf("begin isolated link test transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			t.Errorf("rollback isolated link test transaction: %v", rollbackErr)
		}
	}()

	test(New(
		logger.NewWithText(io.Discard, slog.LevelError, "links-integration-test"),
		tx,
	))
}

func assertStringDataRightTruncation(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("oversized link value was accepted; want PostgreSQL length error")
	}
	var databaseErr *pgconn.PgError
	if !errors.As(err, &databaseErr) {
		t.Fatalf("oversized link error = %v, want PostgreSQL error", err)
	}
	if databaseErr.Code != "22001" {
		t.Fatalf("oversized link SQLSTATE = %s, want 22001", databaseErr.Code)
	}
}

func insertLinkTestStory(t *testing.T, ctx context.Context, tx pgx.Tx, label string) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()

	workspaceID := uuid.New()
	teamID := uuid.New()
	storyID := uuid.New()
	actorID := uuid.New()
	suffix := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, $4)
	`, actorID, "links-actor-"+suffix, "links-actor-"+suffix+"@example.com", "Links test actor "+label); err != nil {
		t.Fatalf("insert actor %s: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, workspaceID, "Links test "+label, "links-test-"+suffix); err != nil {
		t.Fatalf("insert workspace %s: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, workspaceID, actorID); err != nil {
		t.Fatalf("insert workspace membership %s: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, teamID, "Links test "+label, workspaceID, "LNK-"+label+suffix[:6]); err != nil {
		t.Fatalf("insert team %s: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO stories (id, team_id, title, workspace_id)
		VALUES ($1, $2, $3, $4)
	`, storyID, teamID, "Links test "+label, workspaceID); err != nil {
		t.Fatalf("insert story %s: %v", label, err)
	}
	return workspaceID, storyID, actorID
}

func assertStoredLink(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	linkID uuid.UUID,
	wantTitle string,
	wantURL string,
) {
	t.Helper()

	var title *string
	var url string
	if err := tx.QueryRow(
		ctx,
		"SELECT title, url FROM story_links WHERE link_id = $1",
		linkID,
	).Scan(&title, &url); err != nil {
		t.Fatalf("read stored link: %v", err)
	}
	if title == nil || *title != wantTitle || url != wantURL {
		t.Fatalf("stored link = %v/%q, want %q/%q", title, url, wantTitle, wantURL)
	}
}
