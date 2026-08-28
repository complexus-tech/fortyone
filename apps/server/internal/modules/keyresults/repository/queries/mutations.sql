-- name: GetKeyResultCreateScope :one
SELECT objective.team_id
FROM public.objectives AS objective
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
FOR UPDATE OF objective;

-- name: ValidateKeyResultAssignees :one
SELECT NOT EXISTS (
    SELECT candidate.user_id
    FROM unnest(CAST(sqlc.arg(user_ids) AS uuid[])) AS candidate(user_id)
    WHERE NOT EXISTS (
        SELECT 1
        FROM public.users AS account
        INNER JOIN public.workspace_members AS membership
            ON membership.user_id = account.user_id
           AND membership.workspace_id = sqlc.arg(workspace_id)
        INNER JOIN public.team_members AS team_membership
            ON team_membership.user_id = account.user_id
           AND team_membership.team_id = sqlc.arg(team_id)
        WHERE account.user_id = candidate.user_id
          AND account.is_active = TRUE
          AND membership.role IN ('member', 'admin')
    )
) AS valid;

-- name: AllocateKeyResultSequences :one
INSERT INTO public.team_key_result_sequences (
    workspace_id,
    team_id,
    current_sequence
)
SELECT
    team.workspace_id,
    team.team_id,
    CAST(sqlc.arg(sequence_count) AS integer)
FROM public.teams AS team
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND team.team_id = sqlc.arg(team_id)
ON CONFLICT (workspace_id, team_id) DO UPDATE
SET current_sequence = team_key_result_sequences.current_sequence + EXCLUDED.current_sequence
RETURNING current_sequence;

-- name: CreateKeyResult :one
INSERT INTO public.key_results (
    objective_id,
    team_id,
    sequence_id,
    name,
    measurement_type,
    start_value,
    current_value,
    target_value,
    lead,
    start_date,
    end_date,
    created_by
)
SELECT
    objective.objective_id,
    objective.team_id,
    CAST(sqlc.arg(sequence_id) AS integer),
    sqlc.arg(name),
    CAST(sqlc.arg(measurement_type) AS measurement_type),
    CAST(sqlc.arg(start_value) AS double precision),
    CAST(sqlc.arg(current_value) AS double precision),
    CAST(sqlc.arg(target_value) AS double precision),
    CAST(sqlc.narg(lead_user_id) AS uuid),
    CAST(sqlc.arg(start_date) AS date),
    CAST(sqlc.arg(end_date) AS date),
    actor.user_id
FROM public.objectives AS objective
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND objective.team_id = sqlc.arg(team_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
RETURNING
    id,
    sequence_id,
    objective_id,
    name,
    CAST(measurement_type AS text) AS measurement_type,
    CAST(COALESCE(start_value, 0) AS double precision) AS start_value,
    CAST(COALESCE(current_value, 0) AS double precision) AS current_value,
    CAST(COALESCE(target_value, 0) AS double precision) AS target_value,
    lead,
    start_date,
    end_date,
    created_at,
    updated_at,
    created_by;

-- name: AddKeyResultContributor :execrows
INSERT INTO public.key_result_contributors (key_result_id, user_id, created_at, updated_at)
SELECT key_result.id, account.user_id, clock_timestamp(), clock_timestamp()
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
   AND objective.team_id = key_result.team_id
INNER JOIN public.users AS account
    ON account.user_id = sqlc.arg(user_id)
   AND account.is_active = TRUE
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = account.user_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = account.user_id
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
ON CONFLICT (key_result_id, user_id) DO NOTHING;

-- name: CreateKeyResultActivity :execrows
INSERT INTO public.okr_activities (
    objective_id,
    key_result_id,
    user_id,
    activity_type,
    update_type,
    field_changed,
    current_value,
    comment,
    workspace_id
)
SELECT
    objective.objective_id,
    key_result.id,
    actor.user_id,
    CAST(sqlc.arg(activity_type) AS okr_activity_type),
    'key_result'::okr_update_type,
    CAST(sqlc.narg(field_changed) AS text),
    CAST(sqlc.narg(current_value) AS text),
    CAST(sqlc.narg(comment) AS text),
    objective.workspace_id
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
   AND objective.team_id = key_result.team_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = actor.user_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = actor.user_id
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin');

-- name: GetKeyResultForMutation :one
SELECT
    key_result.id,
    key_result.sequence_id,
    key_result.objective_id,
    key_result.name,
    CAST(key_result.measurement_type AS text) AS measurement_type,
    CAST(COALESCE(key_result.start_value, 0) AS double precision) AS start_value,
    CAST(COALESCE(key_result.current_value, 0) AS double precision) AS current_value,
    CAST(COALESCE(key_result.target_value, 0) AS double precision) AS target_value,
    key_result.lead,
    ARRAY(
        SELECT contributor.user_id
        FROM public.key_result_contributors AS contributor
        WHERE contributor.key_result_id = key_result.id
        ORDER BY contributor.created_at, contributor.user_id
    )::uuid[] AS contributor_ids,
    key_result.start_date,
    key_result.end_date,
    key_result.created_at,
    key_result.updated_at,
    key_result.created_by,
    objective.team_id
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
   AND objective.team_id = key_result.team_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
FOR UPDATE OF key_result;

-- name: UpdateKeyResult :one
UPDATE public.key_results AS key_result
SET
    name = CASE WHEN CAST(sqlc.arg(set_name) AS boolean) THEN sqlc.arg(name) ELSE key_result.name END,
    measurement_type = CASE WHEN CAST(sqlc.arg(set_measurement_type) AS boolean) THEN CAST(sqlc.arg(measurement_type) AS measurement_type) ELSE key_result.measurement_type END,
    start_value = CASE WHEN CAST(sqlc.arg(set_start_value) AS boolean) THEN CAST(sqlc.arg(start_value) AS double precision) ELSE key_result.start_value END,
    current_value = CASE WHEN CAST(sqlc.arg(set_current_value) AS boolean) THEN CAST(sqlc.arg(current_value) AS double precision) ELSE key_result.current_value END,
    target_value = CASE WHEN CAST(sqlc.arg(set_target_value) AS boolean) THEN CAST(sqlc.arg(target_value) AS double precision) ELSE key_result.target_value END,
    lead = CASE WHEN CAST(sqlc.arg(set_lead) AS boolean) THEN CAST(sqlc.narg(lead_user_id) AS uuid) ELSE key_result.lead END,
    start_date = CASE WHEN CAST(sqlc.arg(set_start_date) AS boolean) THEN CAST(sqlc.arg(start_date) AS date) ELSE key_result.start_date END,
    end_date = CASE WHEN CAST(sqlc.arg(set_end_date) AS boolean) THEN CAST(sqlc.arg(end_date) AS date) ELSE key_result.end_date END,
    updated_at = GREATEST(clock_timestamp(), key_result.updated_at + INTERVAL '1 microsecond')
FROM public.objectives AS objective,
     public.workspace_members AS membership,
     public.team_members AS team_membership,
     public.users AS actor
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.objective_id = key_result.objective_id
  AND objective.team_id = key_result.team_id
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.workspace_id = objective.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role IN ('member', 'admin')
  AND team_membership.team_id = objective.team_id
  AND team_membership.user_id = membership.user_id
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
RETURNING
    key_result.id,
    key_result.sequence_id,
    key_result.objective_id,
    key_result.name,
    CAST(key_result.measurement_type AS text) AS measurement_type,
    CAST(COALESCE(key_result.start_value, 0) AS double precision) AS start_value,
    CAST(COALESCE(key_result.current_value, 0) AS double precision) AS current_value,
    CAST(COALESCE(key_result.target_value, 0) AS double precision) AS target_value,
    key_result.lead,
    key_result.start_date,
    key_result.end_date,
    key_result.created_at,
    key_result.updated_at,
    key_result.created_by;

-- name: DeleteKeyResultContributors :exec
DELETE FROM public.key_result_contributors AS contributor
USING public.key_results AS key_result, public.objectives AS objective
WHERE contributor.key_result_id = key_result.id
  AND key_result.id = sqlc.arg(key_result_id)
  AND objective.objective_id = key_result.objective_id
  AND objective.workspace_id = sqlc.arg(workspace_id);

-- name: DeleteKeyResult :one
DELETE FROM public.key_results AS key_result
USING public.objectives AS objective,
      public.workspace_members AS membership,
      public.team_members AS team_membership,
      public.users AS actor
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.objective_id = key_result.objective_id
  AND objective.team_id = key_result.team_id
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.workspace_id = objective.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role IN ('member', 'admin')
  AND team_membership.team_id = objective.team_id
  AND team_membership.user_id = membership.user_id
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
RETURNING key_result.objective_id, key_result.name;

-- name: CreateDeletedKeyResultActivity :execrows
INSERT INTO public.okr_activities (
    objective_id,
    key_result_id,
    user_id,
    activity_type,
    update_type,
    field_changed,
    current_value,
    comment,
    workspace_id
)
SELECT
    objective.objective_id,
    NULL,
    actor.user_id,
    'delete'::okr_activity_type,
    'key_result'::okr_update_type,
    'all',
    sqlc.arg(current_value),
    '',
    objective.workspace_id
FROM public.objectives AS objective
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = actor.user_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = actor.user_id
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin');
