package webhooks

import (
	"regexp"
	"strings"
)

var outcomeCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)

func ValidateCompletion(status Status, outcomeCode string) error {
	switch status {
	case StatusCompleted, StatusIgnored, StatusFailed, StatusCancelled:
	default:
		return ErrInvalidState
	}
	outcomeCode = strings.TrimSpace(outcomeCode)
	if outcomeCode == "" {
		if status == StatusFailed || status == StatusCancelled {
			return ErrInvalidState
		}
		return nil
	}
	if !outcomeCodePattern.MatchString(outcomeCode) {
		return ErrInvalidState
	}
	return nil
}
