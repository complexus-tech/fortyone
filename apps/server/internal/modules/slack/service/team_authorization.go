package slack

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) findTeamForActor(ctx context.Context, workspaceID, actorID, teamID uuid.UUID) (slackTeamRecord, error) {
	if workspaceID == uuid.Nil || actorID == uuid.Nil || teamID == uuid.Nil {
		return slackTeamRecord{}, ErrSlackTeamNotAvailable
	}
	teams, err := s.repo.ListWorkspaceTeamsForUser(ctx, workspaceID, actorID)
	if err != nil {
		return slackTeamRecord{}, err
	}
	for _, team := range teams {
		if team.ID == teamID {
			return team, nil
		}
	}
	return slackTeamRecord{}, ErrSlackTeamNotAvailable
}
