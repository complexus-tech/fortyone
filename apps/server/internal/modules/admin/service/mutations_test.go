package admin

import (
	"context"
	"testing"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateWorkspaceTrialUsesDeterministicUTCCommand(t *testing.T) {
	actorID, workspaceID := uuid.New(), uuid.New()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	trialEnd := now.Add(10 * 24 * time.Hour)
	repository := &adminTestRepository{workspace: admindomain.WorkspaceOverview{
		Workspace: admindomain.WorkspaceSummary{ID: workspaceID},
	}}
	service := New(repository, WithNow(func() time.Time { return now }))

	_, err := service.UpdateWorkspaceTrial(context.Background(), actorID, workspaceID, UpdateWorkspaceTrialInput{
		TrialEndsOn: trialEnd, Reason: "  sales extension  ",
	})

	require.NoError(t, err)
	require.Equal(t, actorID, repository.trialCommand.ActorID)
	require.Equal(t, now.UTC(), repository.trialCommand.Now)
	require.Equal(t, trialEnd.UTC(), repository.trialCommand.TrialEndsOn)
	require.Equal(t, "sales extension", repository.trialCommand.Reason)
}

func TestUpdateWorkspaceTrialRejectsPastTimeBeforeRepository(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	repository := &adminTestRepository{}
	service := New(repository, WithNow(func() time.Time { return now }))

	_, err := service.UpdateWorkspaceTrial(context.Background(), uuid.New(), uuid.New(), UpdateWorkspaceTrialInput{
		TrialEndsOn: now, Reason: "extension",
	})

	require.ErrorIs(t, err, ErrInvalidTrialEndsOn)
	require.Equal(t, uuid.Nil, repository.trialCommand.ActorID)
}

func TestUpdateUserStatePreservesTriStatePatch(t *testing.T) {
	actorID, userID := uuid.New(), uuid.New()
	repository := &adminTestRepository{user: admindomain.UserOverview{User: admindomain.UserSummary{ID: userID}}}
	service := New(repository)

	_, err := service.UpdateUserState(context.Background(), actorID, userID, UpdateUserStateInput{
		Patch:  admindomain.UserStatePatch{IsActive: platformpatch.Set(false)},
		Reason: " security review ",
	})

	require.NoError(t, err)
	active, specified := repository.userStateCommand.Patch.IsActive.Value()
	require.True(t, specified)
	require.NotNil(t, active)
	require.False(t, *active)
	_, internalSpecified := repository.userStateCommand.Patch.IsInternal.Value()
	require.False(t, internalSpecified)
}

func TestUpdateUserStateRejectsExplicitNull(t *testing.T) {
	repository := &adminTestRepository{}
	service := New(repository)

	_, err := service.UpdateUserState(context.Background(), uuid.New(), uuid.New(), UpdateUserStateInput{
		Patch:  admindomain.UserStatePatch{IsActive: platformpatch.Clear[bool]()},
		Reason: "security review",
	})

	require.ErrorIs(t, err, ErrInvalidAdminAction)
	require.Equal(t, uuid.Nil, repository.userStateCommand.ActorID)
}

func TestCreateAdminNoteNormalizesWorkspaceTarget(t *testing.T) {
	actorID, workspaceID := uuid.New(), uuid.New()
	repository := &adminTestRepository{note: admindomain.AdminNote{ID: uuid.New()}}
	service := New(repository)

	_, err := service.CreateAdminNote(context.Background(), actorID, CreateAdminNoteInput{
		TargetType: " WORKSPACE ", TargetID: workspaceID, Body: "  Follow up  ",
	})

	require.NoError(t, err)
	require.Equal(t, admindomain.TargetWorkspace, repository.createNoteCommand.TargetType)
	require.Equal(t, workspaceID, *repository.createNoteCommand.WorkspaceID)
	require.Equal(t, "Follow up", repository.createNoteCommand.Body)
}
