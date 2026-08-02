ALTER TYPE public.notification_type ADD VALUE IF NOT EXISTS 'strategy_update';
ALTER TYPE public.entity_type ADD VALUE IF NOT EXISTS 'strategy';

UPDATE public.notification_preferences
SET preferences = jsonb_set(
    preferences,
    '{strategy_update}',
    '{"email": true, "in_app": true}'::jsonb,
    true
)
WHERE NOT preferences ? 'strategy_update';
