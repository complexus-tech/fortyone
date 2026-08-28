package notifications

import (
	"fmt"
	"strings"

	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
)

// PreferenceType includes persisted notification types plus email-only
// communication categories. The latter intentionally cannot be enabled in the
// product inbox.
type PreferenceType string

const (
	PreferenceTypeStoryUpdate             PreferenceType = "story_update"
	PreferenceTypeStoryComment            PreferenceType = "story_comment"
	PreferenceTypeCommentReply            PreferenceType = "comment_reply"
	PreferenceTypeObjectiveUpdate         PreferenceType = "objective_update"
	PreferenceTypeKeyResultUpdate         PreferenceType = "key_result_update"
	PreferenceTypeMention                 PreferenceType = "mention"
	PreferenceTypeFeedbackComment         PreferenceType = "feedback_comment"
	PreferenceTypeFeedbackStatusUpdate    PreferenceType = "feedback_status_update"
	PreferenceTypeFeedbackUpdatePublished PreferenceType = "feedback_update_published"
	PreferenceTypeFeedbackItemMerged      PreferenceType = "feedback_item_merged"
	PreferenceTypeStrategyUpdate          PreferenceType = "strategy_update"
	PreferenceTypeReminders               PreferenceType = "reminders"
	PreferenceTypeWeeklyDigest            PreferenceType = "weekly_digest"
)

var preferenceTypes = [...]PreferenceType{
	PreferenceTypeStoryUpdate,
	PreferenceTypeStoryComment,
	PreferenceTypeCommentReply,
	PreferenceTypeObjectiveUpdate,
	PreferenceTypeKeyResultUpdate,
	PreferenceTypeStrategyUpdate,
	PreferenceTypeMention,
	PreferenceTypeFeedbackComment,
	PreferenceTypeFeedbackStatusUpdate,
	PreferenceTypeFeedbackUpdatePublished,
	PreferenceTypeFeedbackItemMerged,
	PreferenceTypeReminders,
	PreferenceTypeWeeklyDigest,
}

func ParsePreferenceType(value string) (PreferenceType, error) {
	preferenceType := PreferenceType(strings.TrimSpace(value))
	if !preferenceType.Valid() {
		return "", fmt.Errorf("%w: unsupported notification preference type %q", ErrInvalid, value)
	}
	return preferenceType, nil
}

func (preferenceType PreferenceType) Valid() bool {
	for _, candidate := range preferenceTypes {
		if preferenceType == candidate {
			return true
		}
	}
	return false
}

func (preferenceType PreferenceType) SupportsInAppDelivery() bool {
	switch preferenceType {
	case PreferenceTypeStrategyUpdate, PreferenceTypeReminders, PreferenceTypeWeeklyDigest:
		return false
	default:
		return preferenceType.Valid()
	}
}

type Channels struct {
	Email bool `json:"email"`
	InApp bool `json:"in_app"`
}

type PreferenceSet map[PreferenceType]Channels

func DefaultPreferences() PreferenceSet {
	preferences := make(PreferenceSet, len(preferenceTypes))
	for _, preferenceType := range preferenceTypes {
		preferences[preferenceType] = Channels{
			Email: true,
			InApp: preferenceType.SupportsInAppDelivery(),
		}
	}
	return preferences
}

func (preferences PreferenceSet) WithDefaults() PreferenceSet {
	result := DefaultPreferences()
	for preferenceType, channels := range preferences {
		if !preferenceType.Valid() {
			continue
		}
		if !preferenceType.SupportsInAppDelivery() {
			channels.InApp = false
		}
		result[preferenceType] = channels
	}
	return result
}

type ChannelPatch struct {
	Email platformpatch.Field[bool]
	InApp platformpatch.Field[bool]
}

func (patch ChannelPatch) Empty() bool {
	return !patch.Email.Specified() && !patch.InApp.Specified()
}

func (patch ChannelPatch) Normalized(preferenceType PreferenceType) ChannelPatch {
	if value, specified := patch.InApp.Value(); specified && value != nil && *value && !preferenceType.SupportsInAppDelivery() {
		patch.InApp = platformpatch.Set(false)
	}
	return patch
}
