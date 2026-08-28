package slack

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// BuildSlackMutationConfirmationProviderPayload returns a generic Block Kit
// confirmation payload suitable for the same durable provider_payload column
// used by rich story receipts. The opaque token is never rendered as text.
func BuildSlackMutationConfirmationProviderPayload(prompt, confirmationToken, slackUserID string, createAll bool) (SlackProviderPayload, error) {
	confirmLabel := "Confirm"
	confirmAccessibilityLabel := "Confirm story change"
	if createAll {
		confirmLabel = "Create all"
		confirmAccessibilityLabel = "Create all proposed stories"
	}
	return buildSlackMutationActionProviderPayload(
		prompt,
		confirmationToken,
		slackUserID,
		confirmLabel,
		confirmAccessibilityLabel,
		true,
	)
}

// BuildSlackMutationRetryProviderPayload returns a retry-only confirmation for
// a partially applied batch. Cancellation is intentionally omitted because the
// original confirmation has already been consumed.
func BuildSlackMutationRetryProviderPayload(prompt, confirmationToken, slackUserID string) (SlackProviderPayload, error) {
	return buildSlackMutationActionProviderPayload(
		prompt,
		confirmationToken,
		slackUserID,
		"Retry remaining",
		"Retry creating the remaining proposed stories",
		false,
	)
}

func buildSlackMutationActionProviderPayload(
	prompt, confirmationToken, slackUserID, confirmLabel, confirmAccessibilityLabel string,
	includeCancel bool,
) (SlackProviderPayload, error) {
	promptBlocks, err := buildSlackMutationPromptBlocks(prompt)
	if err != nil {
		return SlackProviderPayload{}, err
	}
	confirmationToken = strings.TrimSpace(confirmationToken)
	if confirmationToken == "" {
		return SlackProviderPayload{}, errors.New("slack mutation confirmation token is invalid")
	}
	actionValue, err := encodeSlackMutationActionValue(slackUserID, confirmationToken)
	if err != nil {
		return SlackProviderPayload{}, err
	}
	elements := []SlackBlockElement{{
		Type:               "button",
		ActionID:           slackConfirmMutationActionID,
		Text:               &SlackTextObject{Type: "plain_text", Text: confirmLabel},
		Value:              actionValue,
		Style:              "primary",
		AccessibilityLabel: confirmAccessibilityLabel,
	}}
	if includeCancel {
		elements = append(elements, SlackBlockElement{
			Type:               "button",
			ActionID:           slackCancelMutationActionID,
			Text:               &SlackTextObject{Type: "plain_text", Text: "Cancel"},
			Value:              actionValue,
			AccessibilityLabel: "Cancel story change",
		})
	}
	blocks := append(promptBlocks, SlackBlock{
		Type:     "actions",
		BlockID:  "fortyone_story_mutation_confirmation",
		Elements: elements,
	})
	return SlackProviderPayload{Blocks: blocks}, nil
}

func buildSlackMutationPromptBlocks(prompt string) ([]SlackBlock, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, errors.New("slack mutation confirmation prompt is required")
	}

	const maximumPromptBlocks = 49 // Reserve one block for mutation actions.
	blocks := make([]SlackBlock, 0, 2)
	var section strings.Builder
	sectionRunes := 0
	flush := func() error {
		text := strings.TrimSpace(section.String())
		if text == "" {
			return nil
		}
		if len(blocks) >= maximumPromptBlocks {
			return errors.New("slack mutation confirmation exceeds the 50-block message limit")
		}
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Text: &SlackTextObject{Type: "mrkdwn", Text: text},
		})
		section.Reset()
		sectionRunes = 0
		return nil
	}

	for _, line := range strings.Split(prompt, "\n") {
		lineRunes := utf8.RuneCountInString(line)
		if lineRunes > slackWorkObjectTextFieldLimit {
			return nil, errors.New("slack mutation confirmation contains a line that exceeds the section text limit")
		}
		separatorRunes := 0
		if section.Len() > 0 {
			separatorRunes = 1
		}
		if sectionRunes+separatorRunes+lineRunes > slackWorkObjectTextFieldLimit {
			if err := flush(); err != nil {
				return nil, err
			}
		}
		if section.Len() > 0 {
			section.WriteByte('\n')
			sectionRunes++
		}
		section.WriteString(line)
		sectionRunes += lineRunes
	}
	if err := flush(); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, errors.New("slack mutation confirmation prompt is required")
	}
	return blocks, nil
}

type slackMutationActionValue struct {
	SlackUserID string `json:"slack_user_id"`
	Token       string `json:"token"`
}

func encodeSlackMutationActionValue(slackUserID, token string) (string, error) {
	value := slackMutationActionValue{
		SlackUserID: strings.TrimSpace(slackUserID),
		Token:       strings.TrimSpace(token),
	}
	if value.SlackUserID == "" || value.Token == "" {
		return "", errors.New("slack mutation action actor and token are required")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode Slack mutation action: %w", err)
	}
	result := base64.RawURLEncoding.EncodeToString(encoded)
	if utf8.RuneCountInString(result) > slackButtonValueLimit {
		return "", errors.New("slack mutation confirmation token is invalid")
	}
	return result, nil
}

func decodeSlackMutationActionValue(raw string) (slackMutationActionValue, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return slackMutationActionValue{}, errors.New("invalid Slack mutation action")
	}
	var value slackMutationActionValue
	if err := json.Unmarshal(decoded, &value); err != nil {
		return slackMutationActionValue{}, errors.New("invalid Slack mutation action")
	}
	value.SlackUserID = strings.TrimSpace(value.SlackUserID)
	value.Token = strings.TrimSpace(value.Token)
	if value.SlackUserID == "" || value.Token == "" {
		return slackMutationActionValue{}, errors.New("invalid Slack mutation action")
	}
	return value, nil
}
