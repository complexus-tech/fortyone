package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const userNotWorkspaceMemberMessage = "user not found or not a member of workspace"

var errUserDeactivationTimeRequired = errors.New("user deactivation time is required")

func (r *repo) GetUser(ctx context.Context, userID uuid.UUID) (usersdomain.User, error) {
	row, err := r.queries.GetActiveUserByID(ctx, usersql.GetActiveUserByIDParams{UserID: userID})
	if err != nil {
		return usersdomain.User{}, mapUserNotFound("get active user", err)
	}
	return mapActiveUserByID(row), nil
}

func (r *repo) GetUserByEmail(ctx context.Context, email string) (usersdomain.User, error) {
	row, err := r.queries.GetActiveUserByEmail(ctx, usersql.GetActiveUserByEmailParams{Email: email})
	if err != nil {
		return usersdomain.User{}, mapUserNotFound("get active user by email", err)
	}
	return mapActiveUserByEmail(row), nil
}

func (r *repo) GetUserByEmailAnyStatus(ctx context.Context, email string) (usersdomain.User, error) {
	row, err := r.queries.GetUserByEmailAnyStatus(ctx, usersql.GetUserByEmailAnyStatusParams{Email: email})
	if err != nil {
		return usersdomain.User{}, mapUserNotFound("get user by email", err)
	}
	return mapUserByEmailAnyStatus(row), nil
}

func (r *repo) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]usersdomain.User, error) {
	if len(userIDs) == 0 {
		return []usersdomain.User{}, nil
	}

	rows, err := r.queries.ListUsersByIDs(ctx, usersql.ListUsersByIDsParams{UserIds: userIDs})
	if err != nil {
		return nil, fmt.Errorf("list users by IDs: %w", err)
	}
	return mapUsersByIDs(rows), nil
}

func (r *repo) List(
	ctx context.Context,
	workspaceID uuid.UUID,
	filter usersdomain.ListUsersFilter,
) ([]usersdomain.User, error) {
	pageLimit, pageOffset, err := userPaginationParams(filter)
	if err != nil {
		return nil, err
	}

	teamID := uuid.Nil
	hasTeam := filter.TeamID != nil
	if hasTeam {
		teamID = *filter.TeamID
	}

	rows, err := r.queries.ListWorkspaceUsers(ctx, usersql.ListWorkspaceUsersParams{
		HasTeam:     hasTeam,
		TeamID:      teamID,
		WorkspaceID: workspaceID,
		Search:      strings.TrimSpace(filter.Search),
		PageLimit:   pageLimit,
		PageOffset:  pageOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("list workspace users: %w", err)
	}
	return mapWorkspaceUsers(rows), nil
}

func (r *repo) Create(ctx context.Context, user usersdomain.User) (usersdomain.User, error) {
	lastLoginAt := user.LastLoginAt
	row, err := r.queries.CreateUser(ctx, usersql.CreateUserParams{
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		AvatarURL:   user.AvatarURL,
		Timezone:    user.Timezone,
		LastLoginAt: &lastLoginAt,
	})
	if err != nil {
		if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
			return usersdomain.User{}, usersdomain.ErrEmailTaken
		}
		return usersdomain.User{}, fmt.Errorf("create user: %w", err)
	}
	return mapCreatedUser(row), nil
}

func (r *repo) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	updates usersdomain.UpdateUser,
) (usersdomain.User, error) {
	if !hasUserUpdates(updates) {
		return r.GetUser(ctx, userID)
	}

	params, err := updateActiveUserParams(userID, updates)
	if err != nil {
		return usersdomain.User{}, err
	}
	row, err := r.queries.UpdateActiveUser(ctx, params)
	if err != nil {
		return usersdomain.User{}, mapUserNotFound("update active user", err)
	}
	return mapUpdatedUser(row), nil
}

func (r *repo) ReactivateUserForVerifiedSignIn(
	ctx context.Context,
	input usersdomain.VerifiedSignInReactivation,
) (usersdomain.User, error) {
	if input.UserID == uuid.Nil || input.SignedInAt.IsZero() {
		return usersdomain.User{}, usersdomain.ErrInvalidCredentials
	}
	row, err := r.queries.ReactivateUserForVerifiedSignIn(
		ctx,
		usersql.ReactivateUserForVerifiedSignInParams{
			UserID: input.UserID, SignedInAt: input.SignedInAt.UTC(),
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return usersdomain.User{}, usersdomain.ErrInvalidCredentials
	}
	if err != nil {
		return usersdomain.User{}, fmt.Errorf("reactivate user for verified sign-in: %w", err)
	}
	return mapVerifiedSignInReactivatedUser(row), nil
}

func (r *repo) DeleteUser(ctx context.Context, userID uuid.UUID, deactivatedAt time.Time) error {
	if userID == uuid.Nil {
		return usersdomain.ErrNotFound
	}
	if deactivatedAt.IsZero() {
		return errUserDeactivationTimeRequired
	}
	rows, err := r.queries.DeactivateUser(ctx, usersql.DeactivateUserParams{
		UserID: userID, DeactivatedAt: deactivatedAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	if rows == 0 {
		return usersdomain.ErrNotFound
	}
	return nil
}

func (r *repo) UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return updateLastUsedWorkspace(ctx, r.queries, userID, workspaceID)
}

func updateLastUsedWorkspace(
	ctx context.Context,
	queries usersql.Querier,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) error {
	rows, err := queries.UpdateLastUsedWorkspaceForMember(ctx, usersql.UpdateLastUsedWorkspaceForMemberParams{
		WorkspaceID: &workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("update user's last used workspace: %w", err)
	}
	if rows == 0 {
		return errors.New(userNotWorkspaceMemberMessage)
	}
	return nil
}

func mapUserNotFound(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return usersdomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func userPaginationParams(filter usersdomain.ListUsersFilter) (int32, int32, error) {
	if filter.Limit <= 0 {
		return 0, 0, nil
	}
	if filter.Limit > math.MaxInt32 {
		return 0, 0, fmt.Errorf("user page limit exceeds database range: %d", filter.Limit)
	}
	if filter.Offset < 0 {
		return 0, 0, fmt.Errorf("user page offset cannot be negative: %d", filter.Offset)
	}
	if filter.Offset > math.MaxInt32 {
		return 0, 0, fmt.Errorf("user page offset exceeds database range: %d", filter.Offset)
	}
	return int32(filter.Limit), int32(filter.Offset), nil
}

func hasUserUpdates(updates usersdomain.UpdateUser) bool {
	return updates.Username != nil ||
		updates.FullName != nil ||
		updates.AvatarURL != nil ||
		updates.HasSeenWalkthrough != nil ||
		updates.Timezone != nil ||
		updates.WorkSchedule != nil
}

func updateActiveUserParams(
	userID uuid.UUID,
	updates usersdomain.UpdateUser,
) (usersql.UpdateActiveUserParams, error) {
	params := usersql.UpdateActiveUserParams{UserID: userID}
	if updates.Username != nil {
		params.SetUsername = true
		params.Username = *updates.Username
	}
	if updates.FullName != nil {
		params.SetFullName = true
		params.FullName = *updates.FullName
	}
	if updates.AvatarURL != nil {
		params.SetAvatarURL = true
		params.AvatarURL = *updates.AvatarURL
	}
	if updates.HasSeenWalkthrough != nil {
		params.SetHasSeenWalkthrough = true
		params.HasSeenWalkthrough = *updates.HasSeenWalkthrough
	}
	if updates.Timezone != nil {
		params.SetTimezone = true
		params.Timezone = *updates.Timezone
	}
	if updates.WorkSchedule == nil {
		return params, nil
	}

	workingDays, err := intsToInt16s(updates.WorkSchedule.WorkingDays)
	if err != nil {
		return usersql.UpdateActiveUserParams{}, err
	}
	workingStartMinute, err := intPointerToInt16(updates.WorkSchedule.WorkingStartMinute)
	if err != nil {
		return usersql.UpdateActiveUserParams{}, err
	}
	workingEndMinute, err := intPointerToInt16(updates.WorkSchedule.WorkingEndMinute)
	if err != nil {
		return usersql.UpdateActiveUserParams{}, err
	}
	params.SetWorkSchedule = true
	params.WorkingDays = workingDays
	params.WorkingStartMinute = workingStartMinute
	params.WorkingEndMinute = workingEndMinute
	return params, nil
}

func intsToInt16s(values []int) ([]int16, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]int16, len(values))
	for index, value := range values {
		if value < math.MinInt16 || value > math.MaxInt16 {
			return nil, fmt.Errorf("work schedule value exceeds database range: %d", value)
		}
		result[index] = int16(value)
	}
	return result, nil
}

func intPointerToInt16(value *int) (*int16, error) {
	if value == nil {
		return nil, nil
	}
	if *value < math.MinInt16 || *value > math.MaxInt16 {
		return nil, fmt.Errorf("work schedule value exceeds database range: %d", *value)
	}
	converted := int16(*value)
	return &converted, nil
}
