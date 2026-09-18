-- Autosave revisions remain monotonic for conflict detection. History entries
-- are bounded editing checkpoints, not a log of every autosave/CRDT update.
ALTER TABLE public.document_revisions
    ADD COLUMN started_at timestamptz,
    ADD COLUMN checkpoint_kind text NOT NULL DEFAULT 'edit'
        CHECK (checkpoint_kind IN ('initial', 'edit', 'restore'));
UPDATE public.document_revisions SET started_at = created_at;
ALTER TABLE public.document_revisions ALTER COLUMN started_at SET NOT NULL;
ALTER TABLE public.document_revisions ALTER COLUMN started_at SET DEFAULT CURRENT_TIMESTAMP;

UPDATE public.document_revisions AS revision SET checkpoint_kind = 'initial'
WHERE revision.revision = (SELECT min(first.revision) FROM public.document_revisions AS first WHERE first.document_id = revision.document_id);

-- Consolidate legacy per-save entries: retain the original and the latest
-- entry per editor / ten-minute bucket, then enforce the same ten-entry cap.
WITH ranked AS (
    SELECT document_id, revision, row_number() OVER (
        PARTITION BY document_id, edited_by, floor(extract(epoch FROM created_at) / 600)
        ORDER BY revision DESC
    ) AS position
    FROM public.document_revisions WHERE checkpoint_kind <> 'initial'
)
DELETE FROM public.document_revisions AS revision USING ranked
WHERE revision.document_id = ranked.document_id AND revision.revision = ranked.revision AND ranked.position > 1;
WITH ranked AS (
    SELECT document_id, revision, row_number() OVER (PARTITION BY document_id ORDER BY revision DESC) AS position
    FROM public.document_revisions
)
DELETE FROM public.document_revisions AS revision USING ranked
WHERE revision.document_id = ranked.document_id AND revision.revision = ranked.revision AND ranked.position > 10;

CREATE OR REPLACE FUNCTION public.document_revision_after_write() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    latest public.document_revisions%ROWTYPE;
    checkpoint_time timestamptz := clock_timestamp();
    is_restore boolean := FALSE;
BEGIN
    IF TG_OP = 'UPDATE' THEN
        is_restore := NEW.collaboration_epoch <> OLD.collaboration_epoch;
        -- CRDT bookkeeping, sharing and unchanged saves are not new versions.
        IF NOT is_restore AND NEW.title IS NOT DISTINCT FROM OLD.title
           AND NEW.content_html IS NOT DISTINCT FROM OLD.content_html
           AND NEW.content_text IS NOT DISTINCT FROM OLD.content_text THEN
            RETURN NEW;
        END IF;
    END IF;
    SELECT * INTO latest FROM public.document_revisions
    WHERE document_id = NEW.document_id ORDER BY revision DESC LIMIT 1;

    IF TG_OP = 'UPDATE' AND NOT is_restore
       AND latest.checkpoint_kind = 'edit'
       AND latest.edited_by IS NOT DISTINCT FROM NEW.updated_by
       AND checkpoint_time - latest.started_at < interval '10 minutes'
       AND checkpoint_time - latest.created_at < interval '2 minutes' THEN
        -- Keep the latest saved content in the current editing checkpoint.
        -- Changing its revision invalidates previews of superseded content.
        UPDATE public.document_revisions SET revision = NEW.revision,
            title = NEW.title, content_html = NEW.content_html, content_text = NEW.content_text,
            edited_by = NEW.updated_by, created_at = checkpoint_time
        WHERE document_id = NEW.document_id AND revision = latest.revision;
    ELSE
        INSERT INTO public.document_revisions
            (document_id, revision, title, content_html, content_text, edited_by, created_at, started_at, checkpoint_kind)
        VALUES (NEW.document_id, NEW.revision, NEW.title, NEW.content_html, NEW.content_text,
            NEW.updated_by, checkpoint_time, checkpoint_time,
            CASE WHEN TG_OP = 'INSERT' THEN 'initial' WHEN is_restore THEN 'restore' ELSE 'edit' END);
    END IF;
    DELETE FROM public.document_revisions WHERE document_id = NEW.document_id
      AND revision NOT IN (SELECT revision FROM public.document_revisions WHERE document_id = NEW.document_id ORDER BY revision DESC LIMIT 10);
    RETURN NEW;
END;
$$;
