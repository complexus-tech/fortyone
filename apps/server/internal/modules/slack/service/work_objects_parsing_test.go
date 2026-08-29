package slack

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseFortyOneStoryURL(t *testing.T) {
	t.Parallel()

	link, err := ParseFortyOneStoryURL("https://acme.fortyone.app/work/web-123?from=slack")
	require.NoError(t, err)
	require.Equal(t, "acme", link.WorkspaceSlug)
	require.Equal(t, "WEB-123", link.StoryReference)
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", link.CanonicalURL)
	require.Equal(t, "https://acme.fortyone.app/work/web-123?from=slack", link.PostedURL)
}

func TestParseFortyOneStoryURLRejectsUntrustedOrNonStoryURLs(t *testing.T) {
	t.Parallel()

	invalidURLs := []string{
		"http://acme.fortyone.app/work/WEB-123",
		"https://fortyone.app/work/WEB-123",
		"https://acme.fortyone.app.evil.example/work/WEB-123",
		"https://one.two.fortyone.app/work/WEB-123",
		"https://acme.fortyone.app:8443/work/WEB-123",
		"https://user@acme.fortyone.app/work/WEB-123",
		"https://acme.fortyone.app/feedback/WEB-123",
		"https://acme.fortyone.app/work/WEB-0",
		"https://acme.fortyone.app/work/WEB-123%2Fprivate",
	}
	for _, rawURL := range invalidURLs {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()
			_, err := ParseFortyOneStoryURL(rawURL)
			require.ErrorIs(t, err, ErrInvalidFortyOneStoryURL)
		})
	}
}

func TestParseFortyOneRequestURL(t *testing.T) {
	t.Parallel()

	teamID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	requestID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	link, err := ParseFortyOneRequestURL(
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
	)
	require.NoError(t, err)
	require.Equal(t, "acme", link.WorkspaceSlug)
	require.Equal(t, teamID, link.TeamID)
	require.Equal(t, requestID, link.RequestID)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		link.CanonicalURL,
	)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222?from=slack",
		link.PostedURL,
	)

	trailingSlashLink, err := ParseFortyOneRequestURL(
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222/",
	)
	require.NoError(t, err)
	require.Equal(t, link.CanonicalURL, trailingSlashLink.CanonicalURL)
	require.Equal(t,
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222/",
		trailingSlashLink.PostedURL,
	)
}

func TestParseFortyOneRequestURLRejectsUntrustedOrNonCanonicalURLs(t *testing.T) {
	t.Parallel()

	const canonicalPath = "/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222"
	invalidURLs := []string{
		"http://acme.fortyone.app" + canonicalPath,
		"https://fortyone.app" + canonicalPath,
		"https://acme.fortyone.app.evil.example" + canonicalPath,
		"https://one.two.fortyone.app" + canonicalPath,
		"https://acme.fortyone.app:8443" + canonicalPath,
		"https://user@acme.fortyone.app" + canonicalPath,
		"https://acme.fortyone.app/feedback/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/not-a-uuid/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/00000000-0000-0000-0000-000000000000/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/not-a-uuid",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/00000000-0000-0000-0000-000000000000",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222/activity",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222//",
		"https://acme.fortyone.app//teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/Teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/11111111111141118111111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/{11111111-1111-4111-8111-111111111111}/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/urn:uuid:11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222%2Fprivate",
		"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222%5Cprivate",
		"https://acme.fortyone.app/teams/11111111%2D1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
	}
	for _, rawURL := range invalidURLs {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()
			_, err := ParseFortyOneRequestURL(rawURL)
			require.ErrorIs(t, err, ErrInvalidFortyOneRequestURL)
		})
	}
}

func TestParseFortyOneObjectiveAndSprintURLs(t *testing.T) {
	t.Parallel()

	teamID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	objectiveID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	sprintID := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	objective, err := ParseFortyOneObjectiveURL("https://acme.fortyone.app/teams/" + teamID.String() + "/objectives/" + objectiveID.String() + "?from=slack")
	require.NoError(t, err)
	require.Equal(t, teamID, objective.TeamID)
	require.Equal(t, objectiveID, objective.ObjectiveID)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/objectives/"+objectiveID.String(), objective.CanonicalURL)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/objectives/"+objectiveID.String()+"?from=slack", objective.PostedURL)

	sprint, err := ParseFortyOneSprintURL("https://acme.fortyone.app/teams/" + teamID.String() + "/sprints/" + sprintID.String() + "/stories/")
	require.NoError(t, err)
	require.Equal(t, teamID, sprint.TeamID)
	require.Equal(t, sprintID, sprint.SprintID)
	require.Equal(t, "https://acme.fortyone.app/teams/"+teamID.String()+"/sprints/"+sprintID.String()+"/stories", sprint.CanonicalURL)
}

func TestParseFortyOneObjectiveAndSprintURLsRejectInvalidRoutes(t *testing.T) {
	t.Parallel()

	invalidURLs := []struct {
		name string
		url  string
		err  error
	}{
		{
			name: "objective wrong route",
			url:  "https://acme.fortyone.app/objectives/22222222-2222-4222-8222-222222222222",
			err:  ErrInvalidFortyOneObjectiveURL,
		},
		{
			name: "objective encoded path",
			url:  "https://acme.fortyone.app/teams/11111111%2D1111-4111-8111-111111111111/objectives/22222222-2222-4222-8222-222222222222",
			err:  ErrInvalidFortyOneObjectiveURL,
		},
		{
			name: "sprint missing stories route",
			url:  "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/sprints/33333333-3333-4333-8333-333333333333",
			err:  ErrInvalidFortyOneSprintURL,
		},
		{
			name: "sprint wrong trailing route",
			url:  "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/sprints/33333333-3333-4333-8333-333333333333/analytics",
			err:  ErrInvalidFortyOneSprintURL,
		},
	}
	for _, testCase := range invalidURLs {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if testCase.err == ErrInvalidFortyOneObjectiveURL {
				_, err := ParseFortyOneObjectiveURL(testCase.url)
				require.ErrorIs(t, err, testCase.err)
				return
			}
			_, err := ParseFortyOneSprintURL(testCase.url)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}
