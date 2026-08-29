DROP INDEX IF EXISTS public.idx_feedback_comments_item_parent_created;

ALTER TABLE public.feedback_comments
    DROP CONSTRAINT IF EXISTS feedback_comments_parent_not_self_check,
    DROP CONSTRAINT IF EXISTS feedback_comments_parent_id_fkey,
    DROP COLUMN IF EXISTS parent_id;
