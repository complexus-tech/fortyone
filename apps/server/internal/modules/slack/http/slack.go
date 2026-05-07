package slackhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	log     *logger.Logger
	service *slack.Service
}

func New(log *logger.Logger, service *slack.Service) *Handlers {
	return &Handlers{log: log, service: service}
}

func (h *Handlers) HandleEvents(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleEvents(rawBody)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if response.Challenge != "" {
		return writeRawJSON(w, http.StatusOK, response)
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handlers) HandleCommands(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleCommand(ctx, rawBody)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return writeRawJSON(w, http.StatusOK, response)
}

func (h *Handlers) HandleInteractivity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleInteractivity(ctx, rawBody)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	}
	status := response.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	if len(response.Body) > 0 {
		w.WriteHeader(status)
		_, writeErr := w.Write(response.Body)
		return writeErr
	}
	w.WriteHeader(status)
	return nil
}

func writeRawJSON(w http.ResponseWriter, status int, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}
