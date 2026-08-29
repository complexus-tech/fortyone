//go:build integration

package storiesrepository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStoryCommentReadsPreserveTreeShapeAndLiveTenantVisibility(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 15, 0, 0, 0, time.UTC)
	storyID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime)
	otherStoryID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(time.Minute))

	olderRootID, newestRootID, replyID := uuid.New(), uuid.New(), uuid.New()
	mustMutationExec(
		t,
		ctx,
		postgres.Pool,
		`INSERT INTO story_comments (
			comment_id, content, story_id, commenter_id, created_at, updated_at, parent_id
		) VALUES
			($1, 'Older root', $2, $3, $4, $4, NULL),
			($5, 'Newest root', $2, $3, $6, $6, NULL),
			($7, 'Reply', $2, $8, $9, $9, $5)`,
		olderRootID,
		storyID,
		fixture.actorID,
		baseTime.Add(2*time.Minute),
		newestRootID,
		baseTime.Add(3*time.Minute),
		replyID,
		fixture.assigneeID,
		baseTime.Add(4*time.Minute),
	)
	assertStoryCommentTreeIndexPlans(t, ctx, postgres.Pool, storyID, newestRootID)

	scope := storydomain.ReadScope{
		ActorID: fixture.actorID, WorkspaceID: fixture.workspaceID, UnrestrictedTeamAccess: true,
	}
	comments, hasMore, err := repository.ListVisibleComments(ctx, scope, storyID, 1, 1)
	if err != nil {
		t.Fatalf("list first visible comment page: %v", err)
	}
	if !hasMore || len(comments) != 1 || comments[0].ID != newestRootID {
		t.Fatalf("first comment page = %#v hasMore=%v", comments, hasMore)
	}
	if len(comments[0].SubComments) != 1 || comments[0].SubComments[0].ID != replyID ||
		comments[0].SubComments[0].Parent == nil || *comments[0].SubComments[0].Parent != newestRootID {
		t.Fatalf("comment reply tree = %#v", comments[0].SubComments)
	}

	secondPage, hasMore, err := repository.ListVisibleComments(ctx, scope, storyID, 2, 1)
	if err != nil {
		t.Fatalf("list second visible comment page: %v", err)
	}
	if hasMore || len(secondPage) != 1 || secondPage[0].ID != olderRootID || len(secondPage[0].SubComments) != 0 {
		t.Fatalf("second comment page = %#v hasMore=%v", secondPage, hasMore)
	}

	parentNotificationComment, err := repository.GetVisibleComment(ctx, scope, newestRootID, storyID)
	if err != nil || parentNotificationComment.UserID != fixture.actorID {
		t.Fatalf("load parent comment for notification: comment=%#v error=%v", parentNotificationComment, err)
	}
	visibleReply, err := repository.GetVisibleComment(ctx, scope, replyID, storyID)
	if err != nil || visibleReply.Parent == nil || *visibleReply.Parent != newestRootID {
		t.Fatalf("load visible reply: comment=%#v error=%v", visibleReply, err)
	}

	if _, err := repository.GetVisibleComment(ctx, scope, newestRootID, otherStoryID); !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("cross-story parent lookup error = %v, want not found", err)
	}
	restrictedScope := scope
	restrictedScope.UnrestrictedTeamAccess = false
	restrictedScope.AllowedTeamIDs = []uuid.UUID{fixture.otherTeamID}
	if _, err := repository.GetVisibleComment(ctx, restrictedScope, newestRootID, storyID); !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("credential-restricted comment lookup error = %v, want not found", err)
	}
	if hidden, _, err := repository.ListVisibleComments(ctx, restrictedScope, storyID, 1, 10); err != nil || len(hidden) != 0 {
		t.Fatalf("credential-restricted comment list = %#v error=%v", hidden, err)
	}

	foreignActorScope := scope
	foreignActorScope.ActorID = fixture.foreignActorID
	if _, err := repository.GetVisibleComment(ctx, foreignActorScope, newestRootID, storyID); !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("cross-tenant actor comment lookup error = %v, want not found", err)
	}

	mustMutationExec(
		t, ctx, postgres.Pool,
		"DELETE FROM team_members WHERE team_id = $1 AND user_id = $2",
		fixture.teamID, fixture.actorID,
	)
	if hidden, _, err := repository.ListVisibleComments(ctx, scope, storyID, 1, 10); err != nil || len(hidden) != 0 {
		t.Fatalf("revoked actor comment list = %#v error=%v", hidden, err)
	}
	if _, err := repository.GetVisibleComment(ctx, scope, newestRootID, storyID); !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("revoked actor comment lookup error = %v, want not found", err)
	}
}

func assertStoryCommentTreeIndexPlans(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	storyID uuid.UUID,
	parentID uuid.UUID,
) {
	t.Helper()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin comment index-plan transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
			t.Errorf("roll back comment index-plan transaction: %v", rollbackErr)
		}
	}()
	if _, err := tx.Exec(ctx, "SET LOCAL enable_seqscan = off"); err != nil {
		t.Fatalf("disable sequential scans for deterministic index contract: %v", err)
	}

	tests := []struct {
		name      string
		indexName string
		query     string
		args      []any
	}{
		{
			name:      "root pagination",
			indexName: "idx_story_comments_roots_page",
			query: `EXPLAIN (COSTS OFF)
				SELECT comment_id
				FROM public.story_comments
				WHERE story_id = $1 AND parent_id IS NULL
				ORDER BY created_at DESC, comment_id DESC
				LIMIT 101`,
			args: []any{storyID},
		},
		{
			name:      "reply hydration",
			indexName: "idx_story_comments_replies_page",
			query: `EXPLAIN (COSTS OFF)
				SELECT comment_id
				FROM public.story_comments
				WHERE story_id = $1
				  AND parent_id = ANY(CAST($2 AS uuid[]))
				ORDER BY parent_id, created_at, comment_id`,
			args: []any{storyID, []uuid.UUID{parentID}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows, err := tx.Query(ctx, test.query, test.args...)
			if err != nil {
				t.Fatalf("explain comment query: %v", err)
			}

			var plan strings.Builder
			for rows.Next() {
				var line string
				if err := rows.Scan(&line); err != nil {
					rows.Close()
					t.Fatalf("scan comment query plan: %v", err)
				}
				plan.WriteString(line)
				plan.WriteByte('\n')
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				t.Fatalf("read comment query plan: %v", err)
			}
			if !strings.Contains(plan.String(), test.indexName) {
				t.Fatalf("query plan does not use %s:\n%s", test.indexName, plan.String())
			}
		})
	}
}
