package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Logger provides structured logging with support for context propagation,
// source location tracking, and multiple output formats (JSON and text).
// It wraps the standard library's slog.Handler with additional convenience methods.
type Logger struct {
	handler        slog.Handler
	fallbackWriter io.Writer
}

// NewWithJSON returns a new Logger with a JSON handler. The handler is configured
// to write to the provided writer and only log messages with a level greater than
// or equal to minLevel..
func NewWithJSON(w io.Writer, minLevel slog.Level, serviceName string) *Logger {
	handler := slog.Handler(
		slog.NewJSONHandler(
			w,
			&slog.HandlerOptions{AddSource: true, Level: minLevel, ReplaceAttr: replaceAttribute},
		),
	)

	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}
	handler = handler.WithAttrs(attrs)
	return &Logger{
		handler:        handler,
		fallbackWriter: os.Stderr,
	}
}

// NewWithText returns a new Logger with a text handler.
func NewWithText(w io.Writer, minLevel slog.Level, serviceName string) *Logger {
	handler := slog.Handler(
		slog.NewTextHandler(
			w,
			&slog.HandlerOptions{AddSource: true, Level: minLevel, ReplaceAttr: replaceAttribute},
		),
	)

	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}
	handler = handler.WithAttrs(attrs)
	return &Logger{
		handler:        handler,
		fallbackWriter: os.Stderr,
	}
}

const redactedValue = "[REDACTED]"

// replaceAttribute is the logger's final privacy boundary. Call sites should
// still avoid sensitive fields, but a missed email/token/payload attribute or
// arbitrary error value must not become durable log data.
func replaceAttribute(groups []string, attribute slog.Attr) slog.Attr {
	if attribute.Key == slog.SourceKey {
		if source, ok := attribute.Value.Any().(*slog.Source); ok {
			value := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
			return slog.String("file", value)
		}
	}

	normalizedKey := normalizeAttributeKey(attribute.Key)
	if sensitiveAttributeKey(normalizedKey) || containsSensitiveAttributeGroup(groups) {
		return slog.String(attribute.Key, redactedValue)
	}
	if err, ok := attribute.Value.Any().(error); ok {
		key := attribute.Key
		if key == "!BADKEY" {
			key = "error"
		}
		return safeErrorAttribute(key, err)
	}
	if normalizedKey == "error" || normalizedKey == "err" {
		return slog.String(attribute.Key, redactedValue)
	}
	return attribute
}

func normalizeAttributeKey(key string) string {
	return strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(key)))
}

func sensitiveAttributeKey(key string) bool {
	switch key {
	case "authorization", "proxyauthorization", "cookie", "setcookie",
		"token", "accesstoken", "refreshtoken", "idtoken", "verificationtoken", "invitationtoken",
		"authorizationcode", "oauthcode",
		"secret", "clientsecret", "signingsecret", "password", "passphrase", "apikey",
		"body", "rawbody", "rawmessage", "payload",
		"errormessage", "errordetail", "errorreason", "cause", "stack", "stacktrace", "exceptionmessage",
		"email", "inviteremail", "actoremail", "url", "uri", "requestpath":
		return true
	default:
		return strings.HasSuffix(key, "email") ||
			strings.HasSuffix(key, "emails") ||
			strings.HasSuffix(key, "token") ||
			strings.HasSuffix(key, "secret") ||
			strings.HasSuffix(key, "password") ||
			strings.HasSuffix(key, "cookie") ||
			strings.HasSuffix(key, "payload") ||
			strings.HasSuffix(key, "url") ||
			strings.HasSuffix(key, "uri")
	}
}

func containsSensitiveAttributeGroup(groups []string) bool {
	for _, group := range groups {
		if sensitiveAttributeKey(normalizeAttributeKey(group)) {
			return true
		}
	}
	return false
}

// sanitizeLogArguments captures error values before slog resolves LogValuer.
// Some error implementations also implement LogValuer and could otherwise turn
// themselves into raw strings before replaceAttribute sees them.
func sanitizeLogArguments(args []any) []any {
	attributes := make([]any, 0, len(args))
	for index := 0; index < len(args); {
		switch argument := args[index].(type) {
		case slog.Attr:
			attributes = append(attributes, sanitizeLogAttribute(argument))
			index++
		case string:
			if index+1 >= len(args) {
				attributes = append(attributes, slog.String("!BADKEY", argument))
				index++
				continue
			}
			attributes = append(attributes, sanitizeLogAttribute(slog.Any(argument, args[index+1])))
			index += 2
		default:
			attributes = append(attributes, sanitizeLogAttribute(slog.Any("!BADKEY", argument)))
			index++
		}
	}
	return attributes
}

func sanitizeLogAttribute(attribute slog.Attr) slog.Attr {
	if sensitiveAttributeKey(normalizeAttributeKey(attribute.Key)) {
		return slog.String(attribute.Key, redactedValue)
	}
	if err, ok := attribute.Value.Any().(error); ok {
		key := attribute.Key
		if key == "!BADKEY" {
			key = "error"
		}
		return safeErrorAttribute(key, err)
	}
	if attribute.Value.Kind() != slog.KindGroup {
		return attribute
	}

	children := attribute.Value.Group()
	sanitized := make([]slog.Attr, 0, len(children))
	for _, child := range children {
		sanitized = append(sanitized, sanitizeLogAttribute(child))
	}
	return slog.Attr{Key: attribute.Key, Value: slog.GroupValue(sanitized...)}
}

// write writes a log message to the handler. The message is only written if the
// handler is enabled for the given level. The message is written with the provided
// arguments. The arguments are added to the log record as attributes.
func (l *Logger) write(ctx context.Context, level slog.Level, msg string, args ...any) {

	if !l.handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	if requestID := requestIDFromContext(ctx); requestID != "" {
		r.Add("request_id", requestID)
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		r.Add("trace_id", spanContext.TraceID().String(), "span_id", spanContext.SpanID().String())
	}

	r.Add(sanitizeLogArguments(args)...)

	if err := l.handler.Handle(ctx, r); err != nil {
		fallback := l.fallbackWriter
		if fallback == nil {
			fallback = os.Stderr
		}
		// Never write the handler error text: output failures may include
		// sensitive destination details. A concrete type is enough to make the
		// failure observable without recursively invoking the failed handler.
		_, _ = fmt.Fprintf(
			fallback,
			"level=ERROR msg=%q handler_error_type=%T\n",
			"structured log handler failed",
			err,
		)
	}
}

// Debug writes a debug level log message to the handler.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelDebug, msg, args...)
}

// Info writes an info level log message to the handler.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelInfo, msg, args...)
}

// Warn writes a warn level log message to the handler.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelWarn, msg, args...)
}

// Error writes an error level log message to the handler.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelError, msg, args...)
}
