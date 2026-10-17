package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type runtimeStories struct {
	story      stories.CoreSingleStory
	comment    stories.CoreComment
	err        error
	commentErr error
	actor      auth.Actor
	audience   []uuid.UUID
}

func (s *runtimeStories) Get(ctx context.Context, _, _ uuid.UUID) (stories.CoreSingleStory, error) {
	s.actor, _ = auth.GetActor(ctx)
	return s.story, s.err
}
func (s *runtimeStories) GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error) {
	if s.commentErr != nil {
		return stories.CoreComment{}, s.commentErr
	}
	return s.comment, s.err
}
func (s *runtimeStories) GetCommentThread(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error) {
	return s.comment, s.err
}
func (s *runtimeStories) GetNotificationAudience(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return s.audience, s.err
}

type runtimeAccess bool

func (a runtimeAccess) WorkspaceCanUseMaya(context.Context, uuid.UUID) (bool, error) {
	return bool(a), nil
}

type runtimeMembership struct {
	workspace workspaces.CoreWorkspace
	err       error
}

func (m runtimeMembership) Get(context.Context, uuid.UUID, uuid.UUID) (workspaces.CoreWorkspace, error) {
	return m.workspace, m.err
}

type runtimeWriter struct {
	calls   []comments.CreateCommentCommand
	targets []comments.MentionTarget
	reply   comments.AgentReply
	scope   comments.AgentReplyScope
}

func (w *runtimeWriter) ResolveAgentMentions(_ context.Context, scope comments.AgentReplyScope, _ []uuid.UUID) ([]comments.MentionTarget, error) {
	w.scope = scope
	return w.targets, nil
}

func (w *runtimeWriter) CreateAgentReply(_ context.Context, c comments.AgentReplyCommand) (comments.CoreComment, error) {
	w.calls = append(w.calls, comments.CreateCommentCommand{CreationID: &c.CreationID, WorkspaceID: c.Scope.WorkspaceID, StoryID: c.Scope.StoryID,
		ParentID: &c.Scope.ParentID, Actor: c.Scope.Author, Content: c.Content, MentionedUserIDs: c.MentionedUserIDs})
	return comments.CoreComment{ID: c.CreationID}, nil
}

func (w *runtimeWriter) GetAgentReply(context.Context, comments.AgentReplyScope, uuid.UUID) (comments.AgentReply, error) {
	return w.reply, nil
}

type runtimeTools struct {
	direct bool
	calls  int
	scope  messaging.ToolScope
	err    error
}

func (s *runtimeTools) InspectStoryMutation(_ context.Context, scope messaging.ToolScope, _ string) (messaging.StoryMutationInspection, error) {
	s.scope = scope
	return messaging.StoryMutationInspection{DirectEligible: s.direct, ChangedFields: []string{"description", "status_id"}}, s.err
}
func (s *runtimeTools) ConfirmStoryMutation(_ context.Context, scope messaging.ToolScope, _ string) (messaging.StoryMutationResult, error) {
	s.calls++
	s.scope = scope
	return messaging.StoryMutationResult{Status: "applied", StoryID: scope.SourceScope.StoryID}, s.err
}

func (s *runtimeTools) RebindStoryMutation(context.Context, messaging.ToolScope, messaging.ToolScope, string) (messaging.StoryMutationConfirmation, error) {
	return messaging.StoryMutationConfirmation{Token: "rebound-token"}, s.err
}

func runtimeFixture() (*CommentRuntime, agentruns.Run, *runtimeStories) {
	now := time.Now().UTC()
	run := agentruns.Run{ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(), RequesterID: uuid.New(), SourceID: uuid.New(), EntityID: uuid.New(), Surface: StoryCommentSurface,
		SourceHash: SourceHash("request"), SourceRevision: now.Format(time.RFC3339Nano)}
	store := &runtimeStories{story: stories.CoreSingleStory{ID: run.EntityID, Workspace: run.WorkspaceID, Team: run.TeamID},
		comment: stories.CoreComment{ID: run.SourceID, StoryID: run.EntityID, UserID: run.RequesterID, Comment: "request", UpdatedAt: now}}
	return &CommentRuntime{Stories: store, Access: runtimeAccess(true), Memberships: runtimeMembership{workspace: workspaces.CoreWorkspace{UserRole: "member"}}, Enabled: true, MayaID: uuid.New()}, run, store
}

func TestCommentRuntimeAuthorizationRequiresCurrentMemberRole(t *testing.T) {
	for _, role := range []string{"member", "admin", "guest", "system", ""} {
		t.Run(role, func(t *testing.T) {
			runtime, run, _ := runtimeFixture()
			runtime.Memberships = runtimeMembership{workspace: workspaces.CoreWorkspace{UserRole: role}}
			err := runtime.Authorize(t.Context(), run)
			if role == "member" || role == "admin" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
			}
		})
	}
	runtime, run, _ := runtimeFixture()
	runtime.Memberships = nil
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
	runtime.Memberships = runtimeMembership{err: workspaces.ErrNotFound}
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
	failure := errors.New("membership lookup temporarily unavailable")
	runtime.Memberships = runtimeMembership{err: failure}
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), failure)
	require.NotErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
	now := time.Now()
	runtime.Memberships = runtimeMembership{workspace: workspaces.CoreWorkspace{UserRole: "member", DeletedAt: &now}}
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
}

func TestCommentRuntimeAuthorizationRechecksSourceAndHumanScope(t *testing.T) {
	runtime, run, store := runtimeFixture()
	require.NoError(t, runtime.Authorize(t.Context(), run))
	require.Equal(t, run.RequesterID, store.actor.PrincipalID)
	require.Equal(t, auth.PrincipalHumanUser, store.actor.Kind)
	require.True(t, store.actor.TeamAccess.Allows(run.TeamID))
	require.False(t, store.actor.TeamAccess.Allows(uuid.New()))
	store.comment.Comment = "edited request"
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrSourceChanged)
	run.Outcome = &agentruns.Outcome{Reply: "Applied", Receipt: json.RawMessage(`{"committed":true}`)}
	require.NoError(t, runtime.Authorize(t.Context(), run), "an already committed action still gets its result reply")
	store.err = stories.ErrNotFound
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
}

func TestCommentRuntimeAuthorizationPreservesTransientFailures(t *testing.T) {
	for _, target := range []string{"story", "comment"} {
		t.Run(target, func(t *testing.T) {
			runtime, run, store := runtimeFixture()
			failure := errors.New("database temporarily unavailable")
			if target == "story" {
				store.err = failure
			} else {
				store.commentErr = failure
			}
			err := runtime.Authorize(t.Context(), run)
			require.ErrorIs(t, err, failure)
			require.NotErrorIs(t, err, agentruns.ErrPermissionDenied)
			require.NotErrorIs(t, err, agentruns.ErrSourceChanged)
		})
	}
	runtime, run, store := runtimeFixture()
	store.commentErr = stories.ErrNotFound
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrSourceChanged)
	store.commentErr = stories.ErrStoryReadForbidden
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
}

func TestCommentRuntimeNeverAppliesUnapprovedProposal(t *testing.T) {
	runtime, run, _ := runtimeFixture()
	tools := &runtimeTools{}
	runtime.Tools = tools
	proposal := agentruns.Proposal{Action: json.RawMessage(`{"token":"server-token"}`), RequiresApproval: true}
	_, err := runtime.Apply(t.Context(), run, proposal)
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	require.Zero(t, tools.calls)
	tools.direct = true
	outcome, err := runtime.Apply(t.Context(), run, proposal)
	require.NoError(t, err)
	require.Equal(t, 1, tools.calls)
	require.Equal(t, run.RequesterID, tools.scope.UserID)
	require.Equal(t, run.EntityID, tools.scope.SourceScope.StoryID)
	require.Contains(t, outcome.Reply, "description")
	require.Contains(t, string(outcome.Receipt), runtime.MayaID.String())
}

func TestMayaPublicationUsesStableReplyAndEscapedOutput(t *testing.T) {
	runtime, run, _ := runtimeFixture()
	writer := &runtimeWriter{}
	runtime.Comments = writer
	outcome := agentruns.Outcome{Reply: `<img src=x onerror=alert(1)>`}
	first, err := runtime.Publish(t.Context(), run, outcome)
	require.NoError(t, err)
	second, err := runtime.Publish(t.Context(), run, outcome)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Equal(t, &run.SourceID, writer.calls[0].ParentID)
	require.Equal(t, runtime.MayaID, writer.calls[0].Actor.PrincipalID)
	require.Equal(t, auth.PrincipalSystem, writer.calls[0].Actor.Kind)
	require.Empty(t, writer.calls[0].MentionedUserIDs)
	require.NotContains(t, writer.calls[0].Content, "<img")
	require.Contains(t, writer.calls[0].Content, "&lt;img")
	require.LessOrEqual(t, len([]rune(replyHTML(strings.Repeat("<&", 6000)))), 10000)
}

func TestConversationExcludesFutureCommentsAndMarksOtherAuthors(t *testing.T) {
	now := time.Now()
	source := stories.CoreComment{ID: uuid.New(), UserID: uuid.New(), CreatedAt: now}
	root := stories.CoreComment{ID: uuid.New(), UserID: uuid.New(), CreatedAt: now.Add(-time.Minute), Comment: "Earlier context"}
	root.SubComments = []stories.CoreComment{source, {ID: uuid.New(), CreatedAt: now.Add(time.Minute), Comment: "Future instruction"}}
	turns := commentConversation(root, source, uuid.New())
	require.Len(t, turns, 1)
	require.Contains(t, turns[0].Text, "another teammate")
	require.NotContains(t, turns[0].Text, "Future")
}
