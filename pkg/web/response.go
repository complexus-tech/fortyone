package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Respond sends JSON to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	ctx, span := AddSpan(ctx, "internal.web.Respond", attribute.Int("status", statusCode))
	defer span.End()
	span.AddEvent("serializing response.")

	SetStatusCode(ctx, statusCode)
	if statusCode == http.StatusNoContent {
		span.SetStatus(codes.Error, "No content to send.")
		span.RecordError(errors.New("no content"),
			trace.WithAttributes(
				attribute.Int("status", statusCode),
			))

		w.WriteHeader(statusCode)
		return nil
	}
	span.SetStatus(codes.Ok, "Response sent successfully.")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
