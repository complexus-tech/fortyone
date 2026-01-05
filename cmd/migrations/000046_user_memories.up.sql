-- 000046_user_memories.up.sql
CREATE TABLE public.user_memories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    content text NOT NULL,
    importance integer DEFAULT 1,
    category character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.user_memories ADD CONSTRAINT user_memories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_memories 
    ADD CONSTRAINT user_memories_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_memories 
    ADD CONSTRAINT user_memories_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_user_memories_user_workspace ON public.user_memories USING btree (user_id, workspace_id);
