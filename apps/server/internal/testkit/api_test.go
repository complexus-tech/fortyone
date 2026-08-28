package testkit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
)

type handlerContextKey struct{}

func TestRecordHandlerProvidesDeterministicValuesAndPreservesContext(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC)
	wantErr := errors.New("expected handler error")
	request := httptest.NewRequest(http.MethodPost, "/widgets", nil)
	request = request.WithContext(context.WithValue(request.Context(), handlerContextKey{}, "actor-context"))
	handler := func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
		if got := ctx.Value(handlerContextKey{}); got != "actor-context" {
			t.Fatalf("handler context marker = %v", got)
		}
		if request.Context() != ctx {
			t.Fatal("request and handler contexts differ")
		}
		if got := web.GetTime(ctx); !got.Equal(now) {
			t.Fatalf("handler time = %v, want %v", got, now)
		}
		if web.GetRequestID(ctx) != testHandlerRequestID || web.GetTraceID(ctx) != testHandlerTraceID {
			t.Fatal("deterministic request metadata was not installed")
		}
		if err := web.SetStatusCode(ctx, http.StatusCreated); err != nil {
			t.Fatalf("set response status: %v", err)
		}
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"created":true}`))
		return wantErr
	}

	result := RecordHandler(t, now, handler, request)
	if !errors.Is(result.Err, wantErr) {
		t.Fatalf("handler error = %v, want %v", result.Err, wantErr)
	}
	if result.Response.Code != http.StatusCreated || result.Response.Body.String() != `{"created":true}` {
		t.Fatalf("handler response = status %d, body %q", result.Response.Code, result.Response.Body.String())
	}
	if result.Response.Header().Get("X-Request-ID") != testHandlerRequestID {
		t.Fatal("handler recorder did not mirror the application request ID header")
	}
	if result.Values.StatusCode != http.StatusCreated || !result.Values.Now.Equal(now) {
		t.Fatalf("handler values = %#v", result.Values)
	}
}
