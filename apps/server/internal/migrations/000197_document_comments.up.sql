CREATE TABLE public.document_comment_threads (
    thread_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    quote text NOT NULL DEFAULT '',
    anchor_start integer NOT NULL DEFAULT 0 CHECK (anchor_start >= 0),
    anchor_end integer NOT NULL DEFAULT 0 CHECK (anchor_end >= anchor_start),
    created_by uuid NOT NULL REFERENCES public.users(user_id),
    resolved_at timestamptz,
    resolved_by uuid REFERENCES public.users(user_id),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX document_comment_threads_document
    ON public.document_comment_threads (document_id, resolved_at, created_at DESC);

CREATE TABLE public.document_comments (
    comment_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id uuid NOT NULL REFERENCES public.document_comment_threads(thread_id) ON DELETE CASCADE,
    body text NOT NULL CHECK (length(btrim(body)) > 0),
    created_by uuid NOT NULL REFERENCES public.users(user_id),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX document_comments_thread
    ON public.document_comments (thread_id, created_at, comment_id);
