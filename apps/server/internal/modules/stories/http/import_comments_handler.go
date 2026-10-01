package storieshttp

import (
	"context"
	"errors"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"html"
	"net/http"
	"strings"
	"time"
)

type AppImportComment struct {
	StoryID        uuid.UUID  `json:"storyId" validate:"required"`
	SourceID       string     `json:"sourceId" validate:"required,max=300"`
	Content        string     `json:"content" validate:"required,max=9000"`
	AuthorName     string     `json:"authorName" validate:"required,max=255"`
	CreatedAt      *time.Time `json:"createdAt"`
	ParentSourceID *string    `json:"parentSourceId,omitempty"`
	Format         string     `json:"format" validate:"omitempty,oneof=text html"`
}
type AppImportCommentsRequest struct {
	Items []AppImportComment `json:"items" validate:"required,min=1,max=50,dive"`
}
type AppImportCommentResult struct {
	SourceID  string                   `json:"sourceId"`
	StoryID   uuid.UUID                `json:"storyId"`
	CommentID *uuid.UUID               `json:"commentId"`
	Error     *AppStoryImportItemError `json:"error"`
}
type storyImportCommentService interface {
	CreateCommentImport(context.Context, uuid.UUID, uuid.UUID, stories.CoreNewComment) (stories.CoreComment, error)
}

func importCommentID(workspaceID, storyID uuid.UUID, sourceID string) uuid.UUID {
	return uuid.NewSHA1(workspaceID, []byte("story-import-comment:v1:"+storyID.String()+":"+sourceID))
}
func importCommentContent(item AppImportComment) string {
	sourceTime := "unknown source time"
	if item.CreatedAt != nil {
		sourceTime = item.CreatedAt.UTC().Format(time.RFC3339)
	}
	attribution := "Imported from source · " + item.AuthorName + " · " + sourceTime
	content := item.Content
	if item.Format == "html" {
		content = safeImportedCommentHTML(content)
	}
	if item.Format != "html" {
		content = "<p>" + strings.ReplaceAll(html.EscapeString(content), "\n", "<br>") + "</p>"
	}
	return "<p><em>" + html.EscapeString(attribution) + "</em></p>" + content
}
func (h *Handlers) ImportComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	service, ok := h.storyImporter.(storyImportCommentService)
	if !ok {
		return web.RespondError(ctx, w, errors.New("Comment import unavailable"), http.StatusServiceUnavailable)
	}
	var request AppImportCommentsRequest
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	results := make([]AppImportCommentResult, 0, len(request.Items))
	for _, item := range request.Items {
		result := AppImportCommentResult{SourceID: item.SourceID, StoryID: item.StoryID}
		if err := validateImportSourceNamespace(item.SourceID); err != nil {
			result.Error = &AppStoryImportItemError{Code: "invalid_source_comment", Message: "The source comment ID is invalid."}
			results = append(results, result)
			continue
		}
		id := importCommentID(workspace.ID, item.StoryID, item.SourceID)
		var parentID *uuid.UUID
		if item.ParentSourceID != nil {
			parent := importCommentID(workspace.ID, item.StoryID, *item.ParentSourceID)
			parentID = &parent
		}
		input := stories.CoreNewComment{StoryID: item.StoryID, CreationID: &id, Comment: importCommentContent(item), Parent: parentID}
		created, err := service.CreateCommentImport(ctx, actorID, workspace.ID, input)
		if err != nil {
			result.Error = storyImportItemError(err)
		} else {
			result.CommentID = &created.ID
		}
		results = append(results, result)
	}
	return web.Respond(ctx, w, struct {
		Items []AppImportCommentResult `json:"items"`
	}{Items: results}, http.StatusOK)
}
