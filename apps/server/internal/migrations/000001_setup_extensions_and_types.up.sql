-- 000001_setup_extensions_and_types.up.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

CREATE TYPE public.user_role AS ENUM (
    'admin',
    'member',
    'guest',
    'system'
);

CREATE TYPE public.billing_interval_enum AS ENUM (
    'day',
    'week',
    'month',
    'year'
);

CREATE TYPE public.measurement_type AS ENUM (
    'percentage',
    'number',
    'boolean'
);

CREATE TYPE public.subscription_tier_enum AS ENUM (
    'free',
    'pro',
    'business',
    'enterprise'
);
