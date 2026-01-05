-- 000001_setup_extensions_and_types.up.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

CREATE TYPE public.billing_interval_enum AS ENUM (
    'day',
    'week',
    'month',
    'year'
);

CREATE TYPE public.entity_type AS ENUM (
    'story',
    'comment',
    'objective',
    'key_result'
);

CREATE TYPE public.measurement_type AS ENUM (
    'percentage',
    'number',
    'boolean'
);

CREATE TYPE public.notification_type AS ENUM (
    'story_update',
    'story_comment',
    'comment_reply',
    'objective_update',
    'key_result_update',
    'mention'
);

CREATE TYPE public.objective_health_status AS ENUM (
    'At Risk',
    'On Track',
    'Off Track'
);

CREATE TYPE public.okr_activity_type AS ENUM (
    'create',
    'update',
    'delete'
);

CREATE TYPE public.okr_update_type AS ENUM (
    'objective',
    'key_result'
);

CREATE TYPE public.subscription_tier_enum AS ENUM (
    'free',
    'pro',
    'business',
    'enterprise'
);

CREATE TYPE public.token_type AS ENUM (
    'login',
    'registration'
);

CREATE TYPE public.user_role AS ENUM (
    'admin',
    'member',
    'guest',
    'system'
);
