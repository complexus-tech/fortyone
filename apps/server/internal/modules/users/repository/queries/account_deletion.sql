-- Account deletion is a separate, irreversible path from account deactivation.
-- The subscriber update/delete workers use the same per-email advisory lock.
-- This prevents a previously queued update from recreating an erased contact.
-- name: LockAccountSubscriberLifecycle :exec
SELECT pg_advisory_xact_lock(hashtextextended('subscriber:' || lower(btrim(CAST(sqlc.arg(email) AS text))), 0));

-- Keep only the address necessary to erase the remote contact; the provider
-- dispatcher deletes this row only after a confirmed, idempotent deletion.
-- name: QueueAccountSubscriberDeletion :exec
INSERT INTO public.account_subscriber_deletions (id, email, created_at, updated_at, attempt_count)
VALUES (gen_random_uuid(), lower(btrim(CAST(sqlc.arg(email) AS text))), sqlc.arg(requested_at), sqlc.arg(requested_at), 0);

-- name: LockAccountForDeletion :one
SELECT user_id, email, COALESCE(avatar_url, '') AS avatar_url
FROM public.users
WHERE user_id = sqlc.arg(user_id) AND is_active = TRUE AND is_system = FALSE
FOR UPDATE;

-- Lock all memberships in affected workspaces, not only the deleting member.
-- Together with SERIALIZABLE isolation this prevents concurrent last-admin exits.
-- name: LockAccountDeletionMemberships :many
SELECT member.workspace_id, member.user_id
FROM public.workspace_members AS member
WHERE member.workspace_id IN (
    SELECT own.workspace_id FROM public.workspace_members AS own WHERE own.user_id = sqlc.arg(user_id)
)
ORDER BY member.workspace_id, member.user_id
FOR UPDATE OF member;

-- name: ListAccountDeletionConflicts :many
SELECT workspace.name
FROM public.workspace_members AS own
JOIN public.workspaces AS workspace ON workspace.workspace_id = own.workspace_id
WHERE own.user_id = sqlc.arg(user_id) AND own.role = 'admin' AND workspace.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM public.workspace_members AS other
      JOIN public.users AS account ON account.user_id = other.user_id
      WHERE other.workspace_id = own.workspace_id AND other.user_id <> own.user_id
        AND other.role = 'admin' AND account.is_active = TRUE AND account.is_system = FALSE
  )
ORDER BY workspace.name, workspace.workspace_id;

-- A real active-to-inactive transition runs the existing Drive revocation saga.
-- name: DeactivateAccountForDeletion :exec
UPDATE public.users SET is_active = FALSE, login_reactivation_policy = 'admin_only',
    auth_session_version = auth_session_version + 1, updated_at = sqlc.arg(deleted_at)
WHERE user_id = sqlc.arg(user_id);

-- name: LockAccountDeletionAttachments :many
SELECT attachment.attachment_id
FROM public.attachments AS attachment
WHERE attachment.uploaded_by = sqlc.arg(user_id)
   OR EXISTS (
       SELECT 1 FROM public.document_attachments AS relation
       JOIN public.documents AS document ON document.document_id = relation.document_id
       WHERE relation.attachment_id = attachment.attachment_id
         AND document.created_by = sqlc.arg(user_id) AND document.visibility = 'private'
   )
ORDER BY attachment.attachment_id FOR UPDATE OF attachment;

-- Comments belong to the shared work record. Preserve IDs, prose and the entire
-- reply hierarchy while moving the author FK away from the erased account.
-- name: ReattributeAccountComments :exec
UPDATE public.story_comments SET commenter_id = sqlc.arg(deleted_user_id)
WHERE commenter_id = sqlc.arg(user_id);

-- name: ReattributeAccountFeedbackComments :exec
UPDATE public.feedback_comments SET author_id = sqlc.arg(deleted_user_id)
WHERE author_id = sqlc.arg(user_id)
  OR contributor_id IN (SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id));

-- name: ReattributeAccountFeedbackSubmissions :exec
UPDATE public.feedback_items SET author_id = sqlc.arg(deleted_user_id)
WHERE author_id = sqlc.arg(user_id) OR contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);

-- name: ReattributeAccountIntegrationComments :exec
UPDATE public.integration_request_comments
SET author_user_id = sqlc.arg(deleted_user_id), external_author_id = NULL
WHERE author_user_id = sqlc.arg(user_id);

-- Anonymous contributor keys preserve vote totals without linking the account.
-- name: AnonymizeAccountFeedbackVotes :exec
UPDATE public.feedback_votes AS vote SET user_id = NULL WHERE vote.user_id = sqlc.arg(user_id)
  OR vote.contributor_id IN (SELECT contributor.id FROM public.feedback_contributors AS contributor WHERE contributor.user_id = sqlc.arg(user_id));

-- Existing submissions require a contributor FK. Erase the identity entirely;
-- leave only an anonymous attribution key for retained workspace discussion.
-- name: AnonymizeAccountFeedbackContributors :exec
UPDATE public.feedback_contributors SET user_id = NULL, kind = 'anonymous', email = NULL,
    email_verified_at = NULL, display_name = NULL, avatar_url = NULL, external_id = NULL,
    blocked_reason = NULL, public_masked = FALSE, last_seen_at = NULL
WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountPrivateDocuments :exec
DELETE FROM public.documents WHERE created_by = sqlc.arg(user_id) AND visibility = 'private';

-- name: DeleteAccountOwnedOAuthInstallations :exec
DELETE FROM public.principals AS principal
USING public.oauth_application_installations AS installation, public.oauth_applications AS application
WHERE principal.principal_id = installation.principal_id
  AND installation.application_id = application.application_id
  AND application.owner_user_id = sqlc.arg(user_id);

-- name: DeleteAccountOwnedOAuthApplications :exec
DELETE FROM public.oauth_applications WHERE owner_user_id = sqlc.arg(user_id);

-- name: DeleteAccountPrincipals :exec
DELETE FROM public.principals WHERE subject_user_id = sqlc.arg(user_id);

-- name: DeleteAccountVerificationTokens :exec
DELETE FROM public.verification_tokens WHERE user_id = sqlc.arg(user_id) OR lower(email) = lower(sqlc.arg(email));

-- name: DeleteAccountInvitations :exec
DELETE FROM public.workspace_invitations WHERE inviter_id = sqlc.arg(user_id) OR lower(email) = lower(sqlc.arg(email));

-- Invitation outboxes may outlive their source invitation. Erase address-bearing
-- snapshots before deleting invitation rows, including invitations sent to us.
-- name: DeleteAccountInvitationOutbox :exec
DELETE FROM public.workspace_invitation_outbox
WHERE actor_id = sqlc.arg(user_id) OR invitation_id IN (
    SELECT invitation_id FROM public.workspace_invitations
    WHERE inviter_id = sqlc.arg(user_id) OR lower(email) = lower(sqlc.arg(email))
);

-- Calendar provider mirrors are staged and detached before these local records
-- are erased. Provider outbox records and sealed credentials remain untouched.
-- name: DeleteAccountCalendarBlocks :exec
DELETE FROM public.calendar_schedule_blocks WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountCalendarHistory :exec
DELETE FROM public.calendar_schedule_reschedule_events WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountCalendarAvailability :exec
DELETE FROM public.calendar_busy_windows WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountDriveCreateOperations :exec
DELETE FROM public.google_drive_create_operations WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountDriveImportOperations :exec
DELETE FROM public.google_drive_document_import_operations WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountFeedbackVerifications :exec
DELETE FROM public.feedback_contributor_verifications WHERE lower(email) = lower(sqlc.arg(email));

-- name: DeleteAccountFeedbackDeliveries :exec
DELETE FROM public.feedback_contributor_deliveries WHERE lower(recipient_email) = lower(sqlc.arg(email));

-- name: DeleteAccountAdminNotes :exec
DELETE FROM public.admin_notes WHERE created_by_user_id = sqlc.arg(user_id)
    OR (target_type = 'user' AND target_id = sqlc.arg(user_id));

-- name: DeleteAccountAdminTargetAudits :exec
DELETE FROM public.admin_audit_logs WHERE target_type = 'user' AND target_id = sqlc.arg(user_id);

-- name: AnonymizeAccountAdminAudits :exec
UPDATE public.admin_audit_logs SET actor_user_id = sqlc.arg(deleted_user_id),
    old_value = NULL, new_value = NULL, metadata = '{}', reason = NULL
WHERE actor_user_id = sqlc.arg(user_id);

-- name: RemoveAccountMutationSnapshots :exec
DELETE FROM public.story_mutation_events WHERE actor_id = sqlc.arg(user_id);

-- name: RemoveAccountScheduleSnapshots :exec
DELETE FROM public.story_schedule_transition_outbox WHERE actor_id = sqlc.arg(user_id);

-- name: RemoveAccountAuditSnapshots :exec
DELETE FROM public.audit_events WHERE actor_id = sqlc.arg(user_id);

-- name: RetireAccountDeletionAttachments :exec
WITH retired AS (
    DELETE FROM public.attachments AS attachment
    WHERE attachment.attachment_id = ANY(CAST(sqlc.arg(attachment_ids) AS uuid[]))
      AND NOT EXISTS (SELECT 1 FROM public.story_attachments WHERE attachment_id = attachment.attachment_id)
      AND NOT EXISTS (SELECT 1 FROM public.story_inline_attachments WHERE attachment_id = attachment.attachment_id)
      AND NOT EXISTS (SELECT 1 FROM public.document_attachments WHERE attachment_id = attachment.attachment_id)
      AND NOT EXISTS (SELECT 1 FROM public.feedback_item_attachments WHERE attachment_id = attachment.attachment_id)
    RETURNING attachment.attachment_id, attachment.workspace_id, attachment.blob_name
)
INSERT INTO public.attachment_object_deletion_outbox (
    attachment_id, workspace_id, storage_provider, container_name, blob_name,
    status, attempt_count, next_attempt_at, created_at, updated_at
)
SELECT attachment_id, workspace_id, sqlc.arg(storage_provider), sqlc.arg(container_name), blob_name,
    'pending', 0, sqlc.arg(deleted_at), sqlc.arg(deleted_at), sqlc.arg(deleted_at)
FROM retired;

-- name: QueueAccountAvatarDeletion :exec
INSERT INTO public.attachment_object_deletion_outbox (
    attachment_id, workspace_id, storage_provider, container_name, blob_name,
    status, attempt_count, next_attempt_at, created_at, updated_at
) SELECT sqlc.arg(object_id), '00000000-0000-0000-0000-000000000000',
    sqlc.arg(storage_provider), sqlc.arg(container_name), CAST(sqlc.arg(blob_name) AS text),
    'pending', 0, sqlc.arg(deleted_at), sqlc.arg(deleted_at), sqlc.arg(deleted_at)
WHERE NOT EXISTS (SELECT 1 FROM public.users WHERE avatar_url = CAST(sqlc.arg(blob_name) AS text)
    AND user_id <> sqlc.arg(user_id));

-- No sign-in path can recover this row. Only calendar cleanup still needs its
-- opaque FK key; the finalizer deletes it as soon as remote cleanup is drained.
-- name: AnonymizeAccountAwaitingCleanup :exec
UPDATE public.users SET username = 'deleted-' || CAST(user_id AS text),
    email = CAST(user_id AS text) || '@deleted.accounts.invalid', full_name = NULL, avatar_url = NULL,
    is_active = FALSE, is_internal = FALSE, login_reactivation_policy = 'admin_only',
    last_login_at = NULL, last_used_workspace_id = NULL, has_seen_walkthrough = FALSE,
    timezone = 'UTC', working_days = NULL, working_start_minute = NULL, working_end_minute = NULL,
    github_username = NULL, github_user_id = NULL, github_access_token = NULL,
    github_access_token_envelope_version = 0, github_access_token_generation = NULL,
    inactivity_warning_sent_at = NULL, updated_at = sqlc.arg(deleted_at)
WHERE user_id = sqlc.arg(user_id);

-- name: QueueAccountDeletionFinalization :exec
INSERT INTO public.account_deletion_requests (user_id, requested_at, updated_at)
VALUES (sqlc.arg(user_id), sqlc.arg(requested_at), sqlc.arg(requested_at));

-- name: ListPendingAccountDeletions :many
SELECT user_id FROM public.account_deletion_requests
ORDER BY updated_at, user_id LIMIT CAST(sqlc.arg(batch_size) AS integer);

-- name: LockPendingAccountDeletion :one
SELECT request.user_id FROM public.account_deletion_requests AS request
JOIN public.users AS account ON account.user_id = request.user_id
WHERE request.user_id = sqlc.arg(user_id) AND account.is_active = FALSE
  AND account.login_reactivation_policy = 'admin_only'
FOR UPDATE OF request, account SKIP LOCKED;

-- name: TouchPendingAccountDeletion :exec
UPDATE public.account_deletion_requests SET updated_at = sqlc.arg(updated_at), attempt_count = attempt_count + 1
WHERE user_id = sqlc.arg(user_id);

-- name: PermanentlyDeleteAccount :execrows
DELETE FROM public.users WHERE user_id = sqlc.arg(user_id) AND is_active = FALSE AND is_system = FALSE;

-- name: DeleteAccountPersonalIntegrations :exec
DELETE FROM public.integrations WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountOAuthGrants :exec
DELETE FROM public.oauth_grants WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountExternalIdentities :exec
DELETE FROM public.user_external_identities WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountMemories :exec
DELETE FROM public.user_memories WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountAutomationPreferences :exec
DELETE FROM public.user_automation_preferences WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountTeamOrder :exec
DELETE FROM public.user_team_orders WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountChats :exec
DELETE FROM public.chat_sessions WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountChatApprovals :exec
DELETE FROM public.chat_mutation_approval_executions WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountMessagingDeliveries :exec
DELETE FROM public.messaging_outbound_deliveries WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountMessagingConversations :exec
DELETE FROM public.messaging_conversations WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountMessagingNonces :exec
DELETE FROM public.messaging_nonces WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountMessagingConfirmations :exec
DELETE FROM public.messaging_story_mutation_confirmations WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountMessagingEmailThreads :exec
DELETE FROM public.messaging_email_threads WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountSlackLinks :exec
DELETE FROM public.slack_user_links WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountNotifications :exec
DELETE FROM public.notifications WHERE recipient_id = sqlc.arg(user_id);

-- Other recipients retain their inbox history. Replace the actor snapshot at
-- its typed location; do not rewrite task titles or shared comment prose.
-- name: ReattributeAccountNotifications :exec
UPDATE public.notifications
SET actor_id = sqlc.arg(deleted_user_id),
    message = CASE WHEN jsonb_typeof(message #> '{variables,actor}') = 'object'
        THEN jsonb_set(message, '{variables,actor,value}', to_jsonb(CAST('Former user' AS text)))
        ELSE message END
WHERE actor_id = sqlc.arg(user_id) AND recipient_id <> sqlc.arg(user_id);

-- New snapshots carry an internal identity reference, making duplicate names
-- safe. Legacy assignee variables store only a display name, not an account ID.
-- Leave those legacy snapshots unchanged rather than guess at their identity.
-- name: AnonymizeAccountNotificationReferences :exec
UPDATE public.notifications AS notification
SET message = jsonb_set(jsonb_set(notification.message, '{variables,assignee,value}', to_jsonb(CAST('Former user' AS text))),
    '{identityReferences,assignee}', to_jsonb(CAST(sqlc.arg(deleted_user_id) AS uuid)))
WHERE notification.recipient_id <> sqlc.arg(user_id)
  AND notification.message #> '{identityReferences,assignee}' = to_jsonb(CAST(sqlc.arg(user_id) AS uuid));

-- name: DeleteAccountNotificationPreferences :exec
DELETE FROM public.notification_preferences WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountRoutineEmails :exec
DELETE FROM public.routine_email_deliveries WHERE recipient_id = sqlc.arg(user_id);

-- name: DeleteAccountNotificationEmails :exec
DELETE FROM public.notification_email_receipts WHERE recipient_id = sqlc.arg(user_id);

-- name: DeleteAccountInternalAlerts :exec
DELETE FROM public.internal_slack_alerts WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountEmailAvatarHandles :exec
DELETE FROM public.email_avatar_handles WHERE user_id = sqlc.arg(user_id);

-- name: ReattributeAccountActivities :exec
UPDATE public.story_activities SET user_id = sqlc.arg(deleted_user_id)
WHERE user_id = sqlc.arg(user_id);

-- name: ReattributeAccountObjectiveActivities :exec
UPDATE public.okr_activities SET user_id = sqlc.arg(deleted_user_id)
WHERE user_id = sqlc.arg(user_id);

-- Actor attribution and assignment history are separate. Replace identity
-- references even on edits authored by somebody else, preserving status/date
-- changes, prose, reasons and timestamps. These fields store UUID scalars.
-- name: AnonymizeAccountActivityReferences :exec
UPDATE public.story_activities
SET current_value = CASE WHEN field_changed = 'collaborator_ids'
        THEN replace(current_value, CAST(sqlc.arg(user_id) AS text), CAST(sqlc.arg(deleted_user_id) AS text))
        WHEN new_value = to_jsonb(CAST(sqlc.arg(user_id) AS text))
        OR current_value = CAST(sqlc.arg(user_id) AS text)
        THEN 'Former user' ELSE current_value END,
    old_value = CAST(replace(CAST(old_value AS text), CAST(sqlc.arg(user_id) AS text), CAST(sqlc.arg(deleted_user_id) AS text)) AS jsonb),
    new_value = CAST(replace(CAST(new_value AS text), CAST(sqlc.arg(user_id) AS text), CAST(sqlc.arg(deleted_user_id) AS text)) AS jsonb)
WHERE field_changed IN ('assignee_id', 'reporter_id', 'collaborator_ids')
  AND (CAST(old_value AS text) LIKE '%' || CAST(sqlc.arg(user_id) AS text) || '%'
       OR CAST(new_value AS text) LIKE '%' || CAST(sqlc.arg(user_id) AS text) || '%'
       OR current_value LIKE '%' || CAST(sqlc.arg(user_id) AS text) || '%');

-- name: AnonymizeAccountObjectiveActivityReferences :exec
UPDATE public.okr_activities
SET current_value = replace(current_value, CAST(sqlc.arg(user_id) AS text), CAST(sqlc.arg(deleted_user_id) AS text))
WHERE field_changed IN ('lead', 'lead_user_id', 'created_by', 'contributors')
  AND current_value LIKE '%' || CAST(sqlc.arg(user_id) AS text) || '%';

-- name: DeleteAccountCommentMentions :exec
DELETE FROM public.comment_mentions WHERE mentioned_user_id = sqlc.arg(user_id);

-- name: DeleteAccountStoryCollaborators :exec
DELETE FROM public.story_collaborators WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountStoryWatchers :exec
DELETE FROM public.story_watchers WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountStoryMutes :exec
DELETE FROM public.story_notification_mutes WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountDocumentMemberships :exec
DELETE FROM public.document_members WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountKeyResultContributions :exec
DELETE FROM public.key_result_contributors WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountFeedbackReads :exec
DELETE FROM public.feedback_item_reads WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountFeedbackSubscriptions :exec
DELETE FROM public.feedback_board_subscriptions WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountFeedbackDigests :exec
DELETE FROM public.feedback_digest_deliveries WHERE recipient_id = sqlc.arg(user_id);

-- name: DeleteAccountMayaRuns :exec
DELETE FROM public.maya_agent_runs WHERE triggered_by_user_id = sqlc.arg(user_id);

-- name: DeleteAccountMayaVoiceSessions :exec
DELETE FROM public.maya_realtime_voice_sessions WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountAnalytics :exec
DELETE FROM public.workspace_analytics_events WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountTeamMemberships :exec
DELETE FROM public.team_members WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountWorkspaceMemberships :exec
DELETE FROM public.workspace_members WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountWorkspaceWalkthroughs :exec
DELETE FROM public.user_onboarding_tour_progress WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountGlobalWalkthroughs :exec
DELETE FROM public.user_onboarding_tour_progress_global WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountAIUsageResets :exec
DELETE FROM public.user_ai_usage_resets WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountFigmaConnections :exec
DELETE FROM public.figma_connections WHERE connected_by_user_id = sqlc.arg(user_id);

-- name: DeleteAccountFigmaOAuthStates :exec
DELETE FROM public.figma_oauth_states WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountDriveOAuthStates :exec
DELETE FROM public.google_drive_oauth_states WHERE user_id = sqlc.arg(user_id);

-- name: DeleteAccountDriveAccounts :exec
DELETE FROM public.google_drive_accounts WHERE user_id = sqlc.arg(user_id);

-- name: ReattributeAccountWorkspacesCreatedBy :exec
UPDATE public.workspaces SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: ClearAccountWorkspacesDeletedBy :exec
UPDATE public.workspaces SET deleted_by = NULL WHERE deleted_by = sqlc.arg(user_id);

-- name: ReattributeAccountObjectivesCreatedBy :exec
UPDATE public.objectives SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: ClearAccountObjectivesLeadUserId :exec
UPDATE public.objectives SET lead_user_id = NULL WHERE lead_user_id = sqlc.arg(user_id);

-- name: ReattributeAccountKeyResultsCreatedBy :exec
UPDATE public.key_results SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: ClearAccountKeyResultsLead :exec
UPDATE public.key_results SET lead = NULL WHERE lead = sqlc.arg(user_id);

-- name: ReattributeAccountStoriesReporterId :exec
UPDATE public.stories SET reporter_id = sqlc.arg(deleted_user_id) WHERE reporter_id = sqlc.arg(user_id);

-- name: ClearAccountStoriesAssigneeId :exec
UPDATE public.stories SET assignee_id = NULL WHERE assignee_id = sqlc.arg(user_id);

-- name: ClearAccountAttachmentsUploadedBy :exec
UPDATE public.attachments SET uploaded_by = NULL WHERE uploaded_by = sqlc.arg(user_id);

-- name: ClearAccountIntegrationRequestsAcceptanceStartedByUserId :exec
UPDATE public.integration_requests SET acceptance_started_by_user_id = NULL WHERE acceptance_started_by_user_id = sqlc.arg(user_id);

-- name: DetachAccountDocumentsCreatedBy :exec
UPDATE public.documents SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: DetachAccountDocumentsUpdatedBy :exec
UPDATE public.documents SET updated_by = sqlc.arg(deleted_user_id) WHERE updated_by = sqlc.arg(user_id);

-- name: DetachAccountDocumentRelationshipsCreatedBy :exec
UPDATE public.document_relationships SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: DetachAccountDocumentAttachmentsCreatedBy :exec
UPDATE public.document_attachments SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: DetachAccountStoryInlineAttachmentsCreatedBy :exec
UPDATE public.story_inline_attachments SET created_by = sqlc.arg(deleted_user_id) WHERE created_by = sqlc.arg(user_id);

-- name: DetachAccountStoryFigmaLinksCreatedByUserId :exec
UPDATE public.story_figma_links SET created_by_user_id = sqlc.arg(deleted_user_id) WHERE created_by_user_id = sqlc.arg(user_id);

-- name: DetachAccountGoogleDriveFileReferencesCreatedByUserId :exec
UPDATE public.google_drive_file_references SET created_by_user_id = sqlc.arg(deleted_user_id) WHERE created_by_user_id = sqlc.arg(user_id);

-- name: DetachAccountGoogleDriveDocumentImportsImportedByUserId :exec
UPDATE public.google_drive_document_imports SET imported_by_user_id = sqlc.arg(deleted_user_id) WHERE imported_by_user_id = sqlc.arg(user_id);

-- name: DetachAccountUserAiUsageResetsResetByUserId :exec
UPDATE public.user_ai_usage_resets SET reset_by_user_id = sqlc.arg(deleted_user_id) WHERE reset_by_user_id = sqlc.arg(user_id);

-- name: EraseAccountFeedbackContributorSessions :exec
DELETE FROM public.feedback_contributor_sessions WHERE contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);

-- name: EraseAccountFeedbackItemFollowers :exec
DELETE FROM public.feedback_item_followers WHERE contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);

-- name: EraseAccountFeedbackPortalFollowers :exec
DELETE FROM public.feedback_portal_followers WHERE contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);

-- name: EraseAccountFeedbackContributorPreferences :exec
DELETE FROM public.feedback_contributor_preferences WHERE contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);

-- name: EraseAccountFeedbackContributorUnsubscribeTokens :exec
DELETE FROM public.feedback_contributor_unsubscribe_tokens WHERE contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);

-- name: EraseAccountFeedbackContributorDeliveries :exec
DELETE FROM public.feedback_contributor_deliveries WHERE contributor_id IN (
    SELECT id FROM public.feedback_contributors WHERE user_id = sqlc.arg(user_id)
);
