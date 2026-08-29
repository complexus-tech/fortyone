package mid

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
)

func TestGzipWriterFlushReportsCompressionFailure(t *testing.T) {
	t.Parallel()

	const sensitiveFailure = "writer failed with sensitive-token"
	response := &failingFlushResponseWriter{
		header:           make(http.Header),
		err:              errors.New(sensitiveFailure),
		successfulWrites: 1,
	}
	var logs bytes.Buffer
	log := logger.NewWithJSON(&logs, slog.LevelDebug, "gzip-test")
	compressed := gzip.NewWriter(response)
	if _, err := compressed.Write([]byte("buffered response")); err != nil {
		t.Fatalf("buffer response: %v", err)
	}

	writer := &gzipWriter{
		ResponseWriter: response,
		gz:             compressed,
		ctx:            context.Background(),
		log:            log,
	}
	writer.Flush()

	if !response.flushed {
		t.Fatal("underlying response writer was not flushed")
	}
	if !strings.Contains(logs.String(), "failed to flush gzip writer") {
		t.Fatalf("logs = %q, want gzip flush failure", logs.String())
	}
	if strings.Contains(logs.String(), sensitiveFailure) || strings.Contains(logs.String(), "sensitive-token") {
		t.Fatalf("logs leaked writer error: %s", logs.String())
	}
}

type failingFlushResponseWriter struct {
	header  http.Header
	err     error
	flushed bool
	writes  int

	successfulWrites int
}

func (w *failingFlushResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingFlushResponseWriter) Write(contents []byte) (int, error) {
	w.writes++
	if w.writes > w.successfulWrites {
		return 0, w.err
	}
	return len(contents), nil
}

func (w *failingFlushResponseWriter) WriteHeader(int) {}

func (w *failingFlushResponseWriter) Flush() {
	w.flushed = true
}
