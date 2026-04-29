package github

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserLinkStateRequiresSameUserAndValidReturnPath(t *testing.T) {
	service := &Service{cfg: Config{SecretKey: "test-secret"}}
	userID := uuid.New()

	state, err := service.createUserLinkState(userID, "/settings/account/profile?tab=integrations", time.Now().Add(15*time.Minute))
	require.NoError(t, err)

	returnTo, err := service.verifyUserLinkState(state, userID, time.Now())
	require.NoError(t, err)
	require.Equal(t, "/settings/account/profile?tab=integrations", returnTo)

	_, err = service.verifyUserLinkState(state, uuid.New(), time.Now())
	require.Error(t, err)
}

func TestUserLinkStateRejectsExpiredOrUnsafeReturnPath(t *testing.T) {
	service := &Service{cfg: Config{SecretKey: "test-secret"}}
	userID := uuid.New()

	expiredState, err := service.createUserLinkState(userID, "/settings/account/profile", time.Now().Add(-time.Minute))
	require.NoError(t, err)

	_, err = service.verifyUserLinkState(expiredState, userID, time.Now())
	require.Error(t, err)

	_, err = service.createUserLinkState(userID, "https://evil.example/callback", time.Now().Add(15*time.Minute))
	require.Error(t, err)
}

func TestIsFortyOneAuthoredCommentBody(t *testing.T) {
	require.True(t, isFortyOneAuthoredCommentBody("**Joseph** commented on FortyOne:\n\nLooks good."))
	require.True(t, isFortyOneAuthoredCommentBody("**Joseph** commented via FortyOne:\n\nLooks good."))
	require.False(t, isFortyOneAuthoredCommentBody("**octocat** commented on GitHub issue #12:\n\nLooks good."))
	require.False(t, isFortyOneAuthoredCommentBody("A regular GitHub comment."))
}
