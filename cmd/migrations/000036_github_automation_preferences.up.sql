-- 000036_github_automation_preferences.up.sql
CREATE TABLE "public"."github_automation_preferences" (
    "user_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "auto_create_branch" bool DEFAULT false,
    "auto_create_pr" bool DEFAULT false,
    "auto_move_story_on_pr_merge" bool DEFAULT true,
    "auto_assign_pr_reviewer" bool DEFAULT false,
    "branch_naming_pattern" text DEFAULT 'story/{story-id}-{title}'::text,
    "pr_template" text,
    "default_reviewer_ids" _uuid DEFAULT '{}'::uuid[],
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "github_automation_preferences_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "github_automation_preferences_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("user_id","workspace_id")
);
