package ssegrp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

type Handler struct {
	Log        *logger.Logger
	SSEHub     *sse.Hub
	CorsOrigin string
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

	h.Log.Info(ctx, "SSE connection attempt (hijack) for user in workspace", "userID", userID, "workspaceID", workspace.ID)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		h.Log.Error(ctx, "sse (hijack): ResponseWriter does not support hijacking", "userID", userID, "workspaceID", workspace.ID)
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return fmt.Errorf("hijacking unsupported: ResponseWriter does not implement http.Hijacker")
	}

	conn, bufRW, err := hijacker.Hijack()
	if err != nil {
		h.Log.Error(ctx, "sse (hijack): Failed to hijack connection", "userID", userID, "workspaceID", workspace.ID, "error", err)
		return fmt.Errorf("failed to hijack connection: %w", err)
	}
	defer func() {
		h.Log.Info(ctx, "SSE (hijack): Closing hijacked connection", "userID", userID, "workspaceID", workspace.ID)
		conn.Close()
	}()

	var headers strings.Builder
	headers.WriteString("HTTP/1.1 200 OK\r\n")
	headers.WriteString("Content-Type: text/event-stream\r\n")
	headers.WriteString("Cache-Control: no-cache\r\n")
	headers.WriteString("Connection: keep-alive\r\n")
	if h.CorsOrigin != "" {
		headers.WriteString(fmt.Sprintf("Access-Control-Allow-Origin: %s\r\n", h.CorsOrigin))
	} else {
		headers.WriteString("Access-Control-Allow-Origin: *\r\n")
	}
	headers.WriteString("X-Accel-Buffering: no\r\n")
	headers.WriteString("\r\n")

	_, err = bufRW.WriteString(headers.String())
	if err == nil {
		err = bufRW.Flush()
	}
	if err != nil {
		h.Log.Error(ctx, "sse (hijack): Failed to write initial headers", "userID", userID, "workspaceID", workspace.ID, "error", err)
		return nil
	}

	sseClient := h.SSEHub.RegisterNewClient(userID, workspace.ID)
	h.Log.Info(ctx, "SSE (hijack): Client registered with hub", "userID", userID, "workspaceID", workspace.ID)
	defer func() {
		h.SSEHub.UnregisterClient(sseClient)
		h.Log.Info(ctx, "SSE (hijack): Client unregistered from hub", "userID", userID, "workspaceID", workspace.ID)
	}()

	initialEvent := fmt.Sprintf("event: connected\ndata: {\"status\": \"connected\", \"userID\": \"%s\", \"workspaceID\": \"%s\"}\n\n", userID.String(), workspace.ID.String())
	_, err = bufRW.WriteString(initialEvent)
	if err == nil {
		err = bufRW.Flush()
	}
	if err != nil {
		h.Log.Warn(ctx, "sse (hijack): Error writing initial connected event", "userID", userID, "workspaceID", workspace.ID, "error", err)
		return nil
	}

	keepAliveTicker := time.NewTicker(25 * time.Second)
	defer keepAliveTicker.Stop()

	h.Log.Info(ctx, "SSE (hijack): Event loop starting", "userID", userID, "workspaceID", workspace.ID)

	for {
		select {
		case <-r.Context().Done():
			h.Log.Info(ctx, "SSE (hijack): Request context done", "userID", userID, "workspaceID", workspace.ID, "error", r.Context().Err())
			return nil

		case <-sseClient.Ctx().Done():
			h.Log.Info(ctx, "SSE (hijack): Hub client context done", "userID", userID, "workspaceID", workspace.ID)
			return nil

		case messageData, ok_chan := <-sseClient.Send:
			if !ok_chan {
				h.Log.Info(ctx, "SSE (hijack): Hub send channel closed", "userID", userID, "workspaceID", workspace.ID)
				return nil
			}
			dataLine := fmt.Sprintf("data: %s\n\n", messageData)
			_, err = bufRW.WriteString(dataLine)
			if err == nil {
				err = bufRW.Flush()
			}
			if err != nil {
				h.Log.Warn(ctx, "SSE (hijack): Error writing data event to client", "userID", userID, "workspaceID", workspace.ID, "error", err)
				return nil
			}

		case <-keepAliveTicker.C:
			pingData := ":keep-alive\n\n"
			_, err = bufRW.WriteString(pingData)
			if err == nil {
				err = bufRW.Flush()
			}
			if err != nil {
				h.Log.Warn(ctx, "SSE (hijack): Error writing keep-alive to client", "userID", userID, "workspaceID", workspace.ID, "error", err)
				return nil
			}
		}
	}
}
