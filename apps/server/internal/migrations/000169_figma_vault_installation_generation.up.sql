-- Establish an immutable installation generation for credential AAD and
-- durable webhook fencing. Existing provider-local ciphertext is explicitly
-- marked version 1 and must be upgraded by the bounded Figma credential
-- migration before the new runtime will open it.
ALTER TABLE public.figma_connections
    ADD COLUMN credential_key_version smallint NOT NULL DEFAULT 1,
    ADD COLUMN installation_generation uuid NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE public.figma_connections
    ADD CONSTRAINT figma_connections_credential_key_version_check
    CHECK (credential_key_version > 0);

CREATE UNIQUE INDEX figma_connections_active_generation
    ON public.figma_connections (workspace_id, installation_generation)
    WHERE is_active;

COMMENT ON COLUMN public.figma_connections.credential_key_version IS
    '1 is the isolated legacy envelope; 2 is the shared context-bound credential vault';
COMMENT ON COLUMN public.figma_connections.installation_generation IS
    'Immutable generation that fences OAuth credentials and durable webhook deliveries';
