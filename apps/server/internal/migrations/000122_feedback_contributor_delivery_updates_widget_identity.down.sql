DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.feedback_contributors
        WHERE kind IN ('verified_guest', 'external')
    ) THEN
        RAISE EXCEPTION 'migration 000122 cannot be rolled back while verified guest or external contributors exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.feedback_votes
        WHERE user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'migration 000122 cannot be rolled back while contributor-only votes exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.feedback_comments AS comment
        INNER JOIN public.feedback_contributors AS contributor
            ON contributor.id = comment.contributor_id
        WHERE comment.author_id IS NULL
            AND contributor.kind <> 'anonymous'
    ) THEN
        RAISE EXCEPTION 'migration 000122 cannot be rolled back while contributor-only comments exist';
    END IF;
END
$$;

DROP TABLE IF EXISTS public.feedback_widget_assertion_nonces;
DROP TABLE IF EXISTS public.feedback_widget_signing_secret_rotations;
DROP TABLE IF EXISTS public.feedback_widget_settings;
DROP FUNCTION IF EXISTS public.feedback_allowed_origins_are_valid(text[]);

DROP TABLE IF EXISTS public.feedback_contributor_unsubscribe_tokens;
DROP TABLE IF EXISTS public.feedback_contributor_deliveries;

DROP INDEX IF EXISTS public.idx_feedback_updates_public_portal_published;
DROP INDEX IF EXISTS public.feedback_updates_portal_slug_unique;

ALTER TABLE public.feedback_updates
    DROP CONSTRAINT IF EXISTS feedback_updates_publication_check,
    DROP CONSTRAINT IF EXISTS feedback_updates_published_by_user_id_fkey,
    DROP CONSTRAINT IF EXISTS feedback_updates_cover_image_url_check,
    DROP CONSTRAINT IF EXISTS feedback_updates_summary_check,
    DROP CONSTRAINT IF EXISTS feedback_updates_slug_check,
    DROP COLUMN IF EXISTS published_by_user_id,
    DROP COLUMN IF EXISTS cover_image_url,
    DROP COLUMN IF EXISTS summary,
    DROP COLUMN IF EXISTS slug;

DROP TABLE IF EXISTS public.feedback_contributor_preferences;
DROP TABLE IF EXISTS public.feedback_portal_followers;
DROP TABLE IF EXISTS public.feedback_item_followers;

DROP INDEX IF EXISTS public.idx_feedback_votes_contributor_created;
DROP INDEX IF EXISTS public.feedback_votes_item_user_unique;

ALTER TABLE public.feedback_votes
    DROP CONSTRAINT IF EXISTS feedback_votes_pkey,
    DROP CONSTRAINT IF EXISTS feedback_votes_contributor_id_fkey,
    DROP CONSTRAINT IF EXISTS feedback_votes_user_id_fkey;

ALTER TABLE public.feedback_votes
    ALTER COLUMN user_id SET NOT NULL,
    ADD CONSTRAINT feedback_votes_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    ADD CONSTRAINT feedback_votes_pkey PRIMARY KEY (item_id, user_id),
    DROP COLUMN IF EXISTS contributor_id;

DROP INDEX IF EXISTS public.idx_feedback_comments_contributor_created;

ALTER TABLE public.feedback_comments
    DROP CONSTRAINT IF EXISTS feedback_comments_contributor_id_fkey,
    DROP COLUMN IF EXISTS contributor_id;

DROP TABLE IF EXISTS public.feedback_contributor_sessions;
DROP TABLE IF EXISTS public.feedback_contributor_verifications;

DELETE FROM public.feedback_contributors AS contributor
WHERE contributor.kind = 'anonymous'
    AND NOT EXISTS (
        SELECT 1
        FROM public.feedback_items AS item
        WHERE item.contributor_id = contributor.id
    );

DROP INDEX IF EXISTS public.idx_feedback_contributors_portal_kind_seen;
DROP INDEX IF EXISTS public.feedback_contributors_portal_external_id_unique;
DROP INDEX IF EXISTS public.feedback_contributors_portal_email_unique;

ALTER TABLE public.feedback_contributors
    DROP CONSTRAINT IF EXISTS feedback_contributors_identity_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_public_masked_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_block_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_external_id_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_avatar_url_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_display_name_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_email_check,
    DROP CONSTRAINT IF EXISTS feedback_contributors_kind_check,
    DROP COLUMN IF EXISTS blocked_reason,
    DROP COLUMN IF EXISTS blocked_at,
    DROP COLUMN IF EXISTS last_seen_at,
    DROP COLUMN IF EXISTS external_id,
    DROP COLUMN IF EXISTS public_masked,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS email,
    ADD CONSTRAINT feedback_contributors_kind_check
        CHECK (kind IN ('account', 'anonymous')),
    ADD CONSTRAINT feedback_contributors_identity_check
        CHECK (kind <> 'anonymous' OR user_id IS NULL);

UPDATE public.feedback_portals
SET participation_mode = 'account_required'
WHERE participation_mode = 'verified_guest';

ALTER TABLE public.feedback_portals
    DROP CONSTRAINT IF EXISTS feedback_portals_guest_identity_policy_check,
    DROP CONSTRAINT IF EXISTS feedback_portals_participation_mode_check,
    DROP COLUMN IF EXISTS guest_identity_policy,
    ADD CONSTRAINT feedback_portals_participation_mode_check
        CHECK (participation_mode IN ('account_required', 'anonymous_allowed'));
