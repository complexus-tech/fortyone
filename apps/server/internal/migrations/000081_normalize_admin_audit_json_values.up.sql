CREATE OR REPLACE FUNCTION public.normalize_admin_audit_json_values()
RETURNS trigger AS $$
BEGIN
    NEW.old_value := COALESCE(NEW.old_value, 'null'::jsonb);
    NEW.new_value := COALESCE(NEW.new_value, 'null'::jsonb);
    NEW.metadata := COALESCE(NEW.metadata, '{}'::jsonb);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

UPDATE public.admin_audit_logs
SET old_value = 'null'::jsonb
WHERE old_value IS NULL;

UPDATE public.admin_audit_logs
SET new_value = 'null'::jsonb
WHERE new_value IS NULL;

UPDATE public.admin_audit_logs
SET metadata = '{}'::jsonb
WHERE metadata IS NULL;

DROP TRIGGER IF EXISTS normalize_admin_audit_json_values ON public.admin_audit_logs;
CREATE TRIGGER normalize_admin_audit_json_values
    BEFORE INSERT OR UPDATE OF old_value, new_value, metadata
    ON public.admin_audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION public.normalize_admin_audit_json_values();
