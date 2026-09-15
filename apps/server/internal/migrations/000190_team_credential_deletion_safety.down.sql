DROP TRIGGER IF EXISTS api_credential_final_team_removal_revokes
    ON public.api_credential_team_restrictions;
DROP FUNCTION IF EXISTS public.revoke_api_credential_after_final_team_removal();

-- Existing revocations and their immutable audit events remain in force.
-- Rolling back the trigger must never reactivate an exhausted credential.
