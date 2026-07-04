ALTER TABLE public.feedback_portals
    ALTER COLUMN is_public SET DEFAULT true;

UPDATE public.feedback_portals
SET is_public = true,
    updated_at = now()
WHERE is_public = false;
