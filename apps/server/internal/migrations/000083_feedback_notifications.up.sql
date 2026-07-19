ALTER TYPE public.notification_type
    ADD VALUE IF NOT EXISTS 'feedback_comment';

ALTER TYPE public.notification_type
    ADD VALUE IF NOT EXISTS 'feedback_status_update';

ALTER TYPE public.entity_type
    ADD VALUE IF NOT EXISTS 'feedback';

ALTER TABLE public.notifications
    ADD COLUMN IF NOT EXISTS dedupe_key text;

UPDATE public.notifications
SET dedupe_key = 'legacy:' || notification_id::text
WHERE dedupe_key IS NULL;

ALTER TABLE public.notifications
    ALTER COLUMN dedupe_key SET DEFAULT ('notification:' || gen_random_uuid()::text),
    ALTER COLUMN dedupe_key SET NOT NULL;

-- Some deployed databases may contain the broad unique index implied by the
-- former repository ON CONFLICT clause even though it never existed in the
-- managed migrations. Remove only a unique index whose ordered columns match
-- that exact legacy shape, so separate events for one entity cannot collide.
DO $$
DECLARE
    legacy_index record;
BEGIN
    FOR legacy_index IN
        SELECT
            index_class.relname AS index_name,
            constraint_row.conname AS constraint_name
        FROM pg_index index_row
        INNER JOIN pg_class table_class ON table_class.oid = index_row.indrelid
        INNER JOIN pg_namespace table_namespace ON table_namespace.oid = table_class.relnamespace
        INNER JOIN pg_class index_class ON index_class.oid = index_row.indexrelid
        LEFT JOIN pg_constraint constraint_row ON constraint_row.conindid = index_row.indexrelid
        WHERE table_namespace.nspname = 'public'
            AND table_class.relname = 'notifications'
            AND index_row.indisunique = true
            AND (
                SELECT array_agg(attribute_row.attname ORDER BY index_key.ordinality)
                FROM unnest(index_row.indkey::smallint[]) WITH ORDINALITY AS index_key(attnum, ordinality)
                INNER JOIN pg_attribute attribute_row
                    ON attribute_row.attrelid = index_row.indrelid
                    AND attribute_row.attnum = index_key.attnum
            ) = ARRAY['recipient_id', 'workspace_id', 'entity_id', 'entity_type']::name[]
    LOOP
        IF legacy_index.constraint_name IS NOT NULL THEN
            EXECUTE format(
                'ALTER TABLE public.notifications DROP CONSTRAINT %I',
                legacy_index.constraint_name
            );
        ELSE
            EXECUTE format('DROP INDEX public.%I', legacy_index.index_name);
        END IF;
    END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_dedupe_key
    ON public.notifications (dedupe_key);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_entity_created
    ON public.notifications (recipient_id, entity_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_entity_unread
    ON public.notifications (recipient_id, entity_type)
    WHERE read_at IS NULL;
