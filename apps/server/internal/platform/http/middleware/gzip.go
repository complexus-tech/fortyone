package mid

import (
	"bufio"
	"compress/gzip"
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

// gzipWriter is a simple wrapper that writes to a gzip writer
type gzipWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

// Write compresses data before writing to the response
func (gw *gzipWriter) Write(b []byte) (int, error) {
	return gw.gz.Write(b)
}

// Flush implements the http.Flusher interface
func (gw *gzipWriter) Flush() {
	gw.gz.Flush()
	if flusher, ok := gw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements the http.Hijacker interface
func (gw *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := gw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// shouldCompress returns true if the response should be compressed
func shouldCompress(r *http.Request, contentType string) bool {
	// Skip compression for already compressed content types
	if strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "application/zip") ||
		strings.Contains(contentType, "application/x-gzip") ||
		strings.Contains(contentType, "application/pdf") ||
		strings.Contains(contentType, "application/octet-stream") {
		return false
	}

	// Skip compression for HEAD requests
	if r.Method == http.MethodHead {
		return false
	}

	return true
}

// Gzip returns a middleware that compresses HTTP responses
func Gzip(log *logger.Logger) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Only compress if client accepts gzip encoding
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				return next(ctx, w, r)
			}

			// Check content type if available
			contentType := w.Header().Get("Content-Type")
			if contentType != "" && !shouldCompress(r, contentType) {
				return next(ctx, w, r)
			}

			// Create gzip writer
			gz := gzip.NewWriter(w)
			defer func() {
				if err := gz.Close(); err != nil {
					log.Error(ctx, "failed to close gzip writer", "err", err)
				}
			}()

			// Set headers before passing to the next handler
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")

			// Wrap response writer with gzip writer
			gzw := &gzipWriter{
				ResponseWriter: w,
				gz:             gz,
			}

			// Pass control to the next handler with compressed writer
			return next(ctx, gzw, r)
		}
	}
}
