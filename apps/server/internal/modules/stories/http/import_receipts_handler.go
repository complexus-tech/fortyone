package storieshttp

import (
	"context"
	"errors"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handlers) ImportReceipts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	query := r.URL.Query()
	provider, digest := query.Get("provider"), query.Get("sourceDigest")
	var namespace *string
	if query.Has("sourceNamespace") {
		value := query.Get("sourceNamespace")
		namespace = &value
	}
	validation := AppStoryImportRequest{Provider: provider, SourceDigest: digest, SourceNamespace: namespace}
	if provider != storyImportProviderFile && provider != storyImportProviderJiraCSV || len(digest) != 64 || strings.Trim(digest, "0123456789abcdef") != "" {
		return web.RespondError(ctx, w, errors.New("Invalid import source"), http.StatusBadRequest)
	}
	if namespace != nil {
		if err := validateImportSourceNamespace(*namespace); err != nil {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}
	}
	offset := 0
	if value := query.Get("offset"); value != "" {
		parsed, e := strconv.Atoi(value)
		if e != nil || parsed < 0 || parsed > 100000 {
			return web.RespondError(ctx, w, errors.New("Invalid receipt page"), http.StatusBadRequest)
		}
		offset = parsed
	}
	store, ok := h.storyImporter.(storyImportReceiptService)
	if !ok {
		return web.RespondError(ctx, w, errors.New("Import receipts unavailable"), http.StatusServiceUnavailable)
	}
	items, err := store.ListImportReceipts(ctx, actorID, workspace.ID, validation.Provider, validation.SourceDigest, namespace, offset)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMutationStatus(err))
	}
	return web.Respond(ctx, w, struct {
		Items   any  `json:"items"`
		HasMore bool `json:"hasMore"`
	}{Items: items, HasMore: len(items) == 500}, http.StatusOK)
}
