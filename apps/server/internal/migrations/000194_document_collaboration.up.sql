ALTER TABLE public.documents
    ADD COLUMN revision bigint NOT NULL DEFAULT 1,
    ADD COLUMN collaboration_epoch bigint NOT NULL DEFAULT 1,
    ADD COLUMN collaboration_state bytea,
    ADD COLUMN public_token text UNIQUE;

CREATE TABLE public.document_revisions (
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    revision bigint NOT NULL,
    title text NOT NULL,
    content_html text NOT NULL,
    content_text text NOT NULL,
    edited_by uuid REFERENCES public.users(user_id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (document_id, revision)
);

INSERT INTO public.document_revisions
    (document_id, revision, title, content_html, content_text, edited_by, created_at)
SELECT document_id, revision, title, content_html, content_text, updated_by, updated_at
FROM public.documents;

CREATE TABLE public.document_collaboration_sessions (
    token_hash text PRIMARY KEY,
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    session_version bigint NOT NULL,
    epoch bigint NOT NULL,
    expires_at timestamptz NOT NULL
);
CREATE INDEX document_collaboration_sessions_expiry ON public.document_collaboration_sessions(expires_at);

-- All writers, including imports and duplicates, participate in history.
CREATE FUNCTION public.document_revision_before_write() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND (
        NEW.title IS DISTINCT FROM OLD.title OR
        NEW.content_html IS DISTINCT FROM OLD.content_html OR
        NEW.content_text IS DISTINCT FROM OLD.content_text OR
        NEW.collaboration_state IS DISTINCT FROM OLD.collaboration_state OR
        NEW.collaboration_epoch IS DISTINCT FROM OLD.collaboration_epoch
    ) THEN
        -- Once activated, the CRDT is authoritative. Fence old REST writers.
        IF OLD.collaboration_state IS NOT NULL
           AND NEW.collaboration_state IS NOT DISTINCT FROM OLD.collaboration_state
           AND NEW.collaboration_epoch = OLD.collaboration_epoch THEN
            RAISE EXCEPTION 'document requires collaborative editing' USING ERRCODE = '40001';
        END IF;
        NEW.revision := OLD.revision + 1;
        NEW.updated_at := CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.document_revision_after_write() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' OR NEW.revision <> OLD.revision THEN
        INSERT INTO public.document_revisions
            (document_id, revision, title, content_html, content_text, edited_by)
        VALUES (NEW.document_id, NEW.revision, NEW.title, NEW.content_html,
                NEW.content_text, NEW.updated_by);
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER document_revision_before_write BEFORE UPDATE ON public.documents
FOR EACH ROW EXECUTE FUNCTION public.document_revision_before_write();
CREATE TRIGGER document_revision_after_write AFTER INSERT OR UPDATE ON public.documents
FOR EACH ROW EXECUTE FUNCTION public.document_revision_after_write();
