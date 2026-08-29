package slack

import (
	"errors"
	"time"
)

func buildSlackRequestWorkObject(input SlackRequestWorkObjectInput, includeAppUnfurlURL bool) (SlackWorkObjectEntity, FortyOneRequestLink, error) {
	if !input.AccessGranted {
		return SlackWorkObjectEntity{}, FortyOneRequestLink{}, ErrSlackRequestPreviewAccessDenied
	}
	link, err := ParseFortyOneRequestURL(input.RequestURL)
	if err != nil {
		return SlackWorkObjectEntity{}, FortyOneRequestLink{}, err
	}
	title := truncateSlackWorkObjectText(input.Title, slackWorkObjectTitleLimit)
	if title == "" {
		return SlackWorkObjectEntity{}, FortyOneRequestLink{}, errors.New("slack request Work Object title is required")
	}

	fields := make(map[string]SlackWorkObjectField, 8)
	if description := truncateSlackWorkObjectText(slackWorkObjectDescription(input.Description), slackWorkObjectTextFieldLimit); description != "" {
		fields["description"] = SlackWorkObjectField{Value: description, Format: "markdown"}
	}
	if createdBy := slackWorkObjectUser(input.CreatorSlackUserID, input.CreatorName); createdBy != nil {
		fields["created_by"] = SlackWorkObjectField{Type: slackUserFieldType, User: createdBy}
	}
	if !input.CreatedAt.IsZero() {
		fields["date_created"] = SlackWorkObjectField{Value: input.CreatedAt.UTC().Unix()}
	}
	if !input.UpdatedAt.IsZero() {
		fields["date_updated"] = SlackWorkObjectField{Value: input.UpdatedAt.UTC().Unix()}
	}
	if assignee := slackWorkObjectUser(input.AssigneeSlackUserID, input.AssigneeName); assignee != nil {
		fields["assignee"] = SlackWorkObjectField{Type: slackUserFieldType, User: assignee}
	}
	if status := truncateSlackWorkObjectText(input.Status, 255); status != "" {
		fields["status"] = SlackWorkObjectField{Value: status}
	}
	if input.DueDate != nil && !input.DueDate.IsZero() {
		fields["due_date"] = SlackWorkObjectField{
			Value: input.DueDate.UTC().Format(time.DateOnly),
			Type:  slackDateFieldType,
		}
	}
	if priority := truncateSlackWorkObjectText(input.Priority, 255); priority != "" {
		fields["priority"] = SlackWorkObjectField{Value: priority}
	}

	lastModified := input.UpdatedAt
	if lastModified.IsZero() {
		lastModified = input.CreatedAt
	}
	entity := SlackWorkObjectEntity{
		URL: link.CanonicalURL,
		ExternalRef: SlackWorkObjectExternalRef{
			ID:   slackRequestExternalRefID(link),
			Type: slackRequestExternalRefType,
		},
		EntityType: slackTaskEntityType,
		EntityPayload: SlackWorkObjectEntityPayload{
			Attributes: SlackWorkObjectAttributes{
				Title:                SlackWorkObjectTitle{Text: title},
				MetadataLastModified: unixTimestamp(lastModified),
			},
			Fields: fields,
			Actions: &SlackWorkObjectActions{
				PrimaryActions: []SlackWorkObjectAction{{
					Text:               "Open in FortyOne",
					ActionID:           slackOpenRequestActionID,
					Value:              link.RequestID.String(),
					URL:                link.CanonicalURL,
					AccessibilityLabel: "Open request in FortyOne",
				}},
			},
		},
	}
	if includeAppUnfurlURL {
		entity.AppUnfurlURL = link.PostedURL
	}
	return entity, link, nil
}
