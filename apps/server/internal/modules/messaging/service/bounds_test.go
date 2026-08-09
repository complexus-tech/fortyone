package messaging

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRequestRejectsOversizedCurrentPrompt(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	request.Prompt = strings.Repeat("x", MaximumMessageBytes+1)

	_, err := NormalizeRequest(request)

	require.ErrorIs(t, err, ErrInvalidRequest)
	require.ErrorIs(t, err, ErrMessageTooLarge)
}

func TestNormalizeRequestRetainsPromptAtExactByteLimit(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	request.Prompt = strings.Repeat("x", MaximumMessageBytes)

	normalized, err := NormalizeRequest(request)

	require.NoError(t, err)
	require.Equal(t, request.Prompt, normalized.Prompt)
}

func TestNormalizeRequestTrimsOldestHistoryByAggregateBytes(t *testing.T) {
	t.Parallel()

	turnText := strings.Repeat("x", 12<<10)
	request := boundedTestRequest()
	request.Conversation = []ConversationTurn{
		{Role: RoleUser, Text: "oldest question"},
		{Role: RoleAssistant, Text: "oldest answer"},
		{Role: RoleUser, Text: turnText},
		{Role: RoleAssistant, Text: turnText},
		{Role: RoleUser, Text: turnText},
		{Role: RoleAssistant, Text: turnText},
	}

	normalized, err := NormalizeRequest(request)

	require.NoError(t, err)
	require.Len(t, normalized.Conversation, 4)
	require.Equal(t, RoleUser, normalized.Conversation[0].Role)
	require.Equal(t, request.Conversation[2:], normalized.Conversation)
	require.LessOrEqual(t, conversationBytes(normalized.Conversation), MaximumConversationBytes)
}

func TestNormalizeRequestDropsLeadingAssistantAfterTurnTrim(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	for index := 0; index < MaximumConversationTurns+1; index++ {
		role := RoleUser
		if index%2 == 1 {
			role = RoleAssistant
		}
		request.Conversation = append(request.Conversation, ConversationTurn{Role: role, Text: "turn"})
	}

	normalized, err := NormalizeRequest(request)

	require.NoError(t, err)
	require.Len(t, normalized.Conversation, MaximumConversationTurns-1)
	require.Equal(t, RoleUser, normalized.Conversation[0].Role)
	require.Equal(t, request.Conversation[2:], normalized.Conversation)
}

func TestNormalizeRequestTreatsOversizedHistoryAsContextBoundary(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	request.Conversation = []ConversationTurn{
		{Role: RoleUser, Text: "must not cross the boundary"},
		{Role: RoleAssistant, Text: strings.Repeat("x", MaximumMessageBytes+1)},
		{Role: RoleUser, Text: "new question"},
		{Role: RoleAssistant, Text: "new answer"},
	}

	normalized, err := NormalizeRequest(request)

	require.NoError(t, err)
	require.Equal(t, request.Conversation[2:], normalized.Conversation)
}

func TestNormalizeRequestRejectsInvalidHistoryEvenWhenItWouldBeTrimmed(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	request.Conversation = []ConversationTurn{
		{Role: ConversationRole("system"), Text: "invalid"},
		{Role: RoleUser, Text: strings.Repeat("x", MaximumMessageBytes)},
		{Role: RoleAssistant, Text: strings.Repeat("x", MaximumMessageBytes)},
		{Role: RoleUser, Text: strings.Repeat("x", MaximumMessageBytes)},
		{Role: RoleAssistant, Text: strings.Repeat("x", MaximumMessageBytes)},
	}

	_, err := NormalizeRequest(request)

	require.True(t, errors.Is(err, ErrInvalidRequest))
}

func TestNormalizeRequestPreservesAndCopiesAllowedTeamScope(t *testing.T) {
	t.Parallel()

	teamID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	request := boundedTestRequest()
	request.AllowedTeamIDs = []uuid.UUID{teamID, teamID}

	normalized, err := NormalizeRequest(request)

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{teamID}, normalized.AllowedTeamIDs)
	request.AllowedTeamIDs[0] = uuid.New()
	require.Equal(t, teamID, normalized.AllowedTeamIDs[0], "normalized scope must not alias provider input")

	empty := boundedTestRequest()
	empty.AllowedTeamIDs = []uuid.UUID{}
	normalizedEmpty, err := NormalizeRequest(empty)
	require.NoError(t, err)
	require.NotNil(t, normalizedEmpty.AllowedTeamIDs, "an explicit empty channel policy must not become unrestricted")
	require.Empty(t, normalizedEmpty.AllowedTeamIDs)

	unrestricted, err := NormalizeRequest(boundedTestRequest())
	require.NoError(t, err)
	require.Nil(t, unrestricted.AllowedTeamIDs)
}

func TestNormalizeRequestRejectsNilAllowedTeamID(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	request.AllowedTeamIDs = []uuid.UUID{uuid.Nil}

	_, err := NormalizeRequest(request)

	require.ErrorIs(t, err, ErrInvalidRequest)
}

func TestNormalizeRequestBoundsWorkspaceGuidance(t *testing.T) {
	t.Parallel()

	request := boundedTestRequest()
	request.Guidance = "  Keep answers concise.  "
	normalized, err := NormalizeRequest(request)
	require.NoError(t, err)
	require.Equal(t, "Keep answers concise.", normalized.Guidance)

	request.Guidance = strings.Repeat("x", MaximumGuidanceRunes+1)
	_, err = NormalizeRequest(request)
	require.ErrorIs(t, err, ErrInvalidRequest)
}

func boundedTestRequest() Request {
	return Request{
		WorkspaceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Prompt:      "What is due?",
	}
}

func conversationBytes(turns []ConversationTurn) int {
	total := 0
	for _, turn := range turns {
		total += len(turn.Text)
	}
	return total
}
