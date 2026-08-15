package stories

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidEstimatedDuration   = errors.New("estimated duration minutes must be greater than zero")
	ErrInvalidMinimumFocusBlock   = errors.New("minimum focus block minutes must be greater than zero")
	ErrEstimatedDurationTooLarge  = errors.New("estimated duration minutes must not exceed 2400")
	ErrMinimumFocusBlockTooLarge  = errors.New("minimum focus block minutes must not exceed 2400")
	ErrFocusBlockRequiresDuration = errors.New("minimum focus block minutes require estimated duration minutes")
	ErrFocusBlockExceedsDuration  = errors.New("minimum focus block minutes must not exceed estimated duration minutes")
)

const MaximumEstimatedDurationMinutes = 40 * 60

// ValidateStoryTimeContract validates the user-entered scheduling inputs on a
// story. Both fields are optional, but a minimum focus block is only meaningful
// when the story also has an estimated duration and cannot exceed that duration.
func ValidateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes *int) error {
	if estimatedDurationMinutes != nil && *estimatedDurationMinutes <= 0 {
		return ErrInvalidEstimatedDuration
	}
	if estimatedDurationMinutes != nil && *estimatedDurationMinutes > MaximumEstimatedDurationMinutes {
		return ErrEstimatedDurationTooLarge
	}
	if minimumFocusBlockMinutes != nil && *minimumFocusBlockMinutes <= 0 {
		return ErrInvalidMinimumFocusBlock
	}
	if minimumFocusBlockMinutes != nil && *minimumFocusBlockMinutes > MaximumEstimatedDurationMinutes {
		return ErrMinimumFocusBlockTooLarge
	}
	if estimatedDurationMinutes == nil && minimumFocusBlockMinutes != nil {
		return ErrFocusBlockRequiresDuration
	}
	if estimatedDurationMinutes != nil && minimumFocusBlockMinutes != nil && *minimumFocusBlockMinutes > *estimatedDurationMinutes {
		return ErrFocusBlockExceedsDuration
	}
	return nil
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
