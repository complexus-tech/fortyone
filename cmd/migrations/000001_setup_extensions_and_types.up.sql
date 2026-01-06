-- 000001_setup_extensions_and_types.up.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

CREATE TYPE public.user_role AS ENUM (
    'admin',
    'member',
    'guest',
    'system'
);

CREATE TYPE public.billing_interval_enum AS ENUM (
