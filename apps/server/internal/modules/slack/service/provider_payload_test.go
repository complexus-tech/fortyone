package slack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlackProviderPayloadAcceptsTypedStoryAndRequestWorkObjectIdentities(t *testing.T) {
	t.Parallel()

	storyReceipt, err := BuildSlackStoryCreationReceipt("Joseph", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	_, err = EncodeSlackProviderPayload(storyReceipt.ProviderPayload)
	require.NoError(t, err)

	requestReceipt, err := BuildSlackRequestCreationReceipt("Joseph", SlackRequestWorkObjectInput{
		AccessGranted: true,
		RequestURL:    "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	_, err = EncodeSlackProviderPayload(requestReceipt.ProviderPayload)
	require.NoError(t, err)

	objectiveUnfurl, err := BuildSlackObjectiveUnfurlRequest("C123", "1754700000.123", SlackObjectiveWorkObjectInput{
		AccessGranted: true,
		ObjectiveURL:  "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/objectives/22222222-2222-4222-8222-222222222222",
		Title:         "Improve onboarding",
	})
	require.NoError(t, err)
	_, err = EncodeSlackProviderPayload(SlackProviderPayload{Metadata: objectiveUnfurl.Metadata})
	require.NoError(t, err)

	sprintUnfurl, err := BuildSlackSprintUnfurlRequest("C123", "1754700000.123", SlackSprintWorkObjectInput{
		AccessGranted: true,
		SprintURL:     "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/sprints/33333333-3333-4333-8333-333333333333/stories",
		Title:         "Sprint 12",
	})
	require.NoError(t, err)
	_, err = EncodeSlackProviderPayload(SlackProviderPayload{Metadata: sprintUnfurl.Metadata})
	require.NoError(t, err)
}

func TestSlackProviderPayloadRejectsCrossTypedWorkObjectIdentities(t *testing.T) {
	t.Parallel()

	t.Run("request URL cannot claim story identity", func(t *testing.T) {
		payload := requestProviderPayloadForTest(t)
		payload.Metadata.Entities[0].ExternalRef.Type = slackStoryExternalRefType

		_, err := EncodeSlackProviderPayload(payload)
		require.ErrorContains(t, err, "identity is invalid")
	})

	t.Run("story URL cannot claim request identity", func(t *testing.T) {
		receipt, err := BuildSlackStoryCreationReceipt("Joseph", SlackStoryWorkObjectInput{
			AccessGranted: true,
			StoryURL:      "https://acme.fortyone.app/work/WEB-123",
			Title:         "Fix workspace login",
		})
		require.NoError(t, err)
		receipt.ProviderPayload.Metadata.Entities[0].ExternalRef.Type = slackRequestExternalRefType

		_, err = EncodeSlackProviderPayload(receipt.ProviderPayload)
		require.ErrorContains(t, err, "identity is invalid")
	})

	t.Run("unknown external reference type is rejected", func(t *testing.T) {
		payload := requestProviderPayloadForTest(t)
		payload.Metadata.Entities[0].ExternalRef.Type = "issue"

		_, err := EncodeSlackProviderPayload(payload)
		require.ErrorContains(t, err, "identity is invalid")
	})
}

func TestSlackProviderPayloadRejectsMismatchedRequestIdentity(t *testing.T) {
	t.Parallel()

	t.Run("external reference request ID differs", func(t *testing.T) {
		payload := requestProviderPayloadForTest(t)
		payload.Metadata.Entities[0].ExternalRef.ID =
			"acme:11111111-1111-4111-8111-111111111111:33333333-3333-4333-8333-333333333333"

		_, err := EncodeSlackProviderPayload(payload)
		require.ErrorContains(t, err, "identity is invalid")
	})

	t.Run("canonical URL team ID differs", func(t *testing.T) {
		payload := requestProviderPayloadForTest(t)
		payload.Metadata.Entities[0].URL =
			"https://acme.fortyone.app/teams/33333333-3333-4333-8333-333333333333/requests/22222222-2222-4222-8222-222222222222"

		_, err := EncodeSlackProviderPayload(payload)
		require.ErrorContains(t, err, "identity is invalid")
	})

	t.Run("app unfurl URL points at another request", func(t *testing.T) {
		payload := requestProviderPayloadForTest(t)
		payload.Metadata.Entities[0].AppUnfurlURL =
			"https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/33333333-3333-4333-8333-333333333333"

		_, err := EncodeSlackProviderPayload(payload)
		require.ErrorContains(t, err, "unfurl URL is invalid")
	})
}

func requestProviderPayloadForTest(t *testing.T) SlackProviderPayload {
	t.Helper()
	receipt, err := BuildSlackRequestCreationReceipt("Joseph", SlackRequestWorkObjectInput{
		AccessGranted: true,
		RequestURL:    "https://acme.fortyone.app/teams/11111111-1111-4111-8111-111111111111/requests/22222222-2222-4222-8222-222222222222",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	return receipt.ProviderPayload
}
