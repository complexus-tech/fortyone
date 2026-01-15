-- 000050_insert_system_user.up.sql
INSERT INTO public.users (
    user_id,
    username,
    email,
    full_name,
    is_active,
    is_system,
    timezone
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'maya',
    'maya@fortyone.app',
    'Maya',
    true,
    true,
    'UTC'
) ON CONFLICT (user_id) DO NOTHING;
