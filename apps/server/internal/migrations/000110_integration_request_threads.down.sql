DROP TABLE IF EXISTS public.integration_request_comments;
DROP TABLE IF EXISTS public.integration_request_threads;

ALTER TABLE public.integration_requests
    DROP COLUMN IF EXISTS label_ids;
