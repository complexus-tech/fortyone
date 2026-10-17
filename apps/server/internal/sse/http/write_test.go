package ssehttp

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A real socket pair reproduces a peer that has stopped reading without
// depending on OS-specific TCP send-buffer sizes or a ten-second unit test.
type pipeStreamWriter struct{ net.Conn }

func (pipeStreamWriter) Header() http.Header { return make(http.Header) }
func (pipeStreamWriter) WriteHeader(int)     {}
func (pipeStreamWriter) Flush()              {}

func TestStreamWriteDeadlineReleasesStalledSocket(t *testing.T) {
	writer, reader := net.Pipe()
	defer writer.Close()
	defer reader.Close()
	started := time.Now()
	err := writeStreamEvent(pipeStreamWriter{writer}, "data: stalled\n\n", 25*time.Millisecond)
	var timeout net.Error
	require.ErrorAs(t, err, &timeout)
	require.True(t, timeout.Timeout())
	require.Less(t, time.Since(started), time.Second)
}

type bufferedPipeStreamWriter struct {
	pipeStreamWriter
	buffer bytes.Buffer
}

func (w *bufferedPipeStreamWriter) Write(data []byte) (int, error) { return w.buffer.Write(data) }
func (w *bufferedPipeStreamWriter) Flush() {
	// http.Flusher cannot return the socket error; middleware commonly exposes
	// only this interface even when the underlying HTTP writer has FlushError.
	_, _ = w.Conn.Write(w.buffer.Bytes())
}

func TestStreamFlushDeadlineClosesDespiteErrorHidingMiddleware(t *testing.T) {
	writer, reader := net.Pipe()
	defer writer.Close()
	defer reader.Close()
	response := &bufferedPipeStreamWriter{pipeStreamWriter: pipeStreamWriter{writer}}
	started := time.Now()
	err := writeStreamEvent(response, "data: stalled flush\n\n", 25*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(started), time.Second)
}

type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadlines []time.Time
}

func (w *deadlineRecorder) SetWriteDeadline(deadline time.Time) error {
	w.deadlines = append(w.deadlines, deadline)
	return nil
}

func TestStreamClearsWriteDeadlineBetweenSuccessfulEvents(t *testing.T) {
	response := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	before := time.Now()
	for _, event := range []string{"event: connected\ndata: {}\n\n", "data: message\n\n", ":keep-alive\n\n"} {
		require.NoError(t, writeStreamEvent(response, event, streamWriteTimeout))
	}
	require.Len(t, response.deadlines, 6)
	for index := 0; index < len(response.deadlines); index += 2 {
		require.True(t, response.deadlines[index].After(before.Add(9*time.Second)))
		require.True(t, response.deadlines[index+1].IsZero(), "idle stream must not inherit a write deadline")
	}
	require.True(t, response.Flushed)
}
