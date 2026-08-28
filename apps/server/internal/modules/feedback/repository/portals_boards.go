package feedbackrepository

import (
	"context"
	"errors"
	"strings"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505"

func (r *Repo) GetPortalBySlug(ctx context.Context, slug string) (feedback.CorePortal, error) {
	return r.GetPortalByWorkspaceSlugAndSlug(ctx, slug, slug)
}

func (r *Repo) GetPortalByWorkspaceSlugAndSlug(ctx context.Context, workspaceSlug, slug string) (feedback.CorePortal, error) {
	if strings.TrimSpace(workspaceSlug) != strings.TrimSpace(slug) {
		return feedback.CorePortal{}, feedback.ErrNotFound
	}
	row, err := r.queries.GetPublicFeedbackPortalByWorkspaceSlug(ctx, feedbacksql.GetPublicFeedbackPortalByWorkspaceSlugParams{WorkspaceSlug: strings.TrimSpace(workspaceSlug)})
	if err != nil {
		return feedback.CorePortal{}, normalizeError(err)
	}
	return portalProjection{row.ID, row.WorkspaceID, row.Name, row.Slug, row.IsPublic, row.ParticipationMode,
		row.GuestIdentityPolicy, row.HasPublishedUpdates, row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) GetPortal(ctx context.Context, workspaceID, portalID uuid.UUID) (feedback.CorePortal, error) {
	row, err := r.queries.GetFeedbackPortal(ctx, feedbacksql.GetFeedbackPortalParams{WorkspaceID: workspaceID, PortalID: portalID})
	if err != nil {
		return feedback.CorePortal{}, normalizeError(err)
	}
	return portalProjection{row.ID, row.WorkspaceID, row.Name, row.Slug, row.IsPublic, row.ParticipationMode,
		row.GuestIdentityPolicy, row.HasPublishedUpdates, row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) ListPortals(ctx context.Context, workspaceID uuid.UUID) ([]feedback.CorePortal, error) {
	rows, err := r.queries.ListFeedbackPortals(ctx, feedbacksql.ListFeedbackPortalsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, err
	}
	portals := make([]feedback.CorePortal, 0, len(rows))
	for _, row := range rows {
		portals = append(portals, portalProjection{row.ID, row.WorkspaceID, row.Name, row.Slug, row.IsPublic,
			row.ParticipationMode, row.GuestIdentityPolicy, row.HasPublishedUpdates, row.CreatedAt, row.UpdatedAt}.core())
	}
	return portals, nil
}

func (r *Repo) CreatePortal(ctx context.Context, input feedback.CorePortalInput) (feedback.CorePortal, error) {
	row, err := r.queries.CreateFeedbackPortal(ctx, feedbacksql.CreateFeedbackPortalParams{
		WorkspaceID: input.WorkspaceID, IsPublic: pointerOr(input.IsPublic, false),
		ParticipationMode:   pointerOr(input.ParticipationMode, feedback.ParticipationModeAccountRequired),
		GuestIdentityPolicy: pointerOr(input.GuestIdentityPolicy, feedback.GuestIdentityPolicyShowIdentity),
	})
	if err != nil {
		return feedback.CorePortal{}, normalizeError(err)
	}
	return portalProjection{row.ID, row.WorkspaceID, row.Name, row.Slug, row.IsPublic, row.ParticipationMode,
		row.GuestIdentityPolicy, row.HasPublishedUpdates, row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input feedback.CorePortalInput) (feedback.CorePortal, error) {
	row, err := r.queries.UpdateFeedbackPortal(ctx, feedbacksql.UpdateFeedbackPortalParams{WorkspaceID: workspaceID,
		PortalID: portalID, IsPublic: input.IsPublic, ParticipationMode: input.ParticipationMode,
		GuestIdentityPolicy: input.GuestIdentityPolicy})
	if err != nil {
		return feedback.CorePortal{}, normalizeError(err)
	}
	return portalProjection{row.ID, row.WorkspaceID, row.Name, row.Slug, row.IsPublic, row.ParticipationMode,
		row.GuestIdentityPolicy, row.HasPublishedUpdates, row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) ListBoards(ctx context.Context, portalID uuid.UUID) ([]feedback.CoreBoard, error) {
	rows, err := r.queries.ListFeedbackBoards(ctx, feedbacksql.ListFeedbackBoardsParams{PortalID: portalID})
	if err != nil {
		return nil, err
	}
	boards := make([]feedback.CoreBoard, 0, len(rows))
	for _, row := range rows {
		boards = append(boards, toCoreBoard(row))
	}
	return boards, nil
}

func (r *Repo) GetBoard(ctx context.Context, portalID, boardID uuid.UUID) (feedback.CoreBoard, error) {
	row, err := r.queries.GetFeedbackBoard(ctx, feedbacksql.GetFeedbackBoardParams{PortalID: portalID, BoardID: boardID})
	if err != nil {
		return feedback.CoreBoard{}, normalizeError(err)
	}
	return toCoreBoard(row), nil
}

func (r *Repo) CreateBoard(ctx context.Context, input feedback.CoreBoardInput) (feedback.CoreBoard, error) {
	orderIndex, err := safecast.Int32(input.OrderIndex)
	if err != nil {
		return feedback.CoreBoard{}, errors.Join(feedback.ErrInvalidInput, err)
	}
	var board feedback.CoreBoard
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		row, err := q.CreateFeedbackBoard(ctx, feedbacksql.CreateFeedbackBoardParams{
			Name: input.Name, Slug: input.Slug, Color: input.Color, OrderIndex: orderIndex,
			TeamID: input.TeamID, PortalID: input.PortalID, WorkspaceID: input.WorkspaceID,
		})
		if err != nil {
			return normalizeBoardWriteError(err)
		}
		if _, err = q.AddFeedbackBoardCreatorReviewer(ctx, feedbacksql.AddFeedbackBoardCreatorReviewerParams{
			EmailFrequency: feedback.EmailFrequencyWeekly, CreatorID: input.CreatorID,
			WorkspaceID: input.WorkspaceID, BoardID: row.ID,
		}); err != nil {
			return normalizeError(err)
		}
		board = toCoreBoard(row)
		return nil
	})
	return board, err
}

func (r *Repo) DeleteBoard(ctx context.Context, workspaceID, boardID uuid.UUID) error {
	return r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		contributors, err := q.LockAnonymousFeedbackBoardContributors(ctx, feedbacksql.LockAnonymousFeedbackBoardContributorsParams{WorkspaceID: workspaceID, BoardID: boardID})
		if err != nil {
			return err
		}
		count, err := q.DeleteFeedbackBoard(ctx, feedbacksql.DeleteFeedbackBoardParams{WorkspaceID: workspaceID, BoardID: boardID})
		if err != nil {
			return err
		}
		if err = requireRowsAffected(count); err != nil {
			return err
		}
		if len(contributors) > 0 {
			_, err = q.DeleteOrphanAnonymousFeedbackContributors(ctx, feedbacksql.DeleteOrphanAnonymousFeedbackContributorsParams{ContributorIds: contributors})
		}
		return err
	})
}

func (r *Repo) ListBoardReviewers(ctx context.Context, workspaceID, boardID uuid.UUID) ([]feedback.CoreBoardReviewer, error) {
	rows, err := r.queries.ListFeedbackBoardReviewers(ctx, feedbacksql.ListFeedbackBoardReviewersParams{
		DefaultFrequency: feedback.EmailFrequencyWeekly, WorkspaceID: workspaceID, BoardID: boardID,
	})
	if err != nil {
		return nil, err
	}
	reviewers := make([]feedback.CoreBoardReviewer, 0, len(rows))
	for _, row := range rows {
		frequency, conversionErr := stringValue(row.EmailFrequency)
		if conversionErr != nil {
			return nil, conversionErr
		}
		reviewers = append(reviewers, feedback.CoreBoardReviewer{UserID: row.UserID, Name: row.Name,
			Email: row.Email, AvatarURL: row.AvatarURL, Role: row.Role, EmailFrequency: frequency})
	}
	return reviewers, nil
}

func (r *Repo) SetBoardReviewer(ctx context.Context, input feedback.CoreBoardReviewerInput) (feedback.CoreBoardReviewer, error) {
	if input.EmailFrequency == feedback.EmailFrequencyOff {
		row, err := r.queries.RemoveFeedbackBoardReviewer(ctx, feedbacksql.RemoveFeedbackBoardReviewerParams{
			EmailFrequency: input.EmailFrequency, UserID: input.UserID, WorkspaceID: input.WorkspaceID, BoardID: input.BoardID,
		})
		if err != nil {
			return feedback.CoreBoardReviewer{}, normalizeError(err)
		}
		return feedback.CoreBoardReviewer{UserID: row.UserID, Name: row.Name, Email: row.Email,
			AvatarURL: row.AvatarURL, Role: row.Role, EmailFrequency: row.EmailFrequency}, nil
	}
	row, err := r.queries.SetFeedbackBoardReviewer(ctx, feedbacksql.SetFeedbackBoardReviewerParams{
		UserID: input.UserID, WorkspaceID: input.WorkspaceID, BoardID: input.BoardID, EmailFrequency: input.EmailFrequency,
	})
	if err != nil {
		return feedback.CoreBoardReviewer{}, normalizeError(err)
	}
	return feedback.CoreBoardReviewer{UserID: row.UserID, Name: row.Name, Email: row.Email,
		AvatarURL: row.AvatarURL, Role: row.Role, EmailFrequency: row.EmailFrequency}, nil
}

func pointerOr[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}

func normalizeBoardWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return feedback.ErrBoardExists
	}
	return normalizeError(err)
}
