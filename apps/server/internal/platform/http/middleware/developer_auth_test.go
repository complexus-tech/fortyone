package mid

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type developerResolverFunc func(context.Context, string) (platformauth.Actor, error)

func (resolve developerResolverFunc) ResolveDeveloperCredential(
	ctx context.Context,
	token string,
) (platformauth.Actor, error) {
	return resolve(ctx, token)
}

func TestDeveloperAuthAcceptsAttributedMachineAndOAuthActors(t *testing.T) {
	t.Parallel()

	machine, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalPersonalToken, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	machine, err = machine.WithWorkspace(uuid.New())
	require.NoError(t, err)
	oauthUser, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalOAuthUser, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	oauthApplication, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalOAuthApplication, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	oauthApplication, err = oauthApplication.WithWorkspace(uuid.New())
	require.NoError(t, err)

	for _, actor := range []platformauth.Actor{machine, oauthUser, oauthApplication} {
		actor := actor
		t.Run(string(actor.Kind), func(t *testing.T) {
			t.Parallel()
			var observed platformauth.Actor
			handler := DeveloperAuthWithErrorResponder(
				testMachineLogger(t),
				developerResolverFunc(func(_ context.Context, token string) (platformauth.Actor, error) {
					require.Equal(t, "opaque-developer-token", token)
					return actor, nil
				}),
				webTestErrorResponder,
			)(func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
				var getErr error
				observed, getErr = platformauth.GetActor(ctx)
				writer.WriteHeader(http.StatusNoContent)
				return getErr
			})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/id", nil)
			request.Header.Set("Authorization", "Bearer opaque-developer-token")
			recorder := httptest.NewRecorder()

			require.NoError(t, handler(context.Background(), recorder, request))
			require.Equal(t, http.StatusNoContent, recorder.Code)
			require.Equal(t, actor, observed)
		})
	}
}

func TestDeveloperAuthRejectsUnsupportedOrIncorrectlyBoundActorsWithoutLeakingBearer(t *testing.T) {
	t.Parallel()

	rawToken := "header.sensitive-payload.signature"
	unboundMachine, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalPersonalToken, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	boundOAuth, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalOAuthUser, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	boundOAuth, err = boundOAuth.WithWorkspace(uuid.New())
	require.NoError(t, err)
	unboundOAuthApplication, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalOAuthApplication, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)

	tests := []struct {
		name  string
		actor platformauth.Actor
		err   error
	}{
		{name: "resolver error", err: errors.New("invalid")},
		{name: "unbound machine", actor: unboundMachine},
		{name: "prematurely bound OAuth user", actor: boundOAuth},
		{name: "unbound OAuth application", actor: unboundOAuthApplication},
		{name: "browser user", actor: platformauth.NewHumanActor(uuid.New())},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			log := logger.NewWithJSON(&output, slog.LevelDebug, "developer-auth-test")
			handler := DeveloperAuthWithErrorResponder(
				log,
				developerResolverFunc(func(context.Context, string) (platformauth.Actor, error) {
					return test.actor, test.err
				}),
				webTestErrorResponder,
			)(func(context.Context, http.ResponseWriter, *http.Request) error {
				t.Fatal("next handler must not run")
				return nil
			})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/id?access_token="+rawToken, nil)
			request.Header.Set("Authorization", "Bearer "+rawToken)
			recorder := httptest.NewRecorder()

			require.NoError(t, handler(context.Background(), recorder, request))
			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			require.NotContains(t, recorder.Body.String(), rawToken)
			require.NotContains(t, output.String(), rawToken)
		})
	}
}

func webTestErrorResponder(_ context.Context, writer http.ResponseWriter, _ error, status int) error {
	writer.WriteHeader(status)
	return nil
}
