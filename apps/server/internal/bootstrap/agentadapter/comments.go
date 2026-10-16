// Package agentadapter connects product surfaces to the reusable invocation
// lifecycle. It owns cross-module composition, never persistence or model policy.
package agentadapter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/agentmentions"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const StoryCommentSurface = "story_comment"

type WorkspaceAccess interface {
	WorkspaceCanUseMaya(context.Context, uuid.UUID) (bool, error)
}

type Wakeup interface{ EnqueueAgentInvocation(uuid.UUID) error }

type CommentCreator struct {
	Pool    *pgxpool.Pool
	Base    stories.CommentCreator
	Stories *stories.Service
	Runs    *agentruns.Service
	MayaID  uuid.UUID
	Tasks   Wakeup
	Log     *logger.Logger
}

type commentPayload struct {
	StoryTitle    string `json:"storyTitle"`
	WorkspaceSlug string `json:"workspaceSlug"`
}

func SourceHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// CreateComment commits the source and invocation together. The pending run is
// the durable dispatch record; the immediate queue wake-up is only a latency aid.
func (c *CommentCreator) CreateComment(ctx context.Context, input stories.CreateCommentCommand) (stories.CoreComment, error) {
	explicit := len(input.AgentMentions) > 0
	if !explicit && input.ParentID == nil {
		return c.Base.CreateComment(ctx, input)
	}
	if input.Actor.Kind != auth.PrincipalHumanUser {
		if !explicit {
			return c.Base.CreateComment(ctx, input)
		}
		return stories.CoreComment{}, auth.ErrInvalidActor
	}
	if explicit {
		if _, err := agentmentions.ValidateMaya(input.Content, input.AgentMentions, c.MayaID); err != nil {
			return stories.CoreComment{}, err
		}
	} else {
		parent, err := c.Stories.GetComment(ctx, *input.ParentID, input.StoryID, input.WorkspaceID)
		if err != nil {
			return stories.CoreComment{}, err
		}
		if parent.UserID != c.MayaID {
			return c.Base.CreateComment(ctx, input)
		}
	}
	if c.Runs == nil || c.Pool == nil {
		return stories.CoreComment{}, errors.New("Maya comment requests are unavailable")
	}
	story, err := c.Stories.Get(ctx, input.StoryID, input.WorkspaceID)
	if err != nil {
		return stories.CoreComment{}, err
	}
	rootID := uuid.Nil
	if input.ParentID != nil {
		root, err := c.Stories.GetCommentThread(ctx, *input.ParentID, input.StoryID, input.WorkspaceID)
		if err != nil {
			return stories.CoreComment{}, err
		}
		rootID = root.ID
	}
	workspace, _ := mid.GetWorkspace(ctx)
	payload, err := json.Marshal(commentPayload{StoryTitle: story.Title, WorkspaceSlug: workspace.Slug})
	if err != nil {
		return stories.CoreComment{}, err
	}
	var created stories.CoreComment
	operation := func(txctx context.Context) error {
		var createErr error
		created, createErr = c.Base.CreateComment(txctx, input)
		if createErr != nil {
			return createErr
		}
		if rootID == uuid.Nil {
			rootID = created.ID
		}
		var continuationID *uuid.UUID
		if input.ParentID != nil {
			continuationID, createErr = c.Runs.ResolveContinuation(txctx, agentruns.ContinuationQuery{
				WorkspaceID: input.WorkspaceID, RequesterID: input.Actor.PrincipalID, Surface: StoryCommentSurface,
				EntityID: input.StoryID, ConversationID: rootID, ReplyToID: *input.ParentID, SourceID: created.ID,
			})
			if createErr != nil {
				return createErr
			}
		}
		// A plain reply only invokes Maya when it answers her persisted question.
		// The responding human owns the new run and supplies its permissions.
		if !explicit && continuationID == nil {
			return nil
		}
		run, accepted, acceptErr := c.Runs.Accept(txctx, agentruns.Input{
			WorkspaceID: input.WorkspaceID, RequesterID: input.Actor.PrincipalID, TeamID: story.Team,
			Surface: StoryCommentSurface, SourceID: created.ID, ConversationID: rootID, EntityID: input.StoryID,
			SourceRevision: created.UpdatedAt.UTC().Format(time.RFC3339Nano), SourceHash: SourceHash(created.Comment),
			IdempotencyKey: "comment:" + created.ID.String(), ContinuationID: continuationID, Payload: payload,
		})
		if acceptErr != nil {
			return acceptErr
		}
		if accepted && c.Tasks != nil {
			platformdatabase.AfterCommit(txctx, func() {
				if err := c.Tasks.EnqueueAgentInvocation(run.ID); err != nil && c.Log != nil {
					c.Log.Error(context.WithoutCancel(ctx), "Maya invocation awaits recovery dispatch", "run_id", run.ID, "error", err)
				}
			})
		}
		return nil
	}
	if _, joined := platformdatabase.UnitTransaction(ctx); joined {
		err = operation(ctx)
	} else {
		err = platformdatabase.WithinUnitOfWork(ctx, c.Pool, func(txctx context.Context, _ pgx.Tx) error { return operation(txctx) })
	}
	return created, err
}
