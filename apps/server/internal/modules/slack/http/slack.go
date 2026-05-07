package slackhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
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

func (h *Handlers) GetIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	integration, err := h.service.GetIntegration(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppIntegration(integration), http.StatusOK)
}

func (h *Handlers) GetRequestLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil {
			return web.RespondError(ctx, w, parseErr, http.StatusBadRequest)
		}
		limit = parsed
	}
	logs, err := h.service.GetRequestLogs(ctx, workspace.ID, limit)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppRequestLogs(logs), http.StatusOK)
}

func (h *Handlers) CreateInstallSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	session, err := h.service.CreateInstallSession(ctx, workspace.ID, userID, workspace.Slug)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, AppCreateInstallSession{InstallURL: session.InstallURL}, http.StatusOK)
}

func (h *Handlers) DisconnectWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.DisconnectWorkspace(ctx, workspace.ID); err != nil {
		if slack.IsNotFound(err) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) HandleSetup(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	slackError := r.URL.Query().Get("error")
	redirectURL, err := h.service.HandleSetup(ctx, code, state, slackError)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) ResyncChannels(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.SyncChannels(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) HandleEvents(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	statusCode := http.StatusOK
	outcome := "received"
	errorMessage := ""
	headers := captureSlackHeaders(r.Header)
	var rawBody []byte
	defer func() {
		h.service.RecordRequestLog(ctx, slack.CoreRequestLogInput{
			RequestType:  "events",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		statusCode = http.StatusBadRequest
		outcome = "body_read_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		statusCode = http.StatusUnauthorized
		outcome = "signature_verification_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleEvents(rawBody)
	if err != nil {
		statusCode = http.StatusBadRequest
		outcome = "event_handler_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if response.Challenge != "" {
		outcome = "url_verification_ack"
		return writeRawJSON(w, http.StatusOK, response)
	}
	outcome = "processed"
	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handlers) HandleCommands(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	statusCode := http.StatusOK
	outcome := "received"
	errorMessage := ""
	headers := captureSlackHeaders(r.Header)
	var rawBody []byte
	defer func() {
		h.service.RecordRequestLog(ctx, slack.CoreRequestLogInput{
			RequestType:  "commands",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		statusCode = http.StatusBadRequest
		outcome = "body_read_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		statusCode = http.StatusUnauthorized
		outcome = "signature_verification_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleCommand(ctx, rawBody)
	if err != nil {
		outcome = "command_handler_failed"
		errorMessage = err.Error()
		return writeRawJSON(w, http.StatusOK, slack.CommandResponse{
			ResponseType: "ephemeral",
			Text:         "FortyOne could not process this command. Please try again.",
		})
	}
	outcome = "acknowledged"
	return writeRawJSON(w, http.StatusOK, response)
}

func (h *Handlers) HandleInteractivity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	statusCode := http.StatusOK
	outcome := "received"
	errorMessage := ""
	headers := captureSlackHeaders(r.Header)
	var rawBody []byte
	defer func() {
		h.service.RecordRequestLog(ctx, slack.CoreRequestLogInput{
			RequestType:  "interactivity",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		statusCode = http.StatusBadRequest
		outcome = "body_read_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		statusCode = http.StatusUnauthorized
		outcome = "signature_verification_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleInteractivity(ctx, rawBody)
	if err != nil {
		outcome = "interactivity_handler_failed"
		errorMessage = err.Error()
		w.WriteHeader(http.StatusOK)
		return nil
	}
	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	}
	status := response.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	statusCode = status
	outcome = "processed"
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

func captureSlackHeaders(headers http.Header) map[string]string {
	keys := []string{
		"X-Slack-Request-Timestamp",
		"X-Slack-Signature",
		"X-Slack-Retry-Num",
		"X-Slack-Retry-Reason",
		"User-Agent",
		"Content-Type",
	}
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(headers.Get(key))
		if value == "" {
			continue
		}
		result[key] = value
	}
	return result
}
