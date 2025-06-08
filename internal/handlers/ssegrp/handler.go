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
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// Handler manages SSE operations.
type Handler struct {
	Log        *logger.Logger
	SSEHub     *sse.Hub
	CorsOrigin string
}

func (h *Handler) StreamNotifications(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, _ := mid.GetUserID(ctx)

	// NEW: Get workspace from URL path parameter
	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		h.Log.Error(ctx, "sse: Invalid workspace ID", "userID", userID, "workspaceParam", workspaceIDParam, "error", err)
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return fmt.Errorf("invalid workspace ID: %w", err)
	}

	h.Log.Info(ctx, "SSE connection attempt (hijack) for user in workspace", "userID", userID, "workspaceID", workspaceID)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		h.Log.Error(ctx, "sse (hijack): ResponseWriter does not support hijacking", "userID", userID, "workspaceID", workspaceID)
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return fmt.Errorf("hijacking unsupported: ResponseWriter does not implement http.Hijacker")
	}

	conn, bufRW, err := hijacker.Hijack()
	if err != nil {
		h.Log.Error(ctx, "sse (hijack): Failed to hijack connection", "userID", userID, "workspaceID", workspaceID, "error", err)
		return fmt.Errorf("failed to hijack connection: %w", err)
	}
	defer func() {
		h.Log.Info(ctx, "SSE (hijack): Closing hijacked connection", "userID", userID, "workspaceID", workspaceID)
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
		h.Log.Error(ctx, "sse (hijack): Failed to write initial headers", "userID", userID, "workspaceID", workspaceID, "error", err)
		return nil
	}

	// Register client with both userID and workspaceID
	sseClient := h.SSEHub.RegisterNewClient(userID, workspaceID)
	h.Log.Info(ctx, "SSE (hijack): Client registered with hub", "userID", userID, "workspaceID", workspaceID)
	defer func() {
		h.SSEHub.UnregisterClient(sseClient)
		h.Log.Info(ctx, "SSE (hijack): Client unregistered from hub", "userID", userID, "workspaceID", workspaceID)
	}()

	// Include workspaceID in initial connected event
	initialEvent := fmt.Sprintf("event: connected\ndata: {\"status\": \"connected\", \"userID\": \"%s\", \"workspaceID\": \"%s\"}\n\n", userID.String(), workspaceID.String())
	_, err = bufRW.WriteString(initialEvent)
	if err == nil {
		err = bufRW.Flush()
	}
	if err != nil {
		h.Log.Warn(ctx, "sse (hijack): Error writing initial connected event", "userID", userID, "workspaceID", workspaceID, "error", err)
		return nil
	}

	keepAliveTicker := time.NewTicker(25 * time.Second)
	defer keepAliveTicker.Stop()

	h.Log.Info(ctx, "SSE (hijack): Event loop starting", "userID", userID, "workspaceID", workspaceID)

	for {
		select {
		case <-r.Context().Done():
			h.Log.Info(ctx, "SSE (hijack): Request context done", "userID", userID, "workspaceID", workspaceID, "error", r.Context().Err())
			return nil

		case <-sseClient.Ctx().Done():
			h.Log.Info(ctx, "SSE (hijack): Hub client context done", "userID", userID, "workspaceID", workspaceID)
			return nil

		case messageData, ok_chan := <-sseClient.Send:
			if !ok_chan {
				h.Log.Info(ctx, "SSE (hijack): Hub send channel closed", "userID", userID, "workspaceID", workspaceID)
				return nil
			}
			dataLine := fmt.Sprintf("data: %s\n\n", messageData)
			_, err = bufRW.WriteString(dataLine)
			if err == nil {
				err = bufRW.Flush()
			}
			if err != nil {
				h.Log.Warn(ctx, "SSE (hijack): Error writing data event to client", "userID", userID, "workspaceID", workspaceID, "error", err)
				return nil
			}

		case <-keepAliveTicker.C:
			pingData := ":keep-alive\n\n"
			_, err = bufRW.WriteString(pingData)
			if err == nil {
				err = bufRW.Flush()
			}
			if err != nil {
				h.Log.Warn(ctx, "SSE (hijack): Error writing keep-alive to client", "userID", userID, "workspaceID", workspaceID, "error", err)
				return nil
			}
		}
	}
}
