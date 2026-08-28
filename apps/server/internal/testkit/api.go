package testkit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	testHandlerRequestID = "test-request-id"
	testHandlerTraceID   = "test-trace-id"
)

// HandlerResult preserves the application handler's response, returned error,
// and request values. It deliberately does not translate Err through the full
// application error renderer; handler tests should assert that contract
// explicitly at their owning HTTP module.
type HandlerResult struct {
	Response *httptest.ResponseRecorder
	Err      error
	Values   *web.Values
}

// RecordHandler invokes one web.Handler with deterministic request metadata and
// decision time while preserving all context values already carried by request
// (including authenticated actors). It is intentionally smaller than a full API
// builder so testkit does not depend on every module or bootstrap credential.
func RecordHandler(t testing.TB, now time.Time, handler web.Handler, request *http.Request) HandlerResult {
	t.Helper()
	if handler == nil {
		t.Fatal("record web handler: handler is required")
	}
	if request == nil {
		t.Fatal("record web handler: request is required")
	}

	values := &web.Values{
		TraceID:   testHandlerTraceID,
		RequestID: testHandlerRequestID,
		Tracer:    noop.NewTracerProvider().Tracer("testkit"),
		Now:       now,
	}
	ctx := web.SetValues(request.Context(), values)
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()
	response.Header().Set("X-Request-ID", testHandlerRequestID)
	err := handler(ctx, response, request)

	return HandlerResult{
		Response: response,
		Err:      err,
		Values:   values,
	}
}
