package figma

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestExtractTextContent(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"type":"FRAME",
		"children":[
			{"type":"TEXT","characters":" Checkout "},
			{"type":"FRAME","children":[
				{"type":"TEXT","characters":"Pay now"},
				{"type":"TEXT","characters":"Pay now"},
				{"type":"TEXT","characters":"   "}
			]}
		]
	}`)

	require.Equal(t, []string{"Checkout", "Pay now"}, extractTextContent(payload))
}

func TestExtractTextContentLimitsLargeFrames(t *testing.T) {
	t.Parallel()

	children := make([]figmaNode, 0, maxImportedTextItems+5)
	for index := range maxImportedTextItems + 5 {
		children = append(children, figmaNode{
			Type:       "TEXT",
			Characters: fmt.Sprintf("Label %d", index),
		})
	}

	require.Len(t, collectTextContent(figmaNode{Type: "FRAME", Children: children}), maxImportedTextItems)
}

func TestAPIClientPreservesFigmaRateLimitMetadata(t *testing.T) {
	t.Parallel()

	client := apiClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{"90"}},
			Body:       io.NopCloser(strings.NewReader(`{"message":"rate limited"}`)),
		}, nil
	})}}

	err := client.get(context.Background(), "token", "/v1/files/file-key", &fileResponse{})
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	require.Equal(t, 90*time.Second, apiErr.RetryAfter)
	require.Contains(t, err.Error(), "try again in 1m30s")
}
