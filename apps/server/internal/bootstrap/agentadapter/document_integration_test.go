//go:build integration

package agentadapter

import (
	"context"
	"errors"
	"sync"
	"testing"

	agentrepository "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	documentsrepository "github.com/complexus-tech/projects-api/internal/modules/documents/repository"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type documentAssistantStub struct {
	calls    int
	response messaging.DocumentResponse
	request  messaging.DocumentRequest
}

func (a *documentAssistantStub) RespondDocument(_ context.Context, request messaging.DocumentRequest) (messaging.DocumentResponse, error) {
	a.calls++
	a.request = request
	return a.response, nil
}

type documentAgentFixture struct {
	pool                                   *pgxpool.Pool
	ctx                                    context.Context
	workspace, user, other, maya, document uuid.UUID
	documents                              *documents.Service
	runs                                   *agentruns.Service
	creator                                *DocumentCommentCreator
	runtime                                *DocumentRuntime
	assistant                              *documentAssistantStub
}

func newDocumentAgentFixture(t *testing.T) documentAgentFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	f := documentAgentFixture{pool: postgres.Pool, ctx: t.Context(), workspace: uuid.New(), user: uuid.New(), other: uuid.New(), maya: uuid.MustParse("00000000-0000-0000-0000-000000000001")}
	_, err := f.pool.Exec(f.ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Document Maya test',$2)`, f.workspace, uuid.NewString())
	require.NoError(t, err)
	for _, id := range []uuid.UUID{f.user, f.other} {
		_, err = f.pool.Exec(f.ctx, `INSERT INTO users(user_id,username,email,full_name,is_system) VALUES($1,$2,$3,'Canonical Person',$4)`, id, id.String(), id.String()+"@example.test", id == f.maya)
		require.NoError(t, err)
		if id != f.maya {
			_, err = f.pool.Exec(f.ctx, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, f.workspace, id)
			require.NoError(t, err)
		}
	}
	actor, err := auth.NewHumanActor(f.user).WithWorkspace(f.workspace)
	require.NoError(t, err)
	f.ctx, err = auth.SetActor(f.ctx, actor)
	require.NoError(t, err)
	f.documents = documents.New(documentsrepository.New(f.pool))
	doc, err := f.documents.Create(f.ctx, documents.CoreCreateInput{WorkspaceID: f.workspace, UserID: f.user, Title: "Original document", Visibility: documents.VisibilityWorkspace, ContentHTML: "<p>Original paragraph.</p>", ContentText: "Original paragraph."})
	require.NoError(t, err)
	f.document = doc.ID
	f.runs = agentruns.New(agentrepository.New(f.pool), nil)
	f.creator = &DocumentCommentCreator{Pool: f.pool, Documents: f.documents, Runs: f.runs, MayaID: f.maya, Tasks: &unavailableWakeup{}}
	f.assistant = &documentAssistantStub{}
	f.runtime = &DocumentRuntime{Documents: f.documents, Assistant: f.assistant, Runs: f.runs, Shared: &CommentRuntime{Enabled: true, MayaID: f.maya, Access: runtimeAccess(true), Memberships: workspacesrepository.New(f.pool)}, EditsAvailable: func(context.Context) (bool, error) { return true, nil }}
	return f
}

func (f documentAgentFixture) request(t *testing.T, body string) (documents.CoreCommentThread, agentruns.Run) {
	t.Helper()
	id := uuid.New()
	thread, err := f.creator.CreateComment(f.ctx, documents.CoreCreateCommentInput{WorkspaceID: f.workspace, UserID: f.user, DocumentID: f.document, CreationID: &id, Body: body})
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.document)
	require.NoError(t, err)
	require.NotEmpty(t, runs)
	for _, run := range runs {
		if run.SourceID == thread.Comments[0].ID {
			return thread, run
		}
	}
	t.Fatal("missing invocation")
	return thread, agentruns.Run{}
}

func TestDocumentInvocationCommitsRepliesMentionsAndCrossAuthorContinuation(t *testing.T) {
	f := newDocumentAgentFixture(t)
	f.assistant.response = messaging.DocumentResponse{Reply: "[@Model Label](mention:" + f.other.String() + "), which paragraph should I update?", NeedsInput: true}
	thread, run := f.request(t, "[@Maya](agent:maya) please help with this document")
	worker := agentruns.New(agentrepository.New(f.pool), f.runtime)
	require.NoError(t, worker.Process(f.ctx, run.ID))
	finished, err := f.runs.Get(f.ctx, f.workspace, run.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusNeedsInput, finished.Status)
	require.NotNil(t, finished.ReplyCommentID)
	require.Equal(t, uuid.Nil, finished.TeamID)
	reply, err := f.documents.GetComment(f.ctx, f.workspace, f.user, f.document, *finished.ReplyCommentID)
	require.NoError(t, err)
	require.Contains(t, reply.Comment.Body, "[@Canonical Person](mention:"+f.other.String()+")")
	require.Equal(t, []uuid.UUID{f.other}, reply.Comment.MentionedUserIDs)
	var count int
	err = f.pool.QueryRow(f.ctx, `SELECT count(*) FROM document_comment_notification_receipts WHERE comment_id=$1 AND recipient_id=$2`, reply.Comment.ID, f.other).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.NoError(t, worker.Process(f.ctx, run.ID))
	require.Equal(t, 1, f.assistant.calls)
	actor, err := auth.NewHumanActor(f.other).WithWorkspace(f.workspace)
	require.NoError(t, err)
	otherCtx, err := auth.SetActor(f.ctx, actor)
	require.NoError(t, err)
	creationID := uuid.New()
	input := documents.CoreReplyCommentInput{WorkspaceID: f.workspace, UserID: f.other, DocumentID: f.document, ThreadID: thread.ID, CreationID: &creationID, ReplyToCommentID: finished.ReplyCommentID, Body: "The opening paragraph."}
	answer, err := f.creator.ReplyToComment(otherCtx, input)
	require.NoError(t, err)
	replay, err := f.creator.ReplyToComment(otherCtx, input)
	require.NoError(t, err)
	require.Equal(t, answer.ID, replay.ID)
	all, err := f.runs.List(f.ctx, f.workspace, f.document)
	require.NoError(t, err)
	var children int
	for _, candidate := range all {
		if candidate.ParentRunID != nil && *candidate.ParentRunID == run.ID {
			children++
			require.Equal(t, f.other, candidate.RequesterID)
		}
	}
	require.Equal(t, 1, children)
}

func TestDocumentInvocationStagesExactEditAndWaitsForAuthoritativeReceipt(t *testing.T) {
	f := newDocumentAgentFixture(t)
	_, err := f.pool.Exec(f.ctx, `UPDATE documents SET collaboration_state=$2 WHERE document_id=$1`, f.document, []byte{0, 0})
	require.NoError(t, err)
	title := "Updated title"
	f.assistant.response = messaging.DocumentResponse{Reply: "I can change the title.", Operations: &messaging.DocumentOperations{Version: 1, Title: &title}}
	_, run := f.request(t, "[@Maya](agent:maya) Rename this document to Updated title")
	worker := agentruns.New(agentrepository.New(f.pool), f.runtime)
	require.NoError(t, worker.Process(f.ctx, run.ID))
	pending, err := f.runs.Get(f.ctx, f.workspace, run.ID)
	require.NoError(t, err)
	require.NotNil(t, pending.ActionStartedAt)
	require.Nil(t, pending.Outcome)
	require.Nil(t, pending.ReplyCommentID)
	edit, err := f.documents.GetAgentEdit(f.ctx, f.workspace, run.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", edit.Status)
	doc, err := f.documents.Get(f.ctx, f.workspace, f.user, f.document)
	require.NoError(t, err)
	require.Equal(t, "Original document", doc.Title, "Go must never overwrite collaborative content")
	_, err = f.runs.Cancel(f.ctx, agentruns.Control{WorkspaceID: f.workspace, RequesterID: f.user, RunID: run.ID, ExpectedVersion: pending.Version})
	require.ErrorIs(t, err, agentruns.ErrConflict)
	// This is the collaboration consumer's committed receipt boundary. Its real
	// Yjs write and broadcast are covered by the collaboration integration suite.
	_, err = f.pool.Exec(f.ctx, `UPDATE document_agent_edits SET status='applied',resulting_revision=2,receipt='{"revision":2}',applied_at=clock_timestamp() WHERE invocation_id=$1`, run.ID)
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `UPDATE agent_invocations SET available_at=clock_timestamp() WHERE id=$1`, run.ID)
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, run.ID))
	finished, err := f.runs.Get(f.ctx, f.workspace, run.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, finished.Status)
	require.NotNil(t, finished.ReplyCommentID)
	require.Equal(t, 1, f.assistant.calls)
	var commandCount int
	err = f.pool.QueryRow(f.ctx, `SELECT count(*) FROM document_agent_edits WHERE invocation_id=$1`, run.ID).Scan(&commandCount)
	require.NoError(t, err)
	require.Equal(t, 1, commandCount)
}

func TestDocumentInvocationSourceAndRunRollbackTogether(t *testing.T) {
	f := newDocumentAgentFixture(t)
	rollback := errors.New("rollback source")
	err := database.WithinUnitOfWork(f.ctx, f.pool, func(ctx context.Context, _ pgx.Tx) error {
		_, err := f.creator.CreateComment(ctx, documents.CoreCreateCommentInput{WorkspaceID: f.workspace, UserID: f.user, DocumentID: f.document, Body: "[@Maya](agent:maya) Help"})
		if err != nil {
			return err
		}
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	runs, err := f.runs.List(f.ctx, f.workspace, f.document)
	require.NoError(t, err)
	require.Empty(t, runs)
	threads, err := f.documents.ListComments(f.ctx, f.workspace, f.user, f.document)
	require.NoError(t, err)
	require.Empty(t, threads)
}

func TestDocumentInvocationApprovalRebindsToReplyingTeammateAndKeepsOriginalRevision(t *testing.T) {
	f := newDocumentAgentFixture(t)
	_, err := f.pool.Exec(f.ctx, `UPDATE documents SET collaboration_state=$2 WHERE document_id=$1`, f.document, []byte{0, 0})
	require.NoError(t, err)
	title := "Reviewed title"
	f.assistant.response = messaging.DocumentResponse{Reply: "I suggest a clearer title.", Operations: &messaging.DocumentOperations{Version: 1, Title: &title}}
	thread, run := f.request(t, "[@Maya](agent:maya) Could you suggest a better title?")
	worker := agentruns.New(agentrepository.New(f.pool), f.runtime)
	require.NoError(t, worker.Process(f.ctx, run.ID))
	parent, err := f.runs.Get(f.ctx, f.workspace, run.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusAwaitingApproval, parent.Status)
	require.NotNil(t, parent.ProposalReplyID)
	var count int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM document_agent_edits`).Scan(&count))
	require.Zero(t, count)
	actor, err := auth.NewHumanActor(f.other).WithWorkspace(f.workspace)
	require.NoError(t, err)
	otherCtx, err := auth.SetActor(f.ctx, actor)
	require.NoError(t, err)
	_, err = f.creator.ReplyToComment(otherCtx, documents.CoreReplyCommentInput{WorkspaceID: f.workspace, UserID: f.other, DocumentID: f.document, ThreadID: thread.ID, ReplyToCommentID: parent.ProposalReplyID, Body: "Yes please"})
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.document)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	child := runs[0]
	require.Equal(t, f.other, child.RequesterID)
	require.Equal(t, &parent.ID, child.ParentRunID)
	require.NoError(t, worker.Process(otherCtx, child.ID))
	edit, err := f.documents.GetAgentEdit(f.ctx, f.workspace, child.ID)
	require.NoError(t, err)
	require.Equal(t, f.other, edit.RequesterID)
	reviewed, err := decodeDocumentAction(parent, *parent.Proposal)
	require.NoError(t, err)
	require.Equal(t, reviewed.ExpectedRevision, edit.ExpectedRevision)
	require.Equal(t, 1, f.assistant.calls, "a clear answer reuses exact reviewed values, not another model proposal")
}

func TestDocumentInvocationRechecksWorkspaceRoleBeforeModelOrReply(t *testing.T) {
	f := newDocumentAgentFixture(t)
	f.assistant.response = messaging.DocumentResponse{Reply: "An answer"}
	_, run := f.request(t, "[@Maya](agent:maya) explain this document")
	_, err := f.pool.Exec(f.ctx, `UPDATE workspace_members SET role='guest' WHERE workspace_id=$1 AND user_id=$2`, f.workspace, f.user)
	require.NoError(t, err)
	err = agentruns.New(agentrepository.New(f.pool), f.runtime).Process(f.ctx, run.ID)
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	stored, err := f.runs.Get(f.ctx, f.workspace, run.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusFailed, stored.Status)
	require.Nil(t, stored.ReplyCommentID)
	require.Nil(t, stored.FailureReplyID)
	require.Zero(t, f.assistant.calls)
}

func TestDocumentInvocationConcurrentCreationReplaysOneSourceAndRun(t *testing.T) {
	f := newDocumentAgentFixture(t)
	f.creator.Tasks = nil
	creationID := uuid.New()
	input := documents.CoreCreateCommentInput{WorkspaceID: f.workspace, UserID: f.user, DocumentID: f.document, CreationID: &creationID, Body: "[@Maya](agent:maya) explain the opening paragraph"}
	start := make(chan struct{})
	results := make(chan uuid.UUID, 4)
	failures := make(chan error, 4)
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			<-start
			thread, err := f.creator.CreateComment(f.ctx, input)
			if err != nil {
				failures <- err
				return
			}
			results <- thread.ID
		})
	}
	close(start)
	group.Wait()
	close(results)
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	var first uuid.UUID
	for id := range results {
		if first == uuid.Nil {
			first = id
		}
		require.Equal(t, first, id)
	}
	require.NotEqual(t, uuid.Nil, first)
	runs, err := f.runs.List(f.ctx, f.workspace, f.document)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	threads, err := f.documents.ListComments(f.ctx, f.workspace, f.user, f.document)
	require.NoError(t, err)
	require.Len(t, threads, 1)
}
