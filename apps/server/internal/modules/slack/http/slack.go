package slackhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const maxSlackRequestBodyBytes = 1 << 20
const maxConcurrentSlackRequestLogs = 32

type Handlers struct {
	log             *logger.Logger
	service         *slack.Service
	requestLogSlots chan struct{}
}

func New(log *logger.Logger, service *slack.Service) *Handlers {
	return &Handlers{
		log:             log,
		service:         service,
		requestLogSlots: make(chan struct{}, maxConcurrentSlackRequestLogs),
	}
}

func (h *Handlers) GetIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	integration, err := h.service.GetIntegration(ctx, workspace.ID, userID)
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

func (h *Handlers) CreateAccountLinkSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreateAccountLinkSessionRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	session, err := h.service.CreateAccountLinkSession(ctx, workspace.ID, userID, workspace.Slug, input.ReturnURL)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, AppCreateAccountLinkSession{
		Linked:     session.Linked,
		CanLink:    session.CanLink,
		InstallURL: session.InstallURL,
	}, http.StatusOK)
}

func (h *Handlers) LinkAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var input AppLinkSlackAccountRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	result, err := h.service.LinkSlackAccount(ctx, workspace.ID, userID, input.Token)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	status := "connected"
	if result.AlreadyLinked {
		status = "already_connected"
	}
	return web.Respond(ctx, w, AppLinkSlackAccountResult{
		Status:      status,
		SlackUserID: result.SlackUserID,
	}, http.StatusOK)
}

func (h *Handlers) DisconnectAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if _, err := h.service.DisconnectSlackAccount(ctx, workspace.ID, userID); err != nil {
		if slack.IsNotFound(err) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
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
	verified := false
	defer func() {
		if !verified {
			return
		}
		h.recordRequestLogAsync(ctx, slack.CoreRequestLogInput{
			RequestType:  "events",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := readSlackBody(w, r)
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
	verified = true
	eventCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	response, err := h.service.HandleEvents(eventCtx, rawBody)
	if err != nil {
		statusCode = http.StatusServiceUnavailable
		if errors.Is(err, slack.ErrSlackInvalidEventPayload) {
			statusCode = http.StatusBadRequest
		}
		outcome = "event_handler_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, statusCode)
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
	verified := false
	defer func() {
		if !verified {
			return
		}
		h.recordRequestLogAsync(ctx, slack.CoreRequestLogInput{
			RequestType:  "commands",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := readSlackBody(w, r)
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
	verified = true
	commandCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	response, err := h.service.HandleCommand(commandCtx, rawBody)
	if err != nil {
		outcome = "command_handler_failed"
		errorMessage = err.Error()
		return writeRawJSON(w, http.StatusOK, slack.CommandResponse{
			ResponseType: "ephemeral",
			Text:         "FortyOne could not process this command. Please try again.",
		})
	}
	outcome = "acknowledged"
	if response.ResponseType == "" && response.Text == "" {
		w.WriteHeader(http.StatusOK)
		return nil
	}
	return writeRawJSON(w, http.StatusOK, response)
}

func (h *Handlers) HandleInteractivity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	statusCode := http.StatusOK
	outcome := "received"
	errorMessage := ""
	headers := captureSlackHeaders(r.Header)
	var rawBody []byte
	verified := false
	defer func() {
		if !verified {
			return
		}
		h.recordRequestLogAsync(ctx, slack.CoreRequestLogInput{
			RequestType:  "interactivity",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := readSlackBody(w, r)
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
	verified = true
	interactionCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	response, err := h.service.HandleInteractivity(interactionCtx, rawBody)
	if err != nil {
		outcome = "interactivity_handler_failed"
		errorMessage = err.Error()
		if slackInteractionType(rawBody) == "view_submission" {
			return writeRawJSON(w, http.StatusOK, map[string]any{
				"response_action": "errors",
				"errors": map[string]string{
					"title": "FortyOne could not create this task. Please try again.",
				},
			})
		}
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

func slackInteractionType(rawBody []byte) string {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return ""
	}
	var payload struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(values.Get("payload")), &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Type)
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

func readSlackBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	body := http.MaxBytesReader(w, r.Body, maxSlackRequestBodyBytes)
	defer body.Close()
	return io.ReadAll(body)
}

func (h *Handlers) recordRequestLogAsync(ctx context.Context, input slack.CoreRequestLogInput) {
	select {
	case h.requestLogSlots <- struct{}{}:
	default:
		if h.log != nil {
			h.log.Warn(ctx, "dropping Slack request diagnostic because the bounded logger is full", "request_type", input.RequestType)
		}
		return
	}
	input.RawBody = bytes.Clone(input.RawBody)
	input.Headers = cloneStringMap(input.Headers)
	go func() {
		defer func() { <-h.requestLogSlots }()
		logCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		h.service.RecordRequestLog(logCtx, input)
	}()
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
