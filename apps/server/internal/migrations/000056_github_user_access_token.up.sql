ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS github_access_token text;
