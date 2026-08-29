package messaging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

func decodeBatchStoryProposal(raw json.RawMessage) (batchStoryMutationProposal, error) {
	if len(raw) == 0 {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: batch proposal is missing", ErrInvalidConfirmation)
	}
	var proposal batchStoryMutationProposal
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proposal); err != nil {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: decode batch proposal: %v", ErrInvalidConfirmation, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: trailing batch proposal data", ErrInvalidConfirmation)
	}
	if proposal.Version != batchStoryProposalVersion || len(proposal.Items) == 0 || len(proposal.Items) > maximumBatchStoryCount {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: unsupported or empty batch proposal", ErrInvalidConfirmation)
	}
	sourceURL, err := normalizedSourceURL(proposal.SourceURL)
	if err != nil || sourceURL != proposal.SourceURL {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: invalid batch source URL", ErrInvalidConfirmation)
	}
	for index, item := range proposal.Items {
		title, titleErr := normalizedStoryTitle(item.Title)
		description, descriptionErr := normalizedBatchStoryDescriptionValue(item.Description)
		priority, priorityErr := normalizedStoryPriority(&item.Priority, "")
		if titleErr != nil || title != item.Title || descriptionErr != nil || description != item.Description || priorityErr != nil || priority != item.Priority {
			return batchStoryMutationProposal{}, fmt.Errorf("%w: invalid batch item %d", ErrInvalidConfirmation, index)
		}
		if item.AssigneeID != nil && *item.AssigneeID == uuid.Nil {
			return batchStoryMutationProposal{}, fmt.Errorf("%w: invalid batch item %d assignee", ErrInvalidConfirmation, index)
		}
	}
	return proposal, nil
}

func normalizedBatchStoryDescription(raw *string) (string, error) {
	if raw == nil {
		return "", nil
	}
	return normalizedBatchStoryDescriptionValue(*raw)
}

func normalizedBatchStoryDescriptionValue(raw string) (string, error) {
	description := strings.TrimSpace(raw)
	if len([]rune(description)) > maximumBatchDescriptionRunes {
		return "", fmt.Errorf("description must not exceed %d characters", maximumBatchDescriptionRunes)
	}
	return description, nil
}

func batchStoryDescription(description, sourceURL string) string {
	description = strings.TrimSpace(description)
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return description
	}
	if description == "" {
		return "Source: " + sourceURL
	}
	return description + "\n\nSource: " + sourceURL
}

func batchStoryConfirmationPrompt(team messagingTeam, previews []StoryMutationPreview) string {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Create %d stories in %s (%s)?", len(previews), confirmationPromptText(team.Name), confirmationPromptText(strings.ToUpper(team.Code)))
	for index, preview := range previews {
		assignee := preview.AssigneeName
		if strings.TrimSpace(assignee) == "" {
			assignee = "Unassigned"
		}
		priority := "No Priority"
		if preview.Priority != nil {
			priority = *preview.Priority
		}
		fmt.Fprintf(&prompt, "\n%d. %s — %s, %s", index+1, confirmationPromptText(preview.Title), confirmationPromptText(assignee), confirmationPromptText(priority))
	}
	if len(previews) > 0 && previews[0].SourceURL != "" {
		prompt.WriteString("\n\nThe supporting descriptions and a link to this Slack thread will be attached to the stories.")
	} else {
		prompt.WriteString("\n\nThe supporting descriptions shown in the draft will be attached to the stories.")
	}
	return prompt.String()
}

func confirmationPromptText(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(value)
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func boolPointer(value bool) *bool {
	return &value
}

func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func optionalBoolValue(value *bool) bool {
	return value != nil && *value
}

func normalizeLegacyUpdateTimeActions(claims *storyMutationClaims) {
	if claims == nil || claims.Operation != StoryMutationUpdate {
		return
	}
	if claims.EstimatedDurationAction == "" {
		claims.EstimatedDurationAction = storyTimeActionUnchanged
		if claims.EstimatedDurationMinutes != nil {
			claims.EstimatedDurationAction = storyTimeActionSet
		}
	}
	if claims.MinimumFocusBlockAction == "" {
		claims.MinimumFocusBlockAction = storyTimeActionUnchanged
		if claims.MinimumFocusBlockMinutes != nil {
			claims.MinimumFocusBlockAction = storyTimeActionSet
		}
	}
}
