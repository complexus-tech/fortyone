package googledrive

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func TestGetFileRequestsAndDecodesTransientThumbnailLink(t *testing.T) {
	t.Parallel()

	thumbnailLink := "https://lh3.googleusercontent.com/thumbnail"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Contains(t, request.URL.Query().Get("fields"), "thumbnailLink")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"id":"provider-file",
			"name":"Launch plan",
			"mimeType":"application/vnd.google-apps.document",
			"webViewLink":"https://docs.google.com/document/d/provider-file/edit",
			"thumbnailLink":"  ` + thumbnailLink + `  "
		}`))
	}))
	t.Cleanup(server.Close)
	client := newGoogleClient(server.Client(), Config{}, nil)
	client.driveURL = server.URL

	file, err := client.GetFile(t.Context(), "access-token", "provider-file", nil)

	require.NoError(t, err)
	require.Equal(t, thumbnailLink, file.ThumbnailLink)
	require.NotContains(t, string(file.Metadata), "thumbnail")
}

func TestReadThumbnailUsesCredentialedBoundedImageRequest(t *testing.T) {
	t.Parallel()

	thumbnail := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	httpClient := &http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, "lh3.googleusercontent.com", request.URL.Hostname())
		require.Equal(t, "Bearer access-token", request.Header.Get("Authorization"))
		require.Equal(t, "image/webp,image/png,image/jpeg", request.Header.Get("Accept"))
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Header:        http.Header{"Content-Type": {"image/png; charset=binary"}},
			Body:          io.NopCloser(strings.NewReader(string(thumbnail))),
			ContentLength: int64(len(thumbnail)),
			Request:       request,
		}, nil
	})}
	client := newGoogleClient(httpClient, Config{}, nil)

	preview, err := client.ReadThumbnail(
		t.Context(),
		"access-token",
		"https://lh3.googleusercontent.com/thumbnail?sz=w800",
		1024,
	)

	require.NoError(t, err)
	require.Equal(t, thumbnail, preview.Bytes)
	require.Equal(t, "image/png", preview.ContentType)
}

func TestReadThumbnailRejectsUnsafeURLsBeforeSendingCredentials(t *testing.T) {
	t.Parallel()

	for _, thumbnailURL := range []string{
		"http://lh3.googleusercontent.com/thumbnail",
		"https://googleusercontent.com.attacker.example/thumbnail",
		"https://user:password@lh3.googleusercontent.com/thumbnail",
		"https://lh3.googleusercontent.com:443/thumbnail",
		"https://lh3.googleusercontent.com/thumbnail#fragment",
	} {
		thumbnailURL := thumbnailURL
		t.Run(thumbnailURL, func(t *testing.T) {
			t.Parallel()

			calls := 0
			client := newGoogleClient(&http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, nil
			})}, Config{}, nil)

			_, err := client.ReadThumbnail(t.Context(), "access-token", thumbnailURL, 1024)

			require.Error(t, err)
			require.Zero(t, calls)
		})
	}
}

func TestReadThumbnailRefusesRedirectOutsideGoogleThumbnailHosts(t *testing.T) {
	t.Parallel()

	requestedHosts := make([]string, 0, 2)
	client := newGoogleClient(&http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		requestedHosts = append(requestedHosts, request.URL.Hostname())
		return &http.Response{
			StatusCode: http.StatusFound,
			Status:     "302 Found",
			Header:     http.Header{"Location": {"https://attacker.example/thumbnail"}},
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    request,
		}, nil
	})}, Config{}, nil)

	_, err := client.ReadThumbnail(
		t.Context(),
		"access-token",
		"https://lh3.googleusercontent.com/thumbnail",
		1024,
	)

	require.ErrorContains(t, err, "unsafe Google thumbnail redirect")
	require.Equal(t, []string{"lh3.googleusercontent.com"}, requestedHosts)
}

func TestReadThumbnailRejectsHTMLAndOversizedBodies(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name          string
		contentType   string
		body          string
		contentLength int64
		limit         int64
		targetError   error
	}{
		{
			name: "HTML body with image header", contentType: "image/png",
			body: "<!doctype html><html></html>", contentLength: -1, limit: 1024,
		},
		{
			name: "HTML content type", contentType: "text/html",
			body: "<!doctype html><html></html>", contentLength: -1, limit: 1024,
		},
		{
			name: "declared body too large", contentType: "image/png",
			body: "", contentLength: 1025, limit: 1024, targetError: domain.ErrContentTooLarge,
		},
		{
			name: "streamed body too large", contentType: "image/png",
			body: strings.Repeat("x", 1025), contentLength: -1, limit: 1024, targetError: domain.ErrContentTooLarge,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			client := newGoogleClient(&http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Status:        "200 OK",
					Header:        http.Header{"Content-Type": {test.contentType}},
					Body:          io.NopCloser(strings.NewReader(test.body)),
					ContentLength: test.contentLength,
					Request:       request,
				}, nil
			})}, Config{}, nil)

			_, err := client.ReadThumbnail(
				t.Context(),
				"access-token",
				"https://lh3.googleusercontent.com/thumbnail",
				test.limit,
			)

			if test.targetError != nil {
				require.ErrorIs(t, err, test.targetError)
				return
			}
			require.Error(t, err)
		})
	}
}
