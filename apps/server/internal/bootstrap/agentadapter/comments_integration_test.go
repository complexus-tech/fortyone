//go:build integration

package agentadapter

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	agentrepository "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type unavailableWakeup struct{ calls int }

func (w *unavailableWakeup) EnqueueAgentInvocation(uuid.UUID) error {
	w.calls++
	return errors.New("queue unavailable")
}

// The model response is deterministic; authorization, publication, continuation
// discovery, and transaction boundaries use the real product implementations.
type clarificationRuntime struct{ comments *CommentRuntime }

func (r clarificationRuntime) Authorize(ctx context.Context, run agentruns.Run) error {
	return r.comments.Authorize(ctx, run)
}
func (r clarificationRuntime) Prepare(context.Context, agentruns.Run) (agentruns.Proposal, error) {
	return agentruns.Proposal{Reply: "Which part of the description should I update?", NeedsInput: true}, nil
}
func (r clarificationRuntime) Apply(context.Context, agentruns.Run, agentruns.Proposal) (agentruns.Outcome, error) {
	return agentruns.Outcome{}, errors.New("clarification must not execute an action")
}
func (r clarificationRuntime) Publish(ctx context.Context, run agentruns.Run, outcome agentruns.Outcome) (uuid.UUID, error) {
	return r.comments.Publish(ctx, run, outcome)
}

func TestCommentInvocationCommitsAtomicallyAndSurvivesQueueOutage(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace, team, story, user, maya, mentioned := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err := postgres.Pool.Exec(ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Maya test',$2)`, workspace, uuid.NewString())
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO teams(team_id,name,workspace_id,code,color) VALUES($1,'Engineering',$2,'MAY','#123456')`, team, workspace)
	require.NoError(t, err)
	for _, id := range []uuid.UUID{user, maya, mentioned} {
		_, err = postgres.Pool.Exec(ctx, `INSERT INTO users(user_id,username,email,full_name,is_system) VALUES($1,$2,$3,'Test user',$4)`, id, id.String(), id.String()+"@example.com", id == maya)
		require.NoError(t, err)
		if id != maya {
			_, err = postgres.Pool.Exec(ctx, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, workspace, id)
			require.NoError(t, err)
			_, err = postgres.Pool.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, id)
			require.NoError(t, err)
		}
	}
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO stories(id,team_id,workspace_id,title,reporter_id,sequence_id) VALUES($1,$2,$3,'Original title',$4,1)`, story, team, workspace, user)
	require.NoError(t, err)
	actor, err := auth.NewHumanActor(user).WithWorkspace(workspace)
	require.NoError(t, err)
	ctx, err = auth.SetActor(ctx, actor)
	require.NoError(t, err)
	log := logger.NewWithText(io.Discard, slog.LevelError, "agent-test")
	commentService := comments.New(commentsrepository.New(log, postgres.Pool))
	base, err := providers.NewStoryCommentCreator(commentService)
	require.NoError(t, err)
	storyService := stories.New(log, storiesrepository.New(log, postgres.Pool), nil, nil)
	runs := agentruns.New(agentrepository.New(postgres.Pool), nil)
	wakeup := &unavailableWakeup{}
	creator := &CommentCreator{Pool: postgres.Pool, Base: base, Stories: storyService, Runs: runs, MayaID: maya, Tasks: wakeup, Log: log}
	storyService.ConfigureCommentCreator(creator)
	creationID := uuid.New()
	body := `<p><a data-type="mention" data-mention-kind="agent" data-agent-key="maya" data-id="` + maya.String() + `">@Maya</a> close this story.</p>`
	input := stories.CoreNewComment{CreationID: &creationID, StoryID: story, Comment: body, AgentMentions: []string{"maya"}, Mentions: []uuid.UUID{mentioned}}
	created, err := storyService.CreateComment(ctx, workspace, input)
	require.NoError(t, err)
	require.Equal(t, creationID, created.ID)
	require.Equal(t, 1, wakeup.calls)
	accepted, err := runs.List(ctx, workspace, story)
	require.NoError(t, err)
	require.Len(t, accepted, 1)
	require.Equal(t, agentruns.StatusQueued, accepted[0].Status)
	require.Equal(t, creationID, accepted[0].SourceID)
	// A changed display title after acceptance must not make a network replay
	// create another run or overwrite the context captured with the source.
	_, err = postgres.Pool.Exec(ctx, `UPDATE stories SET title='Changed later' WHERE id=$1`, story)
	require.NoError(t, err)
	replayed, err := storyService.CreateComment(ctx, workspace, input)
	require.NoError(t, err)
	require.True(t, replayed.Replayed)
	require.Equal(t, 1, wakeup.calls)
	accepted, err = runs.List(ctx, workspace, story)
	require.NoError(t, err)
	require.Len(t, accepted, 1)
	require.Contains(t, string(accepted[0].Payload), "Original title")
	// A real Maya reply is a continuation target. Posting the answer completes
	// the old clarification and creates one new run, even when delivery repeats.
	runtime := &CommentRuntime{Stories: storyService, Comments: commentService, Access: runtimeAccess(true), Memberships: workspacesrepository.New(postgres.Pool), Enabled: true, MayaID: maya}
	worker := agentruns.New(agentrepository.New(postgres.Pool), clarificationRuntime{comments: runtime})
	require.NoError(t, worker.Process(ctx, accepted[0].ID))
	clarification, err := runs.Get(ctx, workspace, accepted[0].ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusNeedsInput, clarification.Status)
	require.NotNil(t, clarification.ReplyCommentID)
	reply, err := storyService.GetComment(ctx, *clarification.ReplyCommentID, story, workspace)
	require.NoError(t, err)
	require.Equal(t, maya, reply.UserID)
	assertVisibleReply := func(expectedReplies int) {
		t.Helper()
		// The UI consumes list/thread reads, not GetComment. Re-fetching either
		// surface must retain a published system reply even though Maya does not
		// have human workspace/team membership.
		for range 2 {
			visible, more, err := storyService.GetComments(ctx, story, workspace, 1, 20)
			require.NoError(t, err)
			require.False(t, more)
			require.Len(t, visible, 1)
			require.Equal(t, creationID, visible[0].ID)
			require.Len(t, visible[0].SubComments, expectedReplies)
			require.Equal(t, reply.ID, visible[0].SubComments[0].ID)
			require.Equal(t, reply.Comment, visible[0].SubComments[0].Comment)
			thread, err := storyService.GetCommentThread(ctx, reply.ID, story, workspace)
			require.NoError(t, err)
			require.Equal(t, visible[0], thread)
		}
	}
	assertVisibleReply(1)
	answerID := uuid.New()
	answerActor, err := auth.NewHumanActor(mentioned).WithWorkspace(workspace)
	require.NoError(t, err)
	answerCtx, err := auth.SetActor(ctx, answerActor)
	require.NoError(t, err)
	answer := stories.CoreNewComment{CreationID: &answerID, Parent: clarification.ReplyCommentID, StoryID: story, Comment: "<p>Update only the acceptance criteria.</p>"}
	_, err = storyService.CreateComment(answerCtx, workspace, answer)
	require.NoError(t, err)
	_, err = storyService.CreateComment(answerCtx, workspace, answer)
	require.NoError(t, err)
	assertVisibleReply(2)
	continued, err := runs.Get(ctx, workspace, clarification.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, continued.Status)
	allRuns, err := runs.List(ctx, workspace, story)
	require.NoError(t, err)
	require.Len(t, allRuns, 2)
	require.Equal(t, &clarification.ID, allRuns[0].ParentRunID)
	require.Equal(t, answerID, allRuns[0].SourceID)
	require.Equal(t, mentioned, allRuns[0].RequesterID, "the teammate who answers authorizes the follow-up")
	require.Equal(t, 2, wakeup.calls)
	// Failure to persist the invocation rolls the ordinary comment back.
	creator.Runs = agentruns.New(nil, nil)
	failedID := uuid.New()
	input.CreationID = &failedID
	_, err = storyService.CreateComment(ctx, workspace, input)
	require.Error(t, err)
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM story_comments WHERE comment_id=$1`, failedID).Scan(&count))
	require.Zero(t, count)
	// Worker authorization sees live membership and edited request revisions.
	require.NoError(t, runtime.Authorize(context.Background(), accepted[0]))
	rollback := errors.New("roll back temporary role change")
	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(txctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(txctx, `UPDATE workspace_members SET role='guest' WHERE workspace_id=$1 AND user_id=$2`, workspace, user)
		require.NoError(t, err)
		require.ErrorIs(t, runtime.Authorize(txctx, accepted[0]), agentruns.ErrPermissionDenied, "the fenced unit must see its current membership snapshot")
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	require.NoError(t, runtime.Authorize(context.Background(), accepted[0]))
	// An accepted follow-up cannot run after its requester loses write access,
	// even though the guest can still read the same story and thread.
	child := allRuns[0]
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspace_members SET role='guest' WHERE workspace_id=$1 AND user_id=$2`, workspace, mentioned)
	require.NoError(t, err)
	_, err = storyService.Get(answerCtx, story, workspace)
	require.NoError(t, err)
	require.ErrorIs(t, runtime.Authorize(context.Background(), child), agentruns.ErrPermissionDenied)
	require.ErrorIs(t, worker.Process(ctx, child.ID), agentruns.ErrPermissionDenied)
	denied, err := runs.Get(ctx, workspace, child.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusFailed, denied.Status)
	require.Nil(t, denied.Proposal)
	require.Nil(t, denied.Outcome)
	require.Nil(t, denied.ReplyCommentID)
	require.Nil(t, denied.FailureReplyID)
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, team, user)
	require.NoError(t, err)
	require.ErrorIs(t, runtime.Authorize(context.Background(), accepted[0]), agentruns.ErrPermissionDenied)
}
