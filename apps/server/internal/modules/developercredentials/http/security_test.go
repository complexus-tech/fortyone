package developercredentialshttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type managementServiceSpy struct {
	Service
	createServiceAccountCalled bool
}

func (spy *managementServiceSpy) CreateServiceAccount(
	context.Context,
	developercredentialsdomain.Access,
	developercredentials.CreateServiceAccountInput,
) (developercredentialsdomain.ServiceAccount, error) {
	spy.createServiceAccountCalled = true
	return developercredentialsdomain.ServiceAccount{}, nil
}

func TestManagementHandlerExplicitlyRejectsMachinePrincipal(t *testing.T) {
	t.Parallel()

	actor, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalServiceAccount, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeServiceAccountsManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(uuid.New())
	require.NoError(t, err)
	ctx, err := platformauth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	spy := &managementServiceSpy{}
	handlers := New(spy)
	request := httptest.NewRequest(http.MethodPost, "/workspaces/acme/service-accounts", strings.NewReader(`{
		"name":"forbidden","workspaceRole":"member"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	require.NoError(t, handlers.CreateServiceAccount(ctx, recorder, request))
	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.False(t, spy.createServiceAccountCalled)
}

func TestCredentialResponsesNeverExposeStoredSecretMaterial(t *testing.T) {
	t.Parallel()

	credential := developercredentialsdomain.Credential{
		ID: uuid.New(), PrincipalID: uuid.New(), Kind: developercredentialsdomain.CredentialPersonalAccessToken,
		Name: "CLI", LookupPrefix: "abcdef012345", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
		TeamIDs: []uuid.UUID{}, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	encoded, err := json.Marshal(credentialModel(credential))
	require.NoError(t, err)
	body := string(encoded)
	require.NotContains(t, body, `"token":`)
	require.NotContains(t, body, "digest")
	require.NotContains(t, body, "secret")
	require.Contains(t, body, `"prefix":"abcdef012345"`)

	issued := issuedCredentialModel(developercredentialsdomain.IssuedCredential{
		Credential: credential,
		Token:      developercredentialsdomain.NewPlaintextToken("f41_pat_v1_abcdef012345_show-once"),
	})
	issuedJSON, err := json.Marshal(issued)
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(issuedJSON), "show-once"))
}
