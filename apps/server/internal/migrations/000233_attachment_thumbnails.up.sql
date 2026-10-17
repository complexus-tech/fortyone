CREATE TABLE public.attachment_thumbnails (
    thumbnail_id uuid PRIMARY KEY,
    attachment_id uuid NOT NULL REFERENCES public.attachments(attachment_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    source_version varchar(64) NOT NULL,
    storage_provider varchar(32) NOT NULL,
    container_name varchar(255) NOT NULL,
    blob_name varchar(1024) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
    lease_expires_at timestamptz,
    retry_after timestamptz,
    byte_size integer,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (attachment_id, source_version),
    CHECK (byte_size IS NULL OR byte_size BETWEEN 1 AND 2097152)
);

CREATE INDEX attachment_thumbnails_workspace ON public.attachment_thumbnails(workspace_id, attachment_id);

-- The deletion outbox has no attachment FK. For a derived physical object its
-- identity is the thumbnail UUID, allowing independent durable deletion of the
-- original and each derivative without changing existing original-file jobs.
CREATE FUNCTION public.retire_attachment_thumbnail_object() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO public.attachment_object_deletion_outbox
        (attachment_id, workspace_id, storage_provider, container_name, blob_name)
    VALUES (OLD.thumbnail_id, OLD.workspace_id, OLD.storage_provider, OLD.container_name, OLD.blob_name)
    ON CONFLICT (attachment_id) DO NOTHING;
    RETURN OLD;
END;
$$;

CREATE TRIGGER attachment_thumbnail_retirement
AFTER DELETE ON public.attachment_thumbnails
FOR EACH ROW EXECUTE FUNCTION public.retire_attachment_thumbnail_object();
