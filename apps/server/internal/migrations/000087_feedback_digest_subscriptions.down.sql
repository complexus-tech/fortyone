DROP TABLE IF EXISTS public.feedback_digest_deliveries;
DROP TABLE IF EXISTS public.feedback_board_subscriptions;
DROP INDEX IF EXISTS public.idx_feedback_items_board_external_created;

ALTER TABLE public.feedback_items
    DROP COLUMN IF EXISTS submission_source;
