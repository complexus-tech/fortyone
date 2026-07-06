ALTER TABLE public.integration_requests
    ADD COLUMN estimate_unit int2,
    ADD COLUMN objective_id uuid,
    ADD COLUMN key_result_id uuid,
    ADD COLUMN sprint_id uuid,
    ADD COLUMN start_date date,
    ADD COLUMN end_date date,
    ADD CONSTRAINT integration_requests_estimate_unit_check
        CHECK (estimate_unit IS NULL OR estimate_unit IN (1, 2, 3, 5, 8)),
    ADD CONSTRAINT integration_requests_objective_id_fkey
        FOREIGN KEY (objective_id) REFERENCES public.objectives(objective_id) ON DELETE SET NULL,
    ADD CONSTRAINT integration_requests_key_result_id_fkey
        FOREIGN KEY (key_result_id) REFERENCES public.key_results(id) ON DELETE SET NULL,
    ADD CONSTRAINT integration_requests_sprint_id_fkey
        FOREIGN KEY (sprint_id) REFERENCES public.sprints(sprint_id) ON DELETE SET NULL;

CREATE INDEX idx_integration_requests_objective_id
    ON public.integration_requests USING btree (objective_id);

CREATE INDEX idx_integration_requests_sprint_id
    ON public.integration_requests USING btree (sprint_id);
