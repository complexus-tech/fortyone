-- 000033_verification_tokens.up.sql
CREATE TABLE public.verification_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    token character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    user_id uuid,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    token_type public.token_type NOT NULL
);

ALTER TABLE ONLY public.verification_tokens ADD CONSTRAINT verification_tokens_email_token_unique UNIQUE (email, token);
ALTER TABLE ONLY public.verification_tokens ADD CONSTRAINT verification_tokens_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.verification_tokens 
    ADD CONSTRAINT verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);

CREATE INDEX idx_verification_tokens_email ON public.verification_tokens USING btree (email);
CREATE INDEX idx_verification_tokens_expires_at ON public.verification_tokens USING btree (expires_at);
CREATE INDEX idx_verification_tokens_token ON public.verification_tokens USING btree (token);
CREATE INDEX idx_verification_tokens_token_lookup ON public.verification_tokens USING btree (token);
CREATE INDEX idx_verification_tokens_validation ON public.verification_tokens USING btree (token, expires_at, used_at);
