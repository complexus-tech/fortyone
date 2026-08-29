ALTER TYPE public.notification_type
    ADD VALUE IF NOT EXISTS 'feedback_update_published';

ALTER TABLE public.feedback_portals
    DROP CONSTRAINT feedback_portals_participation_mode_check,
    ADD CONSTRAINT feedback_portals_participation_mode_check
        CHECK (participation_mode IN ('account_required', 'verified_guest', 'anonymous_allowed'));

ALTER TABLE public.feedback_portals
    ADD COLUMN guest_identity_policy text NOT NULL DEFAULT 'show_identity',
    ADD CONSTRAINT feedback_portals_guest_identity_policy_check
        CHECK (guest_identity_policy IN ('show_identity', 'allow_public_masking', 'always_mask_guests'));

ALTER TABLE public.feedback_contributors
    DROP CONSTRAINT feedback_contributors_kind_check,
    DROP CONSTRAINT feedback_contributors_identity_check,
    ADD COLUMN email text,
    ADD COLUMN email_verified_at timestamptz,
    ADD COLUMN display_name text,
    ADD COLUMN avatar_url text,
    ADD COLUMN public_masked boolean NOT NULL DEFAULT false,
    ADD COLUMN external_id text,
    ADD COLUMN last_seen_at timestamptz,
    ADD COLUMN blocked_at timestamptz,
    ADD COLUMN blocked_reason text,
    ADD CONSTRAINT feedback_contributors_kind_check
        CHECK (kind IN ('account', 'verified_guest', 'anonymous', 'external')),
    ADD CONSTRAINT feedback_contributors_email_check
        CHECK (
            email IS NULL
            OR (
                email = lower(btrim(email))
                AND length(email) BETWEEN 3 AND 320
                AND email !~ '[[:space:]]'
            )
        ),
    ADD CONSTRAINT feedback_contributors_display_name_check
        CHECK (display_name IS NULL OR length(btrim(display_name)) BETWEEN 1 AND 200),
    ADD CONSTRAINT feedback_contributors_avatar_url_check
        CHECK (avatar_url IS NULL OR length(btrim(avatar_url)) BETWEEN 1 AND 2048),
    ADD CONSTRAINT feedback_contributors_external_id_check
        CHECK (external_id IS NULL OR (external_id = btrim(external_id) AND length(external_id) BETWEEN 1 AND 512)),
    ADD CONSTRAINT feedback_contributors_block_check
        CHECK (
            (blocked_at IS NULL AND blocked_reason IS NULL)
            OR (
                blocked_at IS NOT NULL
                AND (blocked_reason IS NULL OR length(btrim(blocked_reason)) BETWEEN 1 AND 2000)
            )
        ),
    ADD CONSTRAINT feedback_contributors_public_masked_check
        CHECK (NOT public_masked OR kind = 'verified_guest'),
    ADD CONSTRAINT feedback_contributors_identity_check
        CHECK (
            (
                kind = 'account'
                AND external_id IS NULL
                AND email IS NULL
                AND email_verified_at IS NULL
            )
            OR (
                kind = 'verified_guest'
                AND user_id IS NULL
                AND email IS NOT NULL
                AND email_verified_at IS NOT NULL
                AND external_id IS NULL
            )
            OR (
                kind = 'anonymous'
                AND user_id IS NULL
                AND email IS NULL
                AND email_verified_at IS NULL
                AND display_name IS NULL
                AND avatar_url IS NULL
                AND external_id IS NULL
            )
            OR (
                kind = 'external'
                AND user_id IS NULL
                AND external_id IS NOT NULL
                AND (
                    (email IS NULL AND email_verified_at IS NULL)
                    OR (email IS NOT NULL AND email_verified_at IS NOT NULL)
                )
            )
        );

CREATE UNIQUE INDEX feedback_contributors_portal_email_unique
    ON public.feedback_contributors (portal_id, lower(email))
    WHERE email IS NOT NULL AND kind <> 'external';

CREATE UNIQUE INDEX feedback_contributors_portal_external_id_unique
    ON public.feedback_contributors (portal_id, external_id)
    WHERE external_id IS NOT NULL;

CREATE INDEX idx_feedback_contributors_portal_kind_seen
    ON public.feedback_contributors (portal_id, kind, last_seen_at DESC, id DESC);

CREATE TABLE public.feedback_contributor_verifications (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    portal_id uuid NOT NULL,
    email text NOT NULL,
    display_name text,
    public_masked boolean NOT NULL DEFAULT false,
    token_hash bytea NOT NULL,
    code_hash bytea NOT NULL,
    source text NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    attempt_count integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_contributor_verifications_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_verifications_email_check
        CHECK (
            email = lower(btrim(email))
            AND length(email) BETWEEN 3 AND 320
            AND email !~ '[[:space:]]'
        ),
    CONSTRAINT feedback_contributor_verifications_display_name_check
        CHECK (display_name IS NULL OR length(btrim(display_name)) BETWEEN 1 AND 200),
    CONSTRAINT feedback_contributor_verifications_token_hash_check
        CHECK (octet_length(token_hash) = 32),
    CONSTRAINT feedback_contributor_verifications_code_hash_check
        CHECK (octet_length(code_hash) = 32),
    CONSTRAINT feedback_contributor_verifications_source_check
        CHECK (source IN ('portal', 'widget')),
    CONSTRAINT feedback_contributor_verifications_expiry_check
        CHECK (expires_at > created_at),
    CONSTRAINT feedback_contributor_verifications_consumed_check
        CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    CONSTRAINT feedback_contributor_verifications_attempt_count_check
        CHECK (attempt_count BETWEEN 0 AND 10),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_contributor_verifications_token_hash_key
    ON public.feedback_contributor_verifications (token_hash);

CREATE INDEX idx_feedback_contributor_verifications_active_email
    ON public.feedback_contributor_verifications (portal_id, email, created_at DESC)
    WHERE consumed_at IS NULL;

CREATE INDEX idx_feedback_contributor_verifications_active_code
    ON public.feedback_contributor_verifications (portal_id, email, code_hash)
    WHERE consumed_at IS NULL;

CREATE INDEX idx_feedback_contributor_verifications_expiry
    ON public.feedback_contributor_verifications (expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE public.feedback_contributor_sessions (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    portal_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    source text NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_contributor_sessions_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_sessions_contributor_fkey
        FOREIGN KEY (portal_id, contributor_id)
        REFERENCES public.feedback_contributors(portal_id, id)
        ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_sessions_token_hash_check
        CHECK (octet_length(token_hash) = 32),
    CONSTRAINT feedback_contributor_sessions_source_check
        CHECK (source IN ('portal', 'widget', 'preferences')),
    CONSTRAINT feedback_contributor_sessions_expiry_check
        CHECK (expires_at > created_at),
    CONSTRAINT feedback_contributor_sessions_revoked_check
        CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CONSTRAINT feedback_contributor_sessions_last_used_check
        CHECK (last_used_at IS NULL OR last_used_at >= created_at),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_contributor_sessions_token_hash_key
    ON public.feedback_contributor_sessions (token_hash);

CREATE INDEX idx_feedback_contributor_sessions_active_contributor
    ON public.feedback_contributor_sessions (portal_id, contributor_id, expires_at DESC)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_feedback_contributor_sessions_expiry
    ON public.feedback_contributor_sessions (expires_at)
    WHERE revoked_at IS NULL;

-- Every legacy account participant needs one portal-scoped contributor before
-- comments and votes can make contributor identity authoritative.
WITH account_participants AS (
    SELECT item.portal_id, vote.user_id
    FROM public.feedback_votes AS vote
    INNER JOIN public.feedback_items AS item ON item.id = vote.item_id
    UNION
    SELECT item.portal_id, comment.author_id AS user_id
    FROM public.feedback_comments AS comment
    INNER JOIN public.feedback_items AS item ON item.id = comment.item_id
    WHERE comment.author_id IS NOT NULL
)
INSERT INTO public.feedback_contributors (portal_id, user_id, kind)
SELECT portal_id, user_id, 'account'
FROM account_participants
ON CONFLICT (portal_id, user_id) WHERE user_id IS NOT NULL DO NOTHING;

ALTER TABLE public.feedback_comments
    ADD COLUMN contributor_id uuid;

UPDATE public.feedback_comments AS comment
SET contributor_id = mapping.contributor_id
FROM (
    SELECT
        item.id AS item_id,
        contributor.id AS contributor_id,
        contributor.user_id
    FROM public.feedback_items AS item
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.portal_id = item.portal_id
) AS mapping
WHERE mapping.item_id = comment.item_id
    AND mapping.user_id = comment.author_id
    AND comment.author_id IS NOT NULL;

-- A null legacy author means the user was already deleted. Preserve each
-- comment without inventing a link to another comment or item author.
CREATE TEMPORARY TABLE feedback_anonymous_comment_backfill (
    comment_id uuid PRIMARY KEY,
    contributor_id uuid NOT NULL UNIQUE,
    portal_id uuid NOT NULL
) ON COMMIT DROP;

INSERT INTO feedback_anonymous_comment_backfill (comment_id, contributor_id, portal_id)
SELECT comment.id, gen_random_uuid(), item.portal_id
FROM public.feedback_comments AS comment
INNER JOIN public.feedback_items AS item ON item.id = comment.item_id
WHERE comment.contributor_id IS NULL
    AND comment.author_id IS NULL;

INSERT INTO public.feedback_contributors (id, portal_id, kind)
SELECT contributor_id, portal_id, 'anonymous'
FROM feedback_anonymous_comment_backfill;

UPDATE public.feedback_comments AS comment
SET contributor_id = mapping.contributor_id
FROM feedback_anonymous_comment_backfill AS mapping
WHERE comment.id = mapping.comment_id;

DROP TABLE feedback_anonymous_comment_backfill;

ALTER TABLE public.feedback_comments
    ALTER COLUMN contributor_id SET NOT NULL,
    ADD CONSTRAINT feedback_comments_contributor_id_fkey
        FOREIGN KEY (contributor_id) REFERENCES public.feedback_contributors(id)
        ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX idx_feedback_comments_contributor_created
    ON public.feedback_comments (contributor_id, created_at DESC, id DESC);

ALTER TABLE public.feedback_votes
    ADD COLUMN contributor_id uuid;

UPDATE public.feedback_votes AS vote
SET contributor_id = mapping.contributor_id
FROM (
    SELECT
        item.id AS item_id,
        contributor.id AS contributor_id,
        contributor.user_id
    FROM public.feedback_items AS item
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.portal_id = item.portal_id
) AS mapping
WHERE mapping.item_id = vote.item_id
    AND mapping.user_id = vote.user_id;

ALTER TABLE public.feedback_votes
    DROP CONSTRAINT feedback_votes_pkey,
    DROP CONSTRAINT feedback_votes_user_id_fkey,
    ALTER COLUMN user_id DROP NOT NULL,
    ALTER COLUMN contributor_id SET NOT NULL,
    ADD CONSTRAINT feedback_votes_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    ADD CONSTRAINT feedback_votes_contributor_id_fkey
        FOREIGN KEY (contributor_id) REFERENCES public.feedback_contributors(id)
        ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
    ADD CONSTRAINT feedback_votes_pkey PRIMARY KEY (item_id, contributor_id);

CREATE UNIQUE INDEX feedback_votes_item_user_unique
    ON public.feedback_votes (item_id, user_id)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_feedback_votes_contributor_created
    ON public.feedback_votes (contributor_id, created_at DESC, item_id);

CREATE TABLE public.feedback_item_followers (
    item_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    unsubscribed_at timestamptz,
    CONSTRAINT feedback_item_followers_item_id_fkey
        FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_item_followers_contributor_id_fkey
        FOREIGN KEY (contributor_id) REFERENCES public.feedback_contributors(id) ON DELETE CASCADE,
    CONSTRAINT feedback_item_followers_unsubscribed_check
        CHECK (unsubscribed_at IS NULL OR unsubscribed_at >= created_at),
    PRIMARY KEY (item_id, contributor_id)
);

CREATE INDEX idx_feedback_item_followers_active_contributor
    ON public.feedback_item_followers (contributor_id, created_at DESC, item_id)
    WHERE unsubscribed_at IS NULL;

-- Existing contactable authors should retain the update loop. Truly anonymous
-- authors remain tracking-link only and are intentionally not followed.
INSERT INTO public.feedback_item_followers (item_id, contributor_id, created_at)
SELECT item.id, item.contributor_id, item.created_at
FROM public.feedback_items AS item
INNER JOIN public.feedback_contributors AS contributor ON contributor.id = item.contributor_id
WHERE contributor.kind <> 'anonymous'
ON CONFLICT (item_id, contributor_id) DO NOTHING;

CREATE TABLE public.feedback_portal_followers (
    portal_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    unsubscribed_at timestamptz,
    CONSTRAINT feedback_portal_followers_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_portal_followers_contributor_fkey
        FOREIGN KEY (portal_id, contributor_id)
        REFERENCES public.feedback_contributors(portal_id, id)
        ON DELETE CASCADE,
    CONSTRAINT feedback_portal_followers_unsubscribed_check
        CHECK (unsubscribed_at IS NULL OR unsubscribed_at >= created_at),
    PRIMARY KEY (portal_id, contributor_id)
);

CREATE INDEX idx_feedback_portal_followers_active_contributor
    ON public.feedback_portal_followers (contributor_id, created_at DESC, portal_id)
    WHERE unsubscribed_at IS NULL;

CREATE TABLE public.feedback_contributor_preferences (
    portal_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    email_unsubscribed_at timestamptz,
    last_seen_update_published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_contributor_preferences_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_preferences_contributor_fkey
        FOREIGN KEY (portal_id, contributor_id)
        REFERENCES public.feedback_contributors(portal_id, id)
        ON DELETE CASCADE,
    PRIMARY KEY (portal_id, contributor_id)
);

ALTER TABLE public.feedback_updates
    ADD COLUMN slug text,
    ADD COLUMN summary text,
    ADD COLUMN cover_image_url text,
    ADD COLUMN published_by_user_id uuid;

UPDATE public.feedback_updates
SET slug = COALESCE(
        NULLIF(
            btrim(
                regexp_replace(lower(title), '[^a-z0-9]+', '-', 'g'),
                '-'
            ),
            ''
        ),
        'update'
    ),
    published_at = CASE
        WHEN status = 'published' THEN COALESCE(published_at, updated_at, created_at)
        ELSE published_at
    END,
    published_by_user_id = CASE
        WHEN status = 'published' THEN author_id
        ELSE NULL
    END;

UPDATE public.feedback_updates
SET slug = btrim(left(slug, 200), '-') || '-' || replace(id::text, '-', '');

ALTER TABLE public.feedback_updates
    ALTER COLUMN slug SET NOT NULL,
    ADD CONSTRAINT feedback_updates_slug_check
        CHECK (slug = btrim(slug) AND length(slug) BETWEEN 1 AND 255),
    ADD CONSTRAINT feedback_updates_summary_check
        CHECK (summary IS NULL OR length(btrim(summary)) BETWEEN 1 AND 1000),
    ADD CONSTRAINT feedback_updates_cover_image_url_check
        CHECK (cover_image_url IS NULL OR length(btrim(cover_image_url)) BETWEEN 1 AND 2048),
    ADD CONSTRAINT feedback_updates_published_by_user_id_fkey
        FOREIGN KEY (published_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    ADD CONSTRAINT feedback_updates_publication_check
        CHECK (status <> 'published' OR published_at IS NOT NULL);

CREATE UNIQUE INDEX feedback_updates_portal_slug_unique
    ON public.feedback_updates (portal_id, slug);

CREATE INDEX idx_feedback_updates_public_portal_published
    ON public.feedback_updates (portal_id, published_at DESC, id DESC)
    WHERE status = 'published' AND published_at IS NOT NULL;

CREATE TABLE public.feedback_contributor_unsubscribe_tokens (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    portal_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    item_id uuid,
    delivery_id uuid,
    purpose text NOT NULL,
    token_hash bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_contributor_unsubscribe_tokens_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_unsubscribe_tokens_contributor_fkey
        FOREIGN KEY (portal_id, contributor_id)
        REFERENCES public.feedback_contributors(portal_id, id)
        ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_unsubscribe_tokens_item_id_fkey
        FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_unsubscribe_tokens_purpose_check
        CHECK (purpose IN ('unsubscribe_item', 'unsubscribe_portal', 'all_email', 'manage_preferences')),
    CONSTRAINT feedback_contributor_unsubscribe_tokens_item_scope_check
        CHECK (
            (purpose = 'unsubscribe_item' AND item_id IS NOT NULL)
            OR (purpose IN ('unsubscribe_portal', 'all_email', 'manage_preferences') AND item_id IS NULL)
        ),
    CONSTRAINT feedback_contributor_unsubscribe_tokens_token_hash_check
        CHECK (octet_length(token_hash) = 32),
    CONSTRAINT feedback_contributor_unsubscribe_tokens_expiry_check
        CHECK (expires_at > created_at),
    CONSTRAINT feedback_contributor_unsubscribe_tokens_consumed_check
        CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_contributor_unsubscribe_tokens_token_hash_key
    ON public.feedback_contributor_unsubscribe_tokens (token_hash);

CREATE INDEX idx_feedback_contributor_unsubscribe_tokens_expiry
    ON public.feedback_contributor_unsubscribe_tokens (expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE public.feedback_contributor_deliveries (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    portal_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    item_id uuid,
    update_id uuid,
    event_type text NOT NULL,
    dedupe_key text NOT NULL,
    subject text NOT NULL,
    message text NOT NULL,
    destination_url text NOT NULL,
    channel text NOT NULL DEFAULT 'email',
    recipient_email text NOT NULL,
    event_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL DEFAULT 'queued',
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz DEFAULT now(),
    last_attempt_at timestamptz,
    sent_at timestamptz,
    provider_message_id text,
    final_failure_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_contributor_deliveries_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_deliveries_contributor_fkey
        FOREIGN KEY (portal_id, contributor_id)
        REFERENCES public.feedback_contributors(portal_id, id)
        ON DELETE CASCADE,
    CONSTRAINT feedback_contributor_deliveries_item_id_fkey
        FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE SET NULL,
    CONSTRAINT feedback_contributor_deliveries_update_id_fkey
        FOREIGN KEY (update_id) REFERENCES public.feedback_updates(id) ON DELETE SET NULL,
    CONSTRAINT feedback_contributor_deliveries_event_type_check
        CHECK (length(btrim(event_type)) BETWEEN 1 AND 200),
    CONSTRAINT feedback_contributor_deliveries_dedupe_key_check
        CHECK (length(btrim(dedupe_key)) BETWEEN 1 AND 512),
    CONSTRAINT feedback_contributor_deliveries_subject_check
        CHECK (length(btrim(subject)) BETWEEN 1 AND 500),
    CONSTRAINT feedback_contributor_deliveries_message_check
        CHECK (length(btrim(message)) BETWEEN 1 AND 100000),
    CONSTRAINT feedback_contributor_deliveries_destination_url_check
        CHECK (length(btrim(destination_url)) BETWEEN 1 AND 2048),
    CONSTRAINT feedback_contributor_deliveries_channel_check
        CHECK (channel = 'email'),
    CONSTRAINT feedback_contributor_deliveries_recipient_email_check
        CHECK (
            recipient_email = lower(btrim(recipient_email))
            AND length(recipient_email) BETWEEN 3 AND 320
            AND recipient_email !~ '[[:space:]]'
        ),
    CONSTRAINT feedback_contributor_deliveries_status_check
        CHECK (status IN ('queued', 'processing', 'retrying', 'sent', 'failed', 'suppressed')),
    CONSTRAINT feedback_contributor_deliveries_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT feedback_contributor_deliveries_retry_check
        CHECK (status NOT IN ('queued', 'retrying') OR next_attempt_at IS NOT NULL),
    CONSTRAINT feedback_contributor_deliveries_sent_check
        CHECK ((status = 'sent') = (sent_at IS NOT NULL)),
    CONSTRAINT feedback_contributor_deliveries_failure_check
        CHECK (
            status <> 'failed'
            OR (
                final_failure_reason IS NOT NULL
                AND length(btrim(final_failure_reason)) > 0
            )
        ),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_contributor_deliveries_event_dedupe_key
    ON public.feedback_contributor_deliveries (portal_id, contributor_id, channel, dedupe_key);

CREATE INDEX idx_feedback_contributor_deliveries_ready
    ON public.feedback_contributor_deliveries (status, next_attempt_at, created_at, id)
    WHERE status IN ('queued', 'retrying');

CREATE INDEX idx_feedback_contributor_deliveries_contributor_created
    ON public.feedback_contributor_deliveries (portal_id, contributor_id, created_at DESC, id DESC);

ALTER TABLE public.feedback_contributor_unsubscribe_tokens
    ADD CONSTRAINT feedback_contributor_unsubscribe_tokens_delivery_id_fkey
        FOREIGN KEY (delivery_id) REFERENCES public.feedback_contributor_deliveries(id) ON DELETE CASCADE,
    ADD CONSTRAINT feedback_contributor_unsubscribe_tokens_delivery_scope_check
        CHECK (
            (purpose = 'all_email' AND delivery_id IS NOT NULL)
            OR (purpose <> 'all_email' AND delivery_id IS NULL)
        );

CREATE UNIQUE INDEX feedback_contributor_unsubscribe_tokens_delivery_unique
    ON public.feedback_contributor_unsubscribe_tokens (delivery_id)
    WHERE delivery_id IS NOT NULL;

-- PostgreSQL CHECK constraints cannot inspect each array element directly. A
-- small immutable predicate keeps origin validation in one database contract.
CREATE FUNCTION public.feedback_allowed_origins_are_valid(origins text[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
    SELECT
        cardinality(origins) <= 100
        AND cardinality(origins) = (
            SELECT count(DISTINCT value)
            FROM unnest(origins) AS origin(value)
        )
        AND NOT EXISTS (
            SELECT 1
            FROM unnest(origins) AS origin(value)
            WHERE value IS NULL
                OR value <> btrim(value)
                OR value LIKE '%*%'
                OR NOT (
                    value ~ '^https://[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(:[0-9]{1,5})?$'
                    OR value ~ '^http://(localhost|127[.]0[.]0[.]1|[[]::1[]])(:[0-9]{1,5})?$'
                )
        );
$$;

CREATE TABLE public.feedback_widget_settings (
    portal_id uuid NOT NULL,
    enabled boolean NOT NULL DEFAULT false,
    widget_key_id uuid NOT NULL DEFAULT gen_random_uuid(),
    allowed_origins text[] NOT NULL DEFAULT '{}',
    signing_secret_encrypted text,
    signing_secret_version integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_widget_settings_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_widget_settings_allowed_origins_check
        CHECK (public.feedback_allowed_origins_are_valid(allowed_origins)),
    CONSTRAINT feedback_widget_settings_secret_check
        CHECK (
            (signing_secret_encrypted IS NULL AND signing_secret_version = 0)
            OR (
                signing_secret_encrypted IS NOT NULL
                AND
                length(btrim(signing_secret_encrypted)) > 0
                AND signing_secret_version > 0
            )
        ),
    CONSTRAINT feedback_widget_settings_enabled_check
        CHECK (
            NOT enabled
            OR (
                cardinality(allowed_origins) > 0
                AND signing_secret_encrypted IS NOT NULL
                AND signing_secret_version > 0
            )
        ),
    PRIMARY KEY (portal_id),
    CONSTRAINT feedback_widget_settings_widget_key_id_key UNIQUE (widget_key_id),
    CONSTRAINT feedback_widget_settings_portal_widget_key_key UNIQUE (portal_id, widget_key_id)
);

CREATE TABLE public.feedback_widget_signing_secret_rotations (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    portal_id uuid NOT NULL,
    signing_secret_version integer NOT NULL,
    signing_secret_encrypted text NOT NULL,
    activated_at timestamptz NOT NULL,
    grace_expires_at timestamptz NOT NULL,
    retired_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_widget_signing_secret_rotations_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_widget_settings(portal_id) ON DELETE CASCADE,
    CONSTRAINT feedback_widget_signing_secret_rotations_version_check
        CHECK (signing_secret_version > 0),
    CONSTRAINT feedback_widget_signing_secret_rotations_secret_check
        CHECK (length(btrim(signing_secret_encrypted)) > 0),
    CONSTRAINT feedback_widget_signing_secret_rotations_grace_check
        CHECK (grace_expires_at > activated_at),
    CONSTRAINT feedback_widget_signing_secret_rotations_retired_check
        CHECK (retired_at IS NULL OR retired_at >= activated_at),
    PRIMARY KEY (id),
    CONSTRAINT feedback_widget_signing_secret_rotations_portal_version_key
        UNIQUE (portal_id, signing_secret_version)
);

CREATE INDEX idx_feedback_widget_signing_secret_rotations_grace
    ON public.feedback_widget_signing_secret_rotations (grace_expires_at)
    WHERE retired_at IS NULL;

CREATE TABLE public.feedback_widget_assertion_nonces (
    portal_id uuid NOT NULL,
    widget_key_id uuid NOT NULL,
    signing_secret_version integer NOT NULL,
    nonce_hash bytea NOT NULL,
    parent_origin text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_widget_assertion_nonces_widget_fkey
        FOREIGN KEY (portal_id, widget_key_id)
        REFERENCES public.feedback_widget_settings(portal_id, widget_key_id)
        ON DELETE CASCADE,
    CONSTRAINT feedback_widget_assertion_nonces_version_check
        CHECK (signing_secret_version > 0),
    CONSTRAINT feedback_widget_assertion_nonces_nonce_hash_check
        CHECK (octet_length(nonce_hash) = 32),
    CONSTRAINT feedback_widget_assertion_nonces_parent_origin_check
        CHECK (public.feedback_allowed_origins_are_valid(ARRAY[parent_origin])),
    CONSTRAINT feedback_widget_assertion_nonces_expiry_check
        CHECK (expires_at > created_at),
    PRIMARY KEY (portal_id, nonce_hash)
);

CREATE INDEX idx_feedback_widget_assertion_nonces_expiry
    ON public.feedback_widget_assertion_nonces (expires_at);
