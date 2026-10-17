//go:build integration

package commentsrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestAgentReplyMentionsRequireLiveRequesterAndRecipientStoryAccess(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, teamID, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "agent-mentions")
	foreignWorkspace, _, _ := insertCommentTestWorkspace(t, ctx, postgres.Pool, "foreign-agent-mentions")
	requesterID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "requester")
	targetID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "target")
	outsideTeamID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "outside-team")
	inactiveID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "inactive")
	foreignID := insertCommentTestUser(t, ctx, postgres.Pool, foreignWorkspace, "foreign")
	mayaID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "maya")
	_, err := postgres.Pool.Exec(ctx, `UPDATE users SET is_system=TRUE WHERE user_id=$1`, mayaID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE users SET is_active=FALSE WHERE user_id=$1`, inactiveID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, outsideTeamID)
	require.NoError(t, err)
	insertCommentWebhookEndpoint(t, ctx, postgres.Pool, workspaceID, requesterID)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "agent-mentions"), postgres.Pool))
	requester := restrictedCommentActor(t, requesterID, workspaceID, teamID)
	parent, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: requester, Content: "Please tag Alex"})
	require.NoError(t, err)
	author, err := auth.NewActor(mayaID, auth.PrincipalSystem, uuid.Nil, auth.MustScopeSet(auth.ScopeCommentsWrite), requester.TeamAccess)
	require.NoError(t, err)
	author, err = author.WithWorkspace(workspaceID)
	require.NoError(t, err)
	scope := comments.AgentReplyScope{WorkspaceID: workspaceID, StoryID: storyID, ParentID: parent.ID, Requester: requester, Author: author}
	ids := []uuid.UUID{targetID, outsideTeamID, inactiveID, foreignID, mayaID, uuid.New()}
	targets, err := service.ResolveAgentMentions(ctx, scope, ids)
	require.NoError(t, err)
	require.Equal(t, []comments.MentionTarget{{ID: targetID, Name: "Comments target"}}, targets)
	command := comments.AgentReplyCommand{Scope: scope, CreationID: uuid.New(), Content: "<p>Saved canonical @Alex</p>", MentionedUserIDs: []uuid.UUID{targetID, targetID}}
	created, err := service.CreateAgentReply(ctx, command)
	require.NoError(t, err)
	require.Equal(t, mayaID, created.UserID)
	assertCommentState(t, ctx, postgres.Pool, created.ID, command.Content, []uuid.UUID{targetID})
	assertCommentEventContract(t, ctx, postgres.Pool, created.ID, "comment.created", 1)
	replayed, err := service.CreateAgentReply(ctx, command)
	require.NoError(t, err)
	require.True(t, replayed.Replayed)
	assertCommentEventContract(t, ctx, postgres.Pool, created.ID, "comment.created", 1)
	conflict := command
	conflict.Content = "Different text"
	_, err = service.CreateAgentReply(ctx, conflict)
	require.ErrorIs(t, err, comments.ErrCreationConflict)
	// Ordinary system/provider writes must remain unable to manufacture mentions.
	_, err = service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, ParentID: &parent.ID, Actor: author, Content: "Forbidden provider mention", MentionedUserIDs: []uuid.UUID{targetID}})
	require.ErrorIs(t, err, comments.ErrInvalidMention)
	for _, deniedID := range ids[1:] {
		invalid := command
		invalid.CreationID, invalid.MentionedUserIDs = uuid.New(), []uuid.UUID{targetID, deniedID}
		_, err = service.CreateAgentReply(ctx, invalid)
		require.ErrorIs(t, err, comments.ErrInvalidMention)
		var count int
		require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM story_comments WHERE comment_id=$1`, invalid.CreationID).Scan(&count))
		require.Zero(t, count, "invalid recipient must roll back comment, mentions, and event")
		assertCommentEventContract(t, ctx, postgres.Pool, invalid.CreationID, "comment.created", 0)
	}
	// Recipients are rechecked at the actual write, after earlier resolution.
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, targetID)
	require.NoError(t, err)
	stale := command
	stale.CreationID = uuid.New()
	_, err = service.CreateAgentReply(ctx, stale)
	require.ErrorIs(t, err, comments.ErrInvalidMention)
	persisted, err := service.GetAgentReply(ctx, scope, created.ID)
	require.NoError(t, err)
	require.Equal(t, command.Content, persisted.Comment.Comment)
	require.Empty(t, persisted.MentionedUserIDs, "removed teammate must not receive delayed mention")
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, teamID, targetID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE users SET full_name='Renamed later' WHERE user_id=$1`, targetID)
	require.NoError(t, err)
	persisted, err = service.GetAgentReply(ctx, scope, created.ID)
	require.NoError(t, err)
	require.Equal(t, command.Content, persisted.Comment.Comment)
	require.Equal(t, []uuid.UUID{targetID}, persisted.MentionedUserIDs)
	// Publication joins the invocation's outer transaction instead of committing
	// the reply or its mention rows ahead of the durable publication receipt.
	rolledBack := command
	rolledBack.CreationID = uuid.New()
	rollback := errors.New("rollback invocation receipt")
	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(ctx context.Context, _ pgx.Tx) error {
		_, err := service.CreateAgentReply(ctx, rolledBack)
		require.NoError(t, err)
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	_, err = service.GetAgentReply(ctx, scope, rolledBack.CreationID)
	require.ErrorIs(t, err, comments.ErrNotFound)
	assertCommentEventContract(t, ctx, postgres.Pool, rolledBack.CreationID, "comment.created", 0)
	for _, test := range []struct {
		name  string
		alter func(*comments.AgentReplyScope)
	}{
		{"foreign workspace", func(s *comments.AgentReplyScope) { s.WorkspaceID = foreignWorkspace }},
		{"wrong story", func(s *comments.AgentReplyScope) { s.StoryID = uuid.New() }},
		{"wrong source", func(s *comments.AgentReplyScope) { s.ParentID = uuid.New() }},
		{"different requester", func(s *comments.AgentReplyScope) { s.Requester = commentIntegrationActor(t, targetID, workspaceID) }},
		{"restricted team", func(s *comments.AgentReplyScope) {
			s.Requester = restrictedCommentActor(t, requesterID, workspaceID, uuid.New())
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			denied := command
			denied.CreationID = uuid.New()
			test.alter(&denied.Scope)
			_, err := service.CreateAgentReply(ctx, denied)
			require.Error(t, err)
			_, err = service.ResolveAgentMentions(ctx, denied.Scope, []uuid.UUID{targetID})
			require.Error(t, err)
			_, err = service.GetAgentReply(ctx, denied.Scope, created.ID)
			require.Error(t, err)
		})
	}
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspace_members SET role='guest' WHERE workspace_id=$1 AND user_id=$2`, workspaceID, requesterID)
	require.NoError(t, err)
	_, err = service.CreateAgentReply(ctx, stale)
	require.ErrorIs(t, err, comments.ErrNotFound)
	_, err = service.GetAgentReply(ctx, scope, created.ID)
	require.ErrorIs(t, err, comments.ErrNotFound)
}
