package googledrivehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	googledrive "github.com/complexus-tech/projects-api/internal/modules/googledrive/service"
	"github.com/stretchr/testify/require"
)

func TestCompleteOAuthRequiresAuthenticatedBrowserActor(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(
		http.MethodGet,
		"/integrations/google-drive/callback?code=code&state=opaque-state",
		nil,
	)
	response := httptest.NewRecorder()

	err := New(nil).CompleteOAuth(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
}

func TestWritePreviewResponseUsesBinarySecurityHeaders(t *testing.T) {
	t.Parallel()

	response := httptest.NewRecorder()
	preview := googledrive.Preview{Bytes: []byte("image-bytes"), ContentType: "image/webp"}

	err := writePreviewResponse(response, preview)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, preview.Bytes, response.Body.Bytes())
	require.Equal(t, "image/webp", response.Header().Get("Content-Type"))
	require.Equal(t, "11", response.Header().Get("Content-Length"))
	require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
	require.Equal(t, "default-src 'none'; sandbox", response.Header().Get("Content-Security-Policy"))
	require.Equal(t, "same-site", response.Header().Get("Cross-Origin-Resource-Policy"))
	require.Equal(t, "no-referrer", response.Header().Get("Referrer-Policy"))
	require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
}
