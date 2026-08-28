package developeroauth

import (
	"testing"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/stretchr/testify/require"
)

func TestPublicAPIResourceScopePolicyRequiresAnExplicitCapability(t *testing.T) {
	t.Parallel()

	policy, err := newScopePolicy(PublicAPIResourceScopePolicy())
	require.NoError(t, err)

	scopes, err := policy.normalize([]string{"stories:read"})
	require.NoError(t, err)
	require.Equal(t, []string{developeroauthdomain.ScopeOfflineAccess, "stories:read"}, scopes)
	require.True(t, policy.accepts(scopes))

	_, err = policy.normalize([]string{developeroauthdomain.ScopeOfflineAccess})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidScope)

	_, err = policy.normalize([]string{developeroauthdomain.ScopeMCPAccess})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidScope)
	require.False(t, policy.accepts([]string{developeroauthdomain.ScopeOfflineAccess, "admin:*"}))
}

func TestPublicAPIApplicationActorScopePolicyIsIntentionallyNarrow(t *testing.T) {
	t.Parallel()

	policy, err := newScopePolicy(PublicAPIApplicationActorScopePolicy())
	require.NoError(t, err)

	accepted, err := policy.normalize([]string{string(platformauth.ScopeStoriesWrite)})
	require.NoError(t, err)
	require.Equal(t, []string{string(platformauth.ScopeStoriesWrite)}, accepted)

	for _, denied := range []string{
		developeroauthdomain.ScopeOfflineAccess,
		developeroauthdomain.ScopeMCPAccess,
		string(platformauth.ScopeStoriesRead),
		string(platformauth.ScopeWebhooksManage),
	} {
		_, err := policy.normalize([]string{denied})
		require.ErrorIs(t, err, developeroauthdomain.ErrInvalidScope)
	}
}

func TestDefaultScopePolicyRemainsAudienceBoundToMCP(t *testing.T) {
	t.Parallel()

	policy, err := newScopePolicy(ScopePolicy{})
	require.NoError(t, err)

	scopes, err := policy.normalize(nil)
	require.NoError(t, err)
	require.Equal(t, []string{developeroauthdomain.ScopeMCPAccess, developeroauthdomain.ScopeOfflineAccess}, scopes)
	require.True(t, policy.accepts(scopes))
	require.False(t, policy.accepts([]string{developeroauthdomain.ScopeOfflineAccess, "stories:read"}))
}

func TestScopePolicyRejectsAmbiguousOrUnsatisfiedConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		policy ScopePolicy
	}{
		{name: "empty supported catalog", policy: ScopePolicy{Supported: []string{}}},
		{name: "whitespace", policy: ScopePolicy{Supported: []string{" stories:read"}}},
		{name: "duplicate supported", policy: ScopePolicy{Supported: []string{"stories:read", "stories:read"}}},
		{name: "unknown required", policy: ScopePolicy{Supported: []string{"stories:read"}, Required: []string{"stories:write"}}},
		{name: "duplicate required", policy: ScopePolicy{Supported: []string{"stories:read"}, Required: []string{"stories:read", "stories:read"}}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := newScopePolicy(test.policy)
			require.Error(t, err)
		})
	}
}
