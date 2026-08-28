-- Key-result reads repeat tenant, current membership, product team membership,
-- and credential team restrictions. IDs are never sufficient authorization.

-- name: GetKeyResult :one
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
    key_result.created_by
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
  );

-- name: ListObjectiveKeyResults :many
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
    key_result.created_by
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
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY key_result.created_at DESC, key_result.id DESC;

-- name: ListKeyResults :many
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
    objective.name AS objective_name,
    objective.team_id,
    team.name AS team_name,
    team.code AS team_code,
    objective.workspace_id
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
   AND objective.team_id = key_result.team_id
INNER JOIN public.teams AS team
    ON team.team_id = objective.team_id
   AND team.workspace_id = objective.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(filter_team_ids) AS uuid[])), 0) = 0
      OR objective.team_id = ANY(CAST(sqlc.arg(filter_team_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])), 0) = 0
      OR key_result.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(measurement_types) AS text[])), 0) = 0
      OR CAST(key_result.measurement_type AS text) = ANY(CAST(sqlc.arg(measurement_types) AS text[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(lead_ids) AS uuid[])), 0) = 0
      OR key_result.lead = ANY(CAST(sqlc.arg(lead_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(created_by_ids) AS uuid[])), 0) = 0
      OR key_result.created_by = ANY(CAST(sqlc.arg(created_by_ids) AS uuid[]))
  )
  AND (CAST(sqlc.narg(created_after) AS timestamptz) IS NULL OR key_result.created_at >= CAST(sqlc.narg(created_after) AS timestamptz))
  AND (CAST(sqlc.narg(created_before) AS timestamptz) IS NULL OR key_result.created_at <= CAST(sqlc.narg(created_before) AS timestamptz))
  AND (CAST(sqlc.narg(updated_after) AS timestamptz) IS NULL OR key_result.updated_at >= CAST(sqlc.narg(updated_after) AS timestamptz))
  AND (CAST(sqlc.narg(updated_before) AS timestamptz) IS NULL OR key_result.updated_at <= CAST(sqlc.narg(updated_before) AS timestamptz))
  AND (CAST(sqlc.narg(end_date_after) AS date) IS NULL OR key_result.end_date >= CAST(sqlc.narg(end_date_after) AS date))
  AND (CAST(sqlc.narg(end_date_before) AS date) IS NULL OR key_result.end_date <= CAST(sqlc.narg(end_date_before) AS date))
ORDER BY
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'name_asc' THEN lower(key_result.name) END ASC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'name_desc' THEN lower(key_result.name) END DESC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'created_at_asc' THEN key_result.created_at END ASC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'created_at_desc' THEN key_result.created_at END DESC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'updated_at_asc' THEN key_result.updated_at END ASC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'updated_at_desc' THEN key_result.updated_at END DESC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'objective_name_asc' THEN lower(objective.name) END ASC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'objective_name_desc' THEN lower(objective.name) END DESC,
    key_result.created_at DESC,
    key_result.id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer)
OFFSET CAST(sqlc.arg(result_offset) AS integer);

-- name: CountKeyResults :one
SELECT COUNT(*)
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
WHERE objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR objective.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(filter_team_ids) AS uuid[])), 0) = 0
      OR objective.team_id = ANY(CAST(sqlc.arg(filter_team_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])), 0) = 0
      OR key_result.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(measurement_types) AS text[])), 0) = 0
      OR CAST(key_result.measurement_type AS text) = ANY(CAST(sqlc.arg(measurement_types) AS text[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(lead_ids) AS uuid[])), 0) = 0
      OR key_result.lead = ANY(CAST(sqlc.arg(lead_ids) AS uuid[]))
  )
  AND (
      COALESCE(cardinality(CAST(sqlc.arg(created_by_ids) AS uuid[])), 0) = 0
      OR key_result.created_by = ANY(CAST(sqlc.arg(created_by_ids) AS uuid[]))
  )
  AND (CAST(sqlc.narg(created_after) AS timestamptz) IS NULL OR key_result.created_at >= CAST(sqlc.narg(created_after) AS timestamptz))
  AND (CAST(sqlc.narg(created_before) AS timestamptz) IS NULL OR key_result.created_at <= CAST(sqlc.narg(created_before) AS timestamptz))
  AND (CAST(sqlc.narg(updated_after) AS timestamptz) IS NULL OR key_result.updated_at >= CAST(sqlc.narg(updated_after) AS timestamptz))
  AND (CAST(sqlc.narg(updated_before) AS timestamptz) IS NULL OR key_result.updated_at <= CAST(sqlc.narg(updated_before) AS timestamptz))
  AND (CAST(sqlc.narg(end_date_after) AS date) IS NULL OR key_result.end_date >= CAST(sqlc.narg(end_date_after) AS date))
  AND (CAST(sqlc.narg(end_date_before) AS date) IS NULL OR key_result.end_date <= CAST(sqlc.narg(end_date_before) AS date));
