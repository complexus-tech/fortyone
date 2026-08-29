package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

// EncodeSlackProviderPayload serializes the Slack-only part of a durable
// outbound delivery after validating its Block Kit and Work Object envelope.
func EncodeSlackProviderPayload(payload SlackProviderPayload) ([]byte, error) {
	if err := validateSlackProviderPayload(payload); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Slack provider payload: %w", err)
	}
	return encoded, nil
}

// DecodeSlackProviderPayload restores the exact payload used by a first
// delivery so an outbox retry does not silently lose buttons or Work Objects.
func DecodeSlackProviderPayload(raw []byte) (SlackProviderPayload, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return SlackProviderPayload{}, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var payload SlackProviderPayload
	if err := decoder.Decode(&payload); err != nil {
		return SlackProviderPayload{}, fmt.Errorf("decode Slack provider payload: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return SlackProviderPayload{}, fmt.Errorf("decode Slack provider payload trailing content: %w", err)
		}
		return SlackProviderPayload{}, errors.New("decode Slack provider payload: multiple JSON values")
	}
	if err := validateSlackProviderPayload(payload); err != nil {
		return SlackProviderPayload{}, err
	}
	return payload, nil
}

func validateSlackProviderPayload(payload SlackProviderPayload) error {
	if authorSlackUserID := strings.TrimSpace(payload.AuthorSlackUserID); authorSlackUserID != "" && !validSlackUserID(authorSlackUserID) {
		return errors.New("slack provider payload contains an invalid author user ID")
	}
	if payload.RequestThreadBinding != nil && payload.RequestThreadBinding.IntegrationRequestID == uuid.Nil {
		return errors.New("slack request thread binding requires an integration request ID")
	}
	if payload.Authorization != nil {
		if payload.Authorization.ActorUserID != nil && *payload.Authorization.ActorUserID == uuid.Nil {
			return errors.New("slack delivery authorization contains an invalid actor ID")
		}
		if len(payload.Authorization.AllowedTeamIDs) == 0 || len(payload.Authorization.AllowedTeamIDs) > 1000 {
			return errors.New("slack delivery authorization requires between 1 and 1000 team IDs")
		}
		seen := make(map[string]struct{}, len(payload.Authorization.AllowedTeamIDs))
		for _, teamID := range payload.Authorization.AllowedTeamIDs {
			if teamID == uuid.Nil {
				return errors.New("slack delivery authorization contains an invalid team ID")
			}
			key := teamID.String()
			if _, exists := seen[key]; exists {
				return errors.New("slack delivery authorization contains duplicate team IDs")
			}
			seen[key] = struct{}{}
		}
		if len(payload.Authorization.SharedTeamIDs) > 1000 {
			return errors.New("slack delivery authorization exceeds 1000 shared team IDs")
		}
		shared := make(map[string]struct{}, len(payload.Authorization.SharedTeamIDs))
		for _, teamID := range payload.Authorization.SharedTeamIDs {
			if teamID == uuid.Nil {
				return errors.New("slack delivery authorization contains an invalid shared team ID")
			}
			key := teamID.String()
			if _, exists := shared[key]; exists {
				return errors.New("slack delivery authorization contains duplicate shared team IDs")
			}
			if _, allowed := seen[key]; !allowed {
				return errors.New("slack delivery authorization contains a shared team outside its allowed teams")
			}
			shared[key] = struct{}{}
		}
	}
	if len(payload.Blocks) > 50 {
		return errors.New("slack provider payload exceeds the 50-block message limit")
	}
	for _, block := range payload.Blocks {
		switch block.Type {
		case "section":
			if block.Text == nil || strings.TrimSpace(block.Text.Text) == "" {
				return errors.New("slack section block text is required")
			}
		case "actions":
			if len(block.Elements) == 0 || len(block.Elements) > 25 {
				return errors.New("slack actions block must contain between 1 and 25 elements")
			}
		default:
			return fmt.Errorf("unsupported durable Slack block type %q", block.Type)
		}
		for _, element := range block.Elements {
			if element.Type != "button" || strings.TrimSpace(element.ActionID) == "" || element.Text == nil || strings.TrimSpace(element.Text.Text) == "" {
				return errors.New("durable Slack actions support only identified buttons with text")
			}
			if len([]rune(element.Value)) > slackButtonValueLimit {
				return errors.New("slack button value exceeds the provider limit")
			}
		}
	}
	if payload.Metadata != nil {
		if len(payload.Metadata.Entities) == 0 {
			return errors.New("slack Work Object metadata requires at least one entity")
		}
		for _, entity := range payload.Metadata.Entities {
			if entity.EntityType != slackTaskEntityType || strings.TrimSpace(entity.EntityPayload.Attributes.Title.Text) == "" {
				return errors.New("slack provider payload contains an invalid task Work Object")
			}
			if err := validateSlackProviderWorkObjectIdentity(entity); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSlackProviderWorkObjectIdentity(entity SlackWorkObjectEntity) error {
	switch entity.ExternalRef.Type {
	case slackStoryExternalRefType:
		link, err := ParseFortyOneStoryURL(entity.URL)
		if err != nil || !validSlackStoryExternalRef(link, entity.ExternalRef.ID) {
			return errors.New("slack provider payload Work Object identity is invalid")
		}
		if entity.AppUnfurlURL != "" {
			postedLink, err := ParseFortyOneStoryURL(entity.AppUnfurlURL)
			if err != nil || postedLink.CanonicalURL != entity.URL {
				return errors.New("slack provider payload Work Object unfurl URL is invalid")
			}
		}
	case slackRequestExternalRefType:
		link, err := ParseFortyOneRequestURL(entity.URL)
		if err != nil || link.CanonicalURL != entity.URL || !validSlackRequestExternalRef(link, entity.ExternalRef.ID) {
			return errors.New("slack provider payload Work Object identity is invalid")
		}
		if entity.AppUnfurlURL != "" {
			postedLink, err := ParseFortyOneRequestURL(entity.AppUnfurlURL)
			if err != nil || postedLink.CanonicalURL != entity.URL {
				return errors.New("slack provider payload Work Object unfurl URL is invalid")
			}
		}
	case slackObjectiveExternalRefType:
		link, err := ParseFortyOneObjectiveURL(entity.URL)
		if err != nil || link.CanonicalURL != entity.URL || !validSlackObjectiveExternalRef(link, entity.ExternalRef.ID) {
			return errors.New("slack provider payload Work Object identity is invalid")
		}
		if entity.AppUnfurlURL != "" {
			postedLink, err := ParseFortyOneObjectiveURL(entity.AppUnfurlURL)
			if err != nil || postedLink.CanonicalURL != entity.URL {
				return errors.New("slack provider payload Work Object unfurl URL is invalid")
			}
		}
	case slackSprintExternalRefType:
		link, err := ParseFortyOneSprintURL(entity.URL)
		if err != nil || link.CanonicalURL != entity.URL || !validSlackSprintExternalRef(link, entity.ExternalRef.ID) {
			return errors.New("slack provider payload Work Object identity is invalid")
		}
		if entity.AppUnfurlURL != "" {
			postedLink, err := ParseFortyOneSprintURL(entity.AppUnfurlURL)
			if err != nil || postedLink.CanonicalURL != entity.URL {
				return errors.New("slack provider payload Work Object unfurl URL is invalid")
			}
		}
	default:
		return errors.New("slack provider payload Work Object identity is invalid")
	}
	return nil
}

func validSlackStoryExternalRef(link FortyOneStoryLink, externalRefID string) bool {
	externalRefID = strings.TrimSpace(externalRefID)
	if externalRefID == slackStoryExternalRefID(link, "") {
		return true
	}
	prefix := strings.ToLower(link.WorkspaceSlug) + ":"
	if !strings.HasPrefix(externalRefID, prefix) {
		return false
	}
	_, err := uuid.Parse(strings.TrimPrefix(externalRefID, prefix))
	return err == nil
}
