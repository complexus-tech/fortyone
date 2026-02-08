-- 000052_estimate_units.up.sql

CREATE TABLE "public"."team_estimation_settings" (
    "team_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "scheme" varchar(32) NOT NULL DEFAULT 'points',
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "team_estimation_settings_team_id_fkey"
        FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    CONSTRAINT "team_estimation_settings_workspace_id_fkey"
        FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "team_estimation_settings_scheme_check"
        CHECK ("scheme" IN ('points', 'hours', 'tshirt', 'ideal_days')),
    PRIMARY KEY ("team_id")
);

ALTER TABLE "public"."stories"
    ADD COLUMN "estimate_unit" int2;

ALTER TABLE "public"."stories"
    ADD CONSTRAINT "stories_estimate_unit_check"
        CHECK ("estimate_unit" IS NULL OR "estimate_unit" IN (1, 2, 3, 5, 8));

CREATE INDEX idx_stories_workspace_team_estimate_unit
    ON public.stories USING btree (workspace_id, team_id, estimate_unit)
    WHERE (deleted_at IS NULL);
