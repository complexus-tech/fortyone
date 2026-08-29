package mid

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type machineResolverFunc func(context.Context, string) (platformauth.Actor, error)

func (resolve machineResolverFunc) ResolveMachineCredential(ctx context.Context, token string) (platformauth.Actor, error) {
	return resolve(ctx, token)
}

func TestMachineAuthAcceptsOnlyResolvedMachineActors(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	credentialID := uuid.New()
	principalID := uuid.New()
	actor, err := platformauth.NewActor(
		principalID,
		platformauth.PrincipalServiceAccount,
		credentialID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)

	resolver := machineResolverFunc(func(_ context.Context, token string) (platformauth.Actor, error) {
		require.Equal(t, "f41_sak_v1_public_secret", token)
		return actor, nil
	})
	var observed platformauth.Actor
	handler := MachineAuth(testMachineLogger(t), resolver)(func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
		var getErr error
		observed, getErr = platformauth.GetActor(ctx)
		writer.WriteHeader(http.StatusNoContent)
		return getErr
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/stories", nil)
	request.Header.Set("Authorization", "Bearer f41_sak_v1_public_secret")
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(context.Background(), recorder, request))
	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Equal(t, actor.PrincipalID, observed.PrincipalID)
	require.Equal(t, actor.CredentialID, observed.CredentialID)
	require.Equal(t, actor.WorkspaceID, observed.WorkspaceID)
}

func TestMachineAuthFailsClosedWithoutLeakingBearer(t *testing.T) {
	t.Parallel()

	rawToken := "f41_pat_v1_public_super-secret-material"
	credentiallessActor, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalServiceAccount, uuid.Nil,
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	credentiallessActor, err = credentiallessActor.WithWorkspace(uuid.New())
	require.NoError(t, err)
	workspaceLessActor, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalServiceAccount, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	tests := []struct {
		name     string
		headers  []string
		resolver machineResolverFunc
	}{
		{name: "missing bearer", resolver: rejectingMachineResolver},
		{name: "query bearer ignored", resolver: rejectingMachineResolver},
		{name: "malformed bearer", headers: []string{"Bearer one two"}, resolver: rejectingMachineResolver},
		{name: "duplicate authorization", headers: []string{"Bearer " + rawToken, "Bearer " + rawToken}, resolver: rejectingMachineResolver},
		{name: "verification failure", headers: []string{"Bearer " + rawToken}, resolver: rejectingMachineResolver},
		{name: "human actor rejected", headers: []string{"Bearer " + rawToken}, resolver: func(context.Context, string) (platformauth.Actor, error) {
			return platformauth.NewHumanActor(uuid.New()), nil
		}},
		{name: "credentialless machine actor rejected", headers: []string{"Bearer " + rawToken}, resolver: func(context.Context, string) (platformauth.Actor, error) {
			return credentiallessActor, nil
		}},
		{name: "workspace-less machine actor rejected", headers: []string{"Bearer " + rawToken}, resolver: func(context.Context, string) (platformauth.Actor, error) {
			return workspaceLessActor, nil
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var logOutput bytes.Buffer
			log := logger.NewWithJSON(&logOutput, slog.LevelDebug, "machine-auth-test")
			handler := MachineAuth(log, test.resolver)(func(context.Context, http.ResponseWriter, *http.Request) error {
				t.Fatal("next handler must not run")
				return nil
			})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/stories?token="+rawToken, nil)
			request.AddCookie(&http.Cookie{Name: authCookieName, Value: rawToken})
			for _, header := range test.headers {
				request.Header.Add("Authorization", header)
			}
			recorder := httptest.NewRecorder()

			require.NoError(t, handler(context.Background(), recorder, request))
			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			require.NotContains(t, recorder.Body.String(), rawToken)
			require.NotContains(t, logOutput.String(), rawToken)
		})
	}
}

func TestMachineAuthWithErrorResponderPreservesTransportEnvelope(t *testing.T) {
	t.Parallel()

	called := false
	respond := func(_ context.Context, writer http.ResponseWriter, err error, status int) error {
		called = true
		require.ErrorIs(t, err, errMachineAuthenticationRequired)
		require.Equal(t, http.StatusUnauthorized, status)
		writer.Header().Set("Content-Type", "application/problem+json")
		writer.WriteHeader(status)
		_, writeErr := writer.Write([]byte(`{"error":{"code":"machine_authentication_required"}}`))
		return writeErr
	}
	handler := MachineAuthWithErrorResponder(
		testMachineLogger(t), machineResolverFunc(rejectingMachineResolver), respond,
	)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler must not run")
		return nil
	})
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(context.Background(), recorder, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)))
	require.True(t, called)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.JSONEq(t, `{"error":{"code":"machine_authentication_required"}}`, recorder.Body.String())
}

func TestMachineAuthKeepsLegacyRoutesOnLegacyEnvelope(t *testing.T) {
	t.Parallel()

	handler := MachineAuth(
		testMachineLogger(t), machineResolverFunc(rejectingMachineResolver),
	)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler must not run")
		return nil
	})
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(context.Background(), recorder, httptest.NewRequest(http.MethodGet, "/legacy", nil)))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.JSONEq(t, `{"data":null,"error":{"code":"authentication_required","message":"valid machine bearer credential required","hint":"Authenticate with a FortyOne session that can access the requested workspace."}}`, recorder.Body.String())
}

func rejectingMachineResolver(context.Context, string) (platformauth.Actor, error) {
	return platformauth.Actor{}, errors.New("rejected")
}

func testMachineLogger(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, strings.ReplaceAll(t.Name(), "/", "-"))
}
