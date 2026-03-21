-- GitHub Integration Enhancements
-- Adds: user linking, PR review state, check status, workspace sync settings

-- 1. GitHub user linking on users table
ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS github_user_id bigint,
    ADD COLUMN IF NOT EXISTS github_username text;

CREATE UNIQUE INDEX IF NOT EXISTS users_github_user_id_key
    ON public.users USING btree (github_user_id) WHERE github_user_id IS NOT NULL;

-- 2. PR review and check state on story links
ALTER TABLE public.github_story_links
    ADD COLUMN IF NOT EXISTS review_state text,
    ADD COLUMN IF NOT EXISTS reviews_approved integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reviews_changes_requested integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS check_state text;

-- 3. Workspace-level sync toggles
ALTER TABLE public.github_workspace_settings
    ADD COLUMN IF NOT EXISTS sync_assignees boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS sync_labels boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS auto_populate_pr_body boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS close_on_commit_keywords boolean NOT NULL DEFAULT true;
