-- Secondary story mutations use one explicit authorization read inside their
-- owning transaction. Target rows are locked in stable ID order before any
-- state, activity, or durable-event write occurs.

-- name: AuthorizeSecondaryStoryTargets :many
WITH target AS MATERIALIZED (
    SELECT
        story.id,
        story.workspace_id,
        story.team_id,
        story.reporter_id,
        story.assignee_id,
        story.title,
        story.deleted_at,
        story.archived_at,
        story.updated_at
    FROM public.stories AS story
    WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
      AND story.workspace_id = sqlc.arg(workspace_id)
    ORDER BY story.id
    FOR UPDATE
)
SELECT
    target.id,
    target.workspace_id,
    target.team_id,
    target.reporter_id,
    target.assignee_id,
    target.title,
    target.deleted_at,
    target.archived_at,
    target.updated_at,
    CAST(COALESCE((
        SELECT CAST(workspace_member.role AS text)
        FROM public.workspace_members AS workspace_member
        WHERE workspace_member.workspace_id = target.workspace_id
          AND workspace_member.user_id = sqlc.arg(actor_id)
    ), '') AS text) AS actor_workspace_role,
    CAST((
        (
            CAST(sqlc.arg(actor_kind) AS text) = 'system'
            AND EXISTS (
                SELECT 1
                FROM public.users AS system_actor
                WHERE system_actor.user_id = sqlc.arg(actor_id)
                  AND system_actor.is_active = TRUE
                  AND system_actor.is_system = TRUE
            )
        ) OR (
            CAST(sqlc.arg(actor_kind) AS text) IN ('human_user', 'oauth_user')
            AND EXISTS (
                SELECT 1
                FROM public.users AS account
                INNER JOIN public.workspace_members AS workspace_member
                    ON workspace_member.workspace_id = target.workspace_id
                   AND workspace_member.user_id = account.user_id
                INNER JOIN public.team_members AS team_member
                    ON team_member.team_id = target.team_id
                   AND team_member.user_id = account.user_id
                WHERE account.user_id = sqlc.arg(actor_id)
                  AND account.is_active = TRUE
            )
        ) OR (
            CAST(sqlc.arg(actor_kind) AS text) = 'personal_token'
            AND EXISTS (
                SELECT 1
                FROM public.api_credentials AS credential
                INNER JOIN public.principals AS principal
                    ON principal.principal_id = credential.principal_id
                   AND principal.workspace_id = credential.workspace_id
                INNER JOIN public.users AS account
                    ON account.user_id = principal.subject_user_id
                   AND account.is_active = TRUE
                INNER JOIN public.workspace_members AS workspace_member
                    ON workspace_member.workspace_id = target.workspace_id
                   AND workspace_member.user_id = account.user_id
                INNER JOIN public.team_members AS team_member
                    ON team_member.team_id = target.team_id
                   AND team_member.user_id = account.user_id
                WHERE credential.credential_id = sqlc.arg(actor_credential_id)
                  AND credential.workspace_id = target.workspace_id
                  AND credential.kind = 'personal_access_token'
                  AND credential.revoked_at IS NULL
                  AND credential.expires_at > sqlc.arg(now)
                  AND principal.status = 'active'
                  AND principal.subject_user_id = sqlc.arg(actor_id)
                  AND EXISTS (
                      SELECT 1
                      FROM public.api_credential_scopes AS credential_scope
                      WHERE credential_scope.credential_id = credential.credential_id
                        AND credential_scope.scope = 'stories:write'
                  )
                  AND (
                      NOT EXISTS (
                          SELECT 1
                          FROM public.api_credential_team_restrictions AS restriction
                          WHERE restriction.credential_id = credential.credential_id
                      ) OR EXISTS (
                          SELECT 1
                          FROM public.api_credential_team_restrictions AS restriction
                          WHERE restriction.credential_id = credential.credential_id
                            AND restriction.workspace_id = target.workspace_id
                            AND restriction.team_id = target.team_id
                      )
                  )
            )
        ) OR (
            CAST(sqlc.arg(actor_kind) AS text) = 'service_account'
            AND EXISTS (
                SELECT 1
                FROM public.principals AS principal
                INNER JOIN public.api_credentials AS credential
                    ON credential.principal_id = principal.principal_id
                   AND credential.workspace_id = principal.workspace_id
                WHERE principal.principal_id = sqlc.arg(actor_id)
                  AND principal.workspace_id = target.workspace_id
                  AND principal.kind = 'service_account'
                  AND principal.status = 'active'
                  AND credential.credential_id = sqlc.arg(actor_credential_id)
                  AND credential.kind = 'service_account_key'
                  AND credential.revoked_at IS NULL
                  AND credential.expires_at > sqlc.arg(now)
                  AND EXISTS (
                      SELECT 1
                      FROM public.api_credential_scopes AS credential_scope
                      WHERE credential_scope.credential_id = credential.credential_id
                        AND credential_scope.scope = 'stories:write'
                  )
                  AND (
                      NOT EXISTS (
                          SELECT 1
                          FROM public.api_credential_team_restrictions AS restriction
                          WHERE restriction.credential_id = credential.credential_id
                      ) OR EXISTS (
                          SELECT 1
                          FROM public.api_credential_team_restrictions AS restriction
                          WHERE restriction.credential_id = credential.credential_id
                            AND restriction.workspace_id = target.workspace_id
                            AND restriction.team_id = target.team_id
                      )
                  )
            )
        )
    ) AS boolean) AS actor_authorized
FROM target
ORDER BY target.id;

-- name: SoftDeleteSecondaryStories :many
UPDATE public.stories AS story
SET
    deleted_at = sqlc.arg(changed_at),
    updated_at = sqlc.arg(changed_at)
WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
RETURNING story.id;

-- name: RestoreSecondaryStories :many
UPDATE public.stories AS story
SET
    deleted_at = NULL,
    updated_at = sqlc.arg(changed_at)
WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NOT NULL
RETURNING story.id;

-- name: ArchiveSecondaryStories :many
UPDATE public.stories AS story
SET
    archived_at = sqlc.arg(changed_at),
    updated_at = sqlc.arg(changed_at)
WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
RETURNING story.id;

-- name: UnarchiveSecondaryStories :many
UPDATE public.stories AS story
SET
    archived_at = NULL,
    updated_at = sqlc.arg(changed_at)
WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NOT NULL
RETURNING story.id;

-- name: HardDeleteSecondaryStories :many
DELETE FROM public.stories AS story
WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
  AND story.workspace_id = sqlc.arg(workspace_id)
RETURNING story.id;

-- name: ListStoryLabelsForUpdate :many
SELECT story_label.label_id
FROM public.story_labels AS story_label
WHERE story_label.story_id = sqlc.arg(story_id)
ORDER BY story_label.label_id;

-- name: DeleteStoryLabelsForUpdate :exec
DELETE FROM public.story_labels
WHERE story_id = sqlc.arg(story_id);

-- name: ListStoryCollaboratorsForUpdate :many
SELECT collaborator.user_id
FROM public.story_collaborators AS collaborator
WHERE collaborator.story_id = sqlc.arg(story_id)
ORDER BY collaborator.user_id;

-- name: CountValidStoryCollaborators :one
SELECT CAST(COUNT(DISTINCT team_member.user_id) AS bigint)
FROM unnest(CAST(sqlc.arg(collaborator_ids) AS uuid[])) AS requested(user_id)
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = sqlc.arg(team_id)
   AND team_member.user_id = requested.user_id
INNER JOIN public.users AS account
    ON account.user_id = team_member.user_id
   AND account.is_active = TRUE
   AND account.is_system = FALSE
WHERE team_member.user_id <> COALESCE(sqlc.narg(assignee_id), CAST('00000000-0000-0000-0000-000000000000' AS uuid));

-- name: DeleteStoryCollaboratorsForUpdate :exec
DELETE FROM public.story_collaborators
WHERE story_id = sqlc.arg(story_id);

-- name: InsertStoryCollaboratorsForUpdate :execrows
INSERT INTO public.story_collaborators (story_id, team_id, user_id)
SELECT
    sqlc.arg(story_id),
    sqlc.arg(team_id),
    requested.user_id
FROM unnest(CAST(sqlc.arg(collaborator_ids) AS uuid[])) AS requested(user_id)
ORDER BY requested.user_id;

-- name: GetStoryWatchStateForUpdate :one
SELECT
    story.team_id,
    CAST((
        story.assignee_id = sqlc.arg(actor_id)
        OR EXISTS (
            SELECT 1
            FROM public.story_collaborators AS collaborator
            WHERE collaborator.story_id = story.id
              AND collaborator.user_id = sqlc.arg(actor_id)
        )
    ) AS boolean) AS has_automatic_audience_role
FROM public.stories AS story
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
FOR UPDATE OF story;

-- name: DeleteStoryNotificationMute :exec
DELETE FROM public.story_notification_mutes
WHERE story_id = sqlc.arg(story_id)
  AND user_id = sqlc.arg(actor_id);

-- name: DeleteStoryWatcher :exec
DELETE FROM public.story_watchers
WHERE story_id = sqlc.arg(story_id)
  AND user_id = sqlc.arg(actor_id);

-- name: InsertStoryWatcher :exec
INSERT INTO public.story_watchers (story_id, team_id, user_id)
VALUES (sqlc.arg(story_id), sqlc.arg(team_id), sqlc.arg(actor_id))
ON CONFLICT (story_id, user_id) DO NOTHING;

-- name: InsertStoryNotificationMute :exec
INSERT INTO public.story_notification_mutes (story_id, team_id, user_id)
VALUES (sqlc.arg(story_id), sqlc.arg(team_id), sqlc.arg(actor_id))
ON CONFLICT (story_id, user_id) DO NOTHING;

-- name: ListStoryNotificationAudience :many
WITH target AS MATERIALIZED (
    SELECT
        story.id,
        story.workspace_id,
        story.team_id,
        story.assignee_id
    FROM public.stories AS story
    WHERE story.id = sqlc.arg(story_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
),
audience AS (
    SELECT target.assignee_id AS user_id
    FROM target
    UNION
    SELECT collaborator.user_id
    FROM target
    INNER JOIN public.story_collaborators AS collaborator
        ON collaborator.story_id = target.id
    UNION
    SELECT watcher.user_id
    FROM target
    INNER JOIN public.story_watchers AS watcher
        ON watcher.story_id = target.id
)
SELECT audience.user_id
FROM audience
INNER JOIN target ON TRUE
INNER JOIN public.users AS account
    ON account.user_id = audience.user_id
   AND account.is_active = TRUE
   AND account.is_system = FALSE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = target.workspace_id
   AND workspace_member.user_id = audience.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = target.team_id
   AND team_member.user_id = audience.user_id
WHERE audience.user_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_notification_mutes AS muted
      WHERE muted.story_id = target.id
        AND muted.user_id = audience.user_id
  )
ORDER BY audience.user_id;
