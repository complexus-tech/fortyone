package feedbackrepository

import (
	"context"
	"errors"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) ListStoryLinks(ctx context.Context, portalID uuid.UUID, itemIDs []uuid.UUID) ([]feedback.CoreStoryLink, error) {
	if len(itemIDs) == 0 {
		return []feedback.CoreStoryLink{}, nil
	}
	rows, err := r.queries.ListPublicFeedbackStoryLinks(ctx, feedbacksql.ListPublicFeedbackStoryLinksParams{PortalID: portalID, ItemIds: itemIDs})
	if err != nil {
		return nil, err
	}
	links := make([]feedback.CoreStoryLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, feedback.CoreStoryLink{ID: row.ID, WorkspaceID: row.WorkspaceID, ItemID: row.ItemID,
			StoryID: row.StoryID, StoryTitle: row.StoryTitle, Relationship: row.Relationship, IsPrimary: row.IsPrimary,
			CreatedByUserID: valueOrZero(row.CreatedByUserID), CreatedAt: row.CreatedAt})
	}
	return links, nil
}

func (r *Repo) ListItemStoryLinks(ctx context.Context, workspaceID, itemID uuid.UUID) ([]feedback.CoreStoryLink, error) {
	rows, err := r.queries.ListInternalFeedbackStoryLinks(ctx, feedbacksql.ListInternalFeedbackStoryLinksParams{
		ActorID: workspaceID, WorkspaceID: workspaceID, ItemID: itemID, AllTeams: true,
	})
	if err != nil {
		return nil, err
	}
	links := make([]feedback.CoreStoryLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, feedback.CoreStoryLink{ID: row.ID, WorkspaceID: row.WorkspaceID, ItemID: row.ItemID,
			StoryID: row.StoryID, StoryTitle: row.StoryTitle, Relationship: row.Relationship, IsPrimary: row.IsPrimary,
			CreatedByUserID: valueOrZero(row.CreatedByUserID), CreatedAt: row.CreatedAt})
	}
	return links, nil
}

func (r *Repo) ListItemStoryLinksScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) ([]feedback.CoreStoryLink, error) {
	rows, err := r.queries.ListInternalFeedbackStoryLinks(ctx, feedbacksql.ListInternalFeedbackStoryLinksParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return nil, err
	}
	links := make([]feedback.CoreStoryLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, feedback.CoreStoryLink{ID: row.ID, WorkspaceID: row.WorkspaceID, ItemID: row.ItemID,
			StoryID: row.StoryID, StoryTitle: row.StoryTitle, Relationship: row.Relationship, IsPrimary: row.IsPrimary,
			CreatedByUserID: valueOrZero(row.CreatedByUserID), CreatedAt: row.CreatedAt})
	}
	return links, nil
}

func (r *Repo) ListStoryFeedbackLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]feedback.CoreStoryFeedbackLink, error) {
	rows, err := r.queries.ListStoryFeedbackLinks(ctx, feedbacksql.ListStoryFeedbackLinksParams{
		ActorID: workspaceID, WorkspaceID: workspaceID, StoryID: storyID, AllTeams: true,
	})
	if err != nil {
		return nil, err
	}
	links := make([]feedback.CoreStoryFeedbackLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, feedback.CoreStoryFeedbackLink{ID: row.ID, WorkspaceID: row.WorkspaceID,
			ItemID: row.ItemID, StoryID: row.StoryID, TeamID: row.TeamID, FeedbackTitle: row.FeedbackTitle,
			Relationship: row.Relationship, IsPrimary: row.IsPrimary, CreatedAt: row.CreatedAt})
	}
	return links, nil
}

func (r *Repo) ListStoryFeedbackLinksScoped(ctx context.Context, scope feedback.CoreAccessScope, storyID uuid.UUID) ([]feedback.CoreStoryFeedbackLink, error) {
	rows, err := r.queries.ListStoryFeedbackLinks(ctx, feedbacksql.ListStoryFeedbackLinksParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, StoryID: storyID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return nil, err
	}
	links := make([]feedback.CoreStoryFeedbackLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, feedback.CoreStoryFeedbackLink{ID: row.ID, WorkspaceID: row.WorkspaceID,
			ItemID: row.ItemID, StoryID: row.StoryID, TeamID: row.TeamID, FeedbackTitle: row.FeedbackTitle,
			Relationship: row.Relationship, IsPrimary: row.IsPrimary, CreatedAt: row.CreatedAt})
	}
	return links, nil
}

func (r *Repo) LinkStory(ctx context.Context, input feedback.CoreStoryLinkInput) (feedback.CoreStoryLink, error) {
	row, err := r.queries.LinkFeedbackStory(ctx, feedbacksql.LinkFeedbackStoryParams{
		Relationship: input.Relationship, IsPrimary: input.IsPrimary, ActorID: uuidPointer(input.CreatedByUserID),
		StoryID: input.StoryID, WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, AllTeams: true,
	})
	if err != nil {
		if isPrimaryStoryConflict(err) {
			return feedback.CoreStoryLink{}, feedback.ErrAlreadyPlanned
		}
		return feedback.CoreStoryLink{}, normalizeError(err)
	}
	return feedback.CoreStoryLink{ID: row.ID, WorkspaceID: row.WorkspaceID, ItemID: row.ItemID,
		StoryID: row.StoryID, Relationship: row.Relationship, IsPrimary: row.IsPrimary,
		CreatedByUserID: valueOrZero(row.CreatedByUserID), CreatedAt: row.CreatedAt}, nil
}

func (r *Repo) LinkStoryScoped(ctx context.Context, scope feedback.CoreAccessScope, input feedback.CoreStoryLinkInput) (feedback.CoreStoryLink, error) {
	if input.WorkspaceID != scope.WorkspaceID || input.CreatedByUserID != scope.ActorID {
		return feedback.CoreStoryLink{}, feedback.ErrForbidden
	}
	row, err := r.queries.LinkFeedbackStory(ctx, feedbacksql.LinkFeedbackStoryParams{
		Relationship: input.Relationship, IsPrimary: input.IsPrimary, ActorID: uuidPointer(scope.ActorID),
		StoryID: input.StoryID, WorkspaceID: scope.WorkspaceID, ItemID: input.ItemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		if isPrimaryStoryConflict(err) {
			return feedback.CoreStoryLink{}, feedback.ErrAlreadyPlanned
		}
		return feedback.CoreStoryLink{}, normalizeError(err)
	}
	return feedback.CoreStoryLink{ID: row.ID, WorkspaceID: row.WorkspaceID, ItemID: row.ItemID,
		StoryID: row.StoryID, Relationship: row.Relationship, IsPrimary: row.IsPrimary,
		CreatedByUserID: valueOrZero(row.CreatedByUserID), CreatedAt: row.CreatedAt}, nil
}

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	statusID, err := r.queries.FindFirstFeedbackStatusByCategory(ctx, feedbacksql.FindFirstFeedbackStatusByCategoryParams{TeamID: teamID, Category: &category})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &statusID, nil
}

func (r *Repo) GetStatusCategory(ctx context.Context, teamID, statusID uuid.UUID) (string, error) {
	category, err := r.queries.GetFeedbackStatusCategory(ctx, feedbacksql.GetFeedbackStatusCategoryParams{TeamID: teamID, StatusID: statusID})
	return category, normalizeError(err)
}

func (r *Repo) ListPrimaryStoryItems(ctx context.Context, workspaceID, storyID uuid.UUID) ([]feedback.CoreItem, error) {
	rows, err := r.queries.ListPrimaryStoryFeedbackItems(ctx, feedbacksql.ListPrimaryStoryFeedbackItemsParams{WorkspaceID: workspaceID, StoryID: storyID})
	if err != nil {
		return nil, err
	}
	items := make([]feedback.CoreItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, feedback.CoreItem{ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
			ContributorID: row.ContributorID, AuthorID: valueOrZero(row.AuthorID), Title: row.Title, Slug: row.Slug,
			Status: row.Status, RoadmapSummary: row.RoadmapSummary, UpdatedAt: row.UpdatedAt})
	}
	return items, nil
}

func isPrimaryStoryConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation &&
		pgErr.ConstraintName == "feedback_story_links_one_primary_per_item"
}

func valueOrZero[T any](value *T) (zero T) {
	if value == nil {
		return zero
	}
	return *value
}
