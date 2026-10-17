package agentinvocationshttp

import (
	"encoding/json"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/google/uuid"
)

type chatTaskResponse struct {
	ID         uuid.UUID  `json:"id"`
	Reference  string     `json:"reference"`
	Title      string     `json:"title"`
	TeamID     uuid.UUID  `json:"teamId"`
	Priority   string     `json:"priority"`
	StatusID   uuid.UUID  `json:"statusId"`
	AssigneeID *uuid.UUID `json:"assigneeId,omitempty"`
}

// Only committed task receipts become public cards. Proposal tokens, execution
// identities and provider metadata remain private to the durable run.
func presentChatTasks(run agentruns.Run) []chatTaskResponse {
	if run.Surface != "chat_message" || run.Outcome == nil {
		return nil
	}
	var receipt struct {
		Result struct {
			Operation string `json:"operation"`
			Items     []struct {
				ID         uuid.UUID  `json:"story_id"`
				Reference  string     `json:"reference"`
				Title      string     `json:"title"`
				TeamID     uuid.UUID  `json:"team_id"`
				Priority   string     `json:"priority"`
				StatusID   uuid.UUID  `json:"story_status_id"`
				AssigneeID *uuid.UUID `json:"assignee_id"`
			} `json:"items"`
		} `json:"result"`
	}
	if json.Unmarshal(run.Outcome.Receipt, &receipt) != nil || receipt.Result.Operation != "create_stories" {
		return nil
	}
	tasks := make([]chatTaskResponse, 0, len(receipt.Result.Items))
	for _, item := range receipt.Result.Items {
		if item.ID == uuid.Nil || item.TeamID != run.TeamID {
			continue
		}
		tasks = append(tasks, chatTaskResponse{ID: item.ID, Reference: item.Reference, Title: item.Title, TeamID: item.TeamID, Priority: item.Priority, StatusID: item.StatusID, AssigneeID: item.AssigneeID})
	}
	return tasks
}
