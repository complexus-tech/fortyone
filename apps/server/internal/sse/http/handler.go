package ssehttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const browserSessionRevalidationInterval = 25 * time.Second

type Handler struct {
	Log             *logger.Logger
	SSEHub          *sse.Hub
	Origins         web.OriginPolicy
	BrowserSessions mid.SessionResolver
	WorkspaceAccess mid.WorkspaceResolver

	// keepAlive is injected only by focused stream tests. Production streams
	// leave it nil and use browserSessionRevalidationInterval.
	keepAlive <-chan time.Time
}

func (h *Handler) StreamNotifications(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		h.Log.Error(ctx, "sse: User not authenticated", "error", err)
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return fmt.Errorf("user not authenticated: %w", err)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.Log.Error(ctx, "sse: Invalid workspace", "userID", userID, "error", err)
		http.Error(w, "Invalid workspace", http.StatusBadRequest)
		return fmt.Errorf("invalid workspace: %w", err)
	}

	_, ok := w.(http.Flusher)
	if !ok {
		h.Log.Error(ctx, "sse: ResponseWriter does not support streaming", "userID", userID, "workspaceID", workspace.ID)
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return fmt.Errorf("streaming unsupported: ResponseWriter does not implement http.Flusher")
	}

	sseClient, err := h.SSEHub.RegisterNewClient(userID, workspace.ID)
	if err != nil {
		h.Log.Warn(ctx, "sse: Hub is not accepting clients", "userID", userID, "workspaceID", workspace.ID, "error", err)
		http.Error(w, "Event stream is not ready", http.StatusServiceUnavailable)
		return nil
	}
	defer func() {
		h.SSEHub.UnregisterClient(sseClient)
		h.Log.Info(context.Background(), "SSE client unregistered from hub", "userID", userID, "workspaceID", workspace.ID)
	}()

	return h.serveStream(ctx, w, r, userID, workspace.ID, workspace.Slug, sseClient.Ctx(), sseClient.Send)
}

func (h *Handler) serveStream(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	workspaceSlug string,
	clientCtx context.Context,
	messages <-chan []byte,
) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported: ResponseWriter does not implement http.Flusher")
	}

	// A global HTTP write timeout is useful for ordinary endpoints but would
	// otherwise impose a fixed lifetime on this long-lived response.
	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil && !errors.Is(err, http.ErrNotSupported) {
		h.Log.Warn(ctx, "sse: Failed to clear stream write deadline", "userID", userID, "workspaceID", workspaceID, "error", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	allowedOrigin := h.Origins.AllowedOrigin(r)
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Vary", "Origin")
	}
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	h.Log.Info(ctx, "SSE client registered with hub", "userID", userID, "workspaceID", workspaceID)

	initialEvent := fmt.Sprintf("event: connected\ndata: {\"status\": \"connected\", \"userID\": \"%s\", \"workspaceID\": \"%s\"}\n\n", userID.String(), workspaceID.String())
	if _, err := fmt.Fprint(w, initialEvent); err != nil {
		h.Log.Warn(ctx, "sse: Error writing initial connected event", "userID", userID, "workspaceID", workspaceID, "error", err)
		return nil
	}
	flusher.Flush()

	keepAlive := h.keepAlive
	if keepAlive == nil {
		keepAliveTicker := time.NewTicker(browserSessionRevalidationInterval)
		defer keepAliveTicker.Stop()
		keepAlive = keepAliveTicker.C
	}

	for {
		select {
		case <-r.Context().Done():
			h.Log.Info(context.Background(), "SSE request context done", "userID", userID, "workspaceID", workspaceID, "error", r.Context().Err())
			return nil
		case <-clientCtx.Done():
			h.Log.Info(context.Background(), "SSE hub client context done", "userID", userID, "workspaceID", workspaceID)
			return nil
		case messageData, ok := <-messages:
			if !ok || !h.streamAuthorizationIsCurrent(ctx, r, clientCtx, userID, workspaceID, workspaceSlug) {
				return nil
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", messageData); err != nil {
				h.Log.Warn(ctx, "sse: Error writing data event to client", "userID", userID, "workspaceID", workspaceID, "error", err)
				return nil
			}
			flusher.Flush()
		case <-keepAlive:
			if !h.streamAuthorizationIsCurrent(ctx, r, clientCtx, userID, workspaceID, workspaceSlug) {
				return nil
			}
			if _, err := fmt.Fprint(w, ":keep-alive\n\n"); err != nil {
				h.Log.Warn(ctx, "sse: Error writing keep-alive", "userID", userID, "workspaceID", workspaceID, "error", err)
				return nil
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) streamAuthorizationIsCurrent(
	ctx context.Context,
	r *http.Request,
	clientCtx context.Context,
	expectedUserID uuid.UUID,
	workspaceID uuid.UUID,
	workspaceSlug string,
) bool {
	if ctx.Err() != nil || r.Context().Err() != nil || clientCtx.Err() != nil {
		return false
	}

	userID, ok, err := mid.ResolveSessionUserID(ctx, r, h.BrowserSessions)
	if err != nil {
		// Authentication and cache errors are intentionally confined to server
		// logs. An SSE client learns only that its stream has ended.
		h.Log.Warn(ctx, "sse: Browser session revalidation failed; closing stream", "userID", expectedUserID, "workspaceID", workspaceID, "error", err)
		return false
	}
	if !ok || userID != expectedUserID {
		h.Log.Info(ctx, "sse: Browser session is no longer valid; closing stream", "userID", expectedUserID, "workspaceID", workspaceID)
		return false
	}

	if h.WorkspaceAccess == nil {
		h.Log.Error(ctx, "sse: Workspace access resolver is unavailable; closing stream", "userID", expectedUserID, "workspaceID", workspaceID)
		return false
	}
	workspace, err := h.WorkspaceAccess.ResolveCurrentWorkspace(ctx, workspaceSlug, expectedUserID)
	if err != nil || workspace.ID != workspaceID {
		// Workspace authorization failures are confined to server logs. The
		// stream closes without distinguishing membership removal, workspace
		// deletion, or an unavailable backing store to the client.
		if err != nil {
			h.Log.Warn(ctx, "sse: Workspace access revalidation failed; closing stream", "userID", expectedUserID, "workspaceID", workspaceID, "error", err)
		} else {
			h.Log.Warn(ctx, "sse: Workspace access changed; closing stream", "userID", expectedUserID, "workspaceID", workspaceID)
		}
		return false
	}

	// Cancellation can race the backing-store lookup. Recheck both lifetimes so
	// a successful stale lookup cannot authorize one final write after teardown.
	return ctx.Err() == nil && r.Context().Err() == nil && clientCtx.Err() == nil
}
