-- 000014_key_results.up.sql
CREATE TABLE public.key_results (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    objective_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    measurement_type public.measurement_type NOT NULL,
    start_value numeric,
    target_value numeric,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    current_value numeric,
    created_by uuid,
    lead uuid,
    start_date date NOT NULL,
    end_date date NOT NULL,
    CONSTRAINT check_dates_valid CHECK ((end_date > start_date)),
    CONSTRAINT valid_boolean CHECK (((measurement_type <> 'boolean'::public.measurement_type) OR ((start_value = ANY (ARRAY[(0)::numeric, (1)::numeric])) AND (target_value = ANY (ARRAY[(0)::numeric, (1)::numeric]))))),
    CONSTRAINT valid_percentage CHECK (((measurement_type <> 'percentage'::public.measurement_type) OR ((start_value >= (0)::numeric) AND (start_value <= (100)::numeric) AND (target_value >= (0)::numeric) AND (target_value <= (100)::numeric))))
);

ALTER TABLE ONLY public.key_results ADD CONSTRAINT key_results_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.key_results 
    ADD CONSTRAINT key_results_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(user_id);

ALTER TABLE ONLY public.key_results 
    ADD CONSTRAINT key_results_lead_fkey FOREIGN KEY (lead) REFERENCES public.users(user_id);

ALTER TABLE ONLY public.key_results 
    ADD CONSTRAINT key_results_objective_id_fkey FOREIGN KEY (objective_id) REFERENCES public.objectives(objective_id) ON DELETE CASCADE;

CREATE INDEX idx_key_results_end_date ON public.key_results USING btree (end_date);
CREATE INDEX idx_key_results_lead ON public.key_results USING btree (lead);
CREATE INDEX idx_key_results_start_date ON public.key_results USING btree (start_date);
