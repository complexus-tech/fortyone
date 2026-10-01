CREATE TABLE public.team_automations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
 team_id uuid NOT NULL REFERENCES public.teams(team_id) ON DELETE CASCADE,
 owner_id uuid NOT NULL REFERENCES public.users(user_id),
 kind text NOT NULL CHECK (kind IN ('rule', 'recurrence')),
 name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 100),
 configuration jsonb NOT NULL CHECK (jsonb_typeof(configuration) = 'object' AND octet_length(CAST(configuration AS text)) <= 65536),
 paused boolean NOT NULL DEFAULT FALSE,
 next_run_at timestamptz,
 last_run_at timestamptz,
 last_error text NOT NULL DEFAULT '',
 event_at timestamptz NOT NULL DEFAULT now(),
 event_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
 lease_token uuid,
 lease_until timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 archived_at timestamptz
);
CREATE INDEX team_automations_active_team_idx ON public.team_automations(workspace_id, team_id, created_at DESC, id DESC) WHERE archived_at IS NULL;
CREATE INDEX team_automations_due_idx ON public.team_automations(next_run_at, id) WHERE archived_at IS NULL AND paused = FALSE;
CREATE TABLE public.team_automation_runs (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 automation_id uuid NOT NULL REFERENCES public.team_automations(id) ON DELETE CASCADE,
 occurrence text NOT NULL CHECK (char_length(occurrence) BETWEEN 1 AND 120),
 status text NOT NULL CHECK (status IN ('running', 'succeeded', 'skipped', 'failed')),
 story_id uuid REFERENCES public.stories(id) ON DELETE SET NULL,
 error text NOT NULL DEFAULT '' CHECK (char_length(error) <= 1000),
 lease_token uuid NOT NULL,
 attempt_count integer NOT NULL DEFAULT 1 CHECK (attempt_count BETWEEN 1 AND 3),
 started_at timestamptz NOT NULL DEFAULT now(),
 finished_at timestamptz,
 UNIQUE (automation_id, occurrence)
);
CREATE INDEX team_automation_runs_recent_idx ON public.team_automation_runs(automation_id, started_at DESC, id DESC);
