-- name: CanReadObjectiveStrategy :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
	  AND membership.role IN ('member', 'admin')
) AS can_read;

-- name: GetWorkspaceStrategy :one
SELECT strategy.ultimate_goal, strategy.description
FROM public.workspace_strategies AS strategy
WHERE strategy.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
	  SELECT 1
	  FROM public.workspace_members AS membership
	  INNER JOIN public.users AS actor
		  ON actor.user_id = membership.user_id
		 AND actor.is_active = TRUE
	  WHERE membership.workspace_id = sqlc.arg(workspace_id)
		AND membership.user_id = sqlc.arg(actor_id)
		AND membership.role IN ('member', 'admin')
  );

-- name: ListStrategicPillars :many
SELECT pillar.pillar_id, pillar.name, pillar.description, pillar.order_index
FROM public.strategic_pillars AS pillar
WHERE pillar.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
	  SELECT 1
	  FROM public.workspace_members AS membership
	  INNER JOIN public.users AS actor
		  ON actor.user_id = membership.user_id
		 AND actor.is_active = TRUE
	  WHERE membership.workspace_id = sqlc.arg(workspace_id)
		AND membership.user_id = sqlc.arg(actor_id)
		AND membership.role IN ('member', 'admin')
  )
ORDER BY pillar.order_index, pillar.created_at, pillar.pillar_id;

-- name: ListVisibleStrategyAlignments :many
SELECT alignment.pillar_id, alignment.objective_id
FROM public.strategy_objective_alignments AS alignment
INNER JOIN public.objectives AS objective
    ON objective.objective_id = alignment.objective_id
   AND objective.workspace_id = alignment.workspace_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.workspace_members AS membership
	ON membership.workspace_id = objective.workspace_id
	AND membership.user_id = team_membership.user_id
INNER JOIN public.users AS actor
	ON actor.user_id = membership.user_id
	AND actor.is_active = TRUE
WHERE alignment.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
ORDER BY alignment.pillar_id, alignment.objective_id;

-- name: UpdateWorkspaceStrategy :one
INSERT INTO public.workspace_strategies (workspace_id, ultimate_goal, description)
SELECT sqlc.arg(workspace_id), sqlc.arg(ultimate_goal), CAST(sqlc.narg(description) AS text)
WHERE EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
      AND membership.role IN ('member', 'admin')
)
ON CONFLICT (workspace_id) DO UPDATE SET
    ultimate_goal = EXCLUDED.ultimate_goal,
    description = EXCLUDED.description,
    updated_at = GREATEST(clock_timestamp(), workspace_strategies.updated_at + INTERVAL '1 microsecond')
RETURNING workspace_id;

-- name: CreateStrategicPillar :one
INSERT INTO public.strategic_pillars (workspace_id, name, description, order_index)
SELECT sqlc.arg(workspace_id), sqlc.arg(name), CAST(sqlc.narg(description) AS text), sqlc.arg(order_index)
WHERE EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
      AND membership.role IN ('member', 'admin')
)
RETURNING pillar_id, name, description, order_index;

-- name: UpdateStrategicPillar :one
UPDATE public.strategic_pillars AS pillar
SET
    name = CASE WHEN CAST(sqlc.arg(set_name) AS boolean) THEN sqlc.arg(name) ELSE pillar.name END,
    description = CASE WHEN CAST(sqlc.arg(set_description) AS boolean) THEN CAST(sqlc.narg(description) AS text) ELSE pillar.description END,
    order_index = CASE WHEN CAST(sqlc.arg(set_order_index) AS boolean) THEN sqlc.arg(order_index) ELSE pillar.order_index END,
    updated_at = GREATEST(clock_timestamp(), pillar.updated_at + INTERVAL '1 microsecond')
WHERE pillar.workspace_id = sqlc.arg(workspace_id)
  AND pillar.pillar_id = sqlc.arg(pillar_id)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS actor
          ON actor.user_id = membership.user_id
         AND actor.is_active = TRUE
      WHERE membership.workspace_id = pillar.workspace_id
        AND membership.user_id = sqlc.arg(actor_id)
        AND membership.role IN ('member', 'admin')
  )
RETURNING pillar_id, name, description, order_index;

-- name: DeleteStrategicPillar :one
DELETE FROM public.strategic_pillars AS pillar
USING public.workspace_members AS membership, public.users AS actor
WHERE pillar.workspace_id = sqlc.arg(workspace_id)
  AND pillar.pillar_id = sqlc.arg(pillar_id)
  AND membership.workspace_id = pillar.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role IN ('member', 'admin')
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
RETURNING pillar.pillar_id;

-- name: DeleteObjectiveAlignment :one
WITH authorized_objective AS (
    SELECT objective.objective_id
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
    WHERE objective.workspace_id = sqlc.arg(workspace_id)
      AND objective.objective_id = sqlc.arg(objective_id)
      AND membership.role IN ('member', 'admin')
), deleted_alignment AS (
    DELETE FROM public.strategy_objective_alignments AS alignment
    USING authorized_objective
    WHERE alignment.workspace_id = sqlc.arg(workspace_id)
      AND alignment.objective_id = authorized_objective.objective_id
    RETURNING alignment.objective_id
)
SELECT objective_id FROM authorized_objective;

-- name: AlignObjective :one
INSERT INTO public.strategy_objective_alignments (workspace_id, objective_id, pillar_id)
SELECT sqlc.arg(workspace_id), objective.objective_id, pillar.pillar_id
FROM public.objectives AS objective
INNER JOIN public.strategic_pillars AS pillar
    ON pillar.pillar_id = sqlc.arg(pillar_id)
   AND pillar.workspace_id = objective.workspace_id
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
ON CONFLICT (objective_id) DO UPDATE SET
    pillar_id = EXCLUDED.pillar_id,
    workspace_id = EXCLUDED.workspace_id,
    updated_at = GREATEST(clock_timestamp(), strategy_objective_alignments.updated_at + INTERVAL '1 microsecond')
RETURNING objective_id;
