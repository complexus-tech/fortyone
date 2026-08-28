package slack

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type commandResponseRoundTripper func(*http.Request) (*http.Response, error)

func (roundTrip commandResponseRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

type countingResponseBody struct {
	reader io.Reader
	read   int64
}

func (body *countingResponseBody) Read(destination []byte) (int, error) {
	read, err := body.reader.Read(destination)
	body.read += int64(read)
	return read, err
}

func (*countingResponseBody) Close() error { return nil }

func TestPostCommandResponseBoundsAndRedactsProviderFailureBody(t *testing.T) {
	t.Parallel()
	providerSecret := "xoxb-provider-body-must-not-escape"
	body := &countingResponseBody{
		reader: strings.NewReader(strings.Repeat(providerSecret, 10_000)),
	}
	service := &Service{client: &http.Client{Transport: commandResponseRoundTripper(
		func(request *http.Request) (*http.Response, error) {
			require.Equal(t, "hooks.slack.com", request.URL.Hostname())
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       body,
				Request:    request,
			}, nil
		},
	)}}

	err := service.postCommandResponse(
		context.Background(),
		"https://hooks.slack.com/commands/T1/B1/secret-capability",
		"safe response",
	)

	var statusErr *slackCommandResponseStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, http.StatusBadGateway, statusErr.StatusCode)
	require.NotContains(t, err.Error(), providerSecret)
	require.LessOrEqual(t, body.read, maxSlackCommandResponseDrainBytes)
}

func TestPostCommandResponseDoesNotFollowProviderRedirect(t *testing.T) {
	t.Parallel()
	requests := 0
	service := &Service{client: &http.Client{Transport: commandResponseRoundTripper(
		func(request *http.Request) (*http.Response, error) {
			requests++
			return &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Header: http.Header{
					"Location": []string{"https://attacker.example/collect"},
				},
				Body:    io.NopCloser(strings.NewReader("redirect")),
				Request: request,
			}, nil
		},
	)}}

	err := service.postCommandResponse(
		context.Background(),
		"https://hooks.slack.com/commands/T1/B1/secret-capability",
		"safe response",
	)

	var statusErr *slackCommandResponseStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, http.StatusTemporaryRedirect, statusErr.StatusCode)
	require.Equal(t, 1, requests)
}
