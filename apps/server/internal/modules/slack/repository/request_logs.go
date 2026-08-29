package slackrepository

import (
	"context"
	"fmt"

	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

func (repository *Repo) InsertRequestLog(ctx context.Context, entry SlackRequestLogInsert) error {
	responseCode, err := safecast.Int32(entry.ResponseCode)
	if err != nil {
		return err
	}
	if err := repository.queries.InsertRequestLog(ctx, slacksql.InsertRequestLogParams{
		RequestType: entry.RequestType, Endpoint: entry.Endpoint,
		WorkspaceID: entry.WorkspaceID, SlackTeamID: entry.SlackTeamID,
		SlackUserID: entry.SlackUserID, SlackChannelID: entry.SlackChannel,
		Command: entry.Command, TriggerID: entry.TriggerID, RequestBody: entry.RequestBody,
		Headers: entry.Headers, ResponseCode: responseCode, Outcome: entry.Outcome,
		ErrorMessage: entry.ErrorMessage,
	}); err != nil {
		return fmt.Errorf("insert Slack request log: %w", mapDatabaseError(err))
	}
	return nil
}
