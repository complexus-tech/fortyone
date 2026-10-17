package agentinvocationshttp

import (
	"encoding/json"
	"testing"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatPresentationExposesOnlyCommittedCurrentTeamTaskCards(t *testing.T) {
	t.Parallel()
	team, story, reply := uuid.New(), uuid.New(), uuid.New()
	receipt, err := json.Marshal(map[string]any{"executorId": uuid.New(), "result": map[string]any{"operation": "create_stories", "items": []map[string]any{
		{"story_id": story, "team_id": team, "reference": "PROD-42", "title": "Onboarding checklist", "priority": "High"},
		{"story_id": uuid.New(), "team_id": uuid.New(), "reference": "OTHER-1", "title": "Outside this team"},
	}}})
	require.NoError(t, err)
	run := agentruns.Run{Surface: "chat_message", TeamID: team, ReplyCommentID: &reply, Outcome: &agentruns.Outcome{Receipt: receipt}, Proposal: &agentruns.Proposal{Action: json.RawMessage(`{"token":"secret"}`)}}
	response := present(run)
	require.Equal(t, &reply, response.ReplyMessageID)
	require.Len(t, response.Tasks, 1)
	require.Equal(t, story, response.Tasks[0].ID)
	body, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotContains(t, string(body), "secret")
	require.NotContains(t, string(body), "executorId")
	require.NotContains(t, string(body), "Outside this team")
	run.Outcome = nil
	require.Empty(t, present(run).Tasks)
}
