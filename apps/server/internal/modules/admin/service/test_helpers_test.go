package admin

import (
	"context"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
)

type adminTestRepository struct {
	adminUser            admindomain.UserSummary
	adminUserErr         error
	workspaces           admindomain.ListResult[admindomain.WorkspaceSummary]
	listWorkspacesQuery  admindomain.ListWorkspacesQuery
	workspace            admindomain.WorkspaceOverview
	workspaceMutationErr error
	trialCommand         admindomain.UpdateWorkspaceTrialCommand
	deletedCommand       admindomain.SetWorkspaceDeletedCommand
	users                admindomain.ListResult[admindomain.UserSummary]
	listUsersQuery       admindomain.ListUsersQuery
	user                 admindomain.UserOverview
	userMutationErr      error
	userStateCommand     admindomain.UpdateUserStateCommand
	sessionCommand       admindomain.RequestSessionRevocationCommand
	auditQuery           admindomain.ListAuditLogsQuery
	notesQuery           admindomain.ListAdminNotesQuery
	createNoteCommand    admindomain.CreateAdminNoteCommand
	note                 admindomain.AdminNote
	beginSyncCommand     admindomain.BeginSubscriptionSyncCommand
	finishSyncCommands   []admindomain.FinishSubscriptionSyncCommand
	syncAttempt          admindomain.SubscriptionSyncAttempt
	beginSyncErr         error
	finishSyncErr        error
}

func (repository *adminTestRepository) GetAdminUser(context.Context, uuid.UUID) (admindomain.UserSummary, error) {
	return repository.adminUser, repository.adminUserErr
}

func (repository *adminTestRepository) GetDashboardSummary(context.Context, admindomain.DashboardSummaryQuery) (admindomain.DashboardSummary, error) {
	return admindomain.DashboardSummary{}, repository.adminUserErr
}

func (repository *adminTestRepository) ListWorkspaces(_ context.Context, query admindomain.ListWorkspacesQuery) (admindomain.ListResult[admindomain.WorkspaceSummary], error) {
	repository.listWorkspacesQuery = query
	return repository.workspaces, repository.adminUserErr
}

func (repository *adminTestRepository) GetWorkspaceOverview(context.Context, admindomain.GetWorkspaceQuery) (admindomain.WorkspaceOverview, error) {
	return repository.workspace, repository.adminUserErr
}

func (repository *adminTestRepository) UpdateWorkspaceTrial(_ context.Context, command admindomain.UpdateWorkspaceTrialCommand) (admindomain.WorkspaceOverview, error) {
	repository.trialCommand = command
	if repository.workspaceMutationErr != nil {
		return admindomain.WorkspaceOverview{}, repository.workspaceMutationErr
	}
	trial := command.TrialEndsOn
	repository.workspace.Workspace.TrialEndsOn = &trial
	return repository.workspace, nil
}

func (repository *adminTestRepository) SetWorkspaceDeleted(_ context.Context, command admindomain.SetWorkspaceDeletedCommand) (admindomain.WorkspaceOverview, error) {
	repository.deletedCommand = command
	return repository.workspace, repository.workspaceMutationErr
}

func (repository *adminTestRepository) ListUsers(_ context.Context, query admindomain.ListUsersQuery) (admindomain.ListResult[admindomain.UserSummary], error) {
	repository.listUsersQuery = query
	return repository.users, repository.adminUserErr
}

func (repository *adminTestRepository) GetUserOverview(context.Context, admindomain.GetUserQuery) (admindomain.UserOverview, error) {
	return repository.user, repository.adminUserErr
}

func (repository *adminTestRepository) UpdateUserState(_ context.Context, command admindomain.UpdateUserStateCommand) (admindomain.UserOverview, error) {
	repository.userStateCommand = command
	return repository.user, repository.userMutationErr
}

func (repository *adminTestRepository) RequestSessionRevocation(_ context.Context, command admindomain.RequestSessionRevocationCommand) (admindomain.UserOverview, error) {
	repository.sessionCommand = command
	return repository.user, repository.userMutationErr
}

func (repository *adminTestRepository) ListAuditLogs(_ context.Context, query admindomain.ListAuditLogsQuery) (admindomain.ListResult[admindomain.AuditLog], error) {
	repository.auditQuery = query
	return admindomain.ListResult[admindomain.AuditLog]{}, repository.adminUserErr
}

func (repository *adminTestRepository) ListAdminNotes(_ context.Context, query admindomain.ListAdminNotesQuery) (admindomain.ListResult[admindomain.AdminNote], error) {
	repository.notesQuery = query
	return admindomain.ListResult[admindomain.AdminNote]{}, repository.adminUserErr
}

func (repository *adminTestRepository) CreateAdminNote(_ context.Context, command admindomain.CreateAdminNoteCommand) (admindomain.AdminNote, error) {
	repository.createNoteCommand = command
	return repository.note, repository.userMutationErr
}

func (repository *adminTestRepository) BeginSubscriptionSync(_ context.Context, command admindomain.BeginSubscriptionSyncCommand) (admindomain.SubscriptionSyncAttempt, admindomain.WorkspaceOverview, error) {
	repository.beginSyncCommand = command
	return repository.syncAttempt, repository.workspace, repository.beginSyncErr
}

func (repository *adminTestRepository) FinishSubscriptionSync(_ context.Context, command admindomain.FinishSubscriptionSyncCommand) (admindomain.WorkspaceOverview, error) {
	repository.finishSyncCommands = append(repository.finishSyncCommands, command)
	return repository.workspace, repository.finishSyncErr
}

type adminTestAssetResolver struct {
	profileExpiry   time.Duration
	workspaceExpiry time.Duration
}

func (resolver *adminTestAssetResolver) ResolveProfileImageURL(_ context.Context, avatar string, expiry time.Duration) (string, error) {
	resolver.profileExpiry = expiry
	return "profile:" + avatar, nil
}

func (resolver *adminTestAssetResolver) ResolveWorkspaceLogoURL(_ context.Context, logo string, expiry time.Duration) (string, error) {
	resolver.workspaceExpiry = expiry
	return "workspace:" + logo, nil
}

type adminTestSubscriptionSyncer struct {
	workspaceID uuid.UUID
	err         error
}

func (syncer *adminTestSubscriptionSyncer) SyncSubscription(_ context.Context, workspaceID uuid.UUID) error {
	syncer.workspaceID = workspaceID
	return syncer.err
}
