-- 000010_team_members.up.sql
CREATE TABLE "public"."team_members" (
    "team_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "team_members_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    CONSTRAINT "team_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("team_id","user_id")
);


-- Indices
CREATE INDEX idx_team_members_team_user ON public.team_members USING btree (team_id, user_id);
CREATE INDEX idx_team_members_team_user_id ON public.team_members USING btree (team_id, user_id);
CREATE INDEX idx_team_members_user ON public.team_members USING btree (user_id);
CREATE INDEX idx_team_members_user_id ON public.team_members USING btree (user_id);
