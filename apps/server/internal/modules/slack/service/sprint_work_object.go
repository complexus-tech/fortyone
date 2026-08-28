package slack

import (
	"errors"
)

func buildSlackSprintWorkObject(input SlackSprintWorkObjectInput, includeAppUnfurlURL, compact bool) (SlackWorkObjectEntity, FortyOneSprintLink, error) {
	if !input.AccessGranted {
		return SlackWorkObjectEntity{}, FortyOneSprintLink{}, ErrSlackSprintPreviewAccessDenied
	}
	link, err := ParseFortyOneSprintURL(input.SprintURL)
	if err != nil {
		return SlackWorkObjectEntity{}, FortyOneSprintLink{}, err
	}
	title := truncateSlackWorkObjectText(input.Title, slackWorkObjectTitleLimit)
	if title == "" {
		return SlackWorkObjectEntity{}, FortyOneSprintLink{}, errors.New("slack sprint Work Object title is required")
	}

	fields := make(map[string]SlackWorkObjectField, 2)
	customFields := make([]SlackWorkObjectCustomField, 0, 3)
	displayOrder := make([]string, 0, 4)
	if !compact {
		if goal := truncateSlackWorkObjectText(slackWorkObjectDescription(input.Goal), slackWorkObjectTextFieldLimit); goal != "" {
			fields["goal"] = SlackWorkObjectField{Value: goal, Format: "markdown"}
		}
	}
	if status := truncateSlackWorkObjectText(input.Status, 255); status != "" {
		fields["status"] = SlackWorkObjectField{Value: status}
		displayOrder = append(displayOrder, "status")
	}
	if progress := truncateSlackWorkObjectText(input.Progress, 255); progress != "" {
		customFields = append(customFields, SlackWorkObjectCustomField{
			Key:   "progress",
			Label: "Progress",
			Value: progress,
			Type:  "string",
		})
		displayOrder = append(displayOrder, "progress")
	}
	if appendSlackWorkObjectCustomDateField(&customFields, "start_date", "Start date", input.StartDate) {
		displayOrder = append(displayOrder, "start_date")
	}
	if appendSlackWorkObjectCustomDateField(&customFields, "end_date", "End date", input.EndDate) {
		displayOrder = append(displayOrder, "end_date")
	}

	lastModified := input.UpdatedAt
	if lastModified.IsZero() {
		lastModified = input.CreatedAt
	}
	openAction := SlackWorkObjectAction{
		Text:               "Open in FortyOne",
		ActionID:           slackOpenSprintActionID,
		Value:              link.SprintID.String(),
		URL:                link.CanonicalURL,
		AccessibilityLabel: "Open sprint in FortyOne",
	}
	entity := SlackWorkObjectEntity{
		URL: link.CanonicalURL,
		ExternalRef: SlackWorkObjectExternalRef{
			ID:   slackSprintExternalRefID(link, input.ExternalID),
			Type: slackSprintExternalRefType,
		},
		EntityType: slackTaskEntityType,
		EntityPayload: SlackWorkObjectEntityPayload{
			Attributes: SlackWorkObjectAttributes{
				Title:                SlackWorkObjectTitle{Text: title},
				DisplayID:            "Sprint",
				MetadataLastModified: unixTimestamp(lastModified),
			},
			Fields:       fields,
			CustomFields: customFields,
			DisplayOrder: displayOrder,
			Actions:      &SlackWorkObjectActions{PrimaryActions: []SlackWorkObjectAction{openAction}},
		},
	}
	if includeAppUnfurlURL {
		entity.AppUnfurlURL = link.PostedURL
	}
	return entity, link, nil
}
