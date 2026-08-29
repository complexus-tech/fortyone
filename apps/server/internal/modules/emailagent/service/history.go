package emailagent

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const truncationMarker = "…"

// DefaultHistoryLimits keeps a useful recent exchange and a durable summary
// while placing a hard bound on model context.
func DefaultHistoryLimits() HistoryLimits {
	return HistoryLimits{
		MaxTurns:        12,
		MaxTotalRunes:   12_000,
		MaxTurnRunes:    4_000,
		MaxSummaryRunes: 6_000,
	}
}

// BoundHistory returns the newest visible turns in chronological order. It
// copies its input and never modifies durable caller-owned history.
func BoundHistory(history []HistoryTurn, limits HistoryLimits) HistoryWindow {
	if validateHistoryLimits(limits) != nil || len(history) == 0 {
		return HistoryWindow{Turns: []HistoryTurn{}}
	}

	selectedReverse := make([]HistoryTurn, 0, min(len(history), limits.MaxTurns))
	totalRunes := 0
	omitted := 0
	truncated := false

	for index := len(history) - 1; index >= 0; index-- {
		turn := history[index]
		turn.Text = strings.TrimSpace(turn.Text)
		if turn.Text == "" || (turn.Role != RoleUser && turn.Role != RoleAssistant) {
			omitted++
			truncated = true
			continue
		}
		if len(selectedReverse) == limits.MaxTurns {
			omitted += index + 1
			truncated = true
			break
		}

		boundedText, turnTruncated := truncateRunes(turn.Text, limits.MaxTurnRunes)
		turn.Text = boundedText
		turnRunes := utf8.RuneCountInString(turn.Text)
		remaining := limits.MaxTotalRunes - totalRunes
		if remaining <= 0 {
			omitted += index + 1
			truncated = true
			break
		}
		if turnRunes > remaining {
			if len(selectedReverse) == 0 {
				turn.Text, _ = truncateRunes(turn.Text, remaining)
				selectedReverse = append(selectedReverse, turn)
				omitted += index
			} else {
				omitted += index + 1
			}
			truncated = true
			break
		}
		selectedReverse = append(selectedReverse, turn)
		totalRunes += turnRunes
		truncated = truncated || turnTruncated
	}

	turns := make([]HistoryTurn, len(selectedReverse))
	for index := range selectedReverse {
		turns[len(selectedReverse)-1-index] = selectedReverse[index]
	}
	return HistoryWindow{
		Turns:        turns,
		OmittedTurns: omitted,
		Truncated:    truncated,
	}
}

// BoundSummary trims and bounds a durable summary. The bool reports whether
// content was truncated.
func BoundSummary(summary string, maxRunes int) (string, bool) {
	return truncateRunes(strings.TrimSpace(summary), maxRunes)
}

func validateHistoryLimits(limits HistoryLimits) error {
	if limits.MaxTurns <= 0 {
		return fmt.Errorf("%w: max history turns must be positive", ErrInvalidRequest)
	}
	if limits.MaxTotalRunes <= 0 {
		return fmt.Errorf("%w: max total history runes must be positive", ErrInvalidRequest)
	}
	if limits.MaxTurnRunes <= 0 || limits.MaxTurnRunes > limits.MaxTotalRunes {
		return fmt.Errorf("%w: max turn runes must be positive and no greater than total history runes", ErrInvalidRequest)
	}
	if limits.MaxSummaryRunes <= 0 {
		return fmt.Errorf("%w: max summary runes must be positive", ErrInvalidRequest)
	}
	return nil
}

func truncateRunes(value string, limit int) (string, bool) {
	if limit <= 0 {
		return "", value != ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value, false
	}
	if limit == 1 {
		return truncationMarker, true
	}
	return string(runes[:limit-1]) + truncationMarker, true
}
