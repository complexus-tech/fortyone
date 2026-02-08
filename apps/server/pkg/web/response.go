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
	Message string `json:"message"`
}

type Response struct {
	Data  any          `json:"data"`
	Error *ErrorDetail `json:"error,omitempty"`
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error, statusCode int) error {
	errResponse := Response{
		Error: &ErrorDetail{
			Message: sanitizeErrorMessage(err, statusCode),
		},
	}
	return respond(ctx, w, errResponse, statusCode)
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
