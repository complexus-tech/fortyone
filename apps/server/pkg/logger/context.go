package logger

import "context"

type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID attaches the server-generated safe correlation identifier used
// by HTTP responses and every structured log emitted during the request.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}
