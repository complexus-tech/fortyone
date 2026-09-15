package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountDeletionRepository struct {
	queries                                     *usersql.Queries
	provider, attachmentsBucket, profilesBucket string
}

type AccountDeletionSnapshot struct {
	UserID uuid.UUID
	Email  string
	Avatar string
}

func NewAccountDeletionRepository(pool *pgxpool.Pool, provider, attachmentsBucket, profilesBucket string) (*AccountDeletionRepository, error) {
	if pool == nil || strings.TrimSpace(provider) == "" || strings.TrimSpace(attachmentsBucket) == "" || strings.TrimSpace(profilesBucket) == "" {
		return nil, errors.New("account deletion requires database and object storage routes")
	}
	return &AccountDeletionRepository{queries: usersql.New(pool), provider: provider, attachmentsBucket: attachmentsBucket, profilesBucket: profilesBucket}, nil
}

func (r *AccountDeletionRepository) Lock(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (AccountDeletionSnapshot, error) {
	q := r.queries.WithTx(tx)
	row, err := q.LockAccountForDeletion(ctx, usersql.LockAccountForDeletionParams{UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountDeletionSnapshot{}, usersdomain.ErrNotFound
	}
	if err != nil {
		return AccountDeletionSnapshot{}, fmt.Errorf("lock account deletion: %w", err)
	}
	if _, err := q.LockAccountDeletionMemberships(ctx, usersql.LockAccountDeletionMembershipsParams{UserID: userID}); err != nil {
		return AccountDeletionSnapshot{}, fmt.Errorf("lock account memberships: %w", err)
	}
	conflicts, err := q.ListAccountDeletionConflicts(ctx, usersql.ListAccountDeletionConflictsParams{UserID: userID})
	if err != nil {
		return AccountDeletionSnapshot{}, fmt.Errorf("check account ownership: %w", err)
	}
	if len(conflicts) != 0 {
		return AccountDeletionSnapshot{}, &usersdomain.AccountDeletionConflict{Workspaces: conflicts}
	}
	return AccountDeletionSnapshot{UserID: userID, Email: row.Email, Avatar: row.AvatarURL}, nil
}

func (r *AccountDeletionRepository) Deactivate(ctx context.Context, tx pgx.Tx, command usersdomain.AccountDeletion) error {
	return r.queries.WithTx(tx).DeactivateAccountForDeletion(ctx, usersql.DeactivateAccountForDeletionParams{UserID: command.UserID, DeletedAt: command.RequestedAt})
}

func (r *AccountDeletionRepository) Erase(ctx context.Context, tx pgx.Tx, snapshot AccountDeletionSnapshot, deletedAt time.Time) error {
	q := r.queries.WithTx(tx)
	userID, email := snapshot.UserID, snapshot.Email
	if err := q.LockAccountSubscriberLifecycle(ctx, usersql.LockAccountSubscriberLifecycleParams{Email: email}); err != nil {
		return fmt.Errorf("lock subscriber lifecycle: %w", err)
	}
	if err := q.QueueAccountSubscriberDeletion(ctx, usersql.QueueAccountSubscriberDeletionParams{Email: email, RequestedAt: deletedAt}); err != nil {
		return fmt.Errorf("queue subscriber deletion: %w", err)
	}
	ids, err := q.LockAccountDeletionAttachments(ctx, usersql.LockAccountDeletionAttachmentsParams{UserID: &userID})
	if err != nil {
		return fmt.Errorf("lock account attachments: %w", err)
	}
	if err := eraseAccountContributions(ctx, q, userID, email); err != nil {
		return err
	}
	if err := revokeAccountCredentials(ctx, q, userID, email); err != nil {
		return err
	}
	if err := eraseAccountPersonalRecords(ctx, q, userID, email); err != nil {
		return err
	}
	if err := detachAccountSharedRecords(ctx, q, userID); err != nil {
		return err
	}
	if err := q.RetireAccountDeletionAttachments(ctx, usersql.RetireAccountDeletionAttachmentsParams{
		AttachmentIds: ids, StorageProvider: r.provider, ContainerName: r.attachmentsBucket, DeletedAt: deletedAt,
	}); err != nil {
		return fmt.Errorf("retire account objects: %w", err)
	}
	// Stored uploads use opaque keys. External provider avatar URLs are not our
	// property and must never be converted into arbitrary storage deletions.
	if avatar := strings.TrimSpace(snapshot.Avatar); avatar != "" && !strings.Contains(avatar, "://") {
		if err := q.QueueAccountAvatarDeletion(ctx, usersql.QueueAccountAvatarDeletionParams{
			ObjectID: uuid.New(), UserID: userID, StorageProvider: r.provider, ContainerName: r.profilesBucket, BlobName: avatar, DeletedAt: deletedAt,
		}); err != nil {
			return fmt.Errorf("queue account avatar deletion: %w", err)
		}
	}
	return q.AnonymizeAccountAwaitingCleanup(ctx, usersql.AnonymizeAccountAwaitingCleanupParams{UserID: userID, DeletedAt: deletedAt})
}

func (r *AccountDeletionRepository) Queue(ctx context.Context, tx pgx.Tx, command usersdomain.AccountDeletion) error {
	return r.queries.WithTx(tx).QueueAccountDeletionFinalization(ctx, usersql.QueueAccountDeletionFinalizationParams{UserID: command.UserID, RequestedAt: command.RequestedAt})
}

func (r *AccountDeletionRepository) Delete(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	count, err := r.queries.WithTx(tx).PermanentlyDeleteAccount(ctx, usersql.PermanentlyDeleteAccountParams{UserID: userID})
	if err != nil {
		return fmt.Errorf("permanently delete account: %w", err)
	}
	if count != 1 {
		return usersdomain.ErrNotFound
	}
	return nil
}

func (r *AccountDeletionRepository) Pending(ctx context.Context, limit int32) ([]uuid.UUID, error) {
	return r.queries.ListPendingAccountDeletions(ctx, usersql.ListPendingAccountDeletionsParams{BatchSize: limit})
}

func (r *AccountDeletionRepository) LockPending(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	_, err := r.queries.WithTx(tx).LockPendingAccountDeletion(ctx, usersql.LockPendingAccountDeletionParams{UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *AccountDeletionRepository) TouchPending(ctx context.Context, tx pgx.Tx, userID uuid.UUID, at time.Time) error {
	return r.queries.WithTx(tx).TouchPendingAccountDeletion(ctx, usersql.TouchPendingAccountDeletionParams{UserID: userID, UpdatedAt: at})
}

func eraseAccountContributions(ctx context.Context, q *usersql.Queries, userID uuid.UUID, email string) error {
	deletedUserID := usersdomain.DeletedUserID
	if err := q.ReattributeAccountComments(ctx, usersql.ReattributeAccountCommentsParams{UserID: userID, DeletedUserID: deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountComments: %w", err)
	}
	if err := q.ReattributeAccountFeedbackComments(ctx, usersql.ReattributeAccountFeedbackCommentsParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountFeedbackComments: %w", err)
	}
	if err := q.ReattributeAccountIntegrationComments(ctx, usersql.ReattributeAccountIntegrationCommentsParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("reattribute account integration comments: %w", err)
	}
	if err := q.ReattributeAccountFeedbackSubmissions(ctx, usersql.ReattributeAccountFeedbackSubmissionsParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountFeedbackSubmissions: %w", err)
	}
	if err := q.AnonymizeAccountFeedbackVotes(ctx, usersql.AnonymizeAccountFeedbackVotesParams{UserID: &userID}); err != nil {
		return fmt.Errorf("AnonymizeAccountFeedbackVotes: %w", err)
	}
	if err := q.EraseAccountFeedbackContributorDeliveries(ctx, usersql.EraseAccountFeedbackContributorDeliveriesParams{UserID: &userID}); err != nil {
		return fmt.Errorf("EraseAccountFeedbackContributorDeliveries: %w", err)
	}
	if err := q.EraseAccountFeedbackContributorPreferences(ctx, usersql.EraseAccountFeedbackContributorPreferencesParams{UserID: &userID}); err != nil {
		return fmt.Errorf("EraseAccountFeedbackContributorPreferences: %w", err)
	}
	if err := q.EraseAccountFeedbackContributorSessions(ctx, usersql.EraseAccountFeedbackContributorSessionsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("EraseAccountFeedbackContributorSessions: %w", err)
	}
	if err := q.EraseAccountFeedbackContributorUnsubscribeTokens(ctx, usersql.EraseAccountFeedbackContributorUnsubscribeTokensParams{UserID: &userID}); err != nil {
		return fmt.Errorf("EraseAccountFeedbackContributorUnsubscribeTokens: %w", err)
	}
	if err := q.EraseAccountFeedbackItemFollowers(ctx, usersql.EraseAccountFeedbackItemFollowersParams{UserID: &userID}); err != nil {
		return fmt.Errorf("EraseAccountFeedbackItemFollowers: %w", err)
	}
	if err := q.EraseAccountFeedbackPortalFollowers(ctx, usersql.EraseAccountFeedbackPortalFollowersParams{UserID: &userID}); err != nil {
		return fmt.Errorf("EraseAccountFeedbackPortalFollowers: %w", err)
	}
	if err := q.AnonymizeAccountFeedbackContributors(ctx, usersql.AnonymizeAccountFeedbackContributorsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("AnonymizeAccountFeedbackContributors: %w", err)
	}
	if err := q.DeleteAccountPrivateDocuments(ctx, usersql.DeleteAccountPrivateDocumentsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountPrivateDocuments: %w", err)
	}
	if err := q.DeleteAccountFeedbackVerifications(ctx, usersql.DeleteAccountFeedbackVerificationsParams{Email: email}); err != nil {
		return fmt.Errorf("DeleteAccountFeedbackVerifications: %w", err)
	}
	if err := q.DeleteAccountFeedbackDeliveries(ctx, usersql.DeleteAccountFeedbackDeliveriesParams{Email: email}); err != nil {
		return fmt.Errorf("DeleteAccountFeedbackDeliveries: %w", err)
	}
	if err := q.DeleteAccountAdminNotes(ctx, usersql.DeleteAccountAdminNotesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountAdminNotes: %w", err)
	}
	if err := q.DeleteAccountAdminTargetAudits(ctx, usersql.DeleteAccountAdminTargetAuditsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountAdminTargetAudits: %w", err)
	}
	if err := q.RemoveAccountMutationSnapshots(ctx, usersql.RemoveAccountMutationSnapshotsParams{UserID: userID}); err != nil {
		return fmt.Errorf("RemoveAccountMutationSnapshots: %w", err)
	}
	if err := q.RemoveAccountScheduleSnapshots(ctx, usersql.RemoveAccountScheduleSnapshotsParams{UserID: userID}); err != nil {
		return fmt.Errorf("RemoveAccountScheduleSnapshots: %w", err)
	}
	if err := q.RemoveAccountAuditSnapshots(ctx, usersql.RemoveAccountAuditSnapshotsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("RemoveAccountAuditSnapshots: %w", err)
	}
	if err := q.AnonymizeAccountActivityReferences(ctx, usersql.AnonymizeAccountActivityReferencesParams{UserID: userID.String(), DeletedUserID: deletedUserID.String()}); err != nil {
		return fmt.Errorf("anonymize account activity references: %w", err)
	}
	if err := q.AnonymizeAccountObjectiveActivityReferences(ctx, usersql.AnonymizeAccountObjectiveActivityReferencesParams{UserID: userID.String(), DeletedUserID: deletedUserID.String()}); err != nil {
		return fmt.Errorf("anonymize account objective activity references: %w", err)
	}
	if err := q.ReattributeAccountActivities(ctx, usersql.ReattributeAccountActivitiesParams{UserID: userID, DeletedUserID: deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountActivities: %w", err)
	}
	if err := q.DeleteAccountCommentMentions(ctx, usersql.DeleteAccountCommentMentionsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountCommentMentions: %w", err)
	}
	if err := q.DeleteAccountFeedbackDigests(ctx, usersql.DeleteAccountFeedbackDigestsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountFeedbackDigests: %w", err)
	}
	if err := q.DeleteAccountFeedbackReads(ctx, usersql.DeleteAccountFeedbackReadsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountFeedbackReads: %w", err)
	}
	if err := q.DeleteAccountFeedbackSubscriptions(ctx, usersql.DeleteAccountFeedbackSubscriptionsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountFeedbackSubscriptions: %w", err)
	}
	if err := q.ReattributeAccountObjectiveActivities(ctx, usersql.ReattributeAccountObjectiveActivitiesParams{UserID: userID, DeletedUserID: deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountObjectiveActivities: %w", err)
	}
	return nil
}

func revokeAccountCredentials(ctx context.Context, q *usersql.Queries, userID uuid.UUID, email string) error {
	if err := q.DeleteAccountOwnedOAuthInstallations(ctx, usersql.DeleteAccountOwnedOAuthInstallationsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountOwnedOAuthInstallations: %w", err)
	}
	if err := q.DeleteAccountOwnedOAuthApplications(ctx, usersql.DeleteAccountOwnedOAuthApplicationsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountOwnedOAuthApplications: %w", err)
	}
	if err := q.DeleteAccountPrincipals(ctx, usersql.DeleteAccountPrincipalsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountPrincipals: %w", err)
	}
	if err := q.DeleteAccountVerificationTokens(ctx, usersql.DeleteAccountVerificationTokensParams{UserID: &userID, Email: email}); err != nil {
		return fmt.Errorf("DeleteAccountVerificationTokens: %w", err)
	}
	if err := q.DeleteAccountDriveAccounts(ctx, usersql.DeleteAccountDriveAccountsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountDriveAccounts: %w", err)
	}
	if err := q.DeleteAccountDriveOAuthStates(ctx, usersql.DeleteAccountDriveOAuthStatesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountDriveOAuthStates: %w", err)
	}
	if err := q.DeleteAccountExternalIdentities(ctx, usersql.DeleteAccountExternalIdentitiesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountExternalIdentities: %w", err)
	}
	if err := q.DeleteAccountFigmaConnections(ctx, usersql.DeleteAccountFigmaConnectionsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountFigmaConnections: %w", err)
	}
	if err := q.DeleteAccountFigmaOAuthStates(ctx, usersql.DeleteAccountFigmaOAuthStatesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountFigmaOAuthStates: %w", err)
	}
	if err := q.DeleteAccountOAuthGrants(ctx, usersql.DeleteAccountOAuthGrantsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountOAuthGrants: %w", err)
	}
	if err := q.DeleteAccountPersonalIntegrations(ctx, usersql.DeleteAccountPersonalIntegrationsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountPersonalIntegrations: %w", err)
	}
	return nil
}

func eraseAccountPersonalRecords(ctx context.Context, q *usersql.Queries, userID uuid.UUID, email string) error {
	deletedUserID := usersdomain.DeletedUserID
	if err := q.DeleteAccountInvitationOutbox(ctx, usersql.DeleteAccountInvitationOutboxParams{UserID: userID, Email: email}); err != nil {
		return fmt.Errorf("DeleteAccountInvitationOutbox: %w", err)
	}
	if err := q.DeleteAccountCalendarBlocks(ctx, usersql.DeleteAccountCalendarBlocksParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountCalendarBlocks: %w", err)
	}
	if err := q.DeleteAccountCalendarHistory(ctx, usersql.DeleteAccountCalendarHistoryParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountCalendarHistory: %w", err)
	}
	if err := q.DeleteAccountCalendarAvailability(ctx, usersql.DeleteAccountCalendarAvailabilityParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountCalendarAvailability: %w", err)
	}
	if err := q.DeleteAccountDriveCreateOperations(ctx, usersql.DeleteAccountDriveCreateOperationsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountDriveCreateOperations: %w", err)
	}
	if err := q.DeleteAccountDriveImportOperations(ctx, usersql.DeleteAccountDriveImportOperationsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountDriveImportOperations: %w", err)
	}
	if err := q.DeleteAccountInvitations(ctx, usersql.DeleteAccountInvitationsParams{UserID: userID, Email: email}); err != nil {
		return fmt.Errorf("DeleteAccountInvitations: %w", err)
	}
	if err := q.DeleteAccountAIUsageResets(ctx, usersql.DeleteAccountAIUsageResetsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountAIUsageResets: %w", err)
	}
	if err := q.DeleteAccountAnalytics(ctx, usersql.DeleteAccountAnalyticsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountAnalytics: %w", err)
	}
	if err := q.DeleteAccountAutomationPreferences(ctx, usersql.DeleteAccountAutomationPreferencesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountAutomationPreferences: %w", err)
	}
	if err := q.DeleteAccountChatApprovals(ctx, usersql.DeleteAccountChatApprovalsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountChatApprovals: %w", err)
	}
	if err := q.DeleteAccountChats(ctx, usersql.DeleteAccountChatsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountChats: %w", err)
	}
	if err := q.DeleteAccountDocumentMemberships(ctx, usersql.DeleteAccountDocumentMembershipsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountDocumentMemberships: %w", err)
	}
	if err := q.DeleteAccountEmailAvatarHandles(ctx, usersql.DeleteAccountEmailAvatarHandlesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountEmailAvatarHandles: %w", err)
	}
	if err := q.DeleteAccountGlobalWalkthroughs(ctx, usersql.DeleteAccountGlobalWalkthroughsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountGlobalWalkthroughs: %w", err)
	}
	if err := q.DeleteAccountInternalAlerts(ctx, usersql.DeleteAccountInternalAlertsParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountInternalAlerts: %w", err)
	}
	if err := q.DeleteAccountKeyResultContributions(ctx, usersql.DeleteAccountKeyResultContributionsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountKeyResultContributions: %w", err)
	}
	if err := q.DeleteAccountMayaRuns(ctx, usersql.DeleteAccountMayaRunsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountMayaRuns: %w", err)
	}
	if err := q.DeleteAccountMayaVoiceSessions(ctx, usersql.DeleteAccountMayaVoiceSessionsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountMayaVoiceSessions: %w", err)
	}
	if err := q.DeleteAccountMemories(ctx, usersql.DeleteAccountMemoriesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountMemories: %w", err)
	}
	if err := q.DeleteAccountMessagingConfirmations(ctx, usersql.DeleteAccountMessagingConfirmationsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountMessagingConfirmations: %w", err)
	}
	if err := q.DeleteAccountMessagingConversations(ctx, usersql.DeleteAccountMessagingConversationsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountMessagingConversations: %w", err)
	}
	if err := q.DeleteAccountMessagingDeliveries(ctx, usersql.DeleteAccountMessagingDeliveriesParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountMessagingDeliveries: %w", err)
	}
	if err := q.DeleteAccountMessagingEmailThreads(ctx, usersql.DeleteAccountMessagingEmailThreadsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountMessagingEmailThreads: %w", err)
	}
	if err := q.DeleteAccountMessagingNonces(ctx, usersql.DeleteAccountMessagingNoncesParams{UserID: &userID}); err != nil {
		return fmt.Errorf("DeleteAccountMessagingNonces: %w", err)
	}
	if err := q.DeleteAccountNotificationEmails(ctx, usersql.DeleteAccountNotificationEmailsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountNotificationEmails: %w", err)
	}
	if err := q.DeleteAccountNotificationPreferences(ctx, usersql.DeleteAccountNotificationPreferencesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountNotificationPreferences: %w", err)
	}
	if err := q.AnonymizeAccountNotificationReferences(ctx, usersql.AnonymizeAccountNotificationReferencesParams{UserID: userID, DeletedUserID: deletedUserID}); err != nil {
		return fmt.Errorf("anonymize account notification references: %w", err)
	}
	if err := q.ReattributeAccountNotifications(ctx, usersql.ReattributeAccountNotificationsParams{UserID: userID, DeletedUserID: deletedUserID}); err != nil {
		return fmt.Errorf("reattribute account notifications: %w", err)
	}
	if err := q.DeleteAccountNotifications(ctx, usersql.DeleteAccountNotificationsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountNotifications: %w", err)
	}
	if err := q.DeleteAccountRoutineEmails(ctx, usersql.DeleteAccountRoutineEmailsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountRoutineEmails: %w", err)
	}
	if err := q.DeleteAccountSlackLinks(ctx, usersql.DeleteAccountSlackLinksParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountSlackLinks: %w", err)
	}
	if err := q.DeleteAccountStoryCollaborators(ctx, usersql.DeleteAccountStoryCollaboratorsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountStoryCollaborators: %w", err)
	}
	if err := q.DeleteAccountStoryMutes(ctx, usersql.DeleteAccountStoryMutesParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountStoryMutes: %w", err)
	}
	if err := q.DeleteAccountStoryWatchers(ctx, usersql.DeleteAccountStoryWatchersParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountStoryWatchers: %w", err)
	}
	if err := q.DeleteAccountTeamMemberships(ctx, usersql.DeleteAccountTeamMembershipsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountTeamMemberships: %w", err)
	}
	if err := q.DeleteAccountTeamOrder(ctx, usersql.DeleteAccountTeamOrderParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountTeamOrder: %w", err)
	}
	if err := q.DeleteAccountWorkspaceMemberships(ctx, usersql.DeleteAccountWorkspaceMembershipsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountWorkspaceMemberships: %w", err)
	}
	if err := q.DeleteAccountWorkspaceWalkthroughs(ctx, usersql.DeleteAccountWorkspaceWalkthroughsParams{UserID: userID}); err != nil {
		return fmt.Errorf("DeleteAccountWorkspaceWalkthroughs: %w", err)
	}
	if err := q.DetachAccountUserAiUsageResetsResetByUserId(ctx, usersql.DetachAccountUserAiUsageResetsResetByUserIdParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountUserAiUsageResetsResetByUserId: %w", err)
	}
	return nil
}

func detachAccountSharedRecords(ctx context.Context, q *usersql.Queries, userID uuid.UUID) error {
	deletedUserID := usersdomain.DeletedUserID
	if err := q.AnonymizeAccountAdminAudits(ctx, usersql.AnonymizeAccountAdminAuditsParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("AnonymizeAccountAdminAudits: %w", err)
	}
	if err := q.ClearAccountAttachmentsUploadedBy(ctx, usersql.ClearAccountAttachmentsUploadedByParams{UserID: &userID}); err != nil {
		return fmt.Errorf("ClearAccountAttachmentsUploadedBy: %w", err)
	}
	if err := q.ClearAccountIntegrationRequestsAcceptanceStartedByUserId(ctx, usersql.ClearAccountIntegrationRequestsAcceptanceStartedByUserIdParams{UserID: &userID}); err != nil {
		return fmt.Errorf("ClearAccountIntegrationRequestsAcceptanceStartedByUserId: %w", err)
	}
	if err := q.ReattributeAccountKeyResultsCreatedBy(ctx, usersql.ReattributeAccountKeyResultsCreatedByParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountKeyResultsCreatedBy: %w", err)
	}
	if err := q.ClearAccountKeyResultsLead(ctx, usersql.ClearAccountKeyResultsLeadParams{UserID: &userID}); err != nil {
		return fmt.Errorf("ClearAccountKeyResultsLead: %w", err)
	}
	if err := q.ReattributeAccountObjectivesCreatedBy(ctx, usersql.ReattributeAccountObjectivesCreatedByParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountObjectivesCreatedBy: %w", err)
	}
	if err := q.ClearAccountObjectivesLeadUserId(ctx, usersql.ClearAccountObjectivesLeadUserIdParams{UserID: &userID}); err != nil {
		return fmt.Errorf("ClearAccountObjectivesLeadUserId: %w", err)
	}
	if err := q.ClearAccountStoriesAssigneeId(ctx, usersql.ClearAccountStoriesAssigneeIdParams{UserID: &userID}); err != nil {
		return fmt.Errorf("ClearAccountStoriesAssigneeId: %w", err)
	}
	if err := q.ReattributeAccountStoriesReporterId(ctx, usersql.ReattributeAccountStoriesReporterIdParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountStoriesReporterId: %w", err)
	}
	if err := q.ReattributeAccountWorkspacesCreatedBy(ctx, usersql.ReattributeAccountWorkspacesCreatedByParams{UserID: &userID, DeletedUserID: &deletedUserID}); err != nil {
		return fmt.Errorf("ReattributeAccountWorkspacesCreatedBy: %w", err)
	}
	if err := q.ClearAccountWorkspacesDeletedBy(ctx, usersql.ClearAccountWorkspacesDeletedByParams{UserID: &userID}); err != nil {
		return fmt.Errorf("ClearAccountWorkspacesDeletedBy: %w", err)
	}
	if err := q.DetachAccountDocumentAttachmentsCreatedBy(ctx, usersql.DetachAccountDocumentAttachmentsCreatedByParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountDocumentAttachmentsCreatedBy: %w", err)
	}
	if err := q.DetachAccountDocumentRelationshipsCreatedBy(ctx, usersql.DetachAccountDocumentRelationshipsCreatedByParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountDocumentRelationshipsCreatedBy: %w", err)
	}
	if err := q.DetachAccountDocumentsCreatedBy(ctx, usersql.DetachAccountDocumentsCreatedByParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountDocumentsCreatedBy: %w", err)
	}
	if err := q.DetachAccountDocumentsUpdatedBy(ctx, usersql.DetachAccountDocumentsUpdatedByParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountDocumentsUpdatedBy: %w", err)
	}
	if err := q.DetachAccountGoogleDriveDocumentImportsImportedByUserId(ctx, usersql.DetachAccountGoogleDriveDocumentImportsImportedByUserIdParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountGoogleDriveDocumentImportsImportedByUserId: %w", err)
	}
	if err := q.DetachAccountGoogleDriveFileReferencesCreatedByUserId(ctx, usersql.DetachAccountGoogleDriveFileReferencesCreatedByUserIdParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountGoogleDriveFileReferencesCreatedByUserId: %w", err)
	}
	if err := q.DetachAccountStoryFigmaLinksCreatedByUserId(ctx, usersql.DetachAccountStoryFigmaLinksCreatedByUserIdParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountStoryFigmaLinksCreatedByUserId: %w", err)
	}
	if err := q.DetachAccountStoryInlineAttachmentsCreatedBy(ctx, usersql.DetachAccountStoryInlineAttachmentsCreatedByParams{DeletedUserID: deletedUserID, UserID: userID}); err != nil {
		return fmt.Errorf("DetachAccountStoryInlineAttachmentsCreatedBy: %w", err)
	}
	return nil
}
