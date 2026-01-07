-- 000044_team_story_sequences.up.sql
CREATE TABLE "public"."team_story_sequences" (
    "team_id" uuid NOT NULL,
    "current_sequence" int4 NOT NULL DEFAULT 0,
    "workspace_id" uuid NOT NULL,
    CONSTRAINT "team_story_sequences_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "team_story_sequences_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    PRIMARY KEY ("workspace_id","team_id")
);


-- Indices
CREATE INDEX idx_team_story_sequences_team_id ON public.team_story_sequences USING btree (team_id);
CREATE INDEX idx_team_story_sequences_workspace_team ON public.team_story_sequences USING btree (workspace_id, team_id);
