-- 000017_stories.up.sql
CREATE TABLE "public"."stories" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "sequence_id" int4,
    "team_id" uuid NOT NULL,
    "title" varchar(255) NOT NULL,
    "description" text,
    "description_html" text,
    "parent_id" uuid,
    "objective_id" uuid,
    "status_id" uuid,
    "assignee_id" uuid,
    "blocked_by_id" uuid,
    "blocking_id" uuid,
    "related_id" uuid,
    "reporter_id" uuid,
    "priority" varchar(100),
    "sprint_id" uuid,
    "workspace_id" uuid NOT NULL,
    "start_date" date,
    "end_date" date,
    "estimate" float8,
    "archived_at" timestamptz,
    "is_draft" bool NOT NULL DEFAULT false,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamptz,
    "search_vector" tsvector,
    "key_result_id" uuid,
    "completed_at" timestamptz,
    CONSTRAINT "stories_sprint_id_fkey" FOREIGN KEY ("sprint_id") REFERENCES "public"."sprints"("sprint_id") ON DELETE SET NULL,
    CONSTRAINT "stories_key_result_id_fkey" FOREIGN KEY ("key_result_id") REFERENCES "public"."key_results"("id") ON DELETE SET NULL,
    CONSTRAINT "stories_reporter_id_fkey" FOREIGN KEY ("reporter_id") REFERENCES "public"."users"("user_id") ON DELETE SET NULL,
    CONSTRAINT "stories_objective_id_fkey" FOREIGN KEY ("objective_id") REFERENCES "public"."objectives"("objective_id") ON DELETE SET NULL,
    CONSTRAINT "stories_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "stories_assignee_id_fkey" FOREIGN KEY ("assignee_id") REFERENCES "public"."users"("user_id") ON DELETE SET NULL,
    CONSTRAINT "stories_blocked_by_id_fkey" FOREIGN KEY ("blocked_by_id") REFERENCES "public"."stories"("id") ON DELETE SET NULL,
    CONSTRAINT "stories_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    CONSTRAINT "stories_status_id_fkey" FOREIGN KEY ("status_id") REFERENCES "public"."statuses"("status_id"),
    CONSTRAINT "stories_blocking_id_fkey" FOREIGN KEY ("blocking_id") REFERENCES "public"."stories"("id") ON DELETE SET NULL,
    CONSTRAINT "stories_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."stories"("id") ON DELETE SET NULL,
    CONSTRAINT "stories_related_id_fkey" FOREIGN KEY ("related_id") REFERENCES "public"."stories"("id") ON DELETE SET NULL,
    PRIMARY KEY ("id")
);


-- Indices
CREATE UNIQUE INDEX unique_team_sequence ON public.stories USING btree (team_id, sequence_id);
CREATE INDEX idx_stories_created ON public.stories USING btree (created_at DESC);
CREATE INDEX idx_stories_objective_id ON public.stories USING btree (objective_id);
CREATE INDEX idx_stories_search ON public.stories USING gin (search_vector);
CREATE INDEX idx_stories_search_workspace_team ON public.stories USING btree (workspace_id, team_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_stories_sprint_id ON public.stories USING btree (sprint_id);
CREATE INDEX idx_stories_team_id ON public.stories USING btree (team_id);
CREATE INDEX idx_stories_title_trigram ON public.stories USING gin (title gin_trgm_ops);
CREATE INDEX idx_stories_workspace_deleted_parent_assignee_reporter ON public.stories USING btree (workspace_id, deleted_at, parent_id, assignee_id, reporter_id);
CREATE INDEX idx_stories_workspace_id ON public.stories USING btree (workspace_id);
CREATE INDEX idx_stories_workspace_team ON public.stories USING btree (workspace_id, team_id);
CREATE INDEX idx_stories_workspace_team_deleted_parent ON public.stories USING btree (workspace_id, team_id, deleted_at, parent_id);
CREATE INDEX idx_stories_assignee_sprint ON public.stories USING btree (assignee_id, sprint_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_stories_status_id ON public.stories USING btree (status_id);
CREATE INDEX idx_stories_updated_completed ON public.stories USING btree (updated_at, sprint_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_stories_key_result_id ON public.stories USING btree (key_result_id);
