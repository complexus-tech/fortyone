package slack

import (
	"errors"
	"time"
)

func buildSlackObjectiveWorkObject(input SlackObjectiveWorkObjectInput, includeAppUnfurlURL, compact bool) (SlackWorkObjectEntity, FortyOneObjectiveLink, error) {
	if !input.AccessGranted {
		return SlackWorkObjectEntity{}, FortyOneObjectiveLink{}, ErrSlackObjectivePreviewAccessDenied
	}
	link, err := ParseFortyOneObjectiveURL(input.ObjectiveURL)
	if err != nil {
		return SlackWorkObjectEntity{}, FortyOneObjectiveLink{}, err
	}
	title := truncateSlackWorkObjectText(input.Title, slackWorkObjectTitleLimit)
	if title == "" {
		return SlackWorkObjectEntity{}, FortyOneObjectiveLink{}, errors.New("slack objective Work Object title is required")
	}

	fields := make(map[string]SlackWorkObjectField, 4)
	customFields := make([]SlackWorkObjectCustomField, 0, 2)
	displayOrder := make([]string, 0, 6)
	if !compact {
		if description := truncateSlackWorkObjectText(slackWorkObjectDescription(input.Description), slackWorkObjectTextFieldLimit); description != "" {
			fields["description"] = SlackWorkObjectField{Value: description, Format: "markdown"}
			displayOrder = append(displayOrder, "description")
		}
	}
	if health := truncateSlackWorkObjectText(input.Health, 255); health != "" {
		fields["status"] = SlackWorkObjectField{Label: "Health", Value: health}
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
	if lead := slackWorkObjectUser(input.LeadSlackUserID, input.LeadName); lead != nil {
		fields["assignee"] = SlackWorkObjectField{Label: "Lead", Type: slackUserFieldType, User: lead}
		displayOrder = append(displayOrder, "assignee")
	}
	if appendSlackWorkObjectCustomDateField(&customFields, "start_date", "Start date", input.StartDate) {
		displayOrder = append(displayOrder, "start_date")
	}
	if input.EndDate != nil && !input.EndDate.IsZero() {
		fields["due_date"] = SlackWorkObjectField{
			Label: "End date",
			Value: input.EndDate.UTC().Format(time.DateOnly),
			Type:  slackDateFieldType,
		}
		displayOrder = append(displayOrder, "due_date")
	}

	lastModified := input.UpdatedAt
	if lastModified.IsZero() {
		lastModified = input.CreatedAt
	}
	openAction := SlackWorkObjectAction{
		Text:               "Open in FortyOne",
		ActionID:           slackOpenObjectiveActionID,
		Value:              link.ObjectiveID.String(),
		URL:                link.CanonicalURL,
		AccessibilityLabel: "Open objective in FortyOne",
	}
	entity := SlackWorkObjectEntity{
		URL: link.CanonicalURL,
		ExternalRef: SlackWorkObjectExternalRef{
			ID:   slackObjectiveExternalRefID(link, input.ExternalID),
			Type: slackObjectiveExternalRefType,
		},
		EntityType: slackTaskEntityType,
		EntityPayload: SlackWorkObjectEntityPayload{
			Attributes: SlackWorkObjectAttributes{
				Title:                SlackWorkObjectTitle{Text: title},
				DisplayID:            "Objective",
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
