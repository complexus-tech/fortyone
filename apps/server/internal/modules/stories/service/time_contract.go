package stories

import (
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
)

var (
	ErrInvalidEstimatedDuration   = storydomain.ErrInvalidEstimatedDuration
	ErrInvalidMinimumFocusBlock   = storydomain.ErrInvalidMinimumFocusBlock
	ErrEstimatedDurationTooLarge  = storydomain.ErrEstimatedDurationTooLarge
	ErrMinimumFocusBlockTooLarge  = storydomain.ErrMinimumFocusBlockTooLarge
	ErrFocusBlockRequiresDuration = storydomain.ErrFocusBlockRequiresDuration
	ErrFocusBlockExceedsDuration  = storydomain.ErrFocusBlockExceedsDuration
)

const MaximumEstimatedDurationMinutes = storydomain.MaximumEstimatedDurationMinutes

// ValidateStoryTimeContract validates the user-entered scheduling inputs on a
// story. Both fields are optional, but a minimum focus block is only meaningful
// when the story also has an estimated duration and cannot exceed that duration.
func ValidateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes *int) error {
	return storydomain.ValidateScheduling(estimatedDurationMinutes, minimumFocusBlockMinutes)
}

func applyStoryTimeContractUpdate(story CoreSingleStory, updates map[string]any) error {
	estimatedDuration, estimatedDurationSet, err := storyTimeUpdateValue(
		story.EstimatedDurationMinutes,
		updates,
		"estimated_duration_minutes",
	)
	if err != nil {
		return err
	}
	minimumFocusBlock, minimumFocusBlockSet, err := storyTimeUpdateValue(
		story.MinimumFocusBlockMinutes,
		updates,
		"minimum_focus_block_minutes",
	)
	if err != nil {
		return err
	}
	if !estimatedDurationSet && !minimumFocusBlockSet {
		return nil
	}
	if err := ValidateStoryTimeContract(estimatedDuration, minimumFocusBlock); err != nil {
		return err
	}
	if estimatedDurationSet {
		updates["estimated_duration_minutes"] = estimatedDuration
	}
	if minimumFocusBlockSet {
		updates["minimum_focus_block_minutes"] = minimumFocusBlock
	}
	return nil
}

func storyTimeUpdateValue(current *int, updates map[string]any, field string) (*int, bool, error) {
	raw, ok := updates[field]
	if !ok {
		return current, false, nil
	}
	if raw == nil {
		return nil, true, nil
	}

	var value int
	switch typed := raw.(type) {
	case *int:
		if typed == nil {
			return nil, true, nil
		}
		value = *typed
	case int:
		value = typed
	case int32:
		value = int(typed)
	case int64:
		value = int(typed)
	case float64:
		value = int(typed)
		if float64(value) != typed {
			return nil, false, fmt.Errorf("invalid %s type: %T", field, raw)
		}
	default:
		return nil, false, fmt.Errorf("invalid %s type: %T", field, raw)
	}
	return &value, true, nil
}
