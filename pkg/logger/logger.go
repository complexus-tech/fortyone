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

type Logger struct {
	handler slog.Handler
}

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

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelDebug, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelInfo, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelWarn, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelError, msg, args...)
}
