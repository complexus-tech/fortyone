package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

func (e *FortyOneToolExecutor) joinedTeams(ctx context.Context, scope ToolScope) ([]messagingTeam, map[uuid.UUID]messagingTeam, error) {
	items, err := e.teams.List(ctx, scope.WorkspaceID, scope.UserID, messagingTeamFilter{JoinedOnly: true})
	if err != nil {
		return nil, nil, fmt.Errorf("list joined teams: %w", err)
	}
	joined := make([]messagingTeam, 0, len(items))
	joinedByID := make(map[uuid.UUID]messagingTeam, len(items))
	var allowedTeamIDs map[uuid.UUID]struct{}
	if scope.AllowedTeamIDs != nil {
		allowedTeamIDs = make(map[uuid.UUID]struct{}, len(scope.AllowedTeamIDs))
		for _, teamID := range scope.AllowedTeamIDs {
			if teamID != uuid.Nil {
				allowedTeamIDs[teamID] = struct{}{}
			}
		}
	}
	for _, team := range items {
		if team.ID == uuid.Nil || team.Workspace != scope.WorkspaceID {
			continue
		}
		if allowedTeamIDs != nil {
			if _, allowed := allowedTeamIDs[team.ID]; !allowed {
				continue
			}
		}
		if _, duplicate := joinedByID[team.ID]; duplicate {
			continue
		}
		joined = append(joined, team)
		joinedByID[team.ID] = team
	}
	return joined, joinedByID, nil
}

func accessibleTeamID(raw *string, joined map[uuid.UUID]messagingTeam) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	teamID, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil || teamID == uuid.Nil {
		return nil, fmt.Errorf("%w: team_id must be a UUID or null", ErrInvalidToolArguments)
	}
	if _, ok := joined[teamID]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	return &teamID, nil
}

func normalizedLimit(value *int) (int, error) {
	if value == nil {
		return defaultToolLimit, nil
	}
	if *value < 1 || *value > maxToolLimit {
		return 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidToolArguments, maxToolLimit)
	}
	return *value, nil
}

func decodeToolArguments(raw json.RawMessage, target any, required ...string) error {
	if len(raw) == 0 {
		return fmt.Errorf("%w: arguments are required", ErrInvalidToolArguments)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return fmt.Errorf("%w: arguments must be a JSON object", ErrInvalidToolArguments)
	}
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%w: missing %s", ErrInvalidToolArguments, key)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: arguments contain trailing data", ErrInvalidToolArguments)
	}
	return nil
}

func marshalToolResult(value any) (json.RawMessage, error) {
	result, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal tool result: %w", err)
	}
	return result, nil
}

func storyReference(teamCode string, sequenceID int) string {
	teamCode = strings.ToUpper(strings.TrimSpace(teamCode))
	if teamCode == "" {
		return fmt.Sprintf("#%d", sequenceID)
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID)
}
