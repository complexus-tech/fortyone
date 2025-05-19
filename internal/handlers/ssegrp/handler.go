package ssegrp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

// Handler manages SSE operations.
type Handler struct {
	Log        *logger.Logger
	SSEHub     *sse.Hub
	CorsOrigin string
}

func (h *Handler) StreamNotifications(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, _ := mid.GetUserID(ctx)
	h.Log.Info(ctx, "SSE connection attempt", "userID", userID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if h.CorsOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", h.CorsOrigin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	// Consider if credentials (cookies, auth headers) are sent by the frontend and if this header is needed.
	// w.Header().Set("Access-Control-Allow-Credentials", "true")

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.Log.Error(ctx, "sse: ResponseWriter does not support flushing", "userID", userID)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Streaming unsupported!")
		return fmt.Errorf("streaming unsupported: ResponseWriter does not implement http.Flusher")
	}

	sseClient := h.SSEHub.RegisterNewClient(userID)
	h.Log.Info(ctx, "SSE client registered with hub", "userID", userID)

	defer func() {
		h.SSEHub.UnregisterClient(sseClient)
		h.Log.Info(ctx, "SSE client unregistered from hub", "userID", userID)
	}()

	// 6. Initial Connection Confirmation (Optional but good practice)
	_, err := fmt.Fprintf(w, "event: connected\ndata: {\"status\": \"connected\", \"userID\": \"%s\"}\n\n", userID.String())
	if err != nil {
		h.Log.Warn(ctx, "sse: error writing initial connected event", "userID", userID, "error", err)
		// Client likely disconnected, no need to return error to framework as we've already started streaming.
		// The loop below will handle the disconnect.
		return nil // Or return err if we want to signal this specific failure more strongly upstream before loop
	}
	flusher.Flush()

	keepAliveTicker := time.NewTicker(25 * time.Second)
	defer keepAliveTicker.Stop()

	h.Log.Info(ctx, "SSE event loop starting", "userID", userID)
	for {
		select {
		case <-r.Context().Done(): // HTTP client disconnected (browser tab closed, etc.)
			h.Log.Info(ctx, "SSE client disconnected (request context done)", "userID", userID)
			return nil

		case <-sseClient.Ctx().Done(): // SSE Hub/Client context explicitly cancelled (e.g. server shutdown, or hub logic)
			h.Log.Info(ctx, "SSE client disconnected (hub client context done)", "userID", userID)
			return nil

		case messageData, ok := <-sseClient.Send:
			if !ok { // Channel closed, client was likely unregistered by the hub
				h.Log.Info(ctx, "SSE client send channel closed, assuming client unregistered", "userID", userID)
				return nil
			}
			_, err = fmt.Fprintf(w, "data: %s\n\n", messageData)
			if err != nil {
				h.Log.Warn(ctx, "sse: error writing data event to client", "userID", userID, "error", err)
				return nil
			}
			flusher.Flush()
			h.Log.Debug(ctx, "SSE message sent to client", "userID", userID, "sizeBytes", len(messageData))

		case <-keepAliveTicker.C:
			_, err = fmt.Fprintf(w, ":keep-alive\n\n") // SSE spec: comments start with a colon
			if err != nil {
				h.Log.Warn(ctx, "sse: error writing keep-alive to client", "userID", userID, "error", err)
				return nil // Client likely disconnected
			}
			flusher.Flush()
			// h.Log.Debug(ctx, "SSE keep-alive sent", "userID", userID)
		}
	}
}
