-- name: GetOrCreateAutomationPreferencesForMember :one
INSERT INTO public.user_automation_preferences (
    user_id,
    workspace_id,
    auto_assign_self,
    auto_scheduling,
    assign_self_on_branch_copy,
    move_story_to_started_on_branch,
    open_story_in_dialog,
    created_at,
    updated_at
)
SELECT
    membership.user_id,
    membership.workspace_id,
    FALSE,
    TRUE,
    FALSE,
    FALSE,
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.workspace_members AS membership
INNER JOIN public.users AS account ON account.user_id = membership.user_id
WHERE membership.user_id = sqlc.arg(user_id)
  AND membership.workspace_id = sqlc.arg(workspace_id)
  AND account.is_active = TRUE
ON CONFLICT (user_id, workspace_id) DO UPDATE
SET user_id = EXCLUDED.user_id
RETURNING
    user_id,
    workspace_id,
    COALESCE(auto_assign_self, FALSE)::boolean AS auto_assign_self,
    COALESCE(auto_scheduling, TRUE)::boolean AS auto_scheduling,
    COALESCE(assign_self_on_branch_copy, FALSE)::boolean AS assign_self_on_branch_copy,
    COALESCE(move_story_to_started_on_branch, FALSE)::boolean AS move_story_to_started_on_branch,
    COALESCE(open_story_in_dialog, TRUE)::boolean AS open_story_in_dialog,
    COALESCE(created_at, CURRENT_TIMESTAMP)::timestamptz AS created_at,
    COALESCE(updated_at, CURRENT_TIMESTAMP)::timestamptz AS updated_at;

-- name: UpsertAutomationPreferencesForMember :execrows
INSERT INTO public.user_automation_preferences (
    user_id,
    workspace_id,
    auto_assign_self,
    auto_scheduling,
    assign_self_on_branch_copy,
    move_story_to_started_on_branch,
    open_story_in_dialog,
    created_at,
    updated_at
)
SELECT
    membership.user_id,
    membership.workspace_id,
    CASE
        WHEN CAST(sqlc.arg(set_auto_assign_self) AS boolean)
            THEN CAST(sqlc.arg(auto_assign_self) AS boolean)
        ELSE FALSE
    END,
    CASE
        WHEN CAST(sqlc.arg(set_auto_scheduling) AS boolean)
            THEN CAST(sqlc.arg(auto_scheduling) AS boolean)
        ELSE TRUE
    END,
    CASE
        WHEN CAST(sqlc.arg(set_assign_self_on_branch_copy) AS boolean)
            THEN CAST(sqlc.arg(assign_self_on_branch_copy) AS boolean)
        ELSE FALSE
    END,
    CASE
        WHEN CAST(sqlc.arg(set_move_story_to_started_on_branch) AS boolean)
            THEN CAST(sqlc.arg(move_story_to_started_on_branch) AS boolean)
        ELSE FALSE
    END,
    CASE
        WHEN CAST(sqlc.arg(set_open_story_in_dialog) AS boolean)
            THEN CAST(sqlc.arg(open_story_in_dialog) AS boolean)
        ELSE TRUE
    END,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.workspace_members AS membership
INNER JOIN public.users AS account ON account.user_id = membership.user_id
WHERE membership.user_id = sqlc.arg(user_id)
  AND membership.workspace_id = sqlc.arg(workspace_id)
  AND account.is_active = TRUE
ON CONFLICT (user_id, workspace_id) DO UPDATE
SET
    auto_assign_self = CASE
        WHEN CAST(sqlc.arg(set_auto_assign_self) AS boolean)
            THEN EXCLUDED.auto_assign_self
        ELSE user_automation_preferences.auto_assign_self
    END,
    auto_scheduling = CASE
        WHEN CAST(sqlc.arg(set_auto_scheduling) AS boolean)
            THEN EXCLUDED.auto_scheduling
        ELSE user_automation_preferences.auto_scheduling
    END,
    assign_self_on_branch_copy = CASE
        WHEN CAST(sqlc.arg(set_assign_self_on_branch_copy) AS boolean)
            THEN EXCLUDED.assign_self_on_branch_copy
        ELSE user_automation_preferences.assign_self_on_branch_copy
    END,
    move_story_to_started_on_branch = CASE
        WHEN CAST(sqlc.arg(set_move_story_to_started_on_branch) AS boolean)
            THEN EXCLUDED.move_story_to_started_on_branch
        ELSE user_automation_preferences.move_story_to_started_on_branch
    END,
    open_story_in_dialog = CASE
        WHEN CAST(sqlc.arg(set_open_story_in_dialog) AS boolean)
            THEN EXCLUDED.open_story_in_dialog
        ELSE user_automation_preferences.open_story_in_dialog
    END,
    updated_at = CURRENT_TIMESTAMP;
