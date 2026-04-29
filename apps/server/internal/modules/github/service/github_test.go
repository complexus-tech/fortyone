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

func TestPullRequestWorkflowEventOnlyMatchesWorkflowActions(t *testing.T) {
	eventKey, ok := pullRequestWorkflowEvent("opened", false, false)
	require.True(t, ok)
	require.Equal(t, EventPROpen, eventKey)

	eventKey, ok = pullRequestWorkflowEvent("opened", true, false)
	require.True(t, ok)
	require.Equal(t, EventDraftPROpen, eventKey)

	eventKey, ok = pullRequestWorkflowEvent("ready_for_review", false, false)
	require.True(t, ok)
	require.Equal(t, EventPRReadyForMerge, eventKey)

	eventKey, ok = pullRequestWorkflowEvent("closed", false, true)
	require.True(t, ok)
	require.Equal(t, EventPRMerge, eventKey)

	for _, action := range []string{"edited", "synchronize", "reopened", "closed"} {
		eventKey, ok = pullRequestWorkflowEvent(action, false, false)
		require.False(t, ok, "action %s should not trigger workflow event %q", action, eventKey)
	}
}

func TestFortyOneCommentMarkerIsHiddenAndStripped(t *testing.T) {
	commentID := uuid.New()
	body := buildFortyOneUserCommentBody("Ship it", commentID)

	require.Contains(t, body, "Ship it")
	require.Contains(t, body, "<!-- fortyone:comment:"+commentID.String()+" -->")
	require.True(t, isFortyOneAuthoredCommentBody(body))
	require.Equal(t, "Ship it", stripFortyOneCommentMarker(body))
}
