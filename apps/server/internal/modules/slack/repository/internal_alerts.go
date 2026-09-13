package slackrepository

import (
	"context"
	"errors"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ClaimInternalAlert(ctx context.Context, teamID, channelID string) (*slackdomain.InternalAlert, error) {
	token := uuid.New()
	row, err := r.queries.ClaimInternalSlackAlert(ctx, slacksql.ClaimInternalSlackAlertParams{
		TeamID: teamID, ChannelID: channelID, LeaseToken: &token,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return &slackdomain.InternalAlert{ID: row.ID, LeaseToken: token, DedupeKey: row.DedupeKey,
		Kind: row.Kind, Payload: row.Payload, Attempts: int(row.Attempts)}, nil
}

func (r *Repo) CompleteInternalAlert(ctx context.Context, alert slackdomain.InternalAlert, messageTS string) error {
	rows, err := r.queries.CompleteInternalSlackAlert(ctx, slacksql.CompleteInternalSlackAlertParams{
		ID: alert.ID, LeaseToken: &alert.LeaseToken, MessageTs: messageTS,
	})
	return internalAlertMutationResult(rows, err)
}

func (r *Repo) RetryInternalAlert(ctx context.Context, alert slackdomain.InternalAlert, availableAt time.Time) error {
	rows, err := r.queries.RetryInternalSlackAlert(ctx, slacksql.RetryInternalSlackAlertParams{
		ID: alert.ID, LeaseToken: &alert.LeaseToken, AvailableAt: availableAt,
	})
	return internalAlertMutationResult(rows, err)
}

func internalAlertMutationResult(rows int64, err error) error {
	if err != nil {
		return mapDatabaseError(err)
	}
	if rows != 1 {
		return slackdomain.ErrInternalAlertLeaseLost
	}
	return nil
}
