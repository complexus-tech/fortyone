DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.feedback_item_attachments LIMIT 1)
        OR EXISTS (
            SELECT 1
            FROM public.feedback_items
            WHERE description_html <> ''
            LIMIT 1
        ) THEN
        RAISE EXCEPTION 'refusing to discard feedback rich descriptions or attachment relations';
    END IF;
END;
$$;

DROP TABLE IF EXISTS public.feedback_item_attachments;

ALTER TABLE public.feedback_items
    DROP COLUMN IF EXISTS description_html;
