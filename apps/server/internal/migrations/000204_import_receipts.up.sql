CREATE TABLE public.story_import_receipts (
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    team_id uuid NOT NULL REFERENCES public.teams(team_id) ON DELETE CASCADE,
    creation_key text NOT NULL,
    provider text NOT NULL CHECK (provider IN ('jira_csv', 'file')),
    source_digest text NOT NULL CHECK (length(source_digest) = 64),
    source_namespace text,
    source_key text NOT NULL,
    story_id uuid REFERENCES public.stories(id) ON DELETE SET NULL,
    created boolean NOT NULL DEFAULT false,
    error_code text,
    error_message text,
    source_metadata jsonb NOT NULL DEFAULT '{}',
    attempts integer NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, creation_key)
);
CREATE INDEX story_import_receipts_source_idx ON public.story_import_receipts(workspace_id, source_digest, source_key);
CREATE INDEX story_import_receipts_namespace_idx ON public.story_import_receipts(workspace_id, provider, source_namespace, source_key) WHERE source_namespace IS NOT NULL;
