//go:build integration

package commentsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCommentLifecycleIsTenantSafeAndTransactional(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	workspaceA, teamA, storyA := insertCommentTestWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB, teamB, _ := insertCommentTestWorkspace(t, ctx, postgres.Pool, "b")
	authorA := insertCommentTestUser(t, ctx, postgres.Pool, workspaceA, "author-a")
	memberA := insertCommentTestUser(t, ctx, postgres.Pool, workspaceA, "member-a")
	memberB := insertCommentTestUser(t, ctx, postgres.Pool, workspaceB, "member-b")
	insertCommentWebhookEndpoint(t, ctx, postgres.Pool, workspaceA, authorA)

	log := logger.NewWithText(io.Discard, slog.LevelError, "comments-integration-test")
	repository := New(log, postgres.Pool)
	service := comments.New(repository)
	authorActor := commentIntegrationActor(t, authorA, workspaceA)
	memberActor := commentIntegrationActor(t, memberA, workspaceA)
	memberBActor := commentIntegrationActor(t, memberB, workspaceB)

	if _, err := service.CreateComment(ctx, comments.CreateCommentCommand{
		WorkspaceID: workspaceB, StoryID: storyA, Actor: memberBActor, Content: "cross tenant",
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("cross-workspace create error = %v, want not found", err)
	}
	restrictedActor := restrictedCommentActor(t, authorA, workspaceA, teamB)
	if _, err := service.CreateComment(ctx, comments.CreateCommentCommand{
		WorkspaceID: workspaceA, StoryID: storyA, Actor: restrictedActor, Content: "wrong team",
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("team-restricted create error = %v, want not found", err)
	}

	created, err := service.CreateComment(ctx, comments.CreateCommentCommand{
		WorkspaceID: workspaceA, StoryID: storyA, Actor: authorActor,
		Content: "Original private comment", MentionedUserIDs: []uuid.UUID{memberA, memberA},
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	assertCommentState(t, ctx, postgres.Pool, created.ID, "Original private comment", []uuid.UUID{memberA})
	assertCommentEventContract(t, ctx, postgres.Pool, created.ID, "comment.created", 1)

	if err := service.UpdateComment(ctx, comments.UpdateCommentCommand{
		Scope:   comments.AuthorScope{CommentID: created.ID, WorkspaceID: workspaceA, Actor: memberActor},
		Content: "Unauthorized",
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("non-author update error = %v, want not found", err)
	}
	if err := service.UpdateComment(ctx, comments.UpdateCommentCommand{
		Scope:   comments.AuthorScope{CommentID: created.ID, WorkspaceID: workspaceB, Actor: memberBActor},
		Content: "Cross tenant",
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("cross-workspace update error = %v, want not found", err)
	}
	assertCommentState(t, ctx, postgres.Pool, created.ID, "Original private comment", []uuid.UUID{memberA})

	err = service.UpdateComment(ctx, comments.UpdateCommentCommand{
		Scope:   comments.AuthorScope{CommentID: created.ID, WorkspaceID: workspaceA, Actor: authorActor},
		Content: "Must roll back", MentionedUserIDs: []uuid.UUID{memberA, memberB},
	})
	if !errors.Is(err, comments.ErrInvalidMention) {
		t.Fatalf("cross-workspace mention error = %v, want invalid mention", err)
	}
	assertCommentState(t, ctx, postgres.Pool, created.ID, "Original private comment", []uuid.UUID{memberA})
	assertCommentEventContract(t, ctx, postgres.Pool, created.ID, "comment.updated", 0)

	if err := service.UpdateComment(ctx, comments.UpdateCommentCommand{
		Scope:   comments.AuthorScope{CommentID: created.ID, WorkspaceID: workspaceA, Actor: authorActor},
		Content: "Updated private comment", MentionedUserIDs: []uuid.UUID{authorA},
	}); err != nil {
		t.Fatalf("same-workspace update: %v", err)
	}
	assertCommentState(t, ctx, postgres.Pool, created.ID, "Updated private comment", []uuid.UUID{authorA})
	assertCommentEventContract(t, ctx, postgres.Pool, created.ID, "comment.updated", 1)

	if _, err := postgres.Pool.Exec(ctx, `
		DELETE FROM public.workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceA, authorA); err != nil {
		t.Fatalf("revoke author membership: %v", err)
	}
	if err := service.UpdateComment(ctx, comments.UpdateCommentCommand{
		Scope:   comments.AuthorScope{CommentID: created.ID, WorkspaceID: workspaceA, Actor: authorActor},
		Content: "Revoked update",
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("revoked author update error = %v, want not found", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO public.workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, workspaceA, authorA); err != nil {
		t.Fatalf("restore author membership: %v", err)
	}

	if err := service.DeleteComment(ctx, comments.AuthorScope{
		CommentID: created.ID, WorkspaceID: workspaceA, Actor: memberActor,
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("non-author delete error = %v, want not found", err)
	}
	if err := service.DeleteComment(ctx, comments.AuthorScope{
		CommentID: created.ID, WorkspaceID: workspaceA, Actor: authorActor,
	}); err != nil {
		t.Fatalf("author delete: %v", err)
	}
	if _, err := service.GetComment(ctx, comments.GetCommentQuery{
		CommentID: created.ID, WorkspaceID: workspaceA,
	}); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("deleted comment get error = %v, want not found", err)
	}
	assertCommentEventContract(t, ctx, postgres.Pool, created.ID, "comment.deleted", 1)
	if teamA == uuid.Nil {
		t.Fatal("workspace fixture returned a zero team")
	}
}

func TestCommentOutboxConflictRollsBackMutation(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "rollback")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "rollback-author")
	commentID := insertCommentTestComment(t, ctx, postgres.Pool, storyID, authorID, "before")
	actor := commentIntegrationActor(t, authorID, workspaceID)
	repository := New(logger.NewWithText(io.Discard, slog.LevelError, "comments-rollback-test"), postgres.Pool)
	eventID := uuid.New()

	first := commentsdomain.UpdateCommand{
		Scope:   commentsdomain.ActorScope{CommentID: commentID, WorkspaceID: workspaceID, Actor: actor},
		Content: "first committed",
	}
	if err := repository.UpdateComment(ctx, first, eventID); err != nil {
		t.Fatalf("first update: %v", err)
	}
	second := first
	second.Content = "must roll back"
	if err := repository.UpdateComment(ctx, second, eventID); err == nil {
		t.Fatal("duplicate event id update error = nil")
	}
	assertCommentState(t, ctx, postgres.Pool, commentID, "first committed", nil)
	assertCommentEventContract(t, ctx, postgres.Pool, commentID, "comment.updated", 1)
}

func TestConcurrentCommentUpdatesSerializeWithoutLosingEvents(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "concurrent")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "concurrent-author")
	commentID := insertCommentTestComment(t, ctx, postgres.Pool, storyID, authorID, "before")
	actor := commentIntegrationActor(t, authorID, workspaceID)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "comments-concurrency-test"), postgres.Pool))

	contents := []string{"concurrent one", "concurrent two"}
	errorsByUpdate := make([]error, len(contents))
	var wait sync.WaitGroup
	start := make(chan struct{})
	for index, content := range contents {
		wait.Add(1)
		go func(index int, content string) {
			defer wait.Done()
			<-start
			errorsByUpdate[index] = service.UpdateComment(ctx, comments.UpdateCommentCommand{
				Scope:   comments.AuthorScope{CommentID: commentID, WorkspaceID: workspaceID, Actor: actor},
				Content: content,
			})
		}(index, content)
	}
	close(start)
	wait.Wait()
	for index, err := range errorsByUpdate {
		if err != nil {
			t.Fatalf("concurrent update %d: %v", index, err)
		}
	}

	var storedContent string
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT content FROM public.story_comments WHERE comment_id = $1
	`, commentID).Scan(&storedContent); err != nil {
		t.Fatalf("read concurrent result: %v", err)
	}
	if storedContent != contents[0] && storedContent != contents[1] {
		t.Fatalf("stored concurrent content = %q", storedContent)
	}
	assertCommentEventContract(t, ctx, postgres.Pool, commentID, "comment.updated", 2)
}

func insertCommentTestWorkspace(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	label string,
) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()

	workspaceID := uuid.New()
	teamID := uuid.New()
	storyID := uuid.New()
	suffix := uuid.NewString()
	if _, err := db.Exec(ctx, `
		INSERT INTO public.workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, workspaceID, "Comments test "+label, "comments-test-"+suffix); err != nil {
		t.Fatalf("insert workspace %s: %v", label, err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO public.teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, teamID, "Comments test "+label, workspaceID, "CMT-"+label+suffix[:6]); err != nil {
		t.Fatalf("insert team %s: %v", label, err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO public.stories (id, team_id, title, workspace_id)
		VALUES ($1, $2, $3, $4)
	`, storyID, teamID, "Comments test "+label, workspaceID); err != nil {
		t.Fatalf("insert story %s: %v", label, err)
	}
	return workspaceID, teamID, storyID
}

func insertCommentTestUser(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	workspaceID uuid.UUID,
	label string,
) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	suffix := uuid.NewString()
	if _, err := db.Exec(ctx, `
		INSERT INTO public.users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, $4)
	`, userID, label+"-"+suffix, label+"-"+suffix+"@example.com", "Comments "+label); err != nil {
		t.Fatalf("insert user %s: %v", label, err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO public.workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member %s: %v", label, err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO public.team_members (team_id, user_id)
		SELECT team_id, $2
		FROM public.teams
		WHERE workspace_id = $1
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert team member %s: %v", label, err)
	}
	return userID
}

func insertCommentTestComment(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	storyID uuid.UUID,
	authorID uuid.UUID,
	content string,
) uuid.UUID {
	t.Helper()

	commentID := uuid.New()
	if _, err := db.Exec(ctx, `
		INSERT INTO public.story_comments (comment_id, content, story_id, commenter_id)
		VALUES ($1, $2, $3, $4)
	`, commentID, content, storyID, authorID); err != nil {
		t.Fatalf("insert comment: %v", err)
	}
	return commentID
}

func insertCommentWebhookEndpoint(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	workspaceID, ownerUserID uuid.UUID,
) {
	t.Helper()

	now := time.Now().UTC()
	principalID := uuid.New()
	endpointID := uuid.New()
	if _, err := db.Exec(ctx, `
		INSERT INTO public.principals (
			principal_id, workspace_id, kind, name, subject_user_id,
			status, created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, 'human_user', 'Comment webhook owner', $3, 'active', $3, $4, $4)
	`, principalID, workspaceID, ownerUserID, now); err != nil {
		t.Fatalf("insert webhook principal: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO public.outbound_webhook_endpoints (
			endpoint_id, workspace_id, owner_principal_id, name, endpoint_url,
			status, signing_secret_envelope, created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, 'Comment sink', 'https://example.com/fortyone',
			'active', $4, $5, $6, $6)
	`, endpointID, workspaceID, principalID, strings.Repeat("x", 40), ownerUserID, now); err != nil {
		t.Fatalf("insert webhook endpoint: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO public.outbound_webhook_subscriptions (
			endpoint_id, workspace_id, event_type, created_at
		) VALUES
			($1, $2, 'comment.created', $3),
			($1, $2, 'comment.updated', $3),
			($1, $2, 'comment.deleted', $3)
	`, endpointID, workspaceID, now); err != nil {
		t.Fatalf("insert webhook subscriptions: %v", err)
	}
}

func assertCommentState(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	commentID uuid.UUID,
	wantContent string,
	wantMentions []uuid.UUID,
) {
	t.Helper()

	var content string
	if err := db.QueryRow(ctx, `
		SELECT content FROM public.story_comments WHERE comment_id = $1
	`, commentID).Scan(&content); err != nil {
		t.Fatalf("get stored comment: %v", err)
	}
	if content != wantContent {
		t.Fatalf("stored content = %q, want %q", content, wantContent)
	}

	rows, err := db.Query(ctx, `
		SELECT mentioned_user_id
		FROM public.comment_mentions
		WHERE comment_id = $1
		ORDER BY mentioned_user_id
	`, commentID)
	if err != nil {
		t.Fatalf("get stored mentions: %v", err)
	}
	defer rows.Close()
	mentions := make([]uuid.UUID, 0, len(wantMentions))
	for rows.Next() {
		var mentionID uuid.UUID
		if err := rows.Scan(&mentionID); err != nil {
			t.Fatalf("scan stored mention: %v", err)
		}
		mentions = append(mentions, mentionID)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate stored mentions: %v", err)
	}
	if len(mentions) != len(wantMentions) {
		t.Fatalf("stored mentions = %v, want %v", mentions, wantMentions)
	}
	for _, wantMention := range wantMentions {
		found := false
		for _, mention := range mentions {
			found = found || mention == wantMention
		}
		if !found {
			t.Fatalf("stored mentions = %v, missing %s", mentions, wantMention)
		}
	}
}

func assertCommentEventContract(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	commentID uuid.UUID,
	eventType string,
	wantCount int,
) {
	t.Helper()

	rows, err := db.Query(ctx, `
		SELECT event.payload, delivery.payload_body
		FROM public.outbound_webhook_events AS event
		LEFT JOIN public.outbound_webhook_deliveries AS delivery
			ON delivery.event_id = event.event_id
		WHERE event.subject_id = $1
		  AND event.subject_type = 'comment'
		  AND event.event_type = $2
		ORDER BY event.created_at, event.event_id
	`, commentID, eventType)
	if err != nil {
		t.Fatalf("query %s events: %v", eventType, err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var payload []byte
		var body []byte
		if err := rows.Scan(&payload, &body); err != nil {
			t.Fatalf("scan %s event: %v", eventType, err)
		}
		count++
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatalf("decode %s payload: %v", eventType, err)
		}
		if len(fields) != 3 || fields["comment_id"] == nil || fields["story_id"] == nil || fields["parent_id"] == nil {
			t.Fatalf("%s payload fields = %v", eventType, fields)
		}
		encoded := string(payload) + string(body)
		for _, forbidden := range []string{"content", "mention", "email", "token", "secret", "Original private comment", "Updated private comment"} {
			if strings.Contains(encoded, forbidden) {
				t.Fatalf("%s event exposed forbidden value %q: %s", eventType, forbidden, encoded)
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s events: %v", eventType, err)
	}
	if count != wantCount {
		t.Fatalf("%s event count = %d, want %d", eventType, count, wantCount)
	}
}

func commentIntegrationActor(t *testing.T, userID, workspaceID uuid.UUID) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewHumanActor(userID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind comment actor: %v", err)
	}
	return actor
}

func restrictedCommentActor(
	t *testing.T,
	userID, workspaceID uuid.UUID,
	allowedTeamID uuid.UUID,
) platformauth.Actor {
	t.Helper()
	teamAccess, err := platformauth.RestrictedTeamAccess(allowedTeamID)
	if err != nil {
		t.Fatalf("create restricted team access: %v", err)
	}
	actor, err := platformauth.NewActor(
		userID,
		platformauth.PrincipalHumanUser,
		uuid.Nil,
		platformauth.MustScopeSet(platformauth.ScopeFirstParty),
		teamAccess,
	)
	if err != nil {
		t.Fatalf("create restricted actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind restricted actor: %v", err)
	}
	return actor
}
