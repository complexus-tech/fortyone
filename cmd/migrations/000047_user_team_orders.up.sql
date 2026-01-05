-- 000047_user_team_orders.up.sql
CREATE TABLE public.user_team_orders (
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    team_ids uuid[] DEFAULT '{}'::uuid[] NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.user_team_orders ADD CONSTRAINT user_team_orders_pkey PRIMARY KEY (user_id, workspace_id);

ALTER TABLE ONLY public.user_team_orders 
    ADD CONSTRAINT user_team_orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_team_orders 
    ADD CONSTRAINT user_team_orders_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;
