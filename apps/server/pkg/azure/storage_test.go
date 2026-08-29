package azure

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/stretchr/testify/require"
)

func TestIsNotFoundResponseMakesObjectDeletionIdempotent(t *testing.T) {
	t.Parallel()

	require.True(t, isNotFoundResponse(&azcore.ResponseError{StatusCode: http.StatusNotFound}))
	require.True(t, isNotFoundResponse(errors.Join(
		errors.New("delete blob"),
		&azcore.ResponseError{StatusCode: http.StatusNotFound},
	)))
	require.False(t, isNotFoundResponse(&azcore.ResponseError{StatusCode: http.StatusForbidden}))
	require.False(t, isNotFoundResponse(errors.New("network unavailable")))
	require.False(t, isNotFoundResponse(nil))
}
