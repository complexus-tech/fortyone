package notificationsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const maximumWeeklyDigestRecipientBatchSize = 100

var (
	errWeeklyDigestRepositoryNotConfigured = errors.New("weekly digest repository is not configured")
	errInvalidWeeklyDigestRecipient        = errors.New("weekly digest query returned an invalid recipient")
)

// ListWeeklyDigestRecipients returns one bounded, stable page of recipients
// whose current membership and notification preference allow the digest.
func (repository *Repository) ListWeeklyDigestRecipients(
	ctx context.Context,
	cursor *notificationsdomain.WeeklyDigestCursor,
	limit int,
) ([]notificationsdomain.WeeklyDigestRecipient, error) {
	if repository == nil || repository.queries == nil {
		return nil, errWeeklyDigestRepositoryNotConfigured
	}
	if limit < 1 || limit > maximumWeeklyDigestRecipientBatchSize {
		return nil, fmt.Errorf(
			"%w: weekly digest recipient limit must be between 1 and %d",
			notificationsdomain.ErrInvalid,
			maximumWeeklyDigestRecipientBatchSize,
		)
	}
	resultLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("%w: weekly digest recipient limit: %v", notificationsdomain.ErrInvalid, err)
	}

	params := notificationssql.ListWeeklyDigestRecipientsParams{ResultLimit: resultLimit}
	if cursor != nil {
		if cursor.WorkspaceID == uuid.Nil || cursor.UserID == uuid.Nil {
			return nil, fmt.Errorf("%w: weekly digest cursor requires workspace and user", notificationsdomain.ErrInvalid)
		}
		params.HasCursor = true
		params.AfterWorkspaceID = cursor.WorkspaceID
		params.AfterUserID = cursor.UserID
	}

	rows, err := repository.queries.ListWeeklyDigestRecipients(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list weekly digest recipients: %w", err)
	}
	recipients := make([]notificationsdomain.WeeklyDigestRecipient, 0, len(rows))
	for _, row := range rows {
		userEmail := strings.TrimSpace(row.UserEmail)
		workspaceSlug := strings.TrimSpace(row.WorkspaceSlug)
		if row.UserID == uuid.Nil || row.WorkspaceID == uuid.Nil || userEmail == "" || workspaceSlug == "" {
			return nil, errInvalidWeeklyDigestRecipient
		}
		recipients = append(recipients, notificationsdomain.WeeklyDigestRecipient{
			UserID:        row.UserID,
			UserEmail:     userEmail,
			UserName:      row.UserName,
			WorkspaceID:   row.WorkspaceID,
			WorkspaceName: row.WorkspaceName,
			WorkspaceSlug: workspaceSlug,
		})
	}
	return recipients, nil
}

// GetWeeklyDigestStats loads the current, tenant-scoped aggregate for one
// recipient. The caller supplies the run timestamp so every recipient uses the
// same UTC reporting boundary.
func (repository *Repository) GetWeeklyDigestStats(
	ctx context.Context,
	query notificationsdomain.WeeklyDigestStatsQuery,
) (notificationsdomain.WeeklyDigestStats, error) {
	if repository == nil || repository.queries == nil {
		return notificationsdomain.WeeklyDigestStats{}, errWeeklyDigestRepositoryNotConfigured
	}
	if err := query.Validate(); err != nil {
		return notificationsdomain.WeeklyDigestStats{}, err
	}

	row, err := repository.queries.GetWeeklyDigestStats(ctx, notificationssql.GetWeeklyDigestStatsParams{
		UserID:      query.UserID,
		WorkspaceID: query.WorkspaceID,
		AsOf:        query.AsOf.UTC(),
	})
	if err != nil {
		return notificationsdomain.WeeklyDigestStats{}, fmt.Errorf("get weekly digest stats: %w", err)
	}
	return notificationsdomain.WeeklyDigestStats{
		UnreadNotifications:         int(row.UnreadNotifications),
		UnreadPriorityNotifications: int(row.UnreadPriorityNotifications),
		OverdueStories:              int(row.OverdueStories),
		DueThisWeekStories:          int(row.DueThisWeekStories),
		ObjectiveRisks:              int(row.ObjectiveRisks),
		TeamComments:                int(row.TeamComments),
	}, nil
}
