package adminhttp

import (
	"net/http/httptest"
	"strings"
	"testing"

	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/stretchr/testify/require"
)

func TestUpdateUserStateRequestPreservesOmittedNullAndFalse(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		activeSet     bool
		activePresent bool
	}{
		{name: "omitted", body: `{"reason":"review"}`},
		{name: "null", body: `{"isActive":null,"reason":"review"}`, activePresent: true},
		{name: "false", body: `{"isActive":false,"reason":"review"}`, activeSet: true, activePresent: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("PATCH", "/admin/users/id/state", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			var decoded updateUserStateRequest

			require.NoError(t, web.Decode(request, &decoded))
			value, specified := userStatePatch(decoded).IsActive.Value()
			require.Equal(t, test.activePresent, specified)
			if test.activeSet {
				require.NotNil(t, value)
				require.False(t, *value)
			} else {
				require.Nil(t, value)
			}
		})
	}
}

func TestAdminRequestRejectsUnknownAndTrailingJSON(t *testing.T) {
	tests := []string{
		`{"deleted":true,"reason":"review","unexpected":true}`,
		`{"deleted":true,"reason":"review"} {}`,
	}
	for _, body := range tests {
		request := httptest.NewRequest("PATCH", "/admin/workspaces/id/deleted", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		var decoded updateWorkspaceDeletedRequest

		require.Error(t, web.Decode(request, &decoded))
	}
}

func TestAdminErrorStatusMapsTypedFailures(t *testing.T) {
	require.Equal(t, 400, adminErrorStatus(admin.ErrInvalidFilter))
	require.Equal(t, 403, adminErrorStatus(admin.ErrForbidden))
	require.Equal(t, 404, adminErrorStatus(admin.ErrNotFound))
	require.Equal(t, 409, adminErrorStatus(admin.ErrConflict))
	require.Equal(t, 503, adminErrorStatus(admin.ErrIntegrationUnavailable))
}
