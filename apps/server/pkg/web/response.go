package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
}

type Response struct {
	Data  any          `json:"data"`
	Error *ErrorDetail `json:"error,omitempty"`
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error, statusCode int) error {
	errResponse := Response{
		Error: &ErrorDetail{
			Code:    errorCode(statusCode),
			Message: sanitizeErrorMessage(err, statusCode),
			Hint:    resolutionHint(statusCode),
		},
	}
	return respond(ctx, w, errResponse, statusCode)
}

func errorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "authentication_required"
	case http.StatusForbidden:
		return "permission_denied"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusConflict:
		return "conflict"
	case http.StatusTooManyRequests:
		return "rate_limited"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		if statusCode >= http.StatusInternalServerError {
			return "internal_error"
		}
		return "request_failed"
	}
}

func resolutionHint(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Check the request parameters and body against https://www.fortyone.app/openapi.json."
	case http.StatusUnauthorized:
		return "Authenticate with a FortyOne session that can access the requested workspace."
	case http.StatusForbidden:
		return "Use an account with permission for this workspace resource."
	case http.StatusNotFound:
		return "Check the path and resource identifier, then consult https://www.fortyone.app/openapi.json."
	case http.StatusMethodNotAllowed:
		return "Use one of the methods advertised for this path in https://www.fortyone.app/openapi.json."
	case http.StatusTooManyRequests:
		return "Wait before retrying and honor any Retry-After header."
	case http.StatusServiceUnavailable:
		return "Retry with backoff after the service reports ready."
	default:
		return "Consult https://www.fortyone.app/developers and retry only when it is safe to do so."
	}
}

func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	response := Response{
		Data: data,
	}
	return respond(ctx, w, response, statusCode)
}

func respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	_, span := AddSpan(ctx, "pkg.web.respond", attribute.Int("status", statusCode))
	defer span.End()

	if statusCode == http.StatusNoContent {
		span.SetStatus(codes.Ok, "No content to send")
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to serialize JSON")
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		span.RecordError(err)
		return err
	}

	span.SetStatus(codes.Ok, "Response sent successfully")
	return nil
}

func sanitizeErrorMessage(err error, statusCode int) string {
	if statusCode >= http.StatusInternalServerError {
		return "internal server error"
	}

	if err == nil {
		return defaultErrorMessage(statusCode)
	}

	message := strings.TrimSpace(err.Error())
	if message == "" {
		return defaultErrorMessage(statusCode)
	}

	if looksLikeDatabaseError(message) {
		return defaultErrorMessage(statusCode)
	}

	return message
}

func defaultErrorMessage(statusCode int) string {
	statusText := strings.ToLower(strings.TrimSpace(http.StatusText(statusCode)))
	if statusText == "" {
		return "request failed"
	}
	return statusText
}

func looksLikeDatabaseError(message string) bool {
	lower := strings.ToLower(message)

	dbMarkers := []string{
		"sql:",
		"pq:",
		"database",
		"no rows in result set",
		"duplicate key value",
		"violates unique constraint",
		"violates foreign key constraint",
		"violates check constraint",
	}

	for _, marker := range dbMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	return false
}
