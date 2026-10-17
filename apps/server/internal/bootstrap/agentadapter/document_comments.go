package agentadapter

import (
	"context"
	"errors"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentCommentCreator struct {
	Pool      *pgxpool.Pool
	Documents *documents.Service
	Runs      *agentruns.Service
	MayaID    uuid.UUID
	Tasks     Wakeup
	Log       *logger.Logger
}

func (c *DocumentCommentCreator) CreateComment(ctx context.Context, input documents.CoreCreateCommentInput) (documents.CoreCommentThread, error) {
	if !documents.HasMayaMention(input.Body) {
		return c.Documents.CreateComment(ctx, input)
	}
	var thread documents.CoreCommentThread
	err := c.inTransaction(ctx, func(txCtx context.Context) error {
		var err error
		thread, err = c.Documents.CreateComment(txCtx, input)
		if err != nil {
			return err
		}
		if len(thread.Comments) == 0 {
			return documents.ErrConflict
		}
		return c.accept(txCtx, input.WorkspaceID, input.UserID, input.DocumentID, thread.ID, thread.Comments[0], true)
	})
	return thread, err
}

func (c *DocumentCommentCreator) ReplyToComment(ctx context.Context, input documents.CoreReplyCommentInput) (documents.CoreComment, error) {
	explicit := documents.HasMayaMention(input.Body)
	if !explicit && input.ReplyToCommentID == nil {
		return c.Documents.ReplyToComment(ctx, input)
	}
	if !explicit {
		parent, err := c.Documents.GetComment(ctx, input.WorkspaceID, input.UserID, input.DocumentID, *input.ReplyToCommentID)
		if err != nil {
			return documents.CoreComment{}, err
		}
		if parent.Thread.ID != input.ThreadID {
			return documents.CoreComment{}, documents.ErrNotFound
		}
		if parent.Comment.CreatedBy != c.MayaID {
			return c.Documents.ReplyToComment(ctx, input)
		}
	}
	var comment documents.CoreComment
	err := c.inTransaction(ctx, func(txCtx context.Context) error {
		var err error
		comment, err = c.Documents.ReplyToComment(txCtx, input)
		if err != nil {
			return err
		}
		return c.accept(txCtx, input.WorkspaceID, input.UserID, input.DocumentID, input.ThreadID, comment, explicit)
	})
	return comment, err
}

func (c *DocumentCommentCreator) accept(ctx context.Context, workspaceID, userID, documentID, threadID uuid.UUID, comment documents.CoreComment, explicit bool) error {
	actor, err := auth.GetActor(ctx)
	if err != nil || actor.Kind != auth.PrincipalHumanUser || actor.PrincipalID != userID {
		return auth.ErrInvalidActor
	}
	if c.Runs == nil {
		return agentruns.ErrUnavailable
	}
	var parent *uuid.UUID
	if comment.ReplyToCommentID != nil {
		parent, err = c.Runs.ResolveContinuation(ctx, agentruns.ContinuationQuery{WorkspaceID: workspaceID, RequesterID: userID, Surface: DocumentCommentSurface, EntityID: documentID, ConversationID: threadID, ReplyToID: *comment.ReplyToCommentID, SourceID: comment.ID})
		if err != nil {
			return err
		}
	}
	if !explicit && parent == nil {
		return nil
	}
	run, accepted, err := c.Runs.Accept(ctx, agentruns.Input{WorkspaceID: workspaceID, RequesterID: userID, Surface: DocumentCommentSurface, EntityID: documentID, ConversationID: threadID, SourceID: comment.ID,
		SourceRevision: comment.UpdatedAt.UTC().Format(time.RFC3339Nano), SourceHash: SourceHash(comment.Body), IdempotencyKey: "document-comment:" + comment.ID.String(), ContinuationID: parent})
	if err != nil {
		return err
	}
	if accepted && c.Tasks != nil {
		database.AfterCommit(ctx, func() {
			if err := c.Tasks.EnqueueAgentInvocation(run.ID); err != nil && c.Log != nil {
				c.Log.Error(context.WithoutCancel(ctx), "Maya document invocation awaits recovery dispatch", "run_id", run.ID, "error", err)
			}
		})
	}
	return nil
}

func (c *DocumentCommentCreator) inTransaction(ctx context.Context, operation func(context.Context) error) error {
	if c.Pool == nil || c.Documents == nil {
		return errors.New("Maya document comments are unavailable")
	}
	if _, joined := database.UnitTransaction(ctx); joined {
		return operation(ctx)
	}
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = database.WithinUnitOfWork(ctx, c.Pool, func(txCtx context.Context, _ pgx.Tx) error { return operation(txCtx) })
		if !database.IsRetryableTransactionError(err) {
			return err
		}
	}
	return err
}

type DocumentAccess struct {
	Documents  *documents.Service
	Workspace  WorkspaceAccess
	Configured bool
}

func (a DocumentAccess) CanRead(ctx context.Context, documentID, workspaceID uuid.UUID) error {
	actor, err := auth.GetActor(ctx)
	if err != nil {
		return err
	}
	_, err = a.Documents.Get(ctx, workspaceID, actor.PrincipalID, documentID)
	return err
}
func (a DocumentAccess) Enabled(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	if !a.Configured || a.Workspace == nil {
		return false, nil
	}
	return a.Workspace.WorkspaceCanUseMaya(ctx, workspaceID)
}
