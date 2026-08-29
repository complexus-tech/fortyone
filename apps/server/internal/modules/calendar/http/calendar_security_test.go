package calendarhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMicrosoftNotificationValidationReturnsSafePlainText(t *testing.T) {
	t.Parallel()

	token := `opaque-token-<script>alert("x")</script>`
	request := httptest.NewRequest(
		http.MethodPost,
		"/webhooks/microsoft?validationToken="+url.QueryEscape(token),
		nil,
	)
	response := httptest.NewRecorder()

	err := (&Handlers{}).HandleMicrosoftNotification(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "text/plain; charset=utf-8", response.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "opaque-token-&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;", response.Body.String())
}

func TestMicrosoftNotificationValidationRejectsOversizedToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(
		http.MethodPost,
		"/webhooks/microsoft?validationToken="+strings.Repeat("a", maxMicrosoftValidationTokenBytes+1),
		nil,
	)
	response := httptest.NewRecorder()

	err := (&Handlers{}).HandleMicrosoftNotification(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestMicrosoftNotificationValidationRejectsRepeatedToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(
		http.MethodPost,
		"/webhooks/microsoft?validationToken=first&validationToken=second",
		nil,
	)
	response := httptest.NewRecorder()

	err := (&Handlers{}).HandleMicrosoftNotification(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestParseScheduleRangeRejectsAmbiguousAndInvalidRanges(t *testing.T) {
	t.Parallel()

	valid := httptest.NewRequest(
		http.MethodGet,
		"/schedule?start=2026-08-28T09%3A00%3A00Z&end=2026-08-28T10%3A00%3A00Z",
		nil,
	)
	startAt, endAt, err := parseScheduleRange(valid)
	require.NoError(t, err)
	require.True(t, startAt.Before(endAt))

	for name, target := range map[string]string{
		"missing start":  "/schedule?end=2026-08-28T10%3A00%3A00Z",
		"repeated start": "/schedule?start=2026-08-28T09%3A00%3A00Z&start=2026-08-28T09%3A30%3A00Z&end=2026-08-28T10%3A00%3A00Z",
		"malformed":      "/schedule?start=tomorrow&end=2026-08-28T10%3A00%3A00Z",
		"reverse":        "/schedule?start=2026-08-28T11%3A00%3A00Z&end=2026-08-28T10%3A00%3A00Z",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodGet, target, nil)
			_, _, err := parseScheduleRange(request)
			require.Error(t, err)
		})
	}
}
