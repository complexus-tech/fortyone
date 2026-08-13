package feedbackhttp

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNextPhaseHTTPStatusMappingsFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err    error
		status int
	}{
		{feedback.ErrContributorSessionInvalid, http.StatusUnauthorized},
		{feedback.ErrWidgetAssertionInvalid, http.StatusUnauthorized},
		{feedback.ErrContributorBlocked, http.StatusForbidden},
		{feedback.ErrWidgetOriginNotAllowed, http.StatusForbidden},
		{feedback.ErrVerificationExpired, http.StatusGone},
		{feedback.ErrVerificationConsumed, http.StatusGone},
		{feedback.ErrVerificationAttempts, http.StatusTooManyRequests},
		{feedback.ErrWidgetAssertionReplayed, http.StatusConflict},
		{feedback.ErrMergeConflict, http.StatusConflict},
		{feedback.ErrFeatureUnavailable, http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		require.Equal(t, test.status, httpStatus(test.err), test.err.Error())
	}
}

func TestPrivateAuthorContractIsSeparateFromPublicItem(t *testing.T) {
	t.Parallel()
	email := "private@example.com"
	contributorID := uuid.New()
	publicItem := toAppItem(feedback.CoreItem{
		ID: uuid.New(), ContributorID: contributorID, ParticipantKind: feedback.ContributorKindVerifiedGuest,
		AuthorName: "Anonymous", AuthorEmail: email, AuthorMasked: true,
	}, nil, nil)
	privateAuthor := AppPrivateAuthor{
		ContributorID: contributorID, Kind: feedback.ContributorKindVerifiedGuest,
		DisplayName: "Private Person", Email: &email, PublicMasked: true,
	}

	publicJSON, err := json.Marshal(publicItem)
	require.NoError(t, err)
	require.NotContains(t, string(publicJSON), email)
	require.NotContains(t, string(publicJSON), "privateAuthor")

	privateJSON, err := json.Marshal(privateAuthor)
	require.NoError(t, err)
	require.Contains(t, string(privateJSON), email)
	require.Contains(t, string(privateJSON), contributorID.String())
}

func TestItemCandidateResponseUsesStableCamelCaseContract(t *testing.T) {
	t.Parallel()
	encoded, err := json.Marshal(AppMergeCandidatesPage{Candidates: []AppMergeCandidate{{
		ID: uuid.New(), Slug: "dark-mode", Title: "Dark mode", Status: feedback.StatusCompleted,
		VoteCount: 12, CommentCount: 4,
	}}, HasMore: true})
	require.NoError(t, err)
	text := string(encoded)
	for _, field := range []string{`"candidates"`, `"voteCount"`, `"commentCount"`, `"hasMore"`} {
		require.Contains(t, text, field)
	}
}

func TestPublicGuestSerializersNeverExposeCanonicalContributorID(t *testing.T) {
	t.Parallel()
	contributorID := uuid.New()
	privateAvatar := "private/guest-avatar.webp"
	item := toAppItem(feedback.CoreItem{
		ID: uuid.New(), ContributorID: contributorID, ParticipantKind: feedback.ContributorKindVerifiedGuest,
		AuthorName: "Anonymous", AuthorAvatar: &privateAvatar, AuthorMasked: true,
	}, nil, nil)
	comment := toAppComment(feedback.CoreComment{
		ID: uuid.New(), ContributorID: contributorID, ParticipantKind: feedback.ContributorKindVerifiedGuest,
		AuthorName: "Anonymous", AuthorAvatar: &privateAvatar, AuthorMasked: true,
	})

	require.Nil(t, item.AuthorID)
	require.Nil(t, comment.AuthorID)
	require.Nil(t, item.AuthorAvatar)
	require.Nil(t, comment.AuthorAvatar)
	require.Equal(t, "Anonymous", item.AuthorName)
	require.Equal(t, "Anonymous", comment.AuthorName)
	require.True(t, item.AuthorMasked)
	require.True(t, comment.AuthorMasked)

	encoded, err := json.Marshal(struct {
		Item    AppItem    `json:"item"`
		Comment AppComment `json:"comment"`
	}{Item: item, Comment: comment})
	require.NoError(t, err)
	require.NotContains(t, string(encoded), contributorID.String())
	require.NotContains(t, string(encoded), privateAvatar)
}

func TestAccountSerializerKeepsPublicIdentityWhenGuestPolicyMasksGuests(t *testing.T) {
	t.Parallel()
	accountID := uuid.New()
	avatar := "profiles/account.webp"
	item := toAppItem(feedback.CoreItem{
		ID: uuid.New(), AuthorID: accountID, ParticipantKind: feedback.ContributorKindAccount,
		AuthorName: "Account Person", AuthorAvatar: &avatar,
	}, nil, nil)
	comment := toAppComment(feedback.CoreComment{
		ID: uuid.New(), AuthorID: accountID, ParticipantKind: feedback.ContributorKindAccount,
		AuthorName: "Account Person", AuthorAvatar: &avatar,
	})

	require.Equal(t, accountID, *item.AuthorID)
	require.Equal(t, accountID, *comment.AuthorID)
	require.Equal(t, avatar, *item.AuthorAvatar)
	require.Equal(t, avatar, *comment.AuthorAvatar)
	require.False(t, item.AuthorMasked)
	require.False(t, comment.AuthorMasked)
}

func TestNextPhaseResponseModelsUseStableCamelCaseContract(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	update := toAppUpdate(feedback.CoreFeedbackUpdate{
		ID: uuid.New(), WorkspaceID: uuid.New(), PortalID: uuid.New(), Slug: "dark-mode", Title: "Dark mode",
		Body: "Shipped", Status: feedback.FeedbackUpdateStatusPublished, PublishedAt: &now,
		LinkedItems: []feedback.CoreUpdateItem{{ID: uuid.New(), Slug: "dark-mode-request", Title: "Dark mode", Status: feedback.StatusCompleted}},
		CreatedAt:   now, UpdatedAt: now,
	})
	encoded, err := json.Marshal(AppFeedbackUpdatesPage{Updates: []AppFeedbackUpdate{update}, Page: 1, PageSize: 20, HasMore: false, UnreadCount: 1})
	require.NoError(t, err)
	text := string(encoded)
	for _, field := range []string{`"portalId"`, `"workspaceId"`, `"linkedItems"`, `"publishedAt"`, `"unreadCount"`, `"hasMore"`} {
		require.Contains(t, text, field)
	}
}
