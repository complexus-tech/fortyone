ALTER TABLE public.github_installations
    ADD COLUMN installation_generation uuid NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN installation_authorized_at timestamptz NOT NULL DEFAULT now();

CREATE UNIQUE INDEX github_installations_installation_generation_key
    ON public.github_installations USING btree (installation_generation);

CREATE FUNCTION public.rotate_github_installation_generation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.github_installation_id IS DISTINCT FROM NEW.github_installation_id
        OR OLD.workspace_id IS DISTINCT FROM NEW.workspace_id
        OR OLD.repository_selection IS DISTINCT FROM NEW.repository_selection
        OR OLD.permissions IS DISTINCT FROM NEW.permissions
        OR OLD.events IS DISTINCT FROM NEW.events
        OR OLD.installed_by_user_id IS DISTINCT FROM NEW.installed_by_user_id
        OR OLD.installed_by_github_user_id IS DISTINCT FROM NEW.installed_by_github_user_id
        OR OLD.is_active IS DISTINCT FROM NEW.is_active
        OR OLD.suspended_at IS DISTINCT FROM NEW.suspended_at
        OR OLD.disconnected_at IS DISTINCT FROM NEW.disconnected_at
    THEN
        NEW.installation_generation = gen_random_uuid();
        NEW.installation_authorized_at = now();
    END IF;
    RETURN NEW;
END
$$;

CREATE TRIGGER github_installations_rotate_generation
BEFORE UPDATE ON public.github_installations
FOR EACH ROW
EXECUTE FUNCTION public.rotate_github_installation_generation();

COMMENT ON COLUMN public.github_installations.installation_generation IS
    'Rotated whenever an installation grant is reauthorized; durable webhook work must match the current generation before processing.';
COMMENT ON COLUMN public.github_installations.installation_authorized_at IS
    'Time at which the current installation generation became authorized.';
