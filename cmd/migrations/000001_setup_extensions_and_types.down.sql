-- 000001_setup_extensions_and_types.down.sql
DROP TYPE IF EXISTS public.subscription_tier_enum CASCADE;
DROP TYPE IF EXISTS public.measurement_type CASCADE;
DROP TYPE IF EXISTS public.billing_interval_enum CASCADE;
DROP TYPE IF EXISTS public.user_role CASCADE;

-- Also cleanup other types that might be lingering from table-specific migrations
DROP TYPE IF EXISTS public.token_type CASCADE;
DROP TYPE IF EXISTS public.okr_update_type CASCADE;
DROP TYPE IF EXISTS public.okr_activity_type CASCADE;
DROP TYPE IF EXISTS public.objective_health_status CASCADE;
DROP TYPE IF EXISTS public.notification_type CASCADE;
DROP TYPE IF EXISTS public.entity_type CASCADE;

DROP EXTENSION IF EXISTS pg_trgm CASCADE;
