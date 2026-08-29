//go:build integration

package mentionsrepository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryReplacesMentionsAtomicallyWithinWorkspace(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newMentionFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	if err := repository.SaveMentions(ctx, fixture.workspaceA, fixture.commentA, []uuid.UUID{fixture.memberA2, fixture.memberA1}); err != nil {
		t.Fatalf("save valid mentions: %v", err)
	}
	mentions, err := repository.GetMentions(ctx, fixture.workspaceA, fixture.commentA)
	if err != nil || !slices.Equal(mentions, sortedUUIDs(fixture.memberA1, fixture.memberA2)) {
		t.Fatalf("GetMentions() = %v, %v", mentions, err)
	}

	for name, deniedUser := range map[string]uuid.UUID{
		"cross workspace": fixture.memberB,
		"inactive":        fixture.inactiveA,
		"non member":      fixture.nonMemberA,
	} {
		t.Run(name, func(t *testing.T) {
			err := repository.SaveMentions(ctx, fixture.workspaceA, fixture.commentA, []uuid.UUID{fixture.memberA1, deniedUser})
			if !errors.Is(err, ErrMentionTargetDenied) {
				t.Fatalf("SaveMentions() error = %v", err)
			}
			got, getErr := repository.GetMentions(ctx, fixture.workspaceA, fixture.commentA)
			if getErr != nil || !slices.Equal(got, sortedUUIDs(fixture.memberA1, fixture.memberA2)) {
				t.Fatalf("failed replacement changed mentions = %v, %v", got, getErr)
			}
		})
	}

	if err := repository.SaveMentions(ctx, fixture.workspaceB, fixture.commentA, []uuid.UUID{fixture.memberB}); !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("cross-tenant SaveMentions() error = %v", err)
	}
	if _, err := repository.GetMentions(ctx, fixture.workspaceB, fixture.commentA); !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("cross-tenant GetMentions() error = %v", err)
	}
	if err := repository.DeleteMentions(ctx, fixture.workspaceA, fixture.commentA); err != nil {
		t.Fatalf("DeleteMentions() error = %v", err)
	}
	mentions, err = repository.GetMentions(ctx, fixture.workspaceA, fixture.commentA)
	if err != nil || len(mentions) != 0 {
		t.Fatalf("mentions after delete = %v, %v", mentions, err)
	}
}

func TestConcurrentMentionReplacementNeverPersistsPartialSet(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newMentionFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	sets := [][]uuid.UUID{{fixture.memberA1}, {fixture.memberA2, fixture.memberA1}}
	errorsByWriter := make(chan error, len(sets))
	for _, set := range sets {
		set := append([]uuid.UUID(nil), set...)
		go func() {
			errorsByWriter <- repository.SaveMentions(ctx, fixture.workspaceA, fixture.commentA, set)
		}()
	}
	for range sets {
		if err := <-errorsByWriter; err != nil {
			t.Fatalf("concurrent SaveMentions() error = %v", err)
		}
	}
	mentions, err := repository.GetMentions(ctx, fixture.workspaceA, fixture.commentA)
	if err != nil {
		t.Fatalf("GetMentions() error = %v", err)
	}
	if !slices.Equal(mentions, sortedUUIDs(fixture.memberA1)) &&
		!slices.Equal(mentions, sortedUUIDs(fixture.memberA1, fixture.memberA2)) {
		t.Fatalf("partial concurrent mention set = %v", mentions)
	}
}

type mentionFixture struct {
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	memberA1   uuid.UUID
	memberA2   uuid.UUID
	memberB    uuid.UUID
	inactiveA  uuid.UUID
	nonMemberA uuid.UUID
	commentA   uuid.UUID
}

func newMentionFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) mentionFixture {
	t.Helper()
	workspaceA := insertMentionWorkspace(t, ctx, pool, "a")
	workspaceB := insertMentionWorkspace(t, ctx, pool, "b")
	memberA1 := insertMentionUser(t, ctx, pool, "member-a1", true)
	memberA2 := insertMentionUser(t, ctx, pool, "member-a2", true)
	memberB := insertMentionUser(t, ctx, pool, "member-b", true)
	inactiveA := insertMentionUser(t, ctx, pool, "inactive-a", false)
	nonMemberA := insertMentionUser(t, ctx, pool, "non-member-a", true)
	insertMentionWorkspaceMember(t, ctx, pool, workspaceA, memberA1)
	insertMentionWorkspaceMember(t, ctx, pool, workspaceA, memberA2)
	insertMentionWorkspaceMember(t, ctx, pool, workspaceA, inactiveA)
	insertMentionWorkspaceMember(t, ctx, pool, workspaceB, memberB)
	teamA := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, 'Mentions A', $2, $3, '#000000')
	`, teamA, workspaceA, "M"+uuid.NewString()[:7]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	storyA := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO stories (id, sequence_id, team_id, title, workspace_id, reporter_id)
		VALUES ($1, 1, $2, 'Mentions story', $3, $4)
	`, storyA, teamA, workspaceA, memberA1); err != nil {
		t.Fatalf("insert story: %v", err)
	}
	commentA := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO story_comments (comment_id, content, story_id, commenter_id)
		VALUES ($1, 'Mention people', $2, $3)
	`, commentA, storyA, memberA1); err != nil {
		t.Fatalf("insert comment: %v", err)
	}
	return mentionFixture{
		workspaceA: workspaceA, workspaceB: workspaceB, memberA1: memberA1, memberA2: memberA2,
		memberB: memberB, inactiveA: inactiveA, nonMemberA: nonMemberA, commentA: commentA,
	}
}

func insertMentionWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`,
		id, "Mentions "+label, "mentions-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertMentionUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, FALSE)
	`, id, label+"-"+id.String(), fmt.Sprintf("%s-%s@example.com", label, id), "Mentions "+label, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertMentionWorkspaceMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST('member' AS user_role))
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func sortedUUIDs(values ...uuid.UUID) []uuid.UUID {
	result := append([]uuid.UUID(nil), values...)
	slices.SortFunc(result, func(left, right uuid.UUID) int {
		return slices.Compare(left[:], right[:])
	})
	return result
}
