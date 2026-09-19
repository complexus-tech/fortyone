package expopush

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestSendReturnsInvalidDeviceTokensWithoutFailingAcceptedTickets(t *testing.T) {
	t.Parallel()
	client := New(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, endpoint, request.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"status":"ok","id":"ticket"},{"status":"error","message":"gone","details":{"error":"DeviceNotRegistered"}}]}`)),
		}, nil
	})})

	result, err := client.Send(context.Background(), []Message{
		{To: "ExponentPushToken[one]", Title: "One", Body: "Body"},
		{To: "ExponentPushToken[two]", Title: "Two", Body: "Body"},
	})

	require.NoError(t, err)
	require.Equal(t, []string{"ExponentPushToken[two]"}, result.InvalidTokens)
}

func TestSendRejectsMismatchedTicketCounts(t *testing.T) {
	t.Parallel()
	client := New(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
		}, nil
	})})

	_, err := client.Send(context.Background(), []Message{{To: "ExponentPushToken[one]"}})

	require.ErrorContains(t, err, "0 tickets for 1 messages")
}
