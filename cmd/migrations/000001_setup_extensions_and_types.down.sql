-- 000001_setup_extensions_and_types.down.sql
DROP TYPE IF EXISTS public.user_role;
DROP TYPE IF EXISTS public.token_type;
DROP TYPE IF EXISTS public.subscription_tier_enum;
DROP TYPE IF EXISTS public.okr_update_type;
DROP TYPE IF EXISTS public.okr_activity_type;
DROP TYPE IF EXISTS public.objective_health_status;
DROP TYPE IF EXISTS public.notification_type;
DROP TYPE IF EXISTS public.measurement_type;
DROP TYPE IF EXISTS public.entity_type;
DROP TYPE IF EXISTS public.billing_interval_enum;
DROP EXTENSION IF EXISTS pg_trgm;
