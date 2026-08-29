-- Strategy notifications use a stable entity (the workspace) and a distinct
-- dedupe key for each communication period. Remove legacy uniqueness that
-- treats the workspace entity as a single notification per recipient.
DO $$
DECLARE
    legacy_index record;
BEGIN
    FOR legacy_index IN
        SELECT index_name, constraint_name
        FROM (
            SELECT
                index_class.relname AS index_name,
                constraint_row.conname AS constraint_name,
                (
                    SELECT array_agg(attribute_row.attname ORDER BY index_key.ordinality)
                    FROM unnest(index_row.indkey::smallint[]) WITH ORDINALITY AS index_key(attnum, ordinality)
                    INNER JOIN pg_attribute attribute_row
                        ON attribute_row.attrelid = index_row.indrelid
                        AND attribute_row.attnum = index_key.attnum
                ) AS column_names
            FROM pg_index index_row
            INNER JOIN pg_class table_class ON table_class.oid = index_row.indrelid
            INNER JOIN pg_namespace table_namespace ON table_namespace.oid = table_class.relnamespace
            INNER JOIN pg_class index_class ON index_class.oid = index_row.indexrelid
            LEFT JOIN pg_constraint constraint_row ON constraint_row.conindid = index_row.indexrelid
            WHERE table_namespace.nspname = 'public'
                AND table_class.relname = 'notifications'
                AND index_row.indisunique = true
        ) AS candidate
        WHERE candidate.column_names = ARRAY['recipient_id', 'workspace_id', 'entity_id']::name[]
            OR candidate.column_names = ARRAY['recipient_id', 'workspace_id', 'entity_id', 'entity_type']::name[]
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

