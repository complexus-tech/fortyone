package slack

import (
	"encoding/json"
	"strings"
)

type slackModalPrivateMetadata struct {
	Source         requestSourceContext `json:"source"`
	SelectedTeamID string               `json:"selected_team_id,omitempty"`
}

func parseSlackModalPrivateMetadata(raw string) (slackModalPrivateMetadata, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return slackModalPrivateMetadata{}, nil
	}

	var structured slackModalPrivateMetadata
	if err := json.Unmarshal([]byte(trimmed), &structured); err == nil && !isZeroRequestSource(structured.Source) {
		return structured, nil
	}

	var legacy requestSourceContext
	if err := json.Unmarshal([]byte(trimmed), &legacy); err != nil {
		return slackModalPrivateMetadata{}, err
	}
	return slackModalPrivateMetadata{Source: legacy}, nil
}

type interactionPayload struct {
	Type        string `json:"type"`
	TriggerID   string `json:"trigger_id"`
	ResponseURL string `json:"response_url"`
	ActionID    string `json:"action_id"`
	BlockID     string `json:"block_id"`
	Value       string `json:"value"`
	Team        struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	User struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"user"`
	Channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	Message struct {
		Text     string `json:"text"`
		TS       string `json:"ts"`
		ThreadTS string `json:"thread_ts"`
		User     string `json:"user"`
	} `json:"message"`
	Container struct {
		MessageTS    string `json:"message_ts"`
		ChannelID    string `json:"channel_id"`
		AppUnfurlURL string `json:"app_unfurl_url"`
	} `json:"container"`
	AppUnfurl struct {
		AppUnfurlURL string `json:"app_unfurl_url"`
	} `json:"app_unfurl"`
	View struct {
		ID              string                     `json:"id"`
		Hash            string                     `json:"hash"`
		Type            string                     `json:"type"`
		CallbackID      string                     `json:"callback_id"`
		PrivateMetadata string                     `json:"private_metadata"`
		EntityURL       string                     `json:"entity_url"`
		AppUnfurlURL    string                     `json:"app_unfurl_url"`
		Channel         string                     `json:"channel"`
		MessageTS       string                     `json:"message_ts"`
		ThreadTS        string                     `json:"thread_ts"`
		ExternalRef     SlackWorkObjectExternalRef `json:"external_ref"`
		Blocks          []struct {
			BlockID string `json:"block_id"`
			Element struct {
				Type          string `json:"type"`
				ActionID      string `json:"action_id"`
				InitialOption struct {
					Value string `json:"value"`
				} `json:"initial_option"`
			} `json:"element"`
		} `json:"blocks"`
		State struct {
			Values interactionViewStateValues `json:"values"`
		} `json:"state"`
	} `json:"view"`
	Actions []struct {
		ActionID       string `json:"action_id"`
		BlockID        string `json:"block_id"`
		Type           string `json:"type"`
		Value          string `json:"value"`
		SelectedOption struct {
			Value string `json:"value"`
		} `json:"selected_option"`
		SelectedOptions []struct {
			Value string `json:"value"`
		} `json:"selected_options"`
	} `json:"actions"`
}
