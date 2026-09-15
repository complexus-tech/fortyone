package feedbackhttp

import (
	"encoding/json"
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRetainedFeedbackAuthorHasNoProfileOrPrivateIdentity(t *testing.T) {
	t.Parallel()
	avatar := "profiles/stale-person.png"
	for _, masked := range []bool{false, true} {
		item := toAppItem(feedback.CoreItem{
			AuthorID: uuid.MustParse(formerUserID), AuthorName: "Stale name",
			AuthorAvatar: &avatar, AuthorEmail: "stale@example.com", AuthorMasked: masked,
			Title: "Shared request", Description: "Keep this discussion", VoteCount: 4,
		}, nil, nil)
		comment := toAppComment(feedback.CoreComment{
			AuthorID: uuid.MustParse(formerUserID), AuthorName: "Anonymous",
			AuthorAvatar: &avatar, AuthorMasked: masked, Body: "Shared reply",
		})
		require.Equal(t, "Former user", item.AuthorName)
		require.Equal(t, "Former user", comment.AuthorName)
		require.Nil(t, item.AuthorID)
		require.Nil(t, comment.AuthorID)
		require.Nil(t, item.AuthorAvatar)
		require.Nil(t, comment.AuthorAvatar)
		require.Equal(t, "Keep this discussion", item.Description)
		require.Equal(t, 4, item.VoteCount)
		require.Equal(t, "Shared reply", comment.Body)
		encoded, err := json.Marshal([]any{item, comment})
		require.NoError(t, err)
		for _, private := range []string{formerUserID, avatar, "stale@example.com", "Stale name"} {
			require.NotContains(t, string(encoded), private)
		}
	}
}

func TestAnonymousAndMissingFeedbackAuthorsAreNotFormerUsers(t *testing.T) {
	t.Parallel()
	for _, id := range []uuid.UUID{uuid.Nil, uuid.New()} {
		authorID, name, avatar := presentAuthor(id, "Anonymous", nil, true)
		require.Nil(t, authorID)
		require.Nil(t, avatar)
		require.Equal(t, "Anonymous", name)
	}
}
