package storieshttp

import (
	"context"
	"fmt"
	"net/http"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) GetChecklist(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil || storyID == uuid.Nil {
		return web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
	}
	items, err := h.stories.ListChecklist(ctx, workspace.ID, storyID)
	if err != nil {
		return web.RespondError(ctx, w, err, storyReadStatus(err))
	}
	return web.Respond(ctx, w, items, http.StatusOK)
}

func (h *Handlers) AddChecklistItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Title string `json:"title"`
	}
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return h.mutateChecklist(ctx, w, r, storydomain.ChecklistCommand{Action: storydomain.ChecklistAdd, Title: &req.Title})
}

func (h *Handlers) UpdateChecklistItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Title     web.PatchField[string] `json:"title"`
		Completed web.PatchField[bool]   `json:"completed"`
	}
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	title, titleSet := req.Title.Value()
	completed, completedSet := req.Completed.Value()
	if (titleSet && title == nil) || (completedSet && completed == nil) {
		return web.RespondError(ctx, w, fmt.Errorf("checklist title and completion cannot be null"), http.StatusBadRequest)
	}
	return h.mutateChecklist(ctx, w, r, storydomain.ChecklistCommand{Action: storydomain.ChecklistUpdate, Title: title, Completed: completed})
}

func (h *Handlers) DeleteChecklistItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.mutateChecklist(ctx, w, r, storydomain.ChecklistCommand{Action: storydomain.ChecklistDelete})
}
func (h *Handlers) DeleteChecklist(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.mutateChecklist(ctx, w, r, storydomain.ChecklistCommand{Action: storydomain.ChecklistClear})
}
func (h *Handlers) ConvertChecklistItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.mutateChecklist(ctx, w, r, storydomain.ChecklistCommand{Action: storydomain.ChecklistConvert})
}
func (h *Handlers) ReorderChecklist(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req struct {
		ItemIDs []uuid.UUID `json:"itemIds"`
	}
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return h.mutateChecklist(ctx, w, r, storydomain.ChecklistCommand{Action: storydomain.ChecklistReorder, ItemIDs: req.ItemIDs})
}

func (h *Handlers) mutateChecklist(ctx context.Context, w http.ResponseWriter, r *http.Request, command storydomain.ChecklistCommand) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	command.StoryID, err = uuid.Parse(web.Params(r, "id"))
	if err != nil || command.StoryID == uuid.Nil {
		return web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
	}
	switch command.Action {
	case storydomain.ChecklistUpdate, storydomain.ChecklistDelete, storydomain.ChecklistConvert:
		command.ItemID, err = uuid.Parse(web.Params(r, "itemId"))
		if err != nil || command.ItemID == uuid.Nil {
			return web.RespondError(ctx, w, fmt.Errorf("invalid checklist item id"), http.StatusBadRequest)
		}
	}
	result, itemID, err := h.stories.MutateChecklist(ctx, workspace.ID, command)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMutationStatus(err))
	}
	switch command.Action {
	case storydomain.ChecklistDelete, storydomain.ChecklistClear:
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	case storydomain.ChecklistReorder:
		return web.Respond(ctx, w, result.Items, http.StatusOK)
	default:
		for _, item := range result.Items {
			if item.ID == itemID {
				status := http.StatusOK
				if command.Action == storydomain.ChecklistAdd {
					status = http.StatusCreated
				}
				return web.Respond(ctx, w, item, status)
			}
		}
		return fmt.Errorf("checklist mutation returned no item")
	}
}
