ALTER TABLE public.feedback_portals
    ADD COLUMN participation_mode text NOT NULL DEFAULT 'account_required',
    ADD CONSTRAINT feedback_portals_participation_mode_check
        CHECK (participation_mode IN ('account_required', 'anonymous_allowed'));

CREATE TABLE public.feedback_contributors (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    portal_id uuid NOT NULL,
    user_id uuid,
    kind text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_contributors_portal_id_fkey
        FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_contributors_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT feedback_contributors_kind_check
        CHECK (kind IN ('account', 'anonymous')),
    CONSTRAINT feedback_contributors_identity_check
        CHECK (kind <> 'anonymous' OR user_id IS NULL),
    PRIMARY KEY (id),
    UNIQUE (portal_id, id)
);

CREATE UNIQUE INDEX feedback_contributors_portal_user_unique
    ON public.feedback_contributors (portal_id, user_id)
    WHERE user_id IS NOT NULL;

ALTER TABLE public.feedback_items
    ADD COLUMN contributor_id uuid;

INSERT INTO public.feedback_contributors (portal_id, user_id, kind)
SELECT DISTINCT portal_id, author_id, 'account'
FROM public.feedback_items
WHERE author_id IS NOT NULL
ON CONFLICT (portal_id, user_id) WHERE user_id IS NOT NULL DO NOTHING;

UPDATE public.feedback_items AS item
SET contributor_id = contributor.id
FROM public.feedback_contributors AS contributor
WHERE contributor.portal_id = item.portal_id
    AND contributor.user_id = item.author_id
    AND item.author_id IS NOT NULL;

-- Legacy data permits a null author after user deletion. Preserve those items
-- with one unlinkable anonymous contributor each instead of inventing identity.
CREATE TEMPORARY TABLE feedback_anonymous_item_backfill (
    item_id uuid PRIMARY KEY,
    contributor_id uuid NOT NULL UNIQUE,
    portal_id uuid NOT NULL
);

INSERT INTO feedback_anonymous_item_backfill (item_id, contributor_id, portal_id)
SELECT item.id, gen_random_uuid(), item.portal_id
FROM public.feedback_items AS item
WHERE item.contributor_id IS NULL;

INSERT INTO public.feedback_contributors (id, portal_id, kind)
SELECT contributor_id, portal_id, 'anonymous'
FROM feedback_anonymous_item_backfill;

UPDATE public.feedback_items AS item
SET contributor_id = mapping.contributor_id
FROM feedback_anonymous_item_backfill AS mapping
WHERE item.id = mapping.item_id;

DROP TABLE feedback_anonymous_item_backfill;

ALTER TABLE public.feedback_items
    ALTER COLUMN contributor_id SET NOT NULL,
    ADD CONSTRAINT feedback_items_contributor_fkey
        FOREIGN KEY (portal_id, contributor_id)
        REFERENCES public.feedback_contributors(portal_id, id)
        ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX idx_feedback_items_portal_contributor_created
    ON public.feedback_items (portal_id, contributor_id, created_at DESC, id DESC);
