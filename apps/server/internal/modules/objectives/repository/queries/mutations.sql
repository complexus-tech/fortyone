-- name: GetObjectiveCreateAuthorization :one
SELECT
    EXISTS (
        SELECT 1 FROM public.teams AS team
        WHERE team.team_id = sqlc.arg(team_id)
          AND team.workspace_id = sqlc.arg(workspace_id)
    ) AS team_exists,
    EXISTS (
        SELECT 1
        FROM public.workspace_members AS membership
        INNER JOIN public.team_members AS team_membership
            ON team_membership.team_id = sqlc.arg(team_id)
           AND team_membership.user_id = membership.user_id
        INNER JOIN public.users AS actor
            ON actor.user_id = membership.user_id
           AND actor.is_active = TRUE
        WHERE membership.workspace_id = sqlc.arg(workspace_id)
          AND membership.user_id = sqlc.arg(actor_id)
          AND membership.role IN ('member', 'admin')
    ) AS actor_authorized,
    EXISTS (
        SELECT 1 FROM public.objective_statuses AS status
        WHERE status.status_id = sqlc.arg(status_id)
          AND status.workspace_id = sqlc.arg(workspace_id)
    ) AS status_valid,
    (
        CAST(sqlc.narg(lead_user_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.users AS lead_user
            INNER JOIN public.workspace_members AS membership
                ON membership.user_id = lead_user.user_id
               AND membership.workspace_id = sqlc.arg(workspace_id)
            INNER JOIN public.team_members AS team_membership
                ON team_membership.user_id = lead_user.user_id
               AND team_membership.team_id = sqlc.arg(team_id)
            WHERE lead_user.user_id = CAST(sqlc.narg(lead_user_id) AS uuid)
              AND lead_user.is_active = TRUE
        )
    ) AS lead_valid;

-- name: CountAssignableObjectiveUsers :one
WITH requested_user AS (
    SELECT DISTINCT requested.user_id
    FROM unnest(CAST(sqlc.arg(user_ids) AS uuid[])) AS requested(user_id)
)
SELECT CAST(COUNT(*) AS integer) AS assignable_count
FROM requested_user
INNER JOIN public.users AS account
    ON account.user_id = requested_user.user_id
   AND account.is_active = TRUE
INNER JOIN public.workspace_members AS membership
    ON membership.user_id = account.user_id
   AND membership.workspace_id = sqlc.arg(workspace_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.user_id = account.user_id
   AND team_membership.team_id = sqlc.arg(team_id);

-- name: AllocateObjectiveSequence :one
INSERT INTO public.team_objective_sequences (workspace_id, team_id, current_sequence)
VALUES (sqlc.arg(workspace_id), sqlc.arg(team_id), 1)
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = team_objective_sequences.current_sequence + 1
RETURNING current_sequence;

-- name: CreateObjective :one
INSERT INTO public.objectives (
    sequence_id, name, description, short_summary, lead_user_id, team_id,
    workspace_id, start_date, end_date, is_private, status_id, priority, color, created_by
) SELECT
    sqlc.arg(sequence_id), sqlc.arg(name), CAST(sqlc.narg(description) AS text),
    CAST(sqlc.narg(short_summary) AS text), CAST(sqlc.narg(lead_user_id) AS uuid),
    sqlc.arg(team_id), sqlc.arg(workspace_id), CAST(sqlc.narg(start_date) AS date),
    CAST(sqlc.narg(end_date) AS date), sqlc.arg(is_private), sqlc.arg(status_id),
    CAST(sqlc.narg(priority) AS text), sqlc.arg(color), sqlc.arg(actor_id)
WHERE EXISTS (
	SELECT 1
	FROM public.teams AS team
	INNER JOIN public.workspace_members AS membership
		ON membership.workspace_id = team.workspace_id
	INNER JOIN public.team_members AS team_membership
		ON team_membership.team_id = team.team_id
	   AND team_membership.user_id = membership.user_id
	INNER JOIN public.users AS actor
		ON actor.user_id = membership.user_id
	   AND actor.is_active = TRUE
	WHERE team.team_id = sqlc.arg(team_id)
	  AND team.workspace_id = sqlc.arg(workspace_id)
	  AND membership.user_id = sqlc.arg(actor_id)
	  AND membership.role IN ('member', 'admin')
)
AND EXISTS (
	SELECT 1
	FROM public.objective_statuses AS status
	WHERE status.status_id = sqlc.arg(status_id)
	  AND status.workspace_id = sqlc.arg(workspace_id)
)
AND (
	CAST(sqlc.narg(lead_user_id) AS uuid) IS NULL
	OR EXISTS (
		SELECT 1
		FROM public.users AS lead_user
		INNER JOIN public.workspace_members AS membership
			ON membership.user_id = lead_user.user_id
		   AND membership.workspace_id = sqlc.arg(workspace_id)
		INNER JOIN public.team_members AS team_membership
			ON team_membership.user_id = lead_user.user_id
		   AND team_membership.team_id = sqlc.arg(team_id)
		WHERE lead_user.user_id = CAST(sqlc.narg(lead_user_id) AS uuid)
		  AND lead_user.is_active = TRUE
	)
)
RETURNING
    objective_id, sequence_id, name, description, short_summary, lead_user_id,
    team_id, workspace_id, start_date, end_date, is_private, created_at,
    updated_at, status_id, priority, CAST(COALESCE(CAST(health AS text), '') AS text) AS health, color, created_by;

-- name: AllocateKeyResultSequences :one
INSERT INTO public.team_key_result_sequences (workspace_id, team_id, current_sequence)
VALUES (sqlc.arg(workspace_id), sqlc.arg(team_id), sqlc.arg(sequence_count))
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = team_key_result_sequences.current_sequence + EXCLUDED.current_sequence
RETURNING current_sequence;

-- name: CreateObjectiveKeyResult :one
INSERT INTO public.key_results (
    objective_id, team_id, sequence_id, name, measurement_type, start_value,
    current_value, target_value, lead, start_date, end_date, created_by
) SELECT
    sqlc.arg(objective_id), sqlc.arg(team_id), sqlc.arg(sequence_id), sqlc.arg(name),
    CAST(sqlc.arg(measurement_type) AS measurement_type),
    CAST(sqlc.arg(start_value) AS double precision),
    CAST(sqlc.arg(current_value) AS double precision),
    CAST(sqlc.arg(target_value) AS double precision),
    CAST(sqlc.narg(lead_user_id) AS uuid), sqlc.arg(start_date), sqlc.arg(end_date),
    sqlc.arg(actor_id)
WHERE EXISTS (
	SELECT 1
	FROM public.objectives AS objective
	INNER JOIN public.teams AS team
		ON team.team_id = objective.team_id
	   AND team.workspace_id = objective.workspace_id
	INNER JOIN public.workspace_members AS membership
		ON membership.workspace_id = objective.workspace_id
	INNER JOIN public.team_members AS team_membership
		ON team_membership.team_id = objective.team_id
	   AND team_membership.user_id = membership.user_id
	INNER JOIN public.users AS actor
		ON actor.user_id = membership.user_id
	   AND actor.is_active = TRUE
	WHERE objective.objective_id = sqlc.arg(objective_id)
	  AND objective.workspace_id = sqlc.arg(workspace_id)
	  AND objective.team_id = sqlc.arg(team_id)
	  AND membership.user_id = sqlc.arg(actor_id)
	  AND membership.role IN ('member', 'admin')
)
AND (
	CAST(sqlc.narg(lead_user_id) AS uuid) IS NULL
	OR EXISTS (
		SELECT 1
		FROM public.users AS lead_user
		INNER JOIN public.workspace_members AS membership
			ON membership.user_id = lead_user.user_id
		   AND membership.workspace_id = sqlc.arg(workspace_id)
		INNER JOIN public.team_members AS team_membership
			ON team_membership.user_id = lead_user.user_id
		   AND team_membership.team_id = sqlc.arg(team_id)
		WHERE lead_user.user_id = CAST(sqlc.narg(lead_user_id) AS uuid)
		  AND lead_user.is_active = TRUE
	)
)
RETURNING
    id, sequence_id, objective_id, name, CAST(measurement_type AS text) AS measurement_type,
    CAST(start_value AS double precision) AS start_value,
    CAST(current_value AS double precision) AS current_value,
    CAST(target_value AS double precision) AS target_value,
    lead, start_date, end_date, created_at, updated_at, created_by;

-- name: AddObjectiveKeyResultContributor :one
INSERT INTO public.key_result_contributors (key_result_id, user_id, created_at, updated_at)
SELECT key_result.id, contributor.user_id, clock_timestamp(), clock_timestamp()
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
	ON objective.objective_id = key_result.objective_id
	AND objective.team_id = sqlc.arg(team_id)
	AND objective.workspace_id = sqlc.arg(workspace_id)
INNER JOIN public.users AS contributor
	ON contributor.user_id = sqlc.arg(user_id)
	AND contributor.is_active = TRUE
INNER JOIN public.workspace_members AS contributor_workspace
	ON contributor_workspace.user_id = contributor.user_id
	AND contributor_workspace.workspace_id = objective.workspace_id
INNER JOIN public.team_members AS contributor_team
	ON contributor_team.user_id = contributor.user_id
	AND contributor_team.team_id = objective.team_id
WHERE key_result.id = sqlc.arg(key_result_id)
	AND EXISTS (
		SELECT 1
		FROM public.workspace_members AS actor_workspace
		INNER JOIN public.team_members AS actor_team
			ON actor_team.team_id = objective.team_id
		   AND actor_team.user_id = actor_workspace.user_id
		INNER JOIN public.users AS actor
			ON actor.user_id = actor_workspace.user_id
		   AND actor.is_active = TRUE
		WHERE actor_workspace.workspace_id = objective.workspace_id
		  AND actor_workspace.user_id = sqlc.arg(actor_id)
		  AND actor_workspace.role IN ('member', 'admin')
	)
ON CONFLICT (key_result_id, user_id) DO UPDATE
SET updated_at = key_result_contributors.updated_at
RETURNING user_id;

-- name: GetObjectiveForMutation :one
SELECT
    objective.objective_id,
    objective.updated_at,
    objective.team_id,
	objective.lead_user_id,
	objective.start_date,
	objective.end_date
FROM public.objectives AS objective
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
FOR UPDATE OF objective;

-- name: ValidateObjectivePatchReferences :one
SELECT
    (
        CAST(sqlc.arg(status_specified) AS boolean) = FALSE
        OR EXISTS (
            SELECT 1 FROM public.objective_statuses AS status
            WHERE status.status_id = CAST(sqlc.narg(status_id) AS uuid)
              AND status.workspace_id = sqlc.arg(workspace_id)
        )
    ) AS status_valid,
    (
        CAST(sqlc.arg(lead_specified) AS boolean) = FALSE
        OR CAST(sqlc.narg(lead_user_id) AS uuid) IS NULL
        OR EXISTS (
            SELECT 1
            FROM public.users AS lead_user
            INNER JOIN public.workspace_members AS membership
                ON membership.user_id = lead_user.user_id
               AND membership.workspace_id = sqlc.arg(workspace_id)
            INNER JOIN public.team_members AS team_membership
                ON team_membership.user_id = lead_user.user_id
               AND team_membership.team_id = sqlc.arg(team_id)
            WHERE lead_user.user_id = CAST(sqlc.narg(lead_user_id) AS uuid)
              AND lead_user.is_active = TRUE
        )
    ) AS lead_valid;

-- name: UpdateObjective :one
UPDATE public.objectives AS objective
SET
    name = CASE WHEN CAST(sqlc.arg(set_name) AS boolean) THEN sqlc.arg(name) ELSE objective.name END,
    description = CASE WHEN CAST(sqlc.arg(set_description) AS boolean) THEN CAST(sqlc.narg(description) AS text) ELSE objective.description END,
    short_summary = CASE WHEN CAST(sqlc.arg(set_short_summary) AS boolean) THEN CAST(sqlc.narg(short_summary) AS text) ELSE objective.short_summary END,
    lead_user_id = CASE WHEN CAST(sqlc.arg(set_lead_user_id) AS boolean) THEN CAST(sqlc.narg(lead_user_id) AS uuid) ELSE objective.lead_user_id END,
    start_date = CASE WHEN CAST(sqlc.arg(set_start_date) AS boolean) THEN CAST(sqlc.narg(start_date) AS date) ELSE objective.start_date END,
    end_date = CASE WHEN CAST(sqlc.arg(set_end_date) AS boolean) THEN CAST(sqlc.narg(end_date) AS date) ELSE objective.end_date END,
    is_private = CASE WHEN CAST(sqlc.arg(set_is_private) AS boolean) THEN sqlc.arg(is_private) ELSE objective.is_private END,
    status_id = CASE WHEN CAST(sqlc.arg(set_status_id) AS boolean) THEN sqlc.arg(status_id) ELSE objective.status_id END,
    priority = CASE WHEN CAST(sqlc.arg(set_priority) AS boolean) THEN CAST(sqlc.narg(priority) AS text) ELSE objective.priority END,
    health = CASE WHEN CAST(sqlc.arg(set_health) AS boolean) THEN CAST(sqlc.narg(health) AS objective_health_status) ELSE objective.health END,
    color = CASE WHEN CAST(sqlc.arg(set_color) AS boolean) THEN sqlc.arg(color) ELSE objective.color END,
    updated_at = GREATEST(clock_timestamp(), objective.updated_at + INTERVAL '1 microsecond')
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
	AND EXISTS (
		SELECT 1
		FROM public.workspace_members AS membership
		INNER JOIN public.team_members AS team_membership
			ON team_membership.team_id = objective.team_id
		   AND team_membership.user_id = membership.user_id
		INNER JOIN public.users AS actor
			ON actor.user_id = membership.user_id
		   AND actor.is_active = TRUE
		WHERE membership.workspace_id = objective.workspace_id
		  AND membership.user_id = sqlc.arg(actor_id)
		  AND membership.role IN ('member', 'admin')
	)
	AND (
		CAST(sqlc.arg(set_status_id) AS boolean) = FALSE
		OR EXISTS (
			SELECT 1
			FROM public.objective_statuses AS status
			WHERE status.status_id = CAST(sqlc.narg(status_id) AS uuid)
			  AND status.workspace_id = objective.workspace_id
		)
	)
	AND (
		CAST(sqlc.arg(set_lead_user_id) AS boolean) = FALSE
		OR CAST(sqlc.narg(lead_user_id) AS uuid) IS NULL
		OR EXISTS (
			SELECT 1
			FROM public.users AS lead_user
			INNER JOIN public.workspace_members AS membership
				ON membership.user_id = lead_user.user_id
			   AND membership.workspace_id = objective.workspace_id
			INNER JOIN public.team_members AS team_membership
				ON team_membership.user_id = lead_user.user_id
			   AND team_membership.team_id = objective.team_id
			WHERE lead_user.user_id = CAST(sqlc.narg(lead_user_id) AS uuid)
			  AND lead_user.is_active = TRUE
		)
	)
RETURNING
    objective_id, sequence_id, name, description, short_summary, lead_user_id,
    team_id, workspace_id, start_date, end_date, is_private, created_at,
    updated_at, status_id, priority, CAST(COALESCE(CAST(health AS text), '') AS text) AS health, color, created_by;

-- name: CreateObjectiveActivity :exec
INSERT INTO public.okr_activities (
    objective_id, key_result_id, user_id, activity_type, update_type,
    field_changed, current_value, comment, workspace_id
) VALUES (
    sqlc.arg(objective_id), CAST(sqlc.narg(key_result_id) AS uuid), sqlc.arg(actor_id),
    CAST(sqlc.arg(activity_type) AS okr_activity_type),
    CAST(sqlc.arg(update_type) AS okr_update_type), CAST(sqlc.narg(field_changed) AS text),
    CAST(sqlc.narg(current_value) AS text), CAST(sqlc.narg(comment) AS text),
    sqlc.arg(workspace_id)
);

-- name: DeleteObjective :one
DELETE FROM public.objectives AS objective
USING public.workspace_members AS membership, public.team_members AS team_membership, public.users AS actor
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.workspace_id = objective.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role IN ('member', 'admin')
  AND team_membership.team_id = objective.team_id
  AND team_membership.user_id = membership.user_id
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
RETURNING objective.objective_id;
