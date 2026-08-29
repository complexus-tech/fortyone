DROP INDEX IF EXISTS public.idx_feedback_items_deleted_at;

ALTER TABLE public.feedback_items
    DROP COLUMN IF EXISTS deleted_at;
