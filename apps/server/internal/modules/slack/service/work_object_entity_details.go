package slack

import (
	"errors"
	"strings"
)

func BuildSlackStoryEntityDetailsRequest(triggerID string, input SlackStoryWorkObjectInput) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("slack entity details trigger is required")
	}
	entity, _, err := buildSlackStoryWorkObject(input, false, false)
	if err != nil {
		return SlackEntityDetailsRequest{}, err
	}
	return SlackEntityDetailsRequest{TriggerID: triggerID, Metadata: &entity}, nil
}

// BuildSlackRequestEntityDetailsRequest presents the same read-only request
// entity without app_unfurl_url, as required by entity.presentDetails.
func BuildSlackRequestEntityDetailsRequest(triggerID string, input SlackRequestWorkObjectInput) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("slack entity details trigger is required")
	}
	entity, _, err := buildSlackRequestWorkObject(input, false)
	if err != nil {
		return SlackEntityDetailsRequest{}, err
	}
	return SlackEntityDetailsRequest{TriggerID: triggerID, Metadata: &entity}, nil
}

func BuildSlackObjectiveEntityDetailsRequest(triggerID string, input SlackObjectiveWorkObjectInput) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("slack entity details trigger is required")
	}
	entity, _, err := buildSlackObjectiveWorkObject(input, false, false)
	if err != nil {
		return SlackEntityDetailsRequest{}, err
	}
	return SlackEntityDetailsRequest{TriggerID: triggerID, Metadata: &entity}, nil
}

func BuildSlackSprintEntityDetailsRequest(triggerID string, input SlackSprintWorkObjectInput) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("slack entity details trigger is required")
	}
	entity, _, err := buildSlackSprintWorkObject(input, false, false)
	if err != nil {
		return SlackEntityDetailsRequest{}, err
	}
	return SlackEntityDetailsRequest{TriggerID: triggerID, Metadata: &entity}, nil
}

func BuildSlackStoryAuthenticationEntityDetailsRequest(triggerID, authURL string) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("slack entity details trigger is required")
	}
	if !isSafeFortyOneHTTPSURL(authURL) {
		return SlackEntityDetailsRequest{}, errors.New("invalid FortyOne Slack account-link URL")
	}
	return SlackEntityDetailsRequest{
		TriggerID:        triggerID,
		UserAuthRequired: true,
		UserAuthURL:      strings.TrimSpace(authURL),
	}, nil
}

func BuildSlackStoryEntityDetailsErrorRequest(triggerID, message string) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("slack entity details trigger is required")
	}
	message = truncateSlackWorkObjectText(message, 500)
	if message == "" {
		message = "FortyOne could not save these changes. Refresh the task and try again."
	}
	return SlackEntityDetailsRequest{
		TriggerID: triggerID,
		Error: &SlackEntityDetailsError{
			Status:        "edit_error",
			CustomMessage: message,
		},
	}, nil
}
