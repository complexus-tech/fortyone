-- ListKeyResultNotificationAudience owns the key-result/objective lookup and
-- recipient selection used by the event consumer. Both the event actor and
-- every recipient must still have live access to the exact workspace team.
-- name: ListKeyResultNotificationAudience :many
WITH target AS (
    SELECT
        key_result.id AS key_result_id,
        key_result.name AS key_result_name,
        key_result.lead AS key_result_lead,
        objective.objective_id,
        objective.name AS objective_name,
        objective.lead_user_id AS objective_lead,
        objective.team_id,
        objective.workspace_id
    FROM public.key_results AS key_result
    INNER JOIN public.objectives AS objective
        ON objective.objective_id = key_result.objective_id
       AND objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = objective.workspace_id
       AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
       AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS actor_membership
        ON actor_membership.workspace_id = objective.workspace_id
       AND actor_membership.user_id = actor.user_id
       AND actor_membership.role IN ('admin', 'member', 'guest')
    WHERE key_result.id = CAST(sqlc.arg(key_result_id) AS uuid)
      AND (
          actor_membership.role = 'admin'
          OR EXISTS (
              SELECT 1
              FROM public.team_members AS actor_team_membership
              WHERE actor_team_membership.team_id = objective.team_id
                AND actor_team_membership.user_id = actor.user_id
          )
      )
), audience AS (
    SELECT target.key_result_lead AS recipient_id FROM target
    UNION
    SELECT target.objective_lead AS recipient_id FROM target
    UNION
    SELECT contributor.user_id AS recipient_id
    FROM public.key_result_contributors AS contributor
    INNER JOIN target ON target.key_result_id = contributor.key_result_id
)
SELECT
    recipient.user_id AS recipient_id,
    target.key_result_id,
    target.objective_id,
    target.key_result_name,
    target.objective_name
FROM target
INNER JOIN audience
    ON audience.recipient_id IS NOT NULL
   AND audience.recipient_id <> CAST(sqlc.arg(actor_id) AS uuid)
INNER JOIN public.users AS recipient
    ON recipient.user_id = audience.recipient_id
   AND recipient.is_active = TRUE
INNER JOIN public.workspace_members AS recipient_membership
    ON recipient_membership.workspace_id = target.workspace_id
   AND recipient_membership.user_id = recipient.user_id
   AND recipient_membership.role IN ('admin', 'member', 'guest')
WHERE recipient_membership.role = 'admin'
   OR EXISTS (
       SELECT 1
       FROM public.team_members AS recipient_team_membership
       WHERE recipient_team_membership.team_id = target.team_id
         AND recipient_team_membership.user_id = recipient.user_id
   )
ORDER BY audience.recipient_id;
