package slack

import (
	"fmt"
)

// BuildSlackStoryCreationReceipt builds a Work Object notification while
// preserving the intentionally minimal top line: "Joseph created WEB-123".
func BuildSlackStoryCreationReceipt(creatorName string, input SlackStoryWorkObjectInput) (SlackStoryCreationReceipt, error) {
	entity, link, err := buildSlackStoryWorkObject(input, false, true)
	if err != nil {
		return SlackStoryCreationReceipt{}, err
	}
	disableUnfurls := false
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackStoryCreationReceipt{
		Text: fmt.Sprintf("%s created <%s|%s>", slackWorkObjectCreatorLabel(creatorName), link.CanonicalURL, link.StoryReference),
		ProviderPayload: SlackProviderPayload{
			Metadata:    &metadata,
			UnfurlLinks: &disableUnfurls,
			UnfurlMedia: &disableUnfurls,
		},
	}, nil
}

// BuildSlackRequestCreationReceipt builds a read-only Work Object receipt while
// keeping the request-opening phrase itself as the canonical link.
func BuildSlackRequestCreationReceipt(creatorName string, input SlackRequestWorkObjectInput) (SlackRequestCreationReceipt, error) {
	entity, link, err := buildSlackRequestWorkObject(input, false)
	if err != nil {
		return SlackRequestCreationReceipt{}, err
	}
	disableUnfurls := false
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackRequestCreationReceipt{
		Text: fmt.Sprintf("%s <%s|opened a request>", slackWorkObjectCreatorLabel(creatorName), link.CanonicalURL),
		ProviderPayload: SlackProviderPayload{
			Metadata:    &metadata,
			UnfurlLinks: &disableUnfurls,
			UnfurlMedia: &disableUnfurls,
		},
	}, nil
}
