package slack

import (
	"errors"
	"strings"
	"time"
)

func buildSlackStoryWorkObject(input SlackStoryWorkObjectInput, includeAppUnfurlURL, compact bool) (SlackWorkObjectEntity, FortyOneStoryLink, error) {
	if !input.AccessGranted {
		return SlackWorkObjectEntity{}, FortyOneStoryLink{}, ErrSlackStoryPreviewAccessDenied
	}
	link, err := ParseFortyOneStoryURL(input.StoryURL)
	if err != nil {
		return SlackWorkObjectEntity{}, FortyOneStoryLink{}, err
	}
	title := truncateSlackWorkObjectText(input.Title, slackWorkObjectTitleLimit)
	if title == "" {
		return SlackWorkObjectEntity{}, FortyOneStoryLink{}, errors.New("slack story Work Object title is required")
	}

	fields := make(map[string]SlackWorkObjectField, 8)
	description := truncateSlackWorkObjectText(slackWorkObjectDescription(input.Description), slackWorkObjectTextFieldLimit)
	if !compact && (description != "" || input.Editable) {
		descriptionField := SlackWorkObjectField{Value: description, Format: "markdown"}
		if input.Editable {
			descriptionField.Edit = &SlackWorkObjectEdit{
				Enabled:  true,
				Optional: true,
				Text:     &SlackWorkObjectEditText{MaxLength: modalDescriptionMaxRunes},
			}
		}
		fields["description"] = descriptionField
	}
	if !compact {
		if createdBy := slackWorkObjectUser(input.CreatorSlackUserID, input.CreatorName); createdBy != nil {
			fields["created_by"] = SlackWorkObjectField{Type: slackUserFieldType, User: createdBy}
		}
		if !input.CreatedAt.IsZero() {
			fields["date_created"] = SlackWorkObjectField{Value: input.CreatedAt.UTC().Unix()}
		}
		if !input.UpdatedAt.IsZero() {
			fields["date_updated"] = SlackWorkObjectField{Value: input.UpdatedAt.UTC().Unix()}
		}
	}
	assignee := slackWorkObjectUser(input.AssigneeSlackUserID, input.AssigneeName)
	if assignee != nil || (input.Editable && len(input.AssigneeOptions) > 0) {
		if assignee == nil {
			assignee = &SlackWorkObjectUser{Text: "No assignee"}
		}
		assigneeField := SlackWorkObjectField{Type: slackUserFieldType, User: assignee}
		if input.Editable && validSlackWorkObjectSelectOptions(input.AssigneeOptions) {
			assigneeField.Edit = &SlackWorkObjectEdit{
				Enabled:  true,
				Optional: true,
				Select: &SlackWorkObjectEditSelect{
					CurrentValue:  strings.TrimSpace(input.AssigneeID),
					StaticOptions: cloneSlackWorkObjectSelectOptions(input.AssigneeOptions),
				},
			}
		}
		fields["assignee"] = assigneeField
	}
	if status := truncateSlackWorkObjectText(input.Status, 255); status != "" {
		statusField := SlackWorkObjectField{Value: status, TagColor: normalizeSlackTagColor(input.StatusColor)}
		if input.Editable && strings.TrimSpace(input.StatusID) != "" && validSlackWorkObjectSelectOptions(input.StatusOptions) {
			statusField.Edit = &SlackWorkObjectEdit{
				Enabled: true,
				Select: &SlackWorkObjectEditSelect{
					CurrentValue:  strings.TrimSpace(input.StatusID),
					StaticOptions: cloneSlackWorkObjectSelectOptions(input.StatusOptions),
				},
			}
		}
		fields["status"] = statusField
	}
	if input.DueDate != nil && !input.DueDate.IsZero() {
		fields["due_date"] = SlackWorkObjectField{
			Value: input.DueDate.UTC().Format(time.DateOnly),
			Type:  slackDateFieldType,
			Edit:  slackOptionalWorkObjectEdit(input.Editable),
		}
	} else if input.Editable {
		fields["due_date"] = SlackWorkObjectField{
			Type: slackDateFieldType,
			Edit: slackOptionalWorkObjectEdit(true),
		}
	}
	if priority := truncateSlackWorkObjectText(input.Priority, 255); priority != "" {
		priorityField := SlackWorkObjectField{Value: priority}
		if input.Editable {
			priorityField.Edit = &SlackWorkObjectEdit{
				Enabled: true,
				Select: &SlackWorkObjectEditSelect{
					CurrentValue:  priority,
					StaticOptions: slackWorkObjectPriorityOptions(),
				},
			}
		}
		fields["priority"] = priorityField
	}

	lastModified := input.UpdatedAt
	if lastModified.IsZero() {
		lastModified = input.CreatedAt
	}
	openAction := SlackWorkObjectAction{
		Text:               "Open in FortyOne",
		ActionID:           slackOpenStoryActionID,
		Value:              link.StoryReference,
		URL:                link.CanonicalURL,
		AccessibilityLabel: "Open " + link.StoryReference + " in FortyOne",
	}
	primaryActions := []SlackWorkObjectAction{openAction}
	overflowActions := []SlackWorkObjectAction(nil)
	if includeAppUnfurlURL {
		primaryActions = []SlackWorkObjectAction{
			{
				Text:               "Edit status",
				ActionID:           slackEditStoryStatusActionID,
				Value:              link.StoryReference,
				AccessibilityLabel: "Edit the status of " + link.StoryReference,
			},
			{
				Text:               "Edit priority",
				ActionID:           slackEditStoryPriorityActionID,
				Value:              link.StoryReference,
				AccessibilityLabel: "Edit the priority of " + link.StoryReference,
			},
		}
		overflowActions = []SlackWorkObjectAction{openAction}
	}
	entity := SlackWorkObjectEntity{
		URL: link.CanonicalURL,
		ExternalRef: SlackWorkObjectExternalRef{
			ID:   slackStoryExternalRefID(link, input.ExternalID),
			Type: slackStoryExternalRefType,
		},
		EntityType: slackTaskEntityType,
		EntityPayload: SlackWorkObjectEntityPayload{
			Attributes: SlackWorkObjectAttributes{
				Title: SlackWorkObjectTitle{
					Text: title,
					Edit: slackStoryTitleEdit(input.Editable),
				},
				DisplayID:            link.StoryReference,
				MetadataLastModified: unixTimestamp(lastModified),
			},
			Fields: fields,
			Actions: &SlackWorkObjectActions{
				PrimaryActions:  primaryActions,
				OverflowActions: overflowActions,
			},
		},
	}
	if includeAppUnfurlURL {
		entity.AppUnfurlURL = link.PostedURL
	}
	return entity, link, nil
}
