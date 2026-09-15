package brevo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	brv "github.com/getbrevo/brevo-go/lib"
	"github.com/stretchr/testify/require"
)

type accountCleanupTransport func(*http.Request) (*http.Response, error)

func (transport accountCleanupTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func TestAccountContactDeletionRequiresConfirmation(t *testing.T) {
	for _, status := range []int{204, 404, 200, 202, 401, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			config := brv.NewConfiguration()
			config.HTTPClient = &http.Client{Transport: accountCleanupTransport(func(request *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodDelete, request.Method)
				require.Equal(t, "/v3/contacts/member+test@example.com", request.URL.Path)
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"message":"member+test@example.com"}`))}, nil
			})}
			service := &Service{client: brv.NewAPIClient(config), enabled: true}
			err := service.DeleteAccountContact(context.Background(), "member+test@example.com")
			if status == 204 || status == 404 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "member+test@example.com")
			}
		})
	}
}

func TestAccountContactDeletionDisabledRemainsPending(t *testing.T) {
	require.ErrorIs(t, (&Service{}).DeleteAccountContact(context.Background(), "member@example.com"), ErrAccountCleanupUnavailable)
}

func TestAccountContactDeletionHandlesMissingResponsePrivately(t *testing.T) {
	config := brv.NewConfiguration()
	config.HTTPClient = &http.Client{Transport: accountCleanupTransport(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("provider failure for member@example.com")
	})}
	service := &Service{client: brv.NewAPIClient(config), enabled: true}
	err := service.DeleteAccountContact(context.Background(), "member@example.com")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "member@example.com")
}
