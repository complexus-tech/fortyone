DROP TRIGGER IF EXISTS normalize_admin_audit_json_values ON public.admin_audit_logs;
DROP FUNCTION IF EXISTS public.normalize_admin_audit_json_values();

UPDATE public.admin_audit_logs
SET old_value = NULL
WHERE old_value = 'null'::jsonb;

UPDATE public.admin_audit_logs
SET new_value = NULL
WHERE new_value = 'null'::jsonb;
