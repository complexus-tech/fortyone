CREATE TABLE public.documents (
    document_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    title varchar(255) NOT NULL DEFAULT 'Untitled document',
    content_html text NOT NULL DEFAULT '',
    content_text text NOT NULL DEFAULT '',
    visibility varchar(20) NOT NULL DEFAULT 'workspace'
        CHECK (visibility IN ('workspace', 'restricted', 'private')),
    created_by uuid NOT NULL REFERENCES public.users(user_id),
    updated_by uuid NOT NULL REFERENCES public.users(user_id),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at timestamptz
);

CREATE INDEX documents_workspace_updated
    ON public.documents (workspace_id, updated_at DESC)
    WHERE archived_at IS NULL;

CREATE INDEX documents_created_by
    ON public.documents (workspace_id, created_by, updated_at DESC)
    WHERE archived_at IS NULL;

CREATE INDEX documents_search
    ON public.documents
    USING gin (to_tsvector('english', title || ' ' || content_text))
    WHERE archived_at IS NULL;

CREATE TABLE public.document_members (
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    role varchar(20) NOT NULL DEFAULT 'editor'
        CHECK (role IN ('viewer', 'editor')),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (document_id, user_id)
);

CREATE INDEX document_members_user
    ON public.document_members (user_id, document_id);

CREATE TABLE public.document_relationships (
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    entity_type varchar(20) NOT NULL
        CHECK (entity_type IN ('story', 'objective')),
    entity_id uuid NOT NULL,
    created_by uuid NOT NULL REFERENCES public.users(user_id),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (document_id, entity_type, entity_id)
);

CREATE INDEX document_relationships_entity
    ON public.document_relationships (workspace_id, entity_type, entity_id);
