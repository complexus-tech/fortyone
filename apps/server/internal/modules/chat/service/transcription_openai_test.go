package chatservice

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/stretchr/testify/require"
)

func TestOpenAITranscriberSendsAnAuthenticatedExtensionBearingMultipartFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.NoError(t, r.ParseMultipartForm(1<<20))
		defer r.MultipartForm.RemoveAll()
		require.Equal(t, DefaultTranscriptionModel, r.FormValue("model"))
		require.Equal(t, "json", r.FormValue("response_format"))
		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()
		require.Equal(t, "voice.m4a", header.Filename)
		require.Equal(t, "audio/mp4", header.Header.Get("Content-Type"))
		body, err := io.ReadAll(file)
		require.NoError(t, err)
		require.Equal(t, "audio bytes", string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"  Meeting at ten.  ","usage":{"seconds":3}}`))
	}))
	defer server.Close()
	provider := NewOpenAITranscriber("test-key", "")
	provider.endpoint = server.URL
	text, model, err := provider.Transcribe(t.Context(), domain.TranscriptionAudio{FileName: "voice.mp4", MediaType: "audio/mp4"}, []byte("audio bytes"))
	require.NoError(t, err)
	require.Equal(t, "Meeting at ten.", text)
	require.Equal(t, DefaultTranscriptionModel, model)
}

func TestOpenAITranscriberBoundsResponsesAndNeverRetriesProviderErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"provider error", 429, `{"error":"sensitive provider details"}`},
		{"missing text", 200, `{"other":"value"}`},
		{"invalid json", 200, `not json`},
		{"too large", 200, strings.Repeat("x", (1<<20)+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			provider := NewOpenAITranscriber("test-key", "")
			provider.endpoint = server.URL
			_, _, err := provider.Transcribe(t.Context(), domain.TranscriptionAudio{FileName: "voice.webm", MediaType: "audio/webm"}, []byte("audio"))
			require.Error(t, err)
			require.NotContains(t, err.Error(), "sensitive provider details")
			require.Equal(t, 1, calls)
		})
	}
}

func TestOpenAITranscriberRejectsRedirectsAndPropagatesCancellation(t *testing.T) {
	var leaked bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/target" {
			leaked = true
		}
		http.Redirect(w, r, "/target", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	provider := NewOpenAITranscriber("test-key", "")
	provider.endpoint = server.URL
	input := domain.TranscriptionAudio{FileName: "voice.webm", MediaType: "audio/webm"}
	_, _, err := provider.Transcribe(t.Context(), input, []byte("audio"))
	require.Error(t, err)
	require.False(t, leaked)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, err = provider.Transcribe(ctx, input, []byte("audio"))
	require.True(t, errors.Is(err, context.Canceled))
}
