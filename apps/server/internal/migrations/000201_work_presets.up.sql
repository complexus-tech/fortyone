CREATE TABLE public.work_presets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    team_id uuid NOT NULL REFERENCES public.teams(team_id) ON DELETE CASCADE,
    owner_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('view', 'template')),
    visibility text NOT NULL CHECK (visibility IN ('personal', 'team')),
    name text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 100),
    configuration jsonb NOT NULL CHECK (jsonb_typeof(configuration) = 'object' AND octet_length(configuration::text) <= 65536),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    archived_at timestamptz
);

CREATE INDEX work_presets_active_team ON public.work_presets (workspace_id, team_id, kind, created_at DESC, id DESC) WHERE archived_at IS NULL;
CREATE INDEX work_presets_owner ON public.work_presets (owner_id) WHERE visibility = 'personal' AND archived_at IS NULL;
