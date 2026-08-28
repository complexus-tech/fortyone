-- name: FindSimilarStories :many
WITH input AS (
    SELECT COALESCE(string_agg(token, ' ' ORDER BY token_order), '') AS normalized_title
    FROM unnest(
        regexp_split_to_array(lower(CAST(sqlc.arg(title) AS text)), '[^a-z0-9]+')
    ) WITH ORDINALITY AS tokens(token, token_order)
    WHERE token <> ''
      AND token NOT IN (
          'a', 'add', 'an', 'build', 'create', 'fix', 'for', 'implement',
          'make', 'new', 'please', 'story', 'support', 'task', 'the', 'to', 'update'
      )
), ranked AS (
    SELECT
        story.id,
        story.sequence_id,
        story.title,
        story.team_id,
        story.status_id,
        story.assignee_id,
        story.priority,
        story.updated_at,
        CAST(GREATEST(
            similarity(normalized.normalized_title, input.normalized_title),
            CASE
                WHEN normalized.normalized_title = input.normalized_title THEN 1.0
                ELSE 0.0
            END
        ) AS double precision) AS confidence
    FROM stories AS story
    INNER JOIN team_members AS team_member
        ON team_member.team_id = story.team_id
       AND team_member.user_id = sqlc.arg(actor_id)
    INNER JOIN workspace_members AS workspace_member
        ON workspace_member.workspace_id = story.workspace_id
       AND workspace_member.user_id = sqlc.arg(actor_id)
    INNER JOIN users AS actor
        ON actor.user_id = workspace_member.user_id
       AND actor.is_active = TRUE
       AND actor.is_system = FALSE
    CROSS JOIN input
    CROSS JOIN LATERAL (
        SELECT COALESCE(string_agg(token, ' ' ORDER BY token_order), '') AS normalized_title
        FROM unnest(
            regexp_split_to_array(lower(story.title), '[^a-z0-9]+')
        ) WITH ORDINALITY AS tokens(token, token_order)
        WHERE token <> ''
          AND token NOT IN (
              'a', 'add', 'an', 'build', 'create', 'fix', 'for', 'implement',
              'make', 'new', 'please', 'story', 'support', 'task', 'the', 'to', 'update'
          )
    ) AS normalized
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND input.normalized_title <> ''
      AND normalized.normalized_title <> ''
      AND (
          CAST(sqlc.narg(team_id) AS uuid) IS NULL
          OR story.team_id = CAST(sqlc.narg(team_id) AS uuid)
      )
)
SELECT
    ranked.id,
    ranked.sequence_id,
    ranked.title,
    ranked.team_id,
    ranked.status_id,
    ranked.assignee_id,
    ranked.priority,
    ranked.confidence
FROM ranked
WHERE ranked.confidence >= 0.45
ORDER BY ranked.confidence DESC, ranked.updated_at DESC, ranked.title, ranked.id
LIMIT sqlc.arg(result_limit);

-- name: SearchStories :many
SELECT
    story.id,
    story.sequence_id,
    story.title,
    story.parent_id,
    story.objective_id,
    story.status_id,
    search_status.name AS status_name,
    search_status.color AS status_color,
    search_status.category AS status_category,
    story.assignee_id,
    search_assignee.full_name AS assignee_full_name,
    search_assignee.username AS assignee_username,
    story.reporter_id,
    story.priority,
    story.estimate_unit,
    COALESCE(estimation.scheme, 'tshirt') AS estimate_scheme,
    story.sprint_id,
    story.key_result_id,
    story.team_id,
    search_team.name AS team_name,
    search_team.code AS team_code,
    story.workspace_id,
    story.start_date,
    story.end_date,
    story.estimated_duration_minutes,
    story.minimum_focus_block_minutes,
    story.auto_scheduling_enabled,
    story.auto_scheduling_locked,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_updated_at,
    story.created_at,
    story.updated_at,
    CAST(COALESCE(
        ARRAY(
            SELECT story_label.label_id
            FROM story_labels AS story_label
            WHERE story_label.story_id = story.id
            ORDER BY story_label.label_id
        ),
        CAST(ARRAY[] AS uuid[])
    ) AS uuid[]) AS label_ids,
    COUNT(*) OVER () AS total_count
FROM stories AS story
INNER JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = sqlc.arg(actor_id)
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = sqlc.arg(actor_id)
INNER JOIN users AS actor
    ON actor.user_id = workspace_member.user_id
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
INNER JOIN teams AS search_team
    ON search_team.team_id = story.team_id
   AND search_team.workspace_id = story.workspace_id
LEFT JOIN statuses AS search_status
    ON search_status.status_id = story.status_id
   AND search_status.team_id = story.team_id
   AND search_status.workspace_id = story.workspace_id
LEFT JOIN users AS search_assignee
    ON search_assignee.user_id = story.assignee_id
   AND search_assignee.is_active = TRUE
   AND search_assignee.is_system = FALSE
   AND EXISTS (
       SELECT 1
       FROM workspace_members AS search_assignee_membership
       WHERE search_assignee_membership.workspace_id = story.workspace_id
         AND search_assignee_membership.user_id = search_assignee.user_id
   )
LEFT JOIN team_estimation_settings AS estimation
    ON estimation.team_id = story.team_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(query_text) AS text) = ''
      OR story.search_vector @@ plainto_tsquery('english', CAST(sqlc.arg(query_text) AS text))
      OR story.title ILIKE '%' || CAST(sqlc.arg(query_text) AS text) || '%'
      OR similarity(lower(story.title), lower(CAST(sqlc.arg(query_text) AS text))) >= 0.2
      OR LEAST(
          word_similarity(lower(story.title), lower(CAST(sqlc.arg(query_text) AS text))),
          word_similarity(lower(CAST(sqlc.arg(query_text) AS text)), lower(story.title))
      ) >= 0.25
  )
  AND (
      CAST(sqlc.narg(team_id) AS uuid) IS NULL
      OR story.team_id = CAST(sqlc.narg(team_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(assignee_id) AS uuid) IS NULL
      OR story.assignee_id = CAST(sqlc.narg(assignee_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(status_id) AS uuid) IS NULL
      OR story.status_id = CAST(sqlc.narg(status_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(priority) AS text) IS NULL
      OR story.priority = CAST(sqlc.narg(priority) AS text)
  )
  AND (
      CAST(sqlc.narg(label_id) AS uuid) IS NULL
      OR EXISTS (
          SELECT 1
          FROM story_labels AS filtered_label
          WHERE filtered_label.story_id = story.id
            AND filtered_label.label_id = CAST(sqlc.narg(label_id) AS uuid)
      )
  )
ORDER BY
    CASE WHEN CAST(sqlc.arg(sort_by) AS text) = 'updated' THEN story.updated_at END DESC,
    CASE WHEN CAST(sqlc.arg(sort_by) AS text) = 'created' THEN story.created_at END DESC,
    CASE
        WHEN CAST(sqlc.arg(sort_by) AS text) = 'relevance'
         AND CAST(sqlc.arg(query_text) AS text) <> ''
        THEN GREATEST(
            ts_rank(story.search_vector, plainto_tsquery('english', CAST(sqlc.arg(query_text) AS text))),
            similarity(lower(story.title), lower(CAST(sqlc.arg(query_text) AS text))),
            LEAST(
                word_similarity(lower(story.title), lower(CAST(sqlc.arg(query_text) AS text))),
                word_similarity(lower(CAST(sqlc.arg(query_text) AS text)), lower(story.title))
            )
        )
    END DESC,
    story.created_at DESC,
    story.id
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSearchStories :one
SELECT COUNT(*)
FROM stories AS story
INNER JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = sqlc.arg(actor_id)
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = sqlc.arg(actor_id)
INNER JOIN users AS actor
    ON actor.user_id = workspace_member.user_id
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(query_text) AS text) = ''
      OR story.search_vector @@ plainto_tsquery('english', CAST(sqlc.arg(query_text) AS text))
      OR story.title ILIKE '%' || CAST(sqlc.arg(query_text) AS text) || '%'
      OR similarity(lower(story.title), lower(CAST(sqlc.arg(query_text) AS text))) >= 0.2
      OR LEAST(
          word_similarity(lower(story.title), lower(CAST(sqlc.arg(query_text) AS text))),
          word_similarity(lower(CAST(sqlc.arg(query_text) AS text)), lower(story.title))
      ) >= 0.25
  )
  AND (
      CAST(sqlc.narg(team_id) AS uuid) IS NULL
      OR story.team_id = CAST(sqlc.narg(team_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(assignee_id) AS uuid) IS NULL
      OR story.assignee_id = CAST(sqlc.narg(assignee_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(status_id) AS uuid) IS NULL
      OR story.status_id = CAST(sqlc.narg(status_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(priority) AS text) IS NULL
      OR story.priority = CAST(sqlc.narg(priority) AS text)
  )
  AND (
      CAST(sqlc.narg(label_id) AS uuid) IS NULL
      OR EXISTS (
          SELECT 1
          FROM story_labels AS filtered_label
          WHERE filtered_label.story_id = story.id
            AND filtered_label.label_id = CAST(sqlc.narg(label_id) AS uuid)
      )
  );

-- name: SearchObjectives :many
SELECT
    objective.objective_id,
    objective.name,
    objective.description,
    objective.short_summary,
    objective.lead_user_id,
    search_lead.full_name AS lead_full_name,
    search_lead.username AS lead_username,
    objective.team_id,
    search_team.name AS team_name,
    search_team.code AS team_code,
    objective.workspace_id,
    objective.start_date,
    objective.end_date,
    objective.status_id,
    objective.priority,
    objective.health,
    objective.created_at,
    objective.updated_at,
    COUNT(*) OVER () AS total_count
FROM objectives AS objective
INNER JOIN team_members AS team_member
    ON team_member.team_id = objective.team_id
   AND team_member.user_id = sqlc.arg(actor_id)
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = objective.workspace_id
   AND workspace_member.user_id = sqlc.arg(actor_id)
INNER JOIN users AS actor
    ON actor.user_id = workspace_member.user_id
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
INNER JOIN teams AS search_team
    ON search_team.team_id = objective.team_id
   AND search_team.workspace_id = objective.workspace_id
LEFT JOIN users AS search_lead
    ON search_lead.user_id = objective.lead_user_id
   AND search_lead.is_active = TRUE
   AND search_lead.is_system = FALSE
   AND EXISTS (
       SELECT 1
       FROM workspace_members AS search_lead_membership
       WHERE search_lead_membership.workspace_id = objective.workspace_id
         AND search_lead_membership.user_id = search_lead.user_id
   )
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (
      CAST(sqlc.arg(query_text) AS text) = ''
      OR objective.search_vector @@ plainto_tsquery('english', CAST(sqlc.arg(query_text) AS text))
      OR objective.name ILIKE '%' || CAST(sqlc.arg(query_text) AS text) || '%'
  )
  AND (
      CAST(sqlc.narg(team_id) AS uuid) IS NULL
      OR objective.team_id = CAST(sqlc.narg(team_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(status_id) AS uuid) IS NULL
      OR objective.status_id = CAST(sqlc.narg(status_id) AS uuid)
  )
ORDER BY
    CASE WHEN CAST(sqlc.arg(sort_by) AS text) = 'updated' THEN objective.updated_at END DESC,
    CASE WHEN CAST(sqlc.arg(sort_by) AS text) = 'created' THEN objective.created_at END DESC,
    CASE
        WHEN CAST(sqlc.arg(sort_by) AS text) = 'relevance'
         AND CAST(sqlc.arg(query_text) AS text) <> ''
        THEN ts_rank(objective.search_vector, plainto_tsquery('english', CAST(sqlc.arg(query_text) AS text)))
    END DESC,
    objective.created_at DESC,
    objective.objective_id
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSearchObjectives :one
SELECT COUNT(*)
FROM objectives AS objective
INNER JOIN team_members AS team_member
    ON team_member.team_id = objective.team_id
   AND team_member.user_id = sqlc.arg(actor_id)
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = objective.workspace_id
   AND workspace_member.user_id = sqlc.arg(actor_id)
INNER JOIN users AS actor
    ON actor.user_id = workspace_member.user_id
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (
      CAST(sqlc.arg(query_text) AS text) = ''
      OR objective.search_vector @@ plainto_tsquery('english', CAST(sqlc.arg(query_text) AS text))
      OR objective.name ILIKE '%' || CAST(sqlc.arg(query_text) AS text) || '%'
  )
  AND (
      CAST(sqlc.narg(team_id) AS uuid) IS NULL
      OR objective.team_id = CAST(sqlc.narg(team_id) AS uuid)
  )
  AND (
      CAST(sqlc.narg(status_id) AS uuid) IS NULL
      OR objective.status_id = CAST(sqlc.narg(status_id) AS uuid)
  );
