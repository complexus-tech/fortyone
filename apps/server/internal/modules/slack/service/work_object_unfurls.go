package slack

import (
	"errors"
	"strings"
)

// BuildSlackStoryUnfurlRequest creates a Work Object response for one
// link_shared event after the caller has proven the actor can read the story.
func BuildSlackStoryUnfurlRequest(channelID, messageTS string, input SlackStoryWorkObjectInput) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	entity, _, err := buildSlackStoryWorkObject(input, true, true)
	if err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackChatUnfurlRequest{
		Channel:  strings.TrimSpace(channelID),
		TS:       strings.TrimSpace(messageTS),
		Metadata: &metadata,
	}, nil
}

// BuildSlackRequestUnfurlRequest creates a read-only Work Object response for
// one request link after the caller has proven the actor can read its team.
func BuildSlackRequestUnfurlRequest(channelID, messageTS string, input SlackRequestWorkObjectInput) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	entity, _, err := buildSlackRequestWorkObject(input, true)
	if err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackChatUnfurlRequest{
		Channel:  strings.TrimSpace(channelID),
		TS:       strings.TrimSpace(messageTS),
		Metadata: &metadata,
	}, nil
}

// BuildSlackObjectiveUnfurlRequest creates a read-only Work Object response
// for one objective link after the caller has proven the actor can read it.
func BuildSlackObjectiveUnfurlRequest(channelID, messageTS string, input SlackObjectiveWorkObjectInput) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	entity, _, err := buildSlackObjectiveWorkObject(input, true, true)
	if err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackChatUnfurlRequest{Channel: strings.TrimSpace(channelID), TS: strings.TrimSpace(messageTS), Metadata: &metadata}, nil
}

// BuildSlackSprintUnfurlRequest creates a read-only Work Object response for
// one sprint link after the caller has proven the actor can read it.
func BuildSlackSprintUnfurlRequest(channelID, messageTS string, input SlackSprintWorkObjectInput) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	entity, _, err := buildSlackSprintWorkObject(input, true, true)
	if err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackChatUnfurlRequest{Channel: strings.TrimSpace(channelID), TS: strings.TrimSpace(messageTS), Metadata: &metadata}, nil
}

// BuildSlackStoryAuthenticationUnfurlRequest creates Slack's private account
// linking prompt. Use this only for an unlinked user. A linked user who cannot
// access the story must receive no unfurl at all, avoiding an existence leak.
func BuildSlackStoryAuthenticationUnfurlRequest(channelID, messageTS, authURL string) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	if !isSafeFortyOneHTTPSURL(authURL) {
		return SlackChatUnfurlRequest{}, errors.New("invalid FortyOne Slack account-link URL")
	}
	return SlackChatUnfurlRequest{
		Channel:          strings.TrimSpace(channelID),
		TS:               strings.TrimSpace(messageTS),
		UserAuthRequired: true,
		UserAuthURL:      strings.TrimSpace(authURL),
		UserAuthMessage:  "Connect your FortyOne account to preview this link.",
	}, nil
}
