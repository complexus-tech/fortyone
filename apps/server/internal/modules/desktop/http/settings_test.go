package desktophttp

import (
	"encoding/json"
	"testing"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSettingsUserNeverSerializesSessionOrIntegrationFields(t *testing.T) {
	secret, githubUsername := "credential-must-not-be-cached", "private-integration"
	user := users.CoreUser{ID: uuid.New(), FullName: "Offline user", Email: "user@example.com", Timezone: "Africa/Harare", Token: &secret, GitHubUsername: &githubUsername, IsInternal: true}
	encoded, err := json.Marshal(settingsUser(user))
	require.NoError(t, err)
	require.NotContains(t, string(encoded), secret)
	require.NotContains(t, string(encoded), githubUsername)
	require.NotContains(t, string(encoded), "isInternal")
	require.NotContains(t, string(encoded), "token")
	require.Contains(t, string(encoded), `"fullName":"Offline user"`)
}
