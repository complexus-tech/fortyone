ALTER TABLE public.messaging_outbound_deliveries
    DROP CONSTRAINT messaging_outbound_deliveries_purpose_check,
    ADD CONSTRAINT messaging_outbound_deliveries_purpose_check
    CHECK (purpose IN ('provider_message', 'assistant', 'account_link', 'access', 'creation_confirmation', 'onboarding')),
    ADD CONSTRAINT messaging_outbound_deliveries_onboarding_recipient_check
    CHECK (
        purpose <> 'onboarding'
        OR (
            provider = 'slack'
            AND btrim(external_workspace_id) <> ''
            AND external_recipient_user_id IS NOT NULL
            AND btrim(external_recipient_user_id) <> ''
        )
    );

-- This marker intentionally outlives operational message retention so the
-- welcome remains once-only after reconnects. Persist only a SHA-256 digest of
-- the external Slack identity, never its raw team/user IDs or message content.
CREATE TABLE public.slack_user_onboarding_receipts (
    workspace_id uuid NOT NULL,
    external_identity_digest bytea NOT NULL,
    first_observed_at timestamptz NOT NULL,
    guide_delivered_at timestamptz,
    CONSTRAINT slack_user_onboarding_receipts_workspace_id_fkey
        FOREIGN KEY (workspace_id)
        REFERENCES public.workspaces(workspace_id)
        ON DELETE CASCADE,
    CONSTRAINT slack_user_onboarding_receipts_digest_check
        CHECK (octet_length(external_identity_digest) = 32),
    PRIMARY KEY (workspace_id, external_identity_digest)
);

-- Do not re-introduce existing users when this feature is deployed. Each source
-- below proves prior use without needing to retain another raw external ID in
-- the long-lived receipt. Request logs are written only after Slack signature
-- verification; user links and outbound deliveries are already workspace-bound.
WITH historic_slack_users AS (
    SELECT
        workspace_id,
        btrim(slack_team_id) AS slack_team_id,
        btrim(slack_user_id) AS slack_user_id,
        COALESCE(linked_at, created_at) AS observed_at
    FROM public.slack_user_links
    WHERE btrim(slack_team_id) <> ''
      AND btrim(slack_user_id) <> ''

    UNION ALL

    SELECT
        workspace_id,
        btrim(slack_team_id),
        btrim(slack_user_id),
        created_at
    FROM public.slack_request_logs
    WHERE workspace_id IS NOT NULL
      AND slack_team_id IS NOT NULL
      AND btrim(slack_team_id) <> ''
      AND slack_user_id IS NOT NULL
      AND btrim(slack_user_id) <> ''

    UNION ALL

    SELECT
        workspace_id,
        btrim(external_workspace_id),
        btrim(external_recipient_user_id),
        created_at
    FROM public.messaging_outbound_deliveries
    WHERE provider = 'slack'
      AND btrim(external_workspace_id) <> ''
      AND external_recipient_user_id IS NOT NULL
      AND btrim(external_recipient_user_id) <> ''
)
INSERT INTO public.slack_user_onboarding_receipts (
    workspace_id,
    external_identity_digest,
    first_observed_at,
    guide_delivered_at
)
SELECT
    workspace_id,
    sha256(convert_to(slack_team_id || chr(31) || slack_user_id, 'UTF8')),
    MIN(observed_at),
    NULL
FROM historic_slack_users
GROUP BY workspace_id, slack_team_id, slack_user_id
ON CONFLICT (workspace_id, external_identity_digest) DO NOTHING;

CREATE OR REPLACE FUNCTION public.record_slack_user_onboarding_receipt()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.provider = 'slack'
       AND NEW.purpose = 'onboarding'
       AND NEW.status = 'delivered'
       AND NEW.external_recipient_user_id IS NOT NULL
       AND btrim(NEW.external_recipient_user_id) <> '' THEN
        INSERT INTO public.slack_user_onboarding_receipts (
            workspace_id,
            external_identity_digest,
            first_observed_at,
            guide_delivered_at
        ) VALUES (
            NEW.workspace_id,
            sha256(convert_to(
                btrim(NEW.external_workspace_id) || chr(31) || btrim(NEW.external_recipient_user_id),
                'UTF8'
            )),
            NEW.created_at,
            COALESCE(NEW.delivered_at, NOW())
        )
        ON CONFLICT (workspace_id, external_identity_digest) DO NOTHING;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER messaging_outbound_slack_user_onboarding_receipt
AFTER INSERT OR UPDATE OF status
ON public.messaging_outbound_deliveries
FOR EACH ROW
EXECUTE FUNCTION public.record_slack_user_onboarding_receipt();
