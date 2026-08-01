CREATE TABLE public.workspace_strategies (
    workspace_id uuid PRIMARY KEY REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    ultimate_goal varchar(255) NOT NULL DEFAULT '',
    description text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.strategic_pillars (
    pillar_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    order_index integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX strategic_pillars_workspace_name_unique
    ON public.strategic_pillars (workspace_id, name);
CREATE INDEX strategic_pillars_workspace_order
    ON public.strategic_pillars (workspace_id, order_index, created_at);

CREATE TABLE public.strategy_objective_alignments (
    objective_id uuid PRIMARY KEY REFERENCES public.objectives(objective_id) ON DELETE CASCADE,
    pillar_id uuid NOT NULL REFERENCES public.strategic_pillars(pillar_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX strategy_objective_alignments_pillar
    ON public.strategy_objective_alignments (pillar_id);
