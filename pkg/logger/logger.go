package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

// Logger is a wrapper around slog.Handler.
type Logger struct {
	handler slog.Handler
}

// NewWithJSON returns a new Logger with a JSON handler. The handler is configured
// to write to the provided writer and only log messages with a level greater than
// or equal to minLevel..
func NewWithJSON(w io.Writer, minLevel slog.Level, serviceName string) *Logger {
	// Convert the file name to just the name.ext
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}
	handler := slog.Handler(
		slog.NewJSONHandler(
			w,
			&slog.HandlerOptions{AddSource: true, Level: minLevel, ReplaceAttr: f},
		),
	)

	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}
	handler = handler.WithAttrs(attrs)
	return &Logger{
		handler: handler,
	}
}

// NewWithText returns a new Logger with a text handler.
func NewWithText(w io.Writer, minLevel slog.Level, serviceName string) *Logger {
	// Convert the file name to just the name.ext
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}
		return a
	}
	handler := slog.Handler(
		slog.NewTextHandler(
			w,
			&slog.HandlerOptions{AddSource: true, Level: minLevel, ReplaceAttr: f},
		),
	)

	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}
	handler = handler.WithAttrs(attrs)
	return &Logger{
		handler: handler,
	}
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

	r.Add(args...)

	l.handler.Handle(ctx, r)
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
