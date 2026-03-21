ALTER TABLE public.github_workspace_settings
    DROP COLUMN IF EXISTS sync_assignees,
    DROP COLUMN IF EXISTS sync_labels,
    DROP COLUMN IF EXISTS auto_populate_pr_body,
    DROP COLUMN IF EXISTS close_on_commit_keywords;

ALTER TABLE public.github_story_links
    DROP COLUMN IF EXISTS review_state,
    DROP COLUMN IF EXISTS reviews_approved,
    DROP COLUMN IF EXISTS reviews_changes_requested,
    DROP COLUMN IF EXISTS check_state;

DROP INDEX IF EXISTS users_github_user_id_key;

ALTER TABLE public.users
    DROP COLUMN IF EXISTS github_user_id,
    DROP COLUMN IF EXISTS github_username;
