ALTER TABLE public.objectives
    ADD COLUMN sequence_id int4;

WITH ranked_objectives AS (
    SELECT
        objective_id,
        ROW_NUMBER() OVER (
            PARTITION BY team_id
            ORDER BY created_at ASC, objective_id ASC
        )::int4 AS sequence_id
    FROM public.objectives
)
UPDATE public.objectives o
SET sequence_id = ranked_objectives.sequence_id
FROM ranked_objectives
WHERE ranked_objectives.objective_id = o.objective_id;

ALTER TABLE public.objectives
    ALTER COLUMN sequence_id SET NOT NULL,
    ADD CONSTRAINT objectives_team_sequence_unique
        UNIQUE (team_id, sequence_id);

CREATE TABLE public.team_objective_sequences (
    workspace_id uuid NOT NULL,
    team_id uuid NOT NULL,
    current_sequence int4 NOT NULL DEFAULT 0,
    CONSTRAINT team_objective_sequences_workspace_id_fkey
        FOREIGN KEY (workspace_id)
        REFERENCES public.workspaces(workspace_id)
        ON DELETE CASCADE,
    CONSTRAINT team_objective_sequences_team_id_fkey
        FOREIGN KEY (team_id)
        REFERENCES public.teams(team_id)
        ON DELETE CASCADE,
    PRIMARY KEY (workspace_id, team_id)
);

INSERT INTO public.team_objective_sequences (
    workspace_id,
    team_id,
    current_sequence
)
SELECT
    workspace_id,
    team_id,
    MAX(sequence_id)
FROM public.objectives
GROUP BY workspace_id, team_id;

CREATE INDEX idx_team_objective_sequences_team_id
    ON public.team_objective_sequences USING btree (team_id);
