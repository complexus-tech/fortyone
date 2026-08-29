package emailagent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoundHistoryKeepsNewestCompleteTurnsAndDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	history := []HistoryTurn{
		{Role: RoleUser, Text: "  first  "},
		{Role: RoleAssistant, Text: "second is too long"},
		{Role: ConversationRole("system"), Text: "invalid role"},
		{Role: RoleUser, Text: "third"},
	}
	originalFirst := history[0].Text
	window := BoundHistory(history, HistoryLimits{
		MaxTurns:        3,
		MaxTotalRunes:   15,
		MaxTurnRunes:    10,
		MaxSummaryRunes: 10,
	})

	require.Equal(t, originalFirst, history[0].Text)
	require.Equal(t, []HistoryTurn{
		{Role: RoleAssistant, Text: "second is…"},
		{Role: RoleUser, Text: "third"},
	}, window.Turns)
	require.Equal(t, 2, window.OmittedTurns)
	require.True(t, window.Truncated)
}

func TestBoundHistoryTruncatesNewestTurnWhenItAloneExceedsTotalBudget(t *testing.T) {
	t.Parallel()

	window := BoundHistory([]HistoryTurn{
		{Role: RoleUser, Text: "old"},
		{Role: RoleAssistant, Text: "this newest reply is too long"},
	}, HistoryLimits{
		MaxTurns:        5,
		MaxTotalRunes:   8,
		MaxTurnRunes:    8,
		MaxSummaryRunes: 10,
	})

	require.Len(t, window.Turns, 1)
	require.Equal(t, "this ne…", window.Turns[0].Text)
	require.Equal(t, 1, window.OmittedTurns)
	require.True(t, window.Truncated)
}

func TestWithHistoryLimitsRejectsInvalidBounds(t *testing.T) {
	t.Parallel()

	_, err := New(nil, WithHistoryLimits(HistoryLimits{
		MaxTurns:        2,
		MaxTotalRunes:   10,
		MaxTurnRunes:    11,
		MaxSummaryRunes: 5,
	}))

	require.ErrorIs(t, err, ErrInvalidRequest)
}
