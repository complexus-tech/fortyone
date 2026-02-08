-- 000052_estimate_units.down.sql

DROP INDEX IF EXISTS public.idx_stories_workspace_team_estimate_unit;

ALTER TABLE IF EXISTS "public"."stories"
    DROP CONSTRAINT IF EXISTS "stories_estimate_unit_check";

ALTER TABLE IF EXISTS "public"."stories"
    DROP COLUMN IF EXISTS "estimate_unit";

DROP TABLE IF EXISTS "public"."team_estimation_settings";
