package usersrepository

import (
	"context"
	"errors"
	"fmt"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) GetOnboardingTourProgress(
	ctx context.Context,
	userID uuid.UUID,
	scope usersdomain.OnboardingTourScope,
) (usersdomain.OnboardingTourProgress, error) {
	row, err := r.queries.GetOrCreateOnboardingTourProgressForUser(
		ctx,
		usersql.GetOrCreateOnboardingTourProgressForUserParams{
			UserID:      userID,
			TourKey:     scope.TourKey,
			TourVersion: scope.TourVersion,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usersdomain.OnboardingTourProgress{}, usersdomain.ErrNotFound
		}
		return usersdomain.OnboardingTourProgress{}, fmt.Errorf("get or create onboarding tour progress: %w", err)
	}
	return mapOnboardingTourProgress(
		row.UserID,
		row.TourKey,
		row.TourVersion,
		row.CompletedStepIds,
		row.CompletedActionIds,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *repo) UpdateOnboardingTourProgress(
	ctx context.Context,
	userID uuid.UUID,
	updates usersdomain.UpdateOnboardingTourProgress,
) (usersdomain.OnboardingTourProgress, error) {
	status := ""
	setStatus := updates.Status != nil
	if setStatus {
		status = string(*updates.Status)
	}
	row, err := r.queries.UpsertOnboardingTourProgressForUser(
		ctx,
		usersql.UpsertOnboardingTourProgressForUserParams{
			UserID:             userID,
			TourKey:            updates.TourKey,
			TourVersion:        updates.TourVersion,
			CompletedStepIds:   updates.CompletedStepIDs,
			CompletedActionIds: updates.CompletedActionIDs,
			SetStatus:          setStatus,
			Status:             status,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usersdomain.OnboardingTourProgress{}, usersdomain.ErrNotFound
		}
		return usersdomain.OnboardingTourProgress{}, fmt.Errorf("upsert onboarding tour progress: %w", err)
	}
	return mapOnboardingTourProgress(
		row.UserID,
		row.TourKey,
		row.TourVersion,
		row.CompletedStepIds,
		row.CompletedActionIds,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}
