package commentsrepository

import (
	"os"
	"strings"
	"testing"
)

func TestCommentQueriesRetainWorkspaceAndAuthorScope(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/comments.sql")
	if err != nil {
		t.Fatalf("read comments queries: %v", err)
	}

	queryFile := string(source)
	tests := []struct {
		name  string
		start string
		end   string
		parts []string
	}{
		{
			name:  "get",
			start: "-- name: GetCommentForWorkspace :one",
			end:   "-- name: CreateCommentForActor :one",
			parts: []string{
				"INNER JOIN public.stories AS story ON story.id = comment.story_id",
				"story.workspace_id = sqlc.arg(workspace_id)",
			},
		},
		{
			name:  "create",
			start: "-- name: CreateCommentForActor :one",
			end:   "-- name: UpdateCommentForAuthor :one",
			parts: []string{
				"actor_member.workspace_id = story.workspace_id",
				"actor_user.is_active = TRUE",
				"actor_team_member.team_id = story.team_id",
				"story.workspace_id = sqlc.arg(workspace_id)",
				"story.deleted_at IS NULL",
				"team_access_unrestricted",
				"parent_comment.story_id = story.id",
			},
		},
		{
			name:  "update",
			start: "-- name: UpdateCommentForAuthor :one",
			end:   "-- name: DeleteCommentForAuthor :one",
			parts: []string{
				"comment.commenter_id = sqlc.arg(actor_id)",
				"story.id = comment.story_id",
				"story.workspace_id = sqlc.arg(workspace_id)",
				"actor_member.workspace_id = story.workspace_id",
				"actor_user.is_active = TRUE",
				"actor_team_member.team_id = story.team_id",
			},
		},
		{
			name:  "delete",
			start: "-- name: DeleteCommentForAuthor :one",
			end:   "-- name: DeleteCommentMentionsForAuthor :one",
			parts: []string{
				"comment.commenter_id = sqlc.arg(actor_id)",
				"story.id = comment.story_id",
				"story.workspace_id = sqlc.arg(workspace_id)",
				"actor_member.workspace_id = story.workspace_id",
				"actor_user.is_active = TRUE",
				"actor_team_member.team_id = story.team_id",
			},
		},
		{
			name:  "mention delete",
			start: "-- name: DeleteCommentMentionsForAuthor :one",
			end:   "-- name: InsertCommentMentionsForAuthor :execrows",
			parts: []string{
				"comment.commenter_id = sqlc.arg(actor_id)",
				"story.workspace_id = sqlc.arg(workspace_id)",
				"actor_member.workspace_id = story.workspace_id",
				"actor_user.is_active = TRUE",
				"actor_team_member.team_id = story.team_id",
				"EXISTS(SELECT 1 FROM scoped_comment) AS comment_found",
			},
		},
		{
			name:  "mention insert",
			start: "-- name: InsertCommentMentionsForAuthor :execrows",
			end:   "-- name: AppendCommentMutationEvent :many",
			parts: []string{
				"INNER JOIN public.workspace_members AS member",
				"member.workspace_id = story.workspace_id",
				"mentioned_user.is_active = TRUE",
				"comment.commenter_id = sqlc.arg(actor_id)",
				"story.workspace_id = sqlc.arg(workspace_id)",
				"actor_member.workspace_id = story.workspace_id",
				"actor_user.is_active = TRUE",
				"actor_team_member.team_id = story.team_id",
			},
		},
		{
			name:  "transactional developer event",
			start: "-- name: AppendCommentMutationEvent :many",
			parts: []string{
				"INSERT INTO public.outbound_webhook_events",
				"INSERT INTO public.outbound_webhook_deliveries",
				"endpoint.workspace_id = created_event.workspace_id",
				"subscription.event_type = sqlc.arg(event_type)",
				"principal.status = 'active'",
				"account.is_active = TRUE",
				"membership.user_id IS NOT NULL",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			section := commentQuerySection(t, queryFile, test.start, test.end)
			for _, part := range test.parts {
				if !strings.Contains(section, part) {
					t.Fatalf("query is missing security clause %q", part)
				}
			}
		})
	}
}

func commentQuerySection(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("query marker %q not found", start)
	}
	section := source[startIndex:]
	if end == "" {
		return section
	}
	endIndex := strings.Index(section, end)
	if endIndex < 0 {
		t.Fatalf("query marker %q not found", end)
	}
	return section[:endIndex]
}
