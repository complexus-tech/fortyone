-- An empty restriction set means workspace-wide access for machine credentials.
-- Removing the final team must therefore revoke the credential in the same
-- transaction, including when the restriction disappears through an FK cascade.
CREATE FUNCTION public.revoke_api_credential_after_final_team_removal()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    revoked_credential public.api_credentials%ROWTYPE;
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.credential_id = OLD.credential_id
       AND NEW.workspace_id = OLD.workspace_id
       AND NEW.team_id = OLD.team_id THEN
        RETURN NULL;
    END IF;

    -- Serialize concurrent removals and rotation on the credential. An actual
    -- row version change also makes a stale REPEATABLE READ/SERIALIZABLE
    -- transaction abort instead of overlooking another committed removal.
    -- The following statement receives a fresh READ COMMITTED snapshot after
    -- any lock wait. Already-revoked or concurrently deleted parents need no work.
    UPDATE public.api_credentials AS credential
    SET revoked_at = credential.revoked_at
    WHERE credential.credential_id = OLD.credential_id
      AND credential.workspace_id = OLD.workspace_id
      AND credential.revoked_at IS NULL;

    IF NOT FOUND THEN
        RETURN NULL;
    END IF;

    UPDATE public.api_credentials AS credential
    SET revoked_at = CURRENT_TIMESTAMP,
        revoked_reason = 'team_restrictions_exhausted'
    WHERE credential.credential_id = OLD.credential_id
      AND credential.workspace_id = OLD.workspace_id
      AND credential.revoked_at IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM public.api_credential_team_restrictions AS restriction
          WHERE restriction.credential_id = credential.credential_id
      )
    RETURNING credential.* INTO revoked_credential;

    IF FOUND THEN
        -- This is an automatic database policy action, not an action performed
        -- by the credential owner. The zero UUID identifies the system policy;
        -- the removed team remains an immutable fact even after its deletion.
        INSERT INTO public.developer_credential_audit_events (
            event_id, workspace_id, actor_kind, actor_id, operation,
            subject_type, subject_id, result, reason_code, metadata, created_at
        ) VALUES (
            gen_random_uuid(),
            revoked_credential.workspace_id,
            'system',
            '00000000-0000-0000-0000-000000000000',
            CASE revoked_credential.kind
                WHEN 'personal_access_token' THEN 'personal_token.revoked'
                ELSE 'service_account_key.revoked'
            END,
            'api_credential',
            revoked_credential.credential_id,
            'succeeded',
            'team_restrictions_exhausted',
            jsonb_build_object('removed_team_id', OLD.team_id),
            CURRENT_TIMESTAMP
        );
    END IF;

    RETURN NULL;
END;
$$;

CREATE TRIGGER api_credential_final_team_removal_revokes
AFTER DELETE OR UPDATE OF credential_id, workspace_id, team_id
ON public.api_credential_team_restrictions
FOR EACH ROW
EXECUTE FUNCTION public.revoke_api_credential_after_final_team_removal();
