-- 000012_objective_statuses.up.sql
CREATE TABLE public.objective_statuses (
    status_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    category character varying(255),
    order_index integer,
    workspace_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    is_default boolean DEFAULT false NOT NULL,
    color character varying(16) DEFAULT '#6b665c'::character varying
);

ALTER TABLE ONLY public.objective_statuses ADD CONSTRAINT objective_statuses_pkey PRIMARY KEY (status_id);

ALTER TABLE ONLY public.objective_statuses 
    ADD CONSTRAINT objective_statuses_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_objective_statuses_status_id ON public.objective_statuses USING btree (status_id);
CREATE INDEX idx_objective_statuses_workspace_name ON public.objective_statuses USING btree (workspace_id, name);
CREATE INDEX idx_objective_statuses_workspace_order ON public.objective_statuses USING btree (workspace_id, order_index);
CREATE UNIQUE INDEX unique_default_status_per_workspace ON public.objective_statuses USING btree (workspace_id) WHERE (is_default = true);
