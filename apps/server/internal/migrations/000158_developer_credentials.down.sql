DROP TRIGGER IF EXISTS developer_credential_audit_events_immutable
    ON public.developer_credential_audit_events;
DROP FUNCTION IF EXISTS public.reject_developer_credential_audit_mutation();
DROP TABLE IF EXISTS public.developer_credential_audit_events;
DROP TABLE IF EXISTS public.api_credential_team_restrictions;
ALTER TABLE public.teams
    DROP CONSTRAINT IF EXISTS teams_team_id_workspace_id_key;
DROP TRIGGER IF EXISTS api_credential_scopes_service_account_boundary
    ON public.api_credential_scopes;
DROP FUNCTION IF EXISTS public.enforce_service_account_scope_boundary();
DROP TABLE IF EXISTS public.api_credential_scopes;
DROP TRIGGER IF EXISTS api_credentials_principal_kind_enforced
    ON public.api_credentials;
DROP FUNCTION IF EXISTS public.enforce_api_credential_principal_kind();
DROP TRIGGER IF EXISTS api_credentials_identity_immutable
    ON public.api_credentials;
DROP FUNCTION IF EXISTS public.reject_api_credential_identity_mutation();
DROP TABLE IF EXISTS public.api_credentials;
DROP TRIGGER IF EXISTS principals_identity_immutable ON public.principals;
DROP FUNCTION IF EXISTS public.reject_principal_identity_mutation();
DROP TABLE IF EXISTS public.principals;
