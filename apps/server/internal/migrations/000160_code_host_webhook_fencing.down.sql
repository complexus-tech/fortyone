DROP TRIGGER IF EXISTS github_installations_rotate_generation
    ON public.github_installations;
DROP FUNCTION IF EXISTS public.rotate_github_installation_generation();

DROP INDEX IF EXISTS public.github_installations_installation_generation_key;

ALTER TABLE public.github_installations
    DROP COLUMN IF EXISTS installation_authorized_at,
    DROP COLUMN IF EXISTS installation_generation;
