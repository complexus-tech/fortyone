package web

import (
	"context"
	"encoding/json"
	"net/http"

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
			Message: err.Error(),
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
