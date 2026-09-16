package mayahttp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/google/uuid"
)

func (h *Handlers) executeDeleteStory(ctx context.Context, workspaceID, userID, sessionID uuid.UUID, rawArgs json.RawMessage, isAdmin bool) (AppRealtimeToolResponse, error) {
	var args AppRealtimeDeleteStoryArguments
	if err := decodeRealtimeArguments(rawArgs, &args, "delete_story"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	story, response, err := h.resolveRealtimeStory(ctx, workspaceID, userID, strings.TrimSpace(args.Reference))
	if err != nil || response != nil {
		return responseOrEmpty(response), err
	}
	authorization := stories.BulkDeleteAuthorization{ActorID: userID, IsAdmin: isAdmin}
	result, err := h.confirmAndDeleteRealtimeStory(sessionID, workspaceID, story, args, authorization, func() error {
		// The ordinary story service rechecks current membership, visibility, and
		// admin/creator authorization in the deletion transaction.
		return h.stories.Delete(ctx, story.ID, workspaceID, authorization)
	})
	if err == nil && result.Success {
		h.invalidateStoryListCaches(ctx, workspaceID)
		if h.cache != nil {
			for _, key := range cache.InvalidateStoryKeys(workspaceID, story.ID) {
				var cacheErr error
				if strings.Contains(key, "*") {
					cacheErr = h.cache.DeleteByPattern(ctx, key)
				} else {
					cacheErr = h.cache.Delete(ctx, key)
				}
				if cacheErr != nil {
					h.log.Warn(ctx, "failed to invalidate deleted maya story cache", "story_id", story.ID, "error", cacheErr)
				}
			}
		}
	}
	return result, err
}

func (h *Handlers) confirmAndDeleteRealtimeStory(sessionID, workspaceID uuid.UUID, story stories.CoreSingleStory, args AppRealtimeDeleteStoryArguments, authorization stories.BulkDeleteAuthorization, deleteStory func() error) (AppRealtimeToolResponse, error) {
	if story.Workspace != workspaceID || authorization.ActorID == uuid.Nil || (!authorization.IsAdmin && (story.Reporter == nil || *story.Reporter != authorization.ActorID)) {
		return AppRealtimeToolResponse{Success: false, Error: "Only workspace admins or the story creator can delete this work item."}, nil
	}
	if story.DeletedAt != nil {
		return AppRealtimeToolResponse{Success: false, Error: "This work item is already in trash."}, nil
	}
	reference := storyReference(story.TeamCode, story.SequenceID)
	// Bind the displayed target, version, and actor as well as the voice session.
	// Renaming or modifying the story after preview requires a fresh approval.
	input := struct {
		WorkspaceID uuid.UUID
		ActorID     uuid.UUID
		StoryID     uuid.UUID
		Title       string
		Reference   string
		UpdatedAt   string
	}{workspaceID, authorization.ActorID, story.ID, story.Title, reference, story.UpdatedAt.UTC().Format(time.RFC3339Nano)}
	token, err := h.confirmationToken(sessionID, "delete_story", input)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	preview := AppRealtimeToolResponse{
		Success: false, RequiresConfirmation: true, ConfirmationToken: token,
		Confirmation: &AppRealtimeConfirmation{Title: story.Title, Description: fmt.Sprintf("Move %s to trash.", reference)},
		Message:      fmt.Sprintf("Confirm moving %s: %q to trash.", reference, story.Title),
	}
	if !args.Confirmed {
		return preview, nil
	}
	valid, err := h.validateConfirmationToken(sessionID, "delete_story", input, args.ConfirmationToken)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !valid {
		return preview, nil
	}
	if err := deleteStory(); err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("delete story: %w", err)
	}
	return AppRealtimeToolResponse{Success: true, Message: fmt.Sprintf("Moved %s: %q to trash.", reference, story.Title)}, nil
}
