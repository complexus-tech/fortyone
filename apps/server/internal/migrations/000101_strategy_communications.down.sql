UPDATE public.notification_preferences
SET preferences = preferences - 'strategy_update';

-- PostgreSQL enum values are intentionally retained. Removing an enum value
-- safely requires rebuilding the type and every dependent column.
