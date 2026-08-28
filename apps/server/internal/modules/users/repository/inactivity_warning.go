package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListUserInactivityWarningCandidates returns one stable, bounded page of
// active non-system accounts that remain below the supplied inactivity cutoff.
func (r *repo) ListUserInactivityWarningCandidates(
	ctx context.Context,
	query usersdomain.InactivityWarningQuery,
) ([]usersdomain.InactivityWarningCandidate, error) {
	if r == nil || r.queries == nil {
		return nil, errUserMaintenanceUnavailable
	}
	if err := validateUserInactivityWarningQuery(query); err != nil {
		return nil, err
	}
	databaseBatchSize, err := safecast.Int32(query.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("validate user inactivity warning batch size: %w", err)
	}

	inactiveBefore := query.InactiveBefore.UTC()
	var afterLastLoginAt *time.Time
	var afterUserID uuid.UUID
	if query.Cursor.Valid {
		cursorTime := query.Cursor.LastLoginAt.UTC()
		afterLastLoginAt = &cursorTime
		afterUserID = query.Cursor.UserID
	}
	rows, err := r.queries.ListUserInactivityWarningCandidates(
		ctx,
		usersql.ListUserInactivityWarningCandidatesParams{
			InactiveBefore:   &inactiveBefore,
			HasCursor:        query.Cursor.Valid,
			AfterLastLoginAt: afterLastLoginAt,
			AfterUserID:      afterUserID,
			BatchSize:        databaseBatchSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list user inactivity warning candidates: %w", err)
	}

	candidates := make([]usersdomain.InactivityWarningCandidate, len(rows))
	for index, row := range rows {
		candidate, mapErr := mapUserInactivityWarningCandidate(
			row.UserID,
			row.Email,
			row.FullName,
			row.LastLoginAt,
		)
		if mapErr != nil {
			return nil, fmt.Errorf("map user inactivity warning candidate %s: %w", row.UserID, mapErr)
		}
		candidates[index] = candidate
	}
	return candidates, nil
}

// GetEligibleUserInactivityWarningCandidate rechecks current account state
// immediately before delivery. Account-lifecycle warnings intentionally do not
// consult workspace notification preferences.
func (r *repo) GetEligibleUserInactivityWarningCandidate(
	ctx context.Context,
	eligibility usersdomain.InactivityWarningEligibility,
) (usersdomain.InactivityWarningCandidate, bool, error) {
	if r == nil || r.queries == nil {
		return usersdomain.InactivityWarningCandidate{}, false, errUserMaintenanceUnavailable
	}
	if eligibility.UserID == uuid.Nil || eligibility.InactiveBefore.IsZero() {
		return usersdomain.InactivityWarningCandidate{}, false, errors.New("user warning eligibility requires user ID and inactivity cutoff")
	}

	inactiveBefore := eligibility.InactiveBefore.UTC()
	row, err := r.queries.GetEligibleUserInactivityWarningCandidate(
		ctx,
		usersql.GetEligibleUserInactivityWarningCandidateParams{
			UserID:         eligibility.UserID,
			InactiveBefore: &inactiveBefore,
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return usersdomain.InactivityWarningCandidate{}, false, nil
	}
	if err != nil {
		return usersdomain.InactivityWarningCandidate{}, false, fmt.Errorf("get eligible user inactivity warning candidate: %w", err)
	}
	candidate, err := mapUserInactivityWarningCandidate(
		row.UserID,
		row.Email,
		row.FullName,
		row.LastLoginAt,
	)
	if err != nil {
		return usersdomain.InactivityWarningCandidate{}, false, fmt.Errorf("map eligible user inactivity warning candidate: %w", err)
	}
	return candidate, true, nil
}

// RecordUserInactivityWarning records delivery only while the account still
// satisfies the same active, non-system, inactive, and unwarned predicates.
func (r *repo) RecordUserInactivityWarning(
	ctx context.Context,
	receipt usersdomain.InactivityWarningReceipt,
) (bool, error) {
	if r == nil || r.queries == nil {
		return false, errUserMaintenanceUnavailable
	}
	if receipt.UserID == uuid.Nil || receipt.InactiveBefore.IsZero() || receipt.WarningSentAt.IsZero() {
		return false, errors.New("user warning receipt requires user ID, inactivity cutoff, and delivery time")
	}
	if !receipt.InactiveBefore.Before(receipt.WarningSentAt) {
		return false, errors.New("user warning inactivity cutoff must precede delivery time")
	}

	inactiveBefore := receipt.InactiveBefore.UTC()
	rowsAffected, err := r.queries.MarkUserInactivityWarningSent(
		ctx,
		usersql.MarkUserInactivityWarningSentParams{
			WarningSentAt:  receipt.WarningSentAt.UTC(),
			UserID:         receipt.UserID,
			InactiveBefore: &inactiveBefore,
		},
	)
	if err != nil {
		return false, fmt.Errorf("record user inactivity warning: %w", err)
	}
	if rowsAffected < 0 || rowsAffected > 1 {
		return false, fmt.Errorf("record user inactivity warning: affected %d rows, want 0 or 1", rowsAffected)
	}
	return rowsAffected == 1, nil
}

func validateUserInactivityWarningQuery(query usersdomain.InactivityWarningQuery) error {
	if query.InactiveBefore.IsZero() {
		return errors.New("user warning inactivity cutoff is required")
	}
	if query.BatchSize <= 0 {
		return errors.New("user warning batch size must be positive")
	}
	if query.Cursor.Valid && (query.Cursor.LastLoginAt.IsZero() || query.Cursor.UserID == uuid.Nil) {
		return errors.New("user warning cursor requires last login time and user ID")
	}
	return nil
}

func mapUserInactivityWarningCandidate(
	userID uuid.UUID,
	email string,
	fullName string,
	lastLoginAt *time.Time,
) (usersdomain.InactivityWarningCandidate, error) {
	if userID == uuid.Nil || lastLoginAt == nil || lastLoginAt.IsZero() {
		return usersdomain.InactivityWarningCandidate{}, errors.New("candidate requires user ID and last login time")
	}
	return usersdomain.InactivityWarningCandidate{
		UserID:      userID,
		Email:       email,
		FullName:    fullName,
		LastLoginAt: lastLoginAt.UTC(),
	}, nil
}
