DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.feedback_items
        WHERE merged_into_item_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'migration 000124 cannot be rolled back while merged feedback items exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.feedback_item_merge_outbox
    ) THEN
        RAISE EXCEPTION 'migration 000124 cannot be rolled back while feedback item merge events exist';
    END IF;
END
$$;

DROP TABLE IF EXISTS public.feedback_item_merge_outbox;

DROP INDEX IF EXISTS public.idx_feedback_items_public_active;
DROP INDEX IF EXISTS public.idx_feedback_items_merged_into_item;

ALTER TABLE public.feedback_items
    DROP CONSTRAINT IF EXISTS feedback_items_merge_lifecycle_check,
    DROP CONSTRAINT IF EXISTS feedback_items_merge_not_self_check,
    DROP CONSTRAINT IF EXISTS feedback_items_merged_by_user_id_fkey,
    DROP CONSTRAINT IF EXISTS feedback_items_merged_into_item_fkey,
    DROP CONSTRAINT IF EXISTS feedback_items_workspace_portal_id_key,
    DROP COLUMN IF EXISTS merged_by_user_id,
    DROP COLUMN IF EXISTS merged_at,
    DROP COLUMN IF EXISTS merged_into_item_id;
