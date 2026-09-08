-- Notification identity is the event dedupe key, not the recipient/resource
-- pair. Migration 83 removed only one column ordering of the legacy key;
-- production also used (recipient_id, workspace_id, entity_type, entity_id).
DO $$
DECLARE
    legacy_index record;
BEGIN
    FOR legacy_index IN
        SELECT index_class.relname AS index_name,
               constraint_row.conname AS constraint_name
        FROM pg_index AS index_row
        JOIN pg_class AS index_class ON index_class.oid = index_row.indexrelid
        LEFT JOIN pg_constraint AS constraint_row ON constraint_row.conindid = index_row.indexrelid
        WHERE index_row.indrelid = 'public.notifications'::regclass
          AND index_row.indisunique
          AND NOT index_row.indisprimary
          AND index_row.indpred IS NULL
          AND index_row.indexprs IS NULL
          AND index_row.indnkeyatts = 4
          AND (
              SELECT array_agg(attribute_row.attname ORDER BY attribute_row.attname)
              FROM unnest(index_row.indkey::smallint[]) WITH ORDINALITY AS index_key(attnum, ordinality)
              JOIN pg_attribute AS attribute_row
                ON attribute_row.attrelid = index_row.indrelid
               AND attribute_row.attnum = index_key.attnum
              WHERE index_key.ordinality <= index_row.indnkeyatts
          ) = ARRAY['entity_id', 'entity_type', 'recipient_id', 'workspace_id']::name[]
    LOOP
        IF legacy_index.constraint_name IS NOT NULL THEN
            EXECUTE format('ALTER TABLE public.notifications DROP CONSTRAINT %I', legacy_index.constraint_name);
        ELSE
            EXECUTE format('DROP INDEX public.%I', legacy_index.index_name);
        END IF;
    END LOOP;
END $$;
