package usershttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type accountDeletionStub struct {
	pending bool
	err     error
	called  bool
	userID  uuid.UUID
}

func (stub *accountDeletionStub) Delete(_ context.Context, command usersdomain.AccountDeletion) (bool, error) {
	stub.called = true
	stub.userID = command.UserID
	return stub.pending, stub.err
}

func TestDeleteAccountHTTPUsesAuthenticatedIdentityAndClearsCookieOnlyOnSuccess(t *testing.T) {
	for _, test := range []struct {
		name    string
		pending bool
		err     error
		status  int
	}{
		{"complete", false, nil, http.StatusNoContent},
		{"cleanup", true, nil, http.StatusAccepted},
		{"ownership", false, &usersdomain.AccountDeletionConflict{Workspaces: []string{"My workspace"}}, http.StatusConflict},
		{"failure", false, errors.New("internal secret"), http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &accountDeletionStub{pending: test.pending, err: test.err}
			handler := &Handlers{users: users.New(nil, nil, nil, users.WithAccountDeletion(stub))}
			id := uuid.New()
			ctx := platformauth.SetUserID(t.Context(), id)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodDelete, "https://api.fortyone.app/users/account", nil)
			require.NoError(t, handler.DeleteAccount(ctx, recorder, request))
			require.Equal(t, test.status, recorder.Code)
			require.Equal(t, id, stub.userID)
			cookies := recorder.Result().Cookies()
			if test.err == nil {
				require.NotEmpty(t, cookies)
				require.Equal(t, -1, cookies[0].MaxAge)
			} else {
				require.Empty(t, cookies)
			}
			require.NotContains(t, recorder.Body.String(), "internal secret")
			if test.pending {
				require.Contains(t, recorder.Body.String(), "cleanup_pending")
			}
			if test.status == http.StatusConflict {
				require.Contains(t, recorder.Body.String(), "My workspace")
			}
		})
	}
}

func TestDeleteAccountHTTPRejectsAnonymousRequest(t *testing.T) {
	stub := &accountDeletionStub{}
	handler := &Handlers{users: users.New(nil, nil, nil, users.WithAccountDeletion(stub))}
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.DeleteAccount(t.Context(), recorder, httptest.NewRequest(http.MethodDelete, "/users/account", nil)))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.False(t, stub.called)
}

func TestDeleteAccountHTTPRejectsAccountChangedSinceConfirmation(t *testing.T) {
	stub := &accountDeletionStub{}
	handler := &Handlers{users: users.New(nil, nil, nil, users.WithAccountDeletion(stub))}
	request := httptest.NewRequest(http.MethodDelete, "/users/account", strings.NewReader(`{"expectedUserId":"`+uuid.NewString()+`"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.DeleteAccount(platformauth.SetUserID(t.Context(), uuid.New()), recorder, request))
	require.Equal(t, http.StatusConflict, recorder.Code)
	require.False(t, stub.called)
	require.Empty(t, recorder.Result().Cookies())
}
