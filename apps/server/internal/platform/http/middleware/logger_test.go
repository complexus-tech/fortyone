package mid

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/stretchr/testify/require"
)

func TestRequestLoggerRecordsRoutePatternWithoutBearerPath(t *testing.T) {
	t.Parallel()

	const bearer = "wi1.v1.raw-invitation-bearer.signature"
	var output bytes.Buffer
	log := logger.NewWithJSON(&output, slog.LevelDebug, "request-logger-test")
	request := httptest.NewRequest(http.MethodGet, "/invitations/"+bearer+"?debug=secret", nil)
	request.Pattern = "GET /invitations/{token}"
	request.RemoteAddr = "203.0.113.42:4242"

	handler := Logger(log)(func(context.Context, http.ResponseWriter, *http.Request) error { return nil })
	require.NoError(t, handler(web.SetValues(context.Background(), &web.Values{}), httptest.NewRecorder(), request))
	require.Contains(t, output.String(), "GET /invitations/{token}")
	require.NotContains(t, output.String(), bearer)
	require.NotContains(t, output.String(), "debug=secret")
	require.NotContains(t, output.String(), "203.0.113.42")
}
