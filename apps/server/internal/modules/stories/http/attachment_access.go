package storieshttp

import (
	"context"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/web"
)

// ResolveStoryAttachment provides a durable browser URL. Session cookies are
// authorized by the route before any fresh storage signature is issued.
func (h *Handlers) ResolveStoryAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, storyID, err := storyMediaContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMediaContextHTTPStatus(err))
	}
	attachmentID, err := storyMediaAttachmentID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.stories.AuthorizeAttachmentRead(ctx, storyID, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, storyReadStatus(err))
	}
	file, err := h.attachments.ResolveStoryAttachmentAccessURL(ctx, storyID, attachmentID, workspace.ID, storyMediaAccessTTL)
	if err != nil {
		return web.RespondError(ctx, w, err, storyAttachmentStatus(err))
	}
	return respondStoryAttachmentAccess(ctx, w, r, file.URL)
}

func respondStoryAttachmentAccess(ctx context.Context, w http.ResponseWriter, r *http.Request, accessURL string) error {
	w.Header().Add("Vary", "Accept")
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Cache-Control", "private, no-store, max-age=0")
		return web.Respond(ctx, w, map[string]string{"url": accessURL}, http.StatusOK)
	}
	redirectStoryMedia(w, r, accessURL)
	return nil
}
