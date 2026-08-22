package figmahttp

import (
	"errors"
	"net/http"
	"testing"

	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	"github.com/stretchr/testify/require"
)

func TestProviderErrorStatus(t *testing.T) {
	t.Parallel()

	require.Equal(t, http.StatusTooManyRequests, providerErrorStatus(&figma.APIError{
		StatusCode: http.StatusTooManyRequests,
	}))
	require.Equal(t, http.StatusBadRequest, providerErrorStatus(errors.New("invalid Figma URL")))
}
