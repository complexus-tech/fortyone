-- name: ListMayaWorkFocusCandidates :many
SELECT
    team.workspace_id,
    member.team_id,
    member.user_id,
    member.ai_role_title,
    member.ai_role_description
FROM team_members member
INNER JOIN teams team
    ON team.team_id = member.team_id
INNER JOIN workspaces workspace
    ON workspace.workspace_id = team.workspace_id
INNER JOIN users actor
    ON actor.user_id = member.user_id
WHERE actor.is_active = TRUE
    AND actor.is_system = FALSE
    AND member.ai_role_title = ''
    AND member.ai_role_description = ''
    AND workspace.deleted_at IS NULL
    AND (
        workspace.trial_ends_on > CURRENT_TIMESTAMP
        OR EXISTS (
            SELECT 1
            FROM workspace_subscriptions subscription
            WHERE subscription.workspace_id = workspace.workspace_id
                AND subscription.subscription_tier <> 'free'
                AND subscription.subscription_status IN ('active', 'trialing', 'past_due')
        )
    )
ORDER BY member.inferred_ai_role_generated_at ASC NULLS FIRST, member.updated_at ASC
LIMIT CAST(sqlc.arg(row_limit) AS integer);

-- name: ListMayaWorkFocusEvidence :many
WITH evidence_stories AS (
    SELECT
        story.id,
        story.title,
        story.description,
        story.updated_at
    FROM stories story
    WHERE story.workspace_id = sqlc.arg(workspace_id)
        AND story.team_id = sqlc.arg(team_id)
        AND story.assignee_id = sqlc.arg(user_id)
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND story.is_draft = FALSE
        AND story.updated_at >= sqlc.arg(updated_after)
    ORDER BY story.updated_at DESC, story.id
    LIMIT CAST(sqlc.arg(row_limit) AS integer)
)
SELECT
    story.id AS story_id,
    story.title,
    story.description,
    label.name AS label
FROM evidence_stories story
LEFT JOIN story_labels story_label
    ON story_label.story_id = story.id
LEFT JOIN labels label
    ON label.label_id = story_label.label_id
ORDER BY story.updated_at DESC, story.id, label.name;

-- name: SaveMayaInferredWorkFocus :execrows
UPDATE team_members member
SET inferred_ai_role_title = sqlc.arg(role_title),
    inferred_ai_role_description = sqlc.arg(role_description),
    inferred_ai_role_story_count = sqlc.arg(story_count),
    inferred_ai_role_confidence = sqlc.arg(confidence),
    inferred_ai_role_generated_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
FROM teams team
WHERE member.team_id = team.team_id
    AND member.team_id = sqlc.arg(team_id)
    AND member.user_id = sqlc.arg(user_id)
    AND team.workspace_id = sqlc.arg(workspace_id)
    AND member.ai_role_title = ''
    AND member.ai_role_description = '';
