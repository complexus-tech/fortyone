ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS is_internal bool NOT NULL DEFAULT false;
