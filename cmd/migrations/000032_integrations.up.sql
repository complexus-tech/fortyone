-- 000032_integrations.up.sql
CREATE TABLE public.integrations (
    integration_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    type character varying(255) NOT NULL,
    config jsonb,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.integrations ADD CONSTRAINT integrations_pkey PRIMARY KEY (integration_id);

ALTER TABLE ONLY public.integrations 
    ADD CONSTRAINT integrations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);
