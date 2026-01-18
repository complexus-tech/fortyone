-- 000018_story_activities.up.sql
CREATE TABLE "public"."story_activities" (
    "activity_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "story_id" uuid NOT NULL,
    "activity_type" varchar(50) NOT NULL,
    "field_changed" varchar(50) NOT NULL,
    "current_value" text NOT NULL,
    "user_id" uuid NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "workspace_id" uuid,
    "old_value" jsonb,
    "new_value" jsonb,
    CONSTRAINT "story_activities_story_id_fkey" FOREIGN KEY ("story_id") REFERENCES "public"."stories"("id") ON DELETE CASCADE,
    CONSTRAINT "story_activities_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "story_activities_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("activity_id")
);


-- Indices
CREATE INDEX idx_story_id ON public.story_activities USING btree (story_id);
CREATE INDEX idx_user_id ON public.story_activities USING btree (user_id);
CREATE INDEX idx_story_activities_story_history ON public.story_activities USING btree (story_id, created_at DESC);
CREATE INDEX idx_story_activities_burndown_events ON public.story_activities USING btree (workspace_id, created_at) WHERE ((field_changed)::text = ANY ((ARRAY['status_id'::character varying, 'sprint_id'::character varying])::text[]));
