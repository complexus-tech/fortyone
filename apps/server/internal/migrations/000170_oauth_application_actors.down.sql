DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.oauth_client_secrets)
       OR EXISTS (SELECT 1 FROM public.oauth_application_installations)
       OR EXISTS (
           SELECT 1
           FROM public.oauth_applications
           WHERE owner_workspace_id IS NOT NULL OR owner_user_id IS NOT NULL
       )
       OR EXISTS (
           SELECT 1
           FROM public.principals
           WHERE kind = 'oauth_application'
       )
       OR EXISTS (
           SELECT 1
           FROM public.oauth_audit_events
           WHERE workspace_id IS NOT NULL
              OR installation_id IS NOT NULL
              OR principal_id IS NOT NULL
              OR secret_id IS NOT NULL
              OR actor_kind IS NOT NULL
              OR actor_id IS NOT NULL
              OR actor_credential_id IS NOT NULL
              OR request_id IS NOT NULL
              OR subject_type IS NOT NULL
              OR subject_id IS NOT NULL
       ) THEN
        RAISE EXCEPTION 'migration 000170 is forward-only after OAuth application actor data exists'
            USING ERRCODE = '55000';
    END IF;
END;
$$;

DROP INDEX public.oauth_audit_events_subject_created_idx;
DROP INDEX public.oauth_audit_events_installation_created_idx;
DROP INDEX public.oauth_audit_events_workspace_created_idx;

ALTER TABLE public.oauth_audit_events
    DROP CONSTRAINT oauth_audit_events_subject_check,
    DROP CONSTRAINT oauth_audit_events_request_id_check,
    DROP CONSTRAINT oauth_audit_events_actor_shape_check,
    DROP CONSTRAINT oauth_audit_events_actor_kind_check,
    DROP COLUMN subject_id,
    DROP COLUMN subject_type,
    DROP COLUMN request_id,
    DROP COLUMN actor_credential_id,
    DROP COLUMN actor_id,
    DROP COLUMN actor_kind,
    DROP COLUMN secret_id,
    DROP COLUMN principal_id,
    DROP COLUMN installation_id,
    DROP COLUMN workspace_id;

DROP TABLE public.oauth_application_installation_scopes;
DROP TRIGGER oauth_application_installations_identity_immutable
    ON public.oauth_application_installations;
DROP FUNCTION public.reject_oauth_application_installation_identity_mutation();
DROP TRIGGER oauth_application_installations_principal_enforced
    ON public.oauth_application_installations;
DROP FUNCTION public.enforce_oauth_application_installation_principal();
DROP TABLE public.oauth_application_installations;

DROP TRIGGER oauth_client_secrets_identity_immutable ON public.oauth_client_secrets;
DROP FUNCTION public.reject_oauth_client_secret_identity_mutation();
DROP TRIGGER oauth_client_secrets_lifecycle_fenced ON public.oauth_client_secrets;
DROP FUNCTION public.enforce_oauth_client_secret_lifecycle();
DROP TABLE public.oauth_client_secrets;

DROP TRIGGER oauth_applications_identity_immutable ON public.oauth_applications;
DROP FUNCTION public.reject_oauth_application_identity_mutation();
DROP INDEX public.oauth_applications_owner_created_idx;
ALTER TABLE public.oauth_applications
    DROP CONSTRAINT oauth_applications_owner_shape_check,
    DROP CONSTRAINT oauth_applications_owner_user_id_fkey,
    DROP CONSTRAINT oauth_applications_owner_workspace_id_fkey,
    DROP COLUMN owner_user_id,
    DROP COLUMN owner_workspace_id;

DROP INDEX public.principals_workspace_oauth_applications_idx;
ALTER TABLE public.principals
    DROP CONSTRAINT principals_machine_role_check,
    DROP CONSTRAINT principals_identity_shape_check,
    DROP CONSTRAINT principals_kind_check,
    ADD CONSTRAINT principals_kind_check
        CHECK (kind IN ('human_user', 'service_account')),
    ADD CONSTRAINT principals_identity_shape_check CHECK (
        (
            kind = 'human_user'
            AND subject_user_id IS NOT NULL
            AND workspace_role IS NULL
        )
        OR
        (
            kind = 'service_account'
            AND subject_user_id IS NULL
            AND workspace_role IS NOT NULL
        )
    ),
    ADD CONSTRAINT principals_service_account_role_check CHECK (
        kind <> 'service_account'
        OR workspace_role IN ('guest', 'member')
    );
