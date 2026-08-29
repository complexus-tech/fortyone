package slack

import (
	"context"
	"errors"
	"strings"
)

// slackWorkObjectPublisher owns the narrow Slack Web API calls needed by rich
// story previews. Event normalization and authorization remain outside this
// provider boundary.
type slackWorkObjectPublisher struct {
	client *slackWebClient
}

type slackEntityDetailsPresentationError struct {
	cause error
}

func (e *slackEntityDetailsPresentationError) Error() string {
	return e.cause.Error()
}

func (e *slackEntityDetailsPresentationError) Unwrap() error {
	return e.cause
}

func newSlackWorkObjectPublisher(client *slackWebClient) *slackWorkObjectPublisher {
	return &slackWorkObjectPublisher{client: client}
}

func (p *slackWorkObjectPublisher) Unfurl(ctx context.Context, botToken string, request SlackChatUnfurlRequest) error {
	if p == nil || p.client == nil {
		return errors.New("slack Work Object publisher is not configured")
	}
	if strings.TrimSpace(botToken) == "" {
		return errors.New("slack bot token is required")
	}
	if err := validateSlackUnfurlRequestDestination(request); err != nil {
		return err
	}
	if request.Metadata != nil {
		if request.UserAuthRequired || request.UserAuthURL != "" || request.UserAuthMessage != "" {
			return errors.New("slack unfurl cannot mix Work Object metadata with an authentication prompt")
		}
		if err := validateSlackProviderPayload(SlackProviderPayload{Metadata: request.Metadata}); err != nil {
			return err
		}
	} else if !request.UserAuthRequired || !isSafeFortyOneHTTPSURL(request.UserAuthURL) {
		return errors.New("slack unfurl requires authorized Work Object metadata or a safe account-link prompt")
	}
	return p.client.callJSON(ctx, botToken, "chat.unfurl", request, nil)
}

func (p *slackWorkObjectPublisher) PresentDetails(ctx context.Context, botToken string, request SlackEntityDetailsRequest) error {
	if p == nil || p.client == nil {
		return errors.New("slack Work Object publisher is not configured")
	}
	if strings.TrimSpace(botToken) == "" || strings.TrimSpace(request.TriggerID) == "" {
		return errors.New("slack bot token and entity details trigger are required")
	}
	if request.Metadata != nil {
		if request.UserAuthRequired || request.UserAuthURL != "" || request.Error != nil {
			return errors.New("slack entity details cannot mix Work Object metadata with authentication or error responses")
		}
		if request.Metadata.AppUnfurlURL != "" {
			return errors.New("slack entity details metadata cannot include app_unfurl_url")
		}
		if err := validateSlackProviderPayload(SlackProviderPayload{
			Metadata: &SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{*request.Metadata}},
		}); err != nil {
			return err
		}
	} else if request.Error != nil {
		if request.UserAuthRequired || request.UserAuthURL != "" || request.Error.Status != "edit_error" || strings.TrimSpace(request.Error.CustomMessage) == "" {
			return errors.New("slack entity details edit error is invalid")
		}
	} else if !request.UserAuthRequired || !isSafeFortyOneHTTPSURL(request.UserAuthURL) {
		return errors.New("slack entity details requires authorized Work Object metadata, an edit error, or a safe account-link prompt")
	}
	if err := p.client.callJSON(ctx, botToken, "entity.presentDetails", request, nil); err != nil {
		// Once the provider call is attempted, the trigger must be treated as
		// spent even if the response is lost. Retrying could reuse a single-use
		// trigger or overwrite a successful flexpane refresh.
		return &slackEntityDetailsPresentationError{cause: err}
	}
	return nil
}

func slackEntityDetailsPresentationWasAttempted(err error) bool {
	var presentationErr *slackEntityDetailsPresentationError
	return errors.As(err, &presentationErr)
}

func isSlackEntityDetailsTerminalError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	code, ok := SlackAPIErrorCode(err)
	if !ok {
		return false
	}
	switch strings.TrimSpace(code) {
	case "invalid_trigger_id", "trigger_expired", "trigger_exchanged":
		return true
	default:
		return false
	}
}

func (p *slackWorkObjectPublisher) PostCreationReceipt(
	ctx context.Context,
	botToken, channelID, threadTS, clientMessageID string,
	receipt SlackStoryCreationReceipt,
) (string, error) {
	if p == nil || p.client == nil {
		return "", errors.New("slack Work Object publisher is not configured")
	}
	if strings.TrimSpace(botToken) == "" || strings.TrimSpace(channelID) == "" || strings.TrimSpace(receipt.Text) == "" {
		return "", errors.New("slack bot token, channel, and receipt text are required")
	}
	if _, err := EncodeSlackProviderPayload(receipt.ProviderPayload); err != nil {
		return "", err
	}
	payload := map[string]any{
		"channel": strings.TrimSpace(channelID),
		"text":    strings.TrimSpace(receipt.Text),
	}
	applySlackProviderPayload(payload, receipt.ProviderPayload)
	if threadTS = strings.TrimSpace(threadTS); threadTS != "" {
		payload["thread_ts"] = threadTS
	}
	if clientMessageID = strings.TrimSpace(clientMessageID); clientMessageID != "" {
		payload["client_msg_id"] = clientMessageID
	}
	var response struct {
		TS string `json:"ts"`
	}
	if err := p.client.callJSON(ctx, botToken, "chat.postMessage", payload, &response); err != nil {
		return "", err
	}
	return strings.TrimSpace(response.TS), nil
}

func applySlackProviderPayload(destination map[string]any, payload SlackProviderPayload) {
	if len(payload.Blocks) > 0 {
		destination["blocks"] = payload.Blocks
	}
	if payload.Metadata != nil {
		destination["metadata"] = payload.Metadata
	}
	if payload.UnfurlLinks != nil {
		destination["unfurl_links"] = *payload.UnfurlLinks
	}
	if payload.UnfurlMedia != nil {
		destination["unfurl_media"] = *payload.UnfurlMedia
	}
}
