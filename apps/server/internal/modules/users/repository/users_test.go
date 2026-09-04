package usersrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeUserQueries struct {
	usersql.Querier
	getActiveByID             func(context.Context, usersql.GetActiveUserByIDParams) (usersql.GetActiveUserByIDRow, error)
	resolveSession            func(context.Context, usersql.GetActiveBrowserSessionVersionParams) (int64, error)
	listWorkspace             func(context.Context, usersql.ListWorkspaceUsersParams) ([]usersql.ListWorkspaceUsersRow, error)
	createUser                func(context.Context, usersql.CreateUserParams) (usersql.CreateUserRow, error)
	updateUser                func(context.Context, usersql.UpdateActiveUserParams) (usersql.UpdateActiveUserRow, error)
	updateWorkspace           func(context.Context, usersql.UpdateLastUsedWorkspaceForMemberParams) (int64, error)
	updateMemory              func(context.Context, usersql.UpdateUserMemoryForOwnerParams) (int64, error)
	upsertPreferences         func(context.Context, usersql.UpsertAutomationPreferencesForMemberParams) (int64, error)
	getOnboardingTourProgress func(
		context.Context,
		usersql.GetOrCreateOnboardingTourProgressForUserParams,
	) (usersql.UserOnboardingTourProgressGlobal, error)
	upsertOnboardingTourProgress func(
		context.Context,
		usersql.UpsertOnboardingTourProgressForUserParams,
	) (usersql.UserOnboardingTourProgressGlobal, error)
	verificationLock func(context.Context, usersql.AcquireVerificationTokenIssueLockParams) error
	countIssues      func(context.Context, usersql.CountRecentVerificationTokenIssuesParams) (int32, error)
	createToken      func(context.Context, usersql.CreateVerificationTokenParams) (usersql.CreateVerificationTokenRow, error)
	reactivateUser   func(context.Context, usersql.ReactivateUserForVerifiedSignInParams) (usersql.ReactivateUserForVerifiedSignInRow, error)
	deactivateUser   func(context.Context, usersql.DeactivateUserParams) (int64, error)
}

func (fake fakeUserQueries) GetActiveUserByID(
	ctx context.Context,
	params usersql.GetActiveUserByIDParams,
) (usersql.GetActiveUserByIDRow, error) {
	return fake.getActiveByID(ctx, params)
}

func (fake fakeUserQueries) GetActiveBrowserSessionVersion(
	ctx context.Context,
	params usersql.GetActiveBrowserSessionVersionParams,
) (int64, error) {
	return fake.resolveSession(ctx, params)
}

func TestResolveActiveBrowserSessionVersionMapsAccountState(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	databaseErr := errors.New("database unavailable")
	tests := []struct {
		name       string
		result     int64
		queryErr   error
		wantActive bool
		wantErr    error
	}{
		{name: "active", result: 9, wantActive: true},
		{name: "inactive or unknown", queryErr: pgx.ErrNoRows},
		{name: "database failure", queryErr: databaseErr, wantErr: databaseErr},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := newWithQueries(fakeUserQueries{
				resolveSession: func(_ context.Context, params usersql.GetActiveBrowserSessionVersionParams) (int64, error) {
					if params.UserID != userID {
						t.Fatalf("user id = %s, want %s", params.UserID, userID)
					}
					return test.result, test.queryErr
				},
			})

			version, active, err := repository.ResolveActiveBrowserSessionVersion(context.Background(), userID)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("resolve error = %v, want %v", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve browser session: %v", err)
			}
			if active != test.wantActive {
				t.Fatalf("active = %t, want %t", active, test.wantActive)
			}
			if test.wantActive && version != test.result {
				t.Fatalf("version = %d, want %d", version, test.result)
			}
		})
	}
}

func (fake fakeUserQueries) ListWorkspaceUsers(
	ctx context.Context,
	params usersql.ListWorkspaceUsersParams,
) ([]usersql.ListWorkspaceUsersRow, error) {
	return fake.listWorkspace(ctx, params)
}

func (fake fakeUserQueries) CreateUser(
	ctx context.Context,
	params usersql.CreateUserParams,
) (usersql.CreateUserRow, error) {
	return fake.createUser(ctx, params)
}

func (fake fakeUserQueries) UpdateActiveUser(
	ctx context.Context,
	params usersql.UpdateActiveUserParams,
) (usersql.UpdateActiveUserRow, error) {
	return fake.updateUser(ctx, params)
}

func (fake fakeUserQueries) UpdateLastUsedWorkspaceForMember(
	ctx context.Context,
	params usersql.UpdateLastUsedWorkspaceForMemberParams,
) (int64, error) {
	return fake.updateWorkspace(ctx, params)
}

func (fake fakeUserQueries) UpdateUserMemoryForOwner(
	ctx context.Context,
	params usersql.UpdateUserMemoryForOwnerParams,
) (int64, error) {
	return fake.updateMemory(ctx, params)
}

func (fake fakeUserQueries) UpsertAutomationPreferencesForMember(
	ctx context.Context,
	params usersql.UpsertAutomationPreferencesForMemberParams,
) (int64, error) {
	return fake.upsertPreferences(ctx, params)
}

func (fake fakeUserQueries) GetOrCreateOnboardingTourProgressForUser(
	ctx context.Context,
	params usersql.GetOrCreateOnboardingTourProgressForUserParams,
) (usersql.UserOnboardingTourProgressGlobal, error) {
	return fake.getOnboardingTourProgress(ctx, params)
}

func (fake fakeUserQueries) UpsertOnboardingTourProgressForUser(
	ctx context.Context,
	params usersql.UpsertOnboardingTourProgressForUserParams,
) (usersql.UserOnboardingTourProgressGlobal, error) {
	return fake.upsertOnboardingTourProgress(ctx, params)
}

func (fake fakeUserQueries) AcquireVerificationTokenIssueLock(
	ctx context.Context,
	params usersql.AcquireVerificationTokenIssueLockParams,
) error {
	return fake.verificationLock(ctx, params)
}

func (fake fakeUserQueries) CountRecentVerificationTokenIssues(
	ctx context.Context,
	params usersql.CountRecentVerificationTokenIssuesParams,
) (int32, error) {
	return fake.countIssues(ctx, params)
}

func (fake fakeUserQueries) CreateVerificationToken(
	ctx context.Context,
	params usersql.CreateVerificationTokenParams,
) (usersql.CreateVerificationTokenRow, error) {
	return fake.createToken(ctx, params)
}

func (fake fakeUserQueries) ReactivateUserForVerifiedSignIn(
	ctx context.Context,
	params usersql.ReactivateUserForVerifiedSignInParams,
) (usersql.ReactivateUserForVerifiedSignInRow, error) {
	return fake.reactivateUser(ctx, params)
}

func (fake fakeUserQueries) DeactivateUser(
	ctx context.Context,
	params usersql.DeactivateUserParams,
) (int64, error) {
	return fake.deactivateUser(ctx, params)
}

func TestVerifiedSignInReactivationUsesTypedPolicyGateAndUTCInstant(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	location := time.FixedZone("CAT", 2*60*60)
	signedInAt := time.Date(2026, time.August, 28, 14, 45, 0, 0, location)
	writtenAt := signedInAt.UTC()
	repository := newWithQueries(fakeUserQueries{
		reactivateUser: func(
			_ context.Context,
			params usersql.ReactivateUserForVerifiedSignInParams,
		) (usersql.ReactivateUserForVerifiedSignInRow, error) {
			if params.UserID != userID || !params.SignedInAt.Equal(writtenAt) || params.SignedInAt.Location() != time.UTC {
				t.Fatalf("reactivation params = %#v", params)
			}
			return usersql.ReactivateUserForVerifiedSignInRow{
				UserID: userID, Email: "ada@example.com", IsActive: true,
				LastLoginAt: &writtenAt, CreatedAt: writtenAt.Add(-time.Hour), UpdatedAt: writtenAt,
			}, nil
		},
	})

	user, err := repository.ReactivateUserForVerifiedSignIn(context.Background(), users.VerifiedSignInReactivation{
		UserID: userID, SignedInAt: signedInAt,
	})
	if err != nil {
		t.Fatalf("reactivate user: %v", err)
	}
	if !user.IsActive || user.ID != userID || !user.LastLoginAt.Equal(writtenAt) {
		t.Fatalf("reactivated user = %#v", user)
	}
}

func TestVerifiedSignInReactivationHidesBlockedAndMissingAccountState(t *testing.T) {
	t.Parallel()

	repository := newWithQueries(fakeUserQueries{
		reactivateUser: func(
			context.Context,
			usersql.ReactivateUserForVerifiedSignInParams,
		) (usersql.ReactivateUserForVerifiedSignInRow, error) {
			return usersql.ReactivateUserForVerifiedSignInRow{}, pgx.ErrNoRows
		},
	})

	_, err := repository.ReactivateUserForVerifiedSignIn(context.Background(), users.VerifiedSignInReactivation{
		UserID: uuid.New(), SignedInAt: time.Now(),
	})
	if !errors.Is(err, users.ErrInvalidCredentials) {
		t.Fatalf("reactivation error = %v, want generic ErrInvalidCredentials", err)
	}
}

func TestSelfDeactivationUsesApplicationUTCInstant(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	location := time.FixedZone("CAT", 2*60*60)
	deactivatedAt := time.Date(2026, time.August, 28, 15, 0, 0, 0, location)
	repository := newWithQueries(fakeUserQueries{
		deactivateUser: func(_ context.Context, params usersql.DeactivateUserParams) (int64, error) {
			if params.UserID != userID || !params.DeactivatedAt.Equal(deactivatedAt.UTC()) || params.DeactivatedAt.Location() != time.UTC {
				t.Fatalf("deactivation params = %#v", params)
			}
			return 1, nil
		},
	})

	if err := repository.DeleteUser(context.Background(), userID, deactivatedAt); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
}

func TestListMapsTypedWorkspaceTeamAndPaginationScope(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	userID := uuid.New()
	activityAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	repository := newWithQueries(fakeUserQueries{
		listWorkspace: func(
			_ context.Context,
			params usersql.ListWorkspaceUsersParams,
		) ([]usersql.ListWorkspaceUsersRow, error) {
			if params.WorkspaceID != workspaceID || params.TeamID != teamID || !params.HasTeam {
				t.Fatalf("scope params = %#v", params)
			}
			if params.Search != "Platform" || params.PageLimit != 25 || params.PageOffset != 50 {
				t.Fatalf("filter params = %#v", params)
			}
			return []usersql.ListWorkspaceUsersRow{{
				UserID: userID, Username: "ada", Email: "ada@example.com", IsActive: true,
				Timezone: "Africa/Harare", Role: "admin", TeamAiRoleTitle: "Lead",
				HasLastStoryActivity: true, LastStoryActivityAt: activityAt,
			}}, nil
		},
	})

	result, err := repository.List(context.Background(), workspaceID, users.CoreListUsersFilter{
		TeamID: &teamID,
		Search: " Platform ",
		Limit:  25,
		Offset: 50,
	})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(result) != 1 || result[0].ID != userID || result[0].Role == nil || *result[0].Role != "admin" {
		t.Fatalf("mapped users = %#v", result)
	}
	if result[0].LastStoryActivityAt == nil || !result[0].LastStoryActivityAt.Equal(activityAt) {
		t.Fatalf("last activity = %v, want %v", result[0].LastStoryActivityAt, activityAt)
	}
}

func TestGetActiveUserMapsHiddenOrInactiveAccountToNotFound(t *testing.T) {
	t.Parallel()

	repository := newWithQueries(fakeUserQueries{
		getActiveByID: func(context.Context, usersql.GetActiveUserByIDParams) (usersql.GetActiveUserByIDRow, error) {
			return usersql.GetActiveUserByIDRow{}, pgx.ErrNoRows
		},
	})
	_, err := repository.GetUser(context.Background(), uuid.New())
	if !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("get error = %v, want ErrNotFound", err)
	}
}

func TestCreateMapsUniqueEmailToDomainConflict(t *testing.T) {
	t.Parallel()

	repository := newWithQueries(fakeUserQueries{
		createUser: func(context.Context, usersql.CreateUserParams) (usersql.CreateUserRow, error) {
			return usersql.CreateUserRow{}, &pgconn.PgError{Code: "23505"}
		},
	})
	_, err := repository.Create(context.Background(), users.CoreUser{Email: "taken@example.com"})
	if !errors.Is(err, users.ErrEmailTaken) {
		t.Fatalf("create error = %v, want ErrEmailTaken", err)
	}
}

func TestUpdateMapsOnlyExplicitProfileAndScheduleFields(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	fullName := "Ada Lovelace"
	start := 540
	end := 1020
	repository := newWithQueries(fakeUserQueries{
		updateUser: func(
			_ context.Context,
			params usersql.UpdateActiveUserParams,
		) (usersql.UpdateActiveUserRow, error) {
			if params.UserID != userID || !params.SetFullName || params.FullName != fullName {
				t.Fatalf("profile params = %#v", params)
			}
			if params.SetUsername || params.SetAvatarURL || params.SetTimezone {
				t.Fatalf("unset fields unexpectedly enabled: %#v", params)
			}
			if !params.SetWorkSchedule || len(params.WorkingDays) != 5 || params.WorkingDays[0] != 1 {
				t.Fatalf("schedule params = %#v", params)
			}
			if params.WorkingStartMinute == nil || *params.WorkingStartMinute != int16(start) ||
				params.WorkingEndMinute == nil || *params.WorkingEndMinute != int16(end) {
				t.Fatalf("schedule minutes = %#v", params)
			}
			return usersql.UpdateActiveUserRow{UserID: userID, FullName: &fullName, IsActive: true}, nil
		},
	})

	updated, err := repository.UpdateUser(context.Background(), userID, users.CoreUpdateUser{
		FullName: &fullName,
		WorkSchedule: &users.CoreWorkScheduleOverride{
			WorkingDays:        []int{1, 2, 3, 4, 5},
			WorkingStartMinute: &start,
			WorkingEndMinute:   &end,
		},
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.ID != userID || updated.FullName != fullName {
		t.Fatalf("updated user = %#v", updated)
	}
}

func TestPrivateMemoryMutationUsesOwnerAndWorkspaceScope(t *testing.T) {
	t.Parallel()

	memoryID := uuid.New()
	userID := uuid.New()
	workspaceID := uuid.New()
	content := "private"
	repository := newWithQueries(fakeUserQueries{
		updateMemory: func(
			_ context.Context,
			params usersql.UpdateUserMemoryForOwnerParams,
		) (int64, error) {
			if params.MemoryID != memoryID || params.UserID != userID || params.WorkspaceID != workspaceID {
				t.Fatalf("memory scope = %#v", params)
			}
			return 0, nil
		},
	})
	err := repository.UpdateUserMemory(context.Background(), memoryID, users.UserMemoryScope{
		UserID: userID, WorkspaceID: workspaceID,
	}, users.UpdateUserMemoryItem{Content: &content})
	if !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("memory error = %v, want ErrMemoryNotFound", err)
	}
}

func TestAutomationPreferencesUseTypedPatchPresence(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	workspaceID := uuid.New()
	disabled := false
	repository := newWithQueries(fakeUserQueries{
		upsertPreferences: func(
			_ context.Context,
			params usersql.UpsertAutomationPreferencesForMemberParams,
		) (int64, error) {
			if params.UserID != userID || params.WorkspaceID != workspaceID {
				t.Fatalf("preference scope = %#v", params)
			}
			if !params.SetAutoScheduling || params.AutoScheduling || params.SetAutoAssignSelf {
				t.Fatalf("preference patch = %#v", params)
			}
			return 1, nil
		},
	})
	if err := repository.UpdateAutomationPreferences(
		context.Background(),
		userID,
		workspaceID,
		users.CoreUpdateAutomationPreferences{AutoScheduling: &disabled},
	); err != nil {
		t.Fatalf("update preferences: %v", err)
	}
}

func TestOnboardingTourProgressUsesTypedUserScopedUpsert(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	status := users.CoreOnboardingTourStatusCompleted
	writtenAt := time.Date(2026, time.August, 31, 7, 30, 0, 0, time.UTC)
	storedStepIDs := []string{"create-task"}
	storedActionIDs := []string{"create-task"}
	repository := newWithQueries(fakeUserQueries{
		upsertOnboardingTourProgress: func(
			_ context.Context,
			params usersql.UpsertOnboardingTourProgressForUserParams,
		) (usersql.UserOnboardingTourProgressGlobal, error) {
			if params.UserID != userID || params.TourKey != "workspace-getting-started" || params.TourVersion != "v1" {
				t.Fatalf("onboarding scope = %#v", params)
			}
			if !params.SetStatus || params.Status != string(status) {
				t.Fatalf("onboarding status patch = %#v", params)
			}
			if len(params.CompletedStepIds) != 1 || params.CompletedStepIds[0] != "create-task" ||
				len(params.CompletedActionIds) != 1 || params.CompletedActionIds[0] != "create-task" {
				t.Fatalf("onboarding completion patch = %#v", params)
			}
			return usersql.UserOnboardingTourProgressGlobal{
				UserID:             userID,
				TourKey:            params.TourKey,
				TourVersion:        params.TourVersion,
				CompletedStepIds:   storedStepIDs,
				CompletedActionIds: storedActionIDs,
				Status:             params.Status,
				CreatedAt:          writtenAt,
				UpdatedAt:          writtenAt,
			}, nil
		},
	})

	progress, err := repository.UpdateOnboardingTourProgress(context.Background(), userID, users.CoreUpdateOnboardingTourProgress{
		OnboardingTourScope: users.CoreOnboardingTourScope{
			TourKey:     "workspace-getting-started",
			TourVersion: "v1",
		},
		CompletedStepIDs:   []string{"create-task"},
		CompletedActionIDs: []string{"create-task"},
		Status:             &status,
	})
	if err != nil {
		t.Fatalf("update onboarding progress: %v", err)
	}
	if progress.UserID != userID || progress.Status != status || !progress.UpdatedAt.Equal(writtenAt) {
		t.Fatalf("mapped onboarding progress = %#v", progress)
	}

	progress.CompletedStepIDs[0] = "mutated"
	if storedStepIDs[0] != "create-task" {
		t.Fatalf("mapped progress aliases SQLC row: %#v", storedStepIDs)
	}
}

func TestOnboardingTourProgressHidesInactiveOrMissingUser(t *testing.T) {
	t.Parallel()

	repository := newWithQueries(fakeUserQueries{
		getOnboardingTourProgress: func(
			context.Context,
			usersql.GetOrCreateOnboardingTourProgressForUserParams,
		) (usersql.UserOnboardingTourProgressGlobal, error) {
			return usersql.UserOnboardingTourProgressGlobal{}, pgx.ErrNoRows
		},
	})

	_, err := repository.GetOnboardingTourProgress(
		context.Background(),
		uuid.New(),
		users.CoreOnboardingTourScope{TourKey: "workspace-getting-started", TourVersion: "v1"},
	)
	if !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("get onboarding progress error = %v, want ErrNotFound", err)
	}
}

type transactionStub struct{ pgx.Tx }

func TestCallerOwnedTransactionIsBoundWithoutOwningCommit(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	workspaceID := uuid.New()
	queries := fakeUserQueries{
		updateWorkspace: func(
			_ context.Context,
			params usersql.UpdateLastUsedWorkspaceForMemberParams,
		) (int64, error) {
			if params.UserID != userID || params.WorkspaceID == nil || *params.WorkspaceID != workspaceID {
				t.Fatalf("workspace params = %#v", params)
			}
			return 1, nil
		},
	}
	repository := newWithQueries(queries)
	bindCalls := 0
	repository.bindTransaction = func(tx pgx.Tx) usersql.Querier {
		bindCalls++
		return queries
	}

	transaction, err := repository.BindWorkspaceTransaction(transactionStub{})
	if err != nil {
		t.Fatalf("bind workspace transaction: %v", err)
	}
	if err := transaction.UpdateLastUsedWorkspace(context.Background(), userID, workspaceID); err != nil {
		t.Fatalf("update workspace in caller transaction: %v", err)
	}
	if bindCalls != 1 {
		t.Fatalf("transaction bind calls = %d, want 1", bindCalls)
	}
}

func TestVerificationTokenUniqueViolationMapsToCollision(t *testing.T) {
	t.Parallel()

	queries := fakeUserQueries{
		verificationLock: func(
			_ context.Context,
			params usersql.AcquireVerificationTokenIssueLockParams,
		) error {
			if params.Email != "person@example.com" || params.TokenType != users.TokenTypeLogin {
				t.Fatalf("verification issue lock params = %#v", params)
			}
			return nil
		},
		countIssues: func(context.Context, usersql.CountRecentVerificationTokenIssuesParams) (int32, error) {
			return 0, nil
		},
		createToken: func(context.Context, usersql.CreateVerificationTokenParams) (usersql.CreateVerificationTokenRow, error) {
			return usersql.CreateVerificationTokenRow{}, &pgconn.PgError{Code: "23505"}
		},
	}
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(usersql.Querier) error) error {
		return operation(queries)
	}

	_, err := repository.CreateVerificationToken(context.Background(), users.NewVerificationToken{
		Email: "person@example.com", TokenType: users.TokenTypeLogin,
		TokenDigest: make([]byte, 32), TokenKeyID: "v1", TokenVersion: 1,
		ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(),
		RateLimitSince: time.Now().Add(-time.Hour), MaximumIssues: 3,
	})
	if !errors.Is(err, users.ErrTokenCollision) {
		t.Fatalf("create token error = %v, want ErrTokenCollision", err)
	}
}
