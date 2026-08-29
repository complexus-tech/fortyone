ALTER TABLE public.key_results
    ADD COLUMN team_id uuid,
    ADD COLUMN sequence_id int4;

UPDATE public.key_results kr
SET team_id = o.team_id
FROM public.objectives o
WHERE o.objective_id = kr.objective_id;

WITH ranked_key_results AS (
    SELECT
        kr.id,
        ROW_NUMBER() OVER (
            PARTITION BY kr.team_id
            ORDER BY kr.created_at ASC, kr.id ASC
        )::int4 AS sequence_id
    FROM public.key_results kr
)
UPDATE public.key_results kr
SET sequence_id = ranked_key_results.sequence_id
FROM ranked_key_results
WHERE ranked_key_results.id = kr.id;

ALTER TABLE public.key_results
    ALTER COLUMN team_id SET NOT NULL,
    ALTER COLUMN sequence_id SET NOT NULL,
    ADD CONSTRAINT key_results_team_id_fkey
        FOREIGN KEY (team_id)
        REFERENCES public.teams(team_id)
        ON DELETE CASCADE,
    ADD CONSTRAINT key_results_team_sequence_unique
        UNIQUE (team_id, sequence_id);

CREATE TABLE public.team_key_result_sequences (
    workspace_id uuid NOT NULL,
    team_id uuid NOT NULL,
    current_sequence int4 NOT NULL DEFAULT 0,
    CONSTRAINT team_key_result_sequences_workspace_id_fkey
        FOREIGN KEY (workspace_id)
        REFERENCES public.workspaces(workspace_id)
        ON DELETE CASCADE,
    CONSTRAINT team_key_result_sequences_team_id_fkey
        FOREIGN KEY (team_id)
        REFERENCES public.teams(team_id)
        ON DELETE CASCADE,
    PRIMARY KEY (workspace_id, team_id)
);

INSERT INTO public.team_key_result_sequences (
    workspace_id,
    team_id,
    current_sequence
)
SELECT
    t.workspace_id,
    kr.team_id,
    MAX(kr.sequence_id)
FROM public.key_results kr
INNER JOIN public.teams t ON t.team_id = kr.team_id
GROUP BY t.workspace_id, kr.team_id;

CREATE INDEX idx_key_results_team_id
    ON public.key_results USING btree (team_id);

CREATE INDEX idx_team_key_result_sequences_team_id
    ON public.team_key_result_sequences USING btree (team_id);
