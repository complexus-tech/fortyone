package ssehttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	browserSessionRevalidationInterval = 25 * time.Second
	streamWriteTimeout                 = 10 * time.Second
)

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
	_, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported: ResponseWriter does not implement http.Flusher")
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
	h.Log.Info(ctx, "SSE client registered with hub", "userID", userID, "workspaceID", workspaceID)

	initialEvent := fmt.Sprintf("event: connected\ndata: {\"status\": \"connected\", \"userID\": \"%s\", \"workspaceID\": \"%s\"}\n\n", userID.String(), workspaceID.String())
	if err := writeStreamEvent(w, initialEvent, streamWriteTimeout); err != nil {
		h.Log.Warn(ctx, "sse: Error writing initial connected event", "userID", userID, "workspaceID", workspaceID, "error", err)
		return nil
	}

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
			if !ok {
				return nil
			}
			workspace, authorized := h.currentStreamWorkspace(ctx, r, clientCtx, userID, workspaceID, workspaceSlug)
			if !authorized {
				return nil
			}
			if !h.allowsEvent(messageData, workspace.ChatEnabled) {
				continue
			}
			if err := writeStreamEvent(w, fmt.Sprintf("data: %s\n\n", messageData), streamWriteTimeout); err != nil {
				h.Log.Warn(ctx, "sse: Error writing data event to client", "userID", userID, "workspaceID", workspaceID, "error", err)
				return nil
			}
		case <-keepAlive:
			if _, authorized := h.currentStreamWorkspace(ctx, r, clientCtx, userID, workspaceID, workspaceSlug); !authorized {
				return nil
			}
			if err := writeStreamEvent(w, ":keep-alive\n\n", streamWriteTimeout); err != nil {
				h.Log.Warn(ctx, "sse: Error writing keep-alive", "userID", userID, "workspaceID", workspaceID, "error", err)
				return nil
			}
		}
	}
}

// Bound each write/flush rather than the stream lifetime. A stalled reader
// must release its handler and Redis subscription, while an idle healthy stream
// must remain usable beyond the server's ordinary request write timeout.
func writeStreamEvent(w http.ResponseWriter, event string, timeout time.Duration) error {
	controller := http.NewResponseController(w)
	deadline := time.Now().Add(timeout)
	if err := controller.SetWriteDeadline(deadline); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	if _, err := fmt.Fprint(w, event); err != nil {
		return err
	}
	if err := controller.Flush(); err != nil {
		return err
	}
	// Some middleware preserves http.Flusher but not FlushError. A timed-out
	// flush still needs to close the stream even when its error was not exposed.
	if !time.Now().Before(deadline) {
		return context.DeadlineExceeded
	}
	if err := controller.SetWriteDeadline(time.Time{}); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	return nil
}

func (h *Handler) currentStreamWorkspace(
	ctx context.Context,
	r *http.Request,
	clientCtx context.Context,
	expectedUserID uuid.UUID,
	workspaceID uuid.UUID,
	workspaceSlug string,
) (mid.WorkspaceInfo, bool) {
	if ctx.Err() != nil || r.Context().Err() != nil || clientCtx.Err() != nil {
		return mid.WorkspaceInfo{}, false
	}

	userID, ok, err := mid.ResolveSessionUserID(ctx, r, h.BrowserSessions)
	if err != nil {
		// Authentication and cache errors are intentionally confined to server
		// logs. An SSE client learns only that its stream has ended.
		h.Log.Warn(ctx, "sse: Browser session revalidation failed; closing stream", "userID", expectedUserID, "workspaceID", workspaceID, "error", err)
		return mid.WorkspaceInfo{}, false
	}
	if !ok || userID != expectedUserID {
		h.Log.Info(ctx, "sse: Browser session is no longer valid; closing stream", "userID", expectedUserID, "workspaceID", workspaceID)
		return mid.WorkspaceInfo{}, false
	}

	if h.WorkspaceAccess == nil {
		h.Log.Error(ctx, "sse: Workspace access resolver is unavailable; closing stream", "userID", expectedUserID, "workspaceID", workspaceID)
		return mid.WorkspaceInfo{}, false
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
		return mid.WorkspaceInfo{}, false
	}

	// Cancellation can race the backing-store lookup. Recheck both lifetimes so
	// a successful stale lookup cannot authorize one final write after teardown.
	return workspace, ctx.Err() == nil && r.Context().Err() == nil && clientCtx.Err() == nil
}

// Shared notification streams remain available while Messages is disabled, but
// old replicas and already queued chat events cannot leak through this stream.
func (h *Handler) allowsEvent(payload []byte, chatEnabled bool) bool {
	var event struct {
		Type             string `json:"type"`
		EntityType       string `json:"entityType"`
		StoredEntityType string `json:"entity_type"`
	}
	if json.Unmarshal(payload, &event) != nil {
		return false
	}
	if strings.HasPrefix(event.Type, "chat.") || event.Type == "chat_message" || event.EntityType == "chat_message" || event.StoredEntityType == "chat_message" {
		return chatEnabled
	}
	return true
}
