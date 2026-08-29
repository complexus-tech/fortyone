-- Story mutation queries deliberately repeat their authorization predicate.
-- Credentials only narrow current product membership; they never replace it.

-- name: GetStoryMutationSnapshot :one
WITH target AS (
    SELECT story.*
    FROM public.stories AS story
    WHERE story.id = sqlc.arg(story_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
), authorized AS (
    SELECT target.*
    FROM target
    WHERE (
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
                  )
                  OR EXISTS (
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
                  )
                  OR EXISTS (
                      SELECT 1
                      FROM public.api_credential_team_restrictions AS restriction
                      WHERE restriction.credential_id = credential.credential_id
                        AND restriction.workspace_id = target.workspace_id
                        AND restriction.team_id = target.team_id
                  )
              )
        )
    )
)
SELECT
    id,
    sequence_id,
    team_id,
    title,
    description,
    description_html,
    parent_id,
    objective_id,
    status_id,
    assignee_id,
    reporter_id,
    CAST(COALESCE(priority, 'No Priority') AS text) AS priority,
    sprint_id,
    key_result_id,
    estimate_unit,
    CAST(COALESCE((
        SELECT estimation.scheme
        FROM public.team_estimation_settings AS estimation
        WHERE estimation.team_id = authorized.team_id
          AND estimation.workspace_id = authorized.workspace_id
    ), 'tshirt') AS text) AS estimate_scheme,
    estimated_duration_minutes,
    minimum_focus_block_minutes,
    auto_scheduling_enabled,
    auto_scheduling_locked,
    auto_scheduling_status,
    auto_scheduling_reason,
    auto_scheduling_updated_at,
    start_date,
    end_date,
    completed_at,
    deleted_at,
    archived_at,
    created_at,
    updated_at,
    workspace_id,
    external_creation_key
FROM authorized;

-- name: AuthorizeStoryCreate :one
SELECT
    team.team_id,
    CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS estimate_scheme
FROM public.teams AS team
LEFT JOIN public.team_estimation_settings AS estimation
    ON estimation.team_id = team.team_id
   AND estimation.workspace_id = team.workspace_id
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND (
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
                  ON workspace_member.workspace_id = team.workspace_id
                 AND workspace_member.user_id = account.user_id
              INNER JOIN public.team_members AS team_member
                  ON team_member.team_id = team.team_id
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
                  ON workspace_member.workspace_id = team.workspace_id
                 AND workspace_member.user_id = account.user_id
              INNER JOIN public.team_members AS team_member
                  ON team_member.team_id = team.team_id
                 AND team_member.user_id = account.user_id
              WHERE credential.credential_id = sqlc.arg(actor_credential_id)
                AND credential.workspace_id = team.workspace_id
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
                    )
                    OR EXISTS (
                        SELECT 1
                        FROM public.api_credential_team_restrictions AS restriction
                        WHERE restriction.credential_id = credential.credential_id
                          AND restriction.workspace_id = team.workspace_id
                          AND restriction.team_id = team.team_id
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
                AND principal.workspace_id = team.workspace_id
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
                    )
                    OR EXISTS (
                        SELECT 1
                        FROM public.api_credential_team_restrictions AS restriction
                        WHERE restriction.credential_id = credential.credential_id
                          AND restriction.workspace_id = team.workspace_id
                          AND restriction.team_id = team.team_id
                    )
                )
          )
      ) OR (
          CAST(sqlc.arg(actor_kind) AS text) = 'oauth_application'
          AND EXISTS (
              SELECT 1
              FROM public.oauth_application_installations AS installation
              INNER JOIN public.principals AS principal
                  ON principal.principal_id = installation.principal_id
                 AND principal.workspace_id = installation.workspace_id
              INNER JOIN public.oauth_applications AS application
                  ON application.application_id = installation.application_id
              INNER JOIN public.oauth_application_installation_scopes AS installation_scope
                  ON installation_scope.installation_id = installation.installation_id
              WHERE installation.installation_id = sqlc.arg(actor_credential_id)
                AND installation.workspace_id = team.workspace_id
                AND installation.principal_id = sqlc.arg(actor_id)
                AND installation.status = 'active'
                AND principal.kind = 'oauth_application'
                AND principal.status = 'active'
                AND application.registration_kind = 'confidential'
                AND application.status = 'active'
                AND application.expires_at > sqlc.arg(now)
                AND installation_scope.scope = 'stories:write'
          )
      )
  );

-- name: GetStoryMutationKeyResult :one
SELECT
    key_result.objective_id,
    key_result.name
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid);

-- name: NextStorySequence :one
INSERT INTO public.team_story_sequences (
    workspace_id,
    team_id,
    current_sequence
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    1
)
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = team_story_sequences.current_sequence + 1
RETURNING current_sequence;

-- name: SynchronizeStorySequence :exec
INSERT INTO public.team_story_sequences (
    workspace_id,
    team_id,
    current_sequence
)
SELECT
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    CAST(COALESCE(MAX(story.sequence_id), 0) AS integer)
FROM public.stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.team_id = sqlc.arg(team_id)
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = GREATEST(
    team_story_sequences.current_sequence,
    EXCLUDED.current_sequence
);

-- name: GetStoryIDByCreationKey :one
SELECT story.id
FROM public.stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.external_creation_key = sqlc.arg(external_creation_key)
ORDER BY story.id
LIMIT 1;

-- name: GetOAuthApplicationStoryCreationReplay :one
SELECT
    story.id,
    story.sequence_id,
    story.team_id,
    story.title,
    story.description,
    story.description_html,
    story.parent_id,
    story.objective_id,
    story.status_id,
    story.assignee_id,
    story.reporter_id,
    CAST(COALESCE(story.priority, 'No Priority') AS text) AS priority,
    story.sprint_id,
    story.key_result_id,
    story.estimate_unit,
    CAST(COALESCE((
        SELECT estimation.scheme
        FROM public.team_estimation_settings AS estimation
        WHERE estimation.team_id = story.team_id
          AND estimation.workspace_id = story.workspace_id
    ), 'tshirt') AS text) AS estimate_scheme,
    story.estimated_duration_minutes,
    story.minimum_focus_block_minutes,
    story.auto_scheduling_enabled,
    story.auto_scheduling_locked,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_updated_at,
    story.start_date,
    story.end_date,
    story.completed_at,
    story.deleted_at,
    story.archived_at,
    story.created_at,
    story.updated_at,
    story.workspace_id,
    story.external_creation_key
FROM public.stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.external_creation_key = sqlc.arg(external_creation_key)
  AND story.deleted_at IS NULL
  AND CAST(sqlc.arg(actor_kind) AS text) = 'oauth_application'
  AND EXISTS (
      SELECT 1
      FROM public.oauth_application_installations AS installation
      INNER JOIN public.principals AS principal
          ON principal.principal_id = installation.principal_id
         AND principal.workspace_id = installation.workspace_id
      INNER JOIN public.oauth_applications AS application
          ON application.application_id = installation.application_id
      INNER JOIN public.oauth_application_installation_scopes AS installation_scope
          ON installation_scope.installation_id = installation.installation_id
      WHERE installation.installation_id = sqlc.arg(actor_credential_id)
        AND installation.workspace_id = story.workspace_id
        AND installation.principal_id = sqlc.arg(actor_id)
        AND installation.status = 'active'
        AND principal.kind = 'oauth_application'
        AND principal.status = 'active'
        AND application.registration_kind = 'confidential'
        AND application.status = 'active'
        AND application.expires_at > sqlc.arg(now)
        AND installation_scope.scope = 'stories:write'
  )
ORDER BY story.id
LIMIT 1;

-- name: CreateStoryMutation :one
INSERT INTO public.stories (
    id,
    sequence_id,
    title,
    description,
    description_html,
    parent_id,
    objective_id,
    status_id,
    assignee_id,
    blocked_by_id,
    blocking_id,
    related_id,
    reporter_id,
    priority,
    estimate_unit,
    estimated_duration_minutes,
    minimum_focus_block_minutes,
    auto_scheduling_enabled,
    auto_scheduling_locked,
    auto_scheduling_status,
    auto_scheduling_reason,
    auto_scheduling_updated_at,
    sprint_id,
    key_result_id,
    team_id,
    workspace_id,
    start_date,
    end_date,
    external_creation_key,
    created_at,
    updated_at
)
SELECT
    sqlc.arg(story_id),
    sqlc.arg(sequence_id),
    sqlc.arg(title),
    sqlc.narg(description),
    sqlc.narg(description_html),
    sqlc.narg(parent_id),
    sqlc.narg(objective_id),
    sqlc.narg(status_id),
    sqlc.narg(assignee_id),
    sqlc.narg(blocked_by_id),
    sqlc.narg(blocking_id),
    sqlc.narg(related_id),
    sqlc.narg(reporter_id),
    sqlc.arg(priority),
    sqlc.narg(estimate_unit),
    sqlc.narg(estimated_duration_minutes),
    sqlc.narg(minimum_focus_block_minutes),
    sqlc.arg(auto_scheduling_enabled),
    sqlc.arg(auto_scheduling_locked),
    sqlc.arg(auto_scheduling_status),
    sqlc.narg(auto_scheduling_reason),
    sqlc.narg(auto_scheduling_updated_at),
    sqlc.narg(sprint_id),
    sqlc.narg(key_result_id),
    sqlc.arg(team_id),
    sqlc.arg(workspace_id),
    sqlc.narg(start_date),
    sqlc.narg(end_date),
    sqlc.narg(external_creation_key),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
WHERE (
        CAST(sqlc.narg(status_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.statuses AS status
            WHERE status.status_id = sqlc.narg(status_id)
              AND status.team_id = sqlc.arg(team_id)
        )
    )
  AND (
        CAST(sqlc.narg(assignee_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.users AS assignee
            WHERE assignee.user_id = sqlc.narg(assignee_id)
              AND assignee.is_active = TRUE
              AND (
                  assignee.is_system = TRUE
                  OR (
                      EXISTS (
                          SELECT 1
                          FROM public.workspace_members AS workspace_member
                          WHERE workspace_member.workspace_id = sqlc.arg(workspace_id)
                            AND workspace_member.user_id = assignee.user_id
                      )
                      AND EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_member
                          WHERE team_member.team_id = sqlc.arg(team_id)
                            AND team_member.user_id = assignee.user_id
                      )
                  )
              )
        )
    )
  AND (
        CAST(sqlc.narg(reporter_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.users AS reporter
            WHERE reporter.user_id = sqlc.narg(reporter_id)
              AND reporter.is_active = TRUE
              AND (
                  reporter.is_system = TRUE
                  OR (
                      EXISTS (
                          SELECT 1
                          FROM public.workspace_members AS workspace_member
                          WHERE workspace_member.workspace_id = sqlc.arg(workspace_id)
                            AND workspace_member.user_id = reporter.user_id
                      )
                      AND EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_member
                          WHERE team_member.team_id = sqlc.arg(team_id)
                            AND team_member.user_id = reporter.user_id
                      )
                  )
              )
        )
    )
  AND (
        CAST(sqlc.narg(parent_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.stories AS parent
            WHERE parent.id = sqlc.narg(parent_id)
              AND parent.workspace_id = sqlc.arg(workspace_id)
              AND parent.team_id = sqlc.arg(team_id)
              AND parent.deleted_at IS NULL
        )
    )
  AND (
        CAST(sqlc.narg(objective_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.objectives AS objective
            WHERE objective.objective_id = sqlc.narg(objective_id)
              AND objective.workspace_id = sqlc.arg(workspace_id)
        )
    )
  AND (
        CAST(sqlc.narg(sprint_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.sprints AS sprint
            WHERE sprint.sprint_id = sqlc.narg(sprint_id)
              AND sprint.workspace_id = sqlc.arg(workspace_id)
              AND sprint.team_id = sqlc.arg(team_id)
        )
    )
  AND (
        CAST(sqlc.narg(key_result_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.key_results AS key_result
            INNER JOIN public.objectives AS objective
                ON objective.objective_id = key_result.objective_id
               AND objective.workspace_id = sqlc.arg(workspace_id)
            WHERE key_result.id = sqlc.narg(key_result_id)
              AND key_result.objective_id = sqlc.narg(objective_id)
        )
    )
  AND (
        (
            CAST(sqlc.narg(blocked_by_id) AS uuid) IS NULL
            OR EXISTS (
                SELECT 1
                FROM public.stories AS related_story
                WHERE related_story.id = sqlc.narg(blocked_by_id)
                  AND related_story.workspace_id = sqlc.arg(workspace_id)
                  AND related_story.team_id = sqlc.arg(team_id)
                  AND related_story.deleted_at IS NULL
            )
        )
        AND (
            CAST(sqlc.narg(blocking_id) AS uuid) IS NULL
            OR EXISTS (
                SELECT 1
                FROM public.stories AS related_story
                WHERE related_story.id = sqlc.narg(blocking_id)
                  AND related_story.workspace_id = sqlc.arg(workspace_id)
                  AND related_story.team_id = sqlc.arg(team_id)
                  AND related_story.deleted_at IS NULL
            )
        )
        AND (
            CAST(sqlc.narg(related_id) AS uuid) IS NULL
            OR EXISTS (
                SELECT 1
                FROM public.stories AS related_story
                WHERE related_story.id = sqlc.narg(related_id)
                  AND related_story.workspace_id = sqlc.arg(workspace_id)
                  AND related_story.team_id = sqlc.arg(team_id)
                  AND related_story.deleted_at IS NULL
            )
        )
    )
RETURNING
    id,
    sequence_id,
    title,
    description,
    description_html,
    parent_id,
    objective_id,
    status_id,
    assignee_id,
    blocked_by_id,
    blocking_id,
    related_id,
    reporter_id,
    CAST(COALESCE(priority, 'No Priority') AS text) AS priority,
    estimate_unit,
    estimated_duration_minutes,
    minimum_focus_block_minutes,
    auto_scheduling_enabled,
    auto_scheduling_locked,
    auto_scheduling_status,
    auto_scheduling_reason,
    auto_scheduling_updated_at,
    sprint_id,
    key_result_id,
    team_id,
    workspace_id,
    start_date,
    end_date,
    external_creation_key,
    created_at,
    updated_at;

-- name: InsertAuthorizedStoryLabels :many
INSERT INTO public.story_labels (story_id, label_id)
SELECT
    sqlc.arg(story_id),
    label.label_id
FROM public.labels AS label
WHERE label.label_id = ANY(CAST(sqlc.arg(label_ids) AS uuid[]))
  AND label.workspace_id = sqlc.arg(workspace_id)
  AND (label.team_id = sqlc.arg(team_id) OR label.team_id IS NULL)
ORDER BY label.label_id
ON CONFLICT (story_id, label_id) DO NOTHING
RETURNING label_id;

-- name: ApplyStoryPatch :one
WITH target AS (
    SELECT story.*
    FROM public.stories AS story
    WHERE story.id = sqlc.arg(story_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
), authorized AS (
    SELECT target.*
    FROM target
    WHERE (
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
                  )
                  OR EXISTS (
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
                  )
                  OR EXISTS (
                      SELECT 1
                      FROM public.api_credential_team_restrictions AS restriction
                      WHERE restriction.credential_id = credential.credential_id
                        AND restriction.workspace_id = target.workspace_id
                        AND restriction.team_id = target.team_id
                  )
              )
        )
    )
), valid AS (
    SELECT authorized.*
    FROM authorized
    WHERE (
        NOT CAST(sqlc.arg(set_status_id) AS boolean)
        OR CAST(sqlc.narg(status_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.statuses AS status
            WHERE status.status_id = sqlc.narg(status_id)
              AND status.team_id = authorized.team_id
        )
    )
      AND (
        NOT CAST(sqlc.arg(set_parent_id) AS boolean)
        OR CAST(sqlc.narg(parent_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.stories AS parent
            WHERE parent.id = sqlc.narg(parent_id)
              AND parent.id <> authorized.id
              AND parent.workspace_id = authorized.workspace_id
              AND parent.team_id = authorized.team_id
              AND parent.deleted_at IS NULL
        )
    )
      AND (
        NOT CAST(sqlc.arg(set_assignee_id) AS boolean)
        OR CAST(sqlc.narg(assignee_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.users AS assignee
            WHERE assignee.user_id = sqlc.narg(assignee_id)
              AND assignee.is_active = TRUE
              AND (
                  assignee.is_system = TRUE
                  OR (
                      EXISTS (
                          SELECT 1
                          FROM public.workspace_members AS workspace_member
                          WHERE workspace_member.workspace_id = authorized.workspace_id
                            AND workspace_member.user_id = assignee.user_id
                      )
                      AND EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_member
                          WHERE team_member.team_id = authorized.team_id
                            AND team_member.user_id = assignee.user_id
                      )
                  )
              )
        )
    )
      AND (
        NOT CAST(sqlc.arg(set_objective_id) AS boolean)
        OR CAST(sqlc.narg(objective_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.objectives AS objective
            WHERE objective.objective_id = sqlc.narg(objective_id)
              AND objective.workspace_id = authorized.workspace_id
        )
    )
      AND (
        NOT CAST(sqlc.arg(set_sprint_id) AS boolean)
        OR CAST(sqlc.narg(sprint_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.sprints AS sprint
            WHERE sprint.sprint_id = sqlc.narg(sprint_id)
              AND sprint.workspace_id = authorized.workspace_id
              AND sprint.team_id = authorized.team_id
        )
    )
      AND (
        NOT CAST(sqlc.arg(set_key_result_id) AS boolean)
        OR CAST(sqlc.narg(key_result_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.key_results AS key_result
            INNER JOIN public.objectives AS objective
                ON objective.objective_id = key_result.objective_id
               AND objective.workspace_id = authorized.workspace_id
            WHERE key_result.id = sqlc.narg(key_result_id)
        )
    )
      AND (
        (
            NOT CAST(sqlc.arg(set_objective_id) AS boolean)
            AND NOT CAST(sqlc.arg(set_key_result_id) AS boolean)
        )
        OR CASE
            WHEN CAST(sqlc.arg(set_key_result_id) AS boolean)
                THEN CAST(sqlc.narg(key_result_id) AS uuid)
            ELSE authorized.key_result_id
        END IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.key_results AS key_result
            WHERE key_result.id = CASE
                    WHEN CAST(sqlc.arg(set_key_result_id) AS boolean)
                        THEN CAST(sqlc.narg(key_result_id) AS uuid)
                    ELSE authorized.key_result_id
                END
              AND key_result.objective_id = CASE
                    WHEN CAST(sqlc.arg(set_objective_id) AS boolean)
                        THEN CAST(sqlc.narg(objective_id) AS uuid)
                    ELSE authorized.objective_id
                END
        )
    )
), updated AS (
    UPDATE public.stories AS story
    SET
        title = CASE WHEN CAST(sqlc.arg(set_title) AS boolean) THEN CAST(sqlc.arg(title) AS text) ELSE story.title END,
        estimate_unit = CASE WHEN CAST(sqlc.arg(set_estimate_unit) AS boolean) THEN sqlc.narg(estimate_unit) ELSE story.estimate_unit END,
        estimated_duration_minutes = CASE WHEN CAST(sqlc.arg(set_estimated_duration_minutes) AS boolean) THEN sqlc.narg(estimated_duration_minutes) ELSE story.estimated_duration_minutes END,
        minimum_focus_block_minutes = CASE WHEN CAST(sqlc.arg(set_minimum_focus_block_minutes) AS boolean) THEN sqlc.narg(minimum_focus_block_minutes) ELSE story.minimum_focus_block_minutes END,
        auto_scheduling_enabled = CASE WHEN CAST(sqlc.arg(set_auto_scheduling_enabled) AS boolean) THEN CAST(sqlc.arg(auto_scheduling_enabled) AS boolean) ELSE story.auto_scheduling_enabled END,
        auto_scheduling_locked = CASE WHEN CAST(sqlc.arg(set_auto_scheduling_locked) AS boolean) THEN CAST(sqlc.arg(auto_scheduling_locked) AS boolean) ELSE story.auto_scheduling_locked END,
        auto_scheduling_status = CASE WHEN CAST(sqlc.arg(set_auto_scheduling_status) AS boolean) THEN CAST(sqlc.arg(auto_scheduling_status) AS text) ELSE story.auto_scheduling_status END,
        auto_scheduling_reason = CASE WHEN CAST(sqlc.arg(set_auto_scheduling_reason) AS boolean) THEN sqlc.narg(auto_scheduling_reason) ELSE story.auto_scheduling_reason END,
        auto_scheduling_updated_at = CASE WHEN CAST(sqlc.arg(set_auto_scheduling_updated_at) AS boolean) THEN sqlc.narg(auto_scheduling_updated_at) ELSE story.auto_scheduling_updated_at END,
        description = CASE WHEN CAST(sqlc.arg(set_description) AS boolean) THEN sqlc.narg(description) ELSE story.description END,
        description_html = CASE WHEN CAST(sqlc.arg(set_description_html) AS boolean) THEN sqlc.narg(description_html) ELSE story.description_html END,
        parent_id = CASE WHEN CAST(sqlc.arg(set_parent_id) AS boolean) THEN sqlc.narg(parent_id) ELSE story.parent_id END,
        objective_id = CASE WHEN CAST(sqlc.arg(set_objective_id) AS boolean) THEN sqlc.narg(objective_id) ELSE story.objective_id END,
        status_id = CASE WHEN CAST(sqlc.arg(set_status_id) AS boolean) THEN sqlc.narg(status_id) ELSE story.status_id END,
        assignee_id = CASE WHEN CAST(sqlc.arg(set_assignee_id) AS boolean) THEN sqlc.narg(assignee_id) ELSE story.assignee_id END,
        priority = CASE WHEN CAST(sqlc.arg(set_priority) AS boolean) THEN CAST(sqlc.arg(priority) AS text) ELSE story.priority END,
        sprint_id = CASE WHEN CAST(sqlc.arg(set_sprint_id) AS boolean) THEN sqlc.narg(sprint_id) ELSE story.sprint_id END,
        key_result_id = CASE WHEN CAST(sqlc.arg(set_key_result_id) AS boolean) THEN sqlc.narg(key_result_id) ELSE story.key_result_id END,
        start_date = CASE WHEN CAST(sqlc.arg(set_start_date) AS boolean) THEN sqlc.narg(start_date) ELSE story.start_date END,
        end_date = CASE WHEN CAST(sqlc.arg(set_end_date) AS boolean) THEN sqlc.narg(end_date) ELSE story.end_date END,
        completed_at = CASE WHEN CAST(sqlc.arg(set_completed_at) AS boolean) THEN sqlc.narg(completed_at) ELSE story.completed_at END,
        updated_at = sqlc.arg(updated_at)
    FROM valid
    WHERE story.id = valid.id
      AND story.workspace_id = valid.workspace_id
      AND story.updated_at = sqlc.arg(expected_updated_at)
    RETURNING story.id, story.assignee_id, story.updated_at
), removed_assignee_collaboration AS (
    DELETE FROM public.story_collaborators AS collaborator
    USING updated
    WHERE collaborator.story_id = updated.id
      AND collaborator.user_id = updated.assignee_id
)
SELECT
    EXISTS (SELECT 1 FROM target) AS story_exists,
    EXISTS (SELECT 1 FROM authorized) AS actor_authorized,
    EXISTS (SELECT 1 FROM valid) AS references_valid,
    EXISTS (SELECT 1 FROM updated) AS story_updated;

-- name: DeleteStoryMutation :one
WITH target AS (
    SELECT story.*
    FROM public.stories AS story
    WHERE story.id = sqlc.arg(story_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
), authorized AS (
    SELECT target.*
    FROM target
    WHERE (
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
                  )
                  OR EXISTS (
                      SELECT 1
                      FROM public.api_credential_team_restrictions AS restriction
                      WHERE restriction.credential_id = credential.credential_id
                        AND restriction.workspace_id = target.workspace_id
                        AND restriction.team_id = target.team_id
                  )
              )
        )
    )
), permitted AS (
    SELECT authorized.*
    FROM authorized
    WHERE CAST(sqlc.arg(actor_kind) AS text) = 'system'
       OR authorized.reporter_id = sqlc.arg(actor_id)
       OR EXISTS (
           SELECT 1
           FROM public.workspace_members AS workspace_member
           WHERE workspace_member.workspace_id = authorized.workspace_id
             AND workspace_member.user_id = sqlc.arg(actor_id)
             AND workspace_member.role = 'admin'
       )
), deleted AS (
    UPDATE public.stories AS story
    SET
        deleted_at = sqlc.arg(deleted_at),
        updated_at = sqlc.arg(deleted_at)
    FROM permitted
    WHERE story.id = permitted.id
      AND story.workspace_id = permitted.workspace_id
      AND story.updated_at = sqlc.arg(expected_updated_at)
      AND story.deleted_at IS NULL
    RETURNING story.id, story.deleted_at, story.updated_at
)
SELECT
    EXISTS (SELECT 1 FROM target) AS story_exists,
    EXISTS (SELECT 1 FROM authorized) AS actor_authorized,
    EXISTS (SELECT 1 FROM permitted) AS deletion_permitted,
    EXISTS (SELECT 1 FROM deleted) AS story_deleted,
    (SELECT deleted_at FROM deleted) AS deleted_at;

-- name: UpsertStoryMutationActivity :one
WITH compacted AS (
    UPDATE public.story_activities AS activity
    SET
        current_value = sqlc.arg(current_value),
        new_value = sqlc.arg(new_value),
        reason = COALESCE(sqlc.narg(reason), activity.reason),
        created_at = CAST(sqlc.arg(created_at) AS timestamp)
    WHERE activity.activity_id = (
        SELECT candidate.activity_id
        FROM public.story_activities AS candidate
        WHERE candidate.story_id = sqlc.arg(story_id)
          AND candidate.user_id = sqlc.arg(user_id)
          AND candidate.workspace_id = sqlc.arg(workspace_id)
          AND candidate.activity_type = sqlc.arg(activity_type)
          AND candidate.field_changed = sqlc.arg(field_changed)
          AND candidate.created_at >= CAST(sqlc.arg(created_at) AS timestamp) - INTERVAL '30 seconds'
        ORDER BY candidate.created_at DESC, candidate.activity_id DESC
        LIMIT 1
        FOR UPDATE
    )
      AND CAST(sqlc.arg(compact) AS boolean)
    RETURNING activity.*
), inserted AS (
    INSERT INTO public.story_activities (
        activity_id,
        story_id,
        activity_type,
        field_changed,
        current_value,
        old_value,
        new_value,
        reason,
        user_id,
        workspace_id,
        created_at
    )
    SELECT
        sqlc.arg(activity_id),
        sqlc.arg(story_id),
        sqlc.arg(activity_type),
        sqlc.arg(field_changed),
        sqlc.arg(current_value),
        sqlc.arg(old_value),
        sqlc.arg(new_value),
        sqlc.narg(reason),
        sqlc.arg(user_id),
        sqlc.arg(workspace_id),
        CAST(sqlc.arg(created_at) AS timestamp)
    WHERE NOT EXISTS (SELECT 1 FROM compacted)
    RETURNING story_activities.*
)
SELECT * FROM compacted
UNION ALL
SELECT * FROM inserted;

-- name: InsertStoryMutationEvent :exec
INSERT INTO public.story_mutation_events (
    event_id,
    workspace_id,
    story_id,
    event_type,
    actor_kind,
    actor_id,
    actor_credential_id,
    payload,
    occurred_at,
    status,
    attempt_count,
    next_attempt_at,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(event_id),
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(event_type),
    sqlc.arg(actor_kind),
    sqlc.arg(actor_id),
    sqlc.narg(actor_credential_id),
    sqlc.arg(payload),
    sqlc.arg(occurred_at),
    'pending',
    0,
    sqlc.arg(created_at),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
);

-- name: LockStoryMediaLinks :many
SELECT
    media.attachment_id,
    attachment.mime_type
FROM public.story_inline_attachments AS media
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = media.attachment_id
   AND attachment.workspace_id = sqlc.arg(workspace_id)
WHERE media.story_id = sqlc.arg(story_id)
ORDER BY media.attachment_id
FOR UPDATE OF media, attachment;

-- name: DeleteStoryMediaLink :execrows
DELETE FROM public.story_inline_attachments
WHERE story_id = sqlc.arg(story_id)
  AND attachment_id = sqlc.arg(target_attachment_id);

-- name: StoryMediaAttachmentIsOrphaned :one
SELECT
    NOT EXISTS (
        SELECT 1
        FROM public.story_inline_attachments AS inline_attachment
        WHERE inline_attachment.attachment_id = sqlc.arg(target_attachment_id)
    )
    AND NOT EXISTS (
        SELECT 1
        FROM public.story_attachments AS story_attachment
        WHERE story_attachment.attachment_id = sqlc.arg(target_attachment_id)
    )
    AND NOT EXISTS (
        SELECT 1
        FROM public.document_attachments AS document_attachment
        WHERE document_attachment.attachment_id = sqlc.arg(target_attachment_id)
    ) AS orphaned;
