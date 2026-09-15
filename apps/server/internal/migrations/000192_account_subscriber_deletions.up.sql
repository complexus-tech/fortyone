-- Retain only the delivery address required to finish deleting provider data.
-- No user FK: deletion of the account must not discard its cleanup obligation.
CREATE TABLE public.account_subscriber_deletions (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL CHECK (length(btrim(email)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    lease_token UUID,
    lease_expires_at TIMESTAMPTZ,
    CHECK ((lease_token IS NULL) = (lease_expires_at IS NULL))
);
CREATE INDEX account_subscriber_deletions_ready_idx
    ON public.account_subscriber_deletions (next_attempt_at, created_at, id);
CREATE INDEX account_subscriber_deletions_email_idx
    ON public.account_subscriber_deletions (lower(btrim(email)));
