-- 000054_insert_github_user.up.sql
INSERT INTO public.users (
    user_id,
    username,
    email,
    full_name,
    is_active,
    is_system,
    timezone
) VALUES (
    '00000000-0000-0000-0000-000000000002',
    'github',
    'github@fortyone.app',
    'GitHub',
    true,
    true,
    'UTC'
)
ON CONFLICT (email) DO UPDATE
SET
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    is_active = EXCLUDED.is_active,
    is_system = EXCLUDED.is_system,
    timezone = EXCLUDED.timezone;
