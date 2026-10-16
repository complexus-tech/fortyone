package desktophttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	desktoprepository "github.com/complexus-tech/projects-api/internal/modules/desktop/repository"
	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	storieshttp "github.com/complexus-tech/projects-api/internal/modules/stories/http"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const maximumOperationBytes = 15 << 20

var errInvalidOperation = errors.New("invalid desktop operation")

type mutation func(context.Context) (uuid.UUID, error)
type OperationResult struct {
	OperationVersion *string   `json:"operationVersion"`
	ResourceID       uuid.UUID `json:"resourceId"`
	Resource         any       `json:"resource,omitempty"`
	Replayed         bool      `json:"replayed"`
}

func (h *Handlers) Apply(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	workspace, userID, err := requestIdentity(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumOperationBytes)
	var operation Operation
	if err := web.DecodeWithLimit(r, &operation, maximumOperationBytes); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, web.ErrRequestBodyTooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		return web.RespondError(ctx, w, err, status)
	}
	if operation.Kind != "attachment.upload" && len(operation.Payload) > 4<<20 {
		return web.RespondError(ctx, w, web.ErrRequestBodyTooLarge, http.StatusRequestEntityTooLarge)
	}
	command, err := canonicalOperation(operation)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	apply, err := h.prepareMutation(operation, workspace.ID, userID, workspace.Slug)
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	receipt, replayed, err := h.Operations.Execute(ctx, userID, workspace.ID, operation.ID, command, func(ctx context.Context) (desktoprepository.Result, error) {
		resourceID, err := apply(ctx)
		if err != nil {
			return desktoprepository.Result{}, err
		}
		version, err := h.committedOperationVersion(ctx, workspace.ID, userID, operation.Kind, resourceID)
		return desktoprepository.Result{ResourceID: resourceID, Version: version}, err
	})
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	resourceID := receipt.ResourceID
	// Replays pass the current resource permission gate, not only the workspace
	// gate. Never return a historical private document after access was revoked.
	resource, readErr := h.operationResource(ctx, workspace.ID, userID, operation, resourceID)
	if readErr != nil && replayed {
		return web.RespondError(ctx, w, readErr, operationStatus(readErr))
	}
	if readErr != nil && h.Log != nil {
		h.Log.Warn(ctx, "desktop mutation committed; response refresh failed", "error", readErr)
	}
	if h.Cache != nil && (strings.HasPrefix(operation.Kind, "story.") || operation.Kind == "comment.create") {
		patterns := []string{fmt.Sprintf(cache.StoryListKey+"*", workspace.ID), fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID)}
		targetID := resourceID
		if operation.Kind == "comment.create" {
			var input struct {
				StoryID uuid.UUID `json:"storyId"`
			}
			_ = json.Unmarshal(operation.Payload, &input)
			targetID = input.StoryID
		}
		patterns = append(patterns, cache.InvalidateStoryKeys(workspace.ID, targetID)...)
		for _, pattern := range patterns {
			var invalidateErr error
			if strings.Contains(pattern, "*") {
				invalidateErr = h.Cache.DeleteByPattern(ctx, pattern)
			} else {
				invalidateErr = h.Cache.Delete(ctx, pattern)
			}
			if invalidateErr != nil && h.Log != nil {
				h.Log.Warn(ctx, "desktop mutation cache invalidation failed", "error", invalidateErr)
			}
		}
	}
	return web.Respond(ctx, w, OperationResult{ResourceID: resourceID, OperationVersion: receipt.Version, Resource: resource, Replayed: replayed}, http.StatusOK)
}

func canonicalOperation(operation Operation) ([]byte, error) {
	if operation.ID == uuid.Nil || operation.ResourceID == uuid.Nil || len(operation.Payload) == 0 {
		return nil, errInvalidOperation
	}
	switch operation.Kind {
	case "story.create", "story.update", "comment.create", "document.create", "document.update", "attachment.upload":
	default:
		return nil, errInvalidOperation
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(operation.Payload, &payload); err != nil || payload == nil {
		return nil, errInvalidOperation
	}
	// Canonical map ordering and compact RawMessages make receipt identity stable
	// across JSON whitespace/property-order changes without coercing integers.
	canonical, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	operation.Payload = canonical
	return json.Marshal(operation)
}

func decodePayload(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: %s", errInvalidOperation, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errInvalidOperation
	}
	return nil
}

func (h *Handlers) prepareMutation(operation Operation, workspaceID, userID uuid.UUID, workspaceSlug string) (mutation, error) {
	switch operation.Kind {
	case "attachment.upload":
		return h.prepareAttachment(operation, workspaceID, userID)
	case "story.create":
		var input storieshttp.AppNewStory
		if err := decodePayload(operation.Payload, &input); err != nil {
			return nil, err
		}
		if input.Priority == "" {
			input.Priority = "No Priority"
		}
		if err := web.ValidateStruct(&input); err != nil {
			return nil, fmt.Errorf("%w: %v", errInvalidOperation, err)
		}
		if strings.TrimSpace(input.Title) == "" || len([]rune(input.Title)) > 255 {
			return nil, errInvalidOperation
		}
		create := storieshttp.ToCoreNewStory(input, userID)
		key := fmt.Sprintf("desktop:%s:%s", userID, operation.ID)
		create.CreationKey = &key
		return func(ctx context.Context) (uuid.UUID, error) {
			story, err := h.Stories.Create(ctx, create, workspaceID)
			return story.ID, err
		}, nil
	case "story.update":
		if operation.BaseVersion == nil {
			return nil, errInvalidOperation
		}
		expected, err := time.Parse(time.RFC3339Nano, *operation.BaseVersion)
		if err != nil {
			return nil, errInvalidOperation
		}
		patch, err := decodeStoryPatch(operation.Payload)
		if err != nil {
			return nil, err
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(operation.Payload, &fields); err != nil {
			return nil, errInvalidOperation
		}
		reconcile, references, err := storieshttp.StoryMediaReconciliationRequest(fields, workspaceSlug, operation.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errInvalidOperation, err)
		}
		return func(ctx context.Context) (uuid.UUID, error) {
			if reconcile {
				return operation.ResourceID, h.Stories.UpdatePatchWithMediaReconciliationIfUnchanged(ctx, operation.ResourceID, workspaceID, expected, patch, references)
			}
			return operation.ResourceID, h.Stories.UpdatePatchIfUnchanged(ctx, operation.ResourceID, workspaceID, expected, patch)
		}, nil
	case "comment.create":
		var input struct {
			StoryID       uuid.UUID   `json:"storyId"`
			Comment       string      `json:"comment"`
			Body          string      `json:"body"`
			ParentID      *uuid.UUID  `json:"parentId"`
			Mentions      []uuid.UUID `json:"mentions"`
			AgentMentions []string    `json:"agentMentions"`
			CreationID    *uuid.UUID  `json:"creationId"`
		}
		if err := decodePayload(operation.Payload, &input); err != nil {
			return nil, err
		}
		if input.Comment != "" && input.Body != "" && input.Comment != input.Body {
			return nil, errInvalidOperation
		}
		if input.CreationID != nil && *input.CreationID != operation.ResourceID {
			return nil, errInvalidOperation
		}
		if input.Comment == "" {
			input.Comment = input.Body
		}
		if input.StoryID == uuid.Nil || strings.TrimSpace(input.Comment) == "" || len(input.Comment) > 100_000 || len(input.Mentions) > 100 {
			return nil, errInvalidOperation
		}
		return func(ctx context.Context) (uuid.UUID, error) {
			comment, err := h.Stories.CreateComment(ctx, workspaceID, stories.CoreNewComment{StoryID: input.StoryID, Comment: input.Comment, Parent: input.ParentID, Mentions: input.Mentions, AgentMentions: input.AgentMentions, CreationID: &operation.ResourceID})
			return comment.ID, err
		}, nil
	case "document.create":
		var input struct {
			Title       string                    `json:"title"`
			ContentHTML string                    `json:"contentHtml"`
			ContentText string                    `json:"contentText"`
			Visibility  documentdomain.Visibility `json:"visibility"`
		}
		if err := decodePayload(operation.Payload, &input); err != nil {
			return nil, err
		}
		if input.Visibility == "" {
			input.Visibility = documentdomain.VisibilityPrivate
		}
		return func(ctx context.Context) (uuid.UUID, error) {
			document, err := h.Documents.Create(ctx, documents.CoreCreateInput{WorkspaceID: workspaceID, UserID: userID, Title: input.Title, ContentHTML: input.ContentHTML, ContentText: input.ContentText, Visibility: input.Visibility})
			return document.ID, err
		}, nil
	case "document.update":
		if operation.BaseVersion == nil {
			return nil, errInvalidOperation
		}
		revision, err := strconv.ParseInt(*operation.BaseVersion, 10, 64)
		if err != nil || revision < 1 {
			return nil, errInvalidOperation
		}
		var input struct {
			Title              *string `json:"title"`
			ContentHTML        *string `json:"contentHtml"`
			ContentText        *string `json:"contentText"`
			CollaborationEpoch *int64  `json:"collaborationEpoch"`
		}
		if err := decodePayload(operation.Payload, &input); err != nil {
			return nil, err
		}
		if input.Title == nil && input.ContentHTML == nil && input.ContentText == nil {
			return nil, errInvalidOperation
		}
		return func(ctx context.Context) (uuid.UUID, error) {
			current, err := h.Documents.Get(ctx, workspaceID, userID, operation.ResourceID)
			if err != nil {
				return uuid.Nil, err
			}
			if current.Collaborative || (input.CollaborationEpoch != nil && *input.CollaborationEpoch != current.CollaborationEpoch) {
				return uuid.Nil, documents.ErrConflict
			}
			document, err := h.Documents.Update(ctx, documents.CoreUpdateInput{WorkspaceID: workspaceID, UserID: userID, DocumentID: operation.ResourceID, ExpectedRevision: revision, Title: input.Title, ContentHTML: input.ContentHTML, ContentText: input.ContentText})
			return document.ID, err
		}, nil
	}
	return nil, errInvalidOperation
}

func (h *Handlers) operationResource(ctx context.Context, workspaceID, userID uuid.UUID, operation Operation, id uuid.UUID) (any, error) {
	switch operation.Kind {
	case "attachment.upload":
		return h.attachmentResource(ctx, workspaceID, userID, operation, id)
	case "story.create", "story.update":
		story, err := h.Stories.Get(ctx, id, workspaceID)
		if err != nil {
			return nil, err
		}
		return mapStory(story), nil
	case "document.create", "document.update":
		document, err := h.Documents.Get(ctx, workspaceID, userID, id)
		if err != nil {
			return nil, err
		}
		return mapDocument(document), nil
	case "comment.create":
		var input struct {
			StoryID uuid.UUID `json:"storyId"`
		}
		_ = json.Unmarshal(operation.Payload, &input)
		comment, err := h.Stories.GetComment(ctx, id, input.StoryID, workspaceID)
		if err != nil {
			return nil, err
		}
		return mapComment(comment), nil
	}
	return nil, errInvalidOperation
}

func requestIdentity(ctx context.Context) (mid.WorkspaceInfo, uuid.UUID, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return workspace, uuid.Nil, err
	}
	userID, err := mid.GetUserID(ctx)
	return workspace, userID, err
}
func operationStatus(err error) int {
	switch {
	case errors.Is(err, commentsdomain.ErrCreationConflict), errors.Is(err, desktoprepository.ErrOperationConflict), errors.Is(err, stories.ErrStoryChanged), errors.Is(err, documents.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, commentsdomain.ErrNotFound), errors.Is(err, attachments.ErrNotFound), errors.Is(err, stories.ErrNotFound), errors.Is(err, documents.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, commentsdomain.ErrForbidden), errors.Is(err, attachments.ErrUnauthorized), errors.Is(err, stories.ErrStoryMutationForbidden), errors.Is(err, stories.ErrStoryReadForbidden), errors.Is(err, documents.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, attachments.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, commentsdomain.ErrInvalidComment), errors.Is(err, commentsdomain.ErrInvalidMention), errors.Is(err, attachments.ErrInvalidFile), errors.Is(err, attachments.ErrInvalidFileType), errors.Is(err, errInvalidOperation), errors.Is(err, stories.ErrInvalidStoryMutation), errors.Is(err, stories.ErrInvalidStoryReference), errors.Is(err, documents.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// These reads join the unit transaction. PostgreSQL's row write lock ensures
// the recorded version is ours, even if another writer commits immediately
// after us. Only this value may advance an outbox dependency's expected base.
func (h *Handlers) committedOperationVersion(ctx context.Context, workspaceID, userID uuid.UUID, kind string, resourceID uuid.UUID) (*string, error) {
	var version string
	switch kind {
	case "story.create", "story.update":
		story, err := h.Stories.Get(ctx, resourceID, workspaceID)
		if err != nil {
			return nil, err
		}
		version = story.UpdatedAt.UTC().Format(time.RFC3339Nano)
	case "document.create", "document.update":
		document, err := h.Documents.Get(ctx, workspaceID, userID, resourceID)
		if err != nil {
			return nil, err
		}
		version = strconv.FormatInt(document.Revision, 10)
	default:
		return nil, nil
	}
	return &version, nil
}
