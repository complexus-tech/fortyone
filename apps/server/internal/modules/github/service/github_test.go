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

func TestGitHubActivityReasonsExplainAutomationSource(t *testing.T) {
	require.Equal(
		t,
		"GitHub issue details changed, so FortyOne synced the linked story.",
		githubIssueSyncReason(),
	)
	require.Equal(
		t,
		"A GitHub workflow automation matched a repository event and moved this story to the configured status.",
		githubWorkflowAutomationReason(),
	)
}

func TestFortyOneCommentMarkerIsHiddenAndStripped(t *testing.T) {
	commentID := uuid.New()
	body := buildFortyOneUserCommentBody("Ship it", commentID)

	require.Contains(t, body, "Ship it")
	require.Contains(t, body, "<!-- fortyone:comment:"+commentID.String()+" -->")
	require.True(t, isFortyOneAuthoredCommentBody(body))
	require.Equal(t, "Ship it", stripFortyOneCommentMarker(body))
}

func TestGitHubPriorityFromLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   []string
		expected string
	}{
		{name: "urgent beats lower priority labels", labels: []string{"enhancement", "priority: high", "P0"}, expected: "Urgent"},
		{name: "high maps common priority labels", labels: []string{"type: bug", "priority/high"}, expected: "High"},
		{name: "medium maps p2 labels", labels: []string{"P2"}, expected: "Medium"},
		{name: "low maps low labels", labels: []string{"good first issue", "priority: low"}, expected: "Low"},
		{name: "unknown labels are not treated as priority", labels: []string{"bug", "backend"}, expected: "No Priority"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, githubPriorityFromLabelNames(tt.labels))
		})
	}
}

func TestGitHubPriorityUpdateFromLabelsOnlyUpdatesWhenPriorityLabelExists(t *testing.T) {
	priority, ok := githubPriorityUpdateFromLabelNames([]string{"bug", "backend"})
	require.False(t, ok)
	require.Equal(t, "No Priority", priority)

	priority, ok = githubPriorityUpdateFromLabelNames([]string{"bug", "P2"})
	require.True(t, ok)
	require.Equal(t, "Medium", priority)
}

func TestIssueActionMatchesStoryStatusCategory(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		category string
		expected bool
	}{
		{name: "closed issue already completed locally", action: "closed", category: "completed", expected: true},
		{name: "closed issue still unstarted locally", action: "closed", category: "unstarted", expected: false},
		{name: "reopened issue already active locally", action: "reopened", category: "started", expected: true},
		{name: "reopened issue already unstarted locally", action: "reopened", category: "unstarted", expected: true},
		{name: "reopened issue still completed locally", action: "reopened", category: "completed", expected: false},
		{name: "reopened issue still cancelled locally", action: "reopened", category: "cancelled", expected: false},
		{name: "edited issue never matches workflow state", action: "edited", category: "completed", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, issueActionMatchesStoryStatusCategory(tt.action, tt.category))
		})
	}
}
