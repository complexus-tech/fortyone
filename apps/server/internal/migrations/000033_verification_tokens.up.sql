-- 000033_verification_tokens.up.sql
DROP TYPE IF EXISTS "public"."token_type";
CREATE TYPE "public"."token_type" AS ENUM ('login', 'registration');

-- Table Definition
CREATE TABLE "public"."verification_tokens" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "token" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "user_id" uuid,
    "expires_at" timestamptz NOT NULL,
    "used_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "token_type" "public"."token_type" NOT NULL,
    CONSTRAINT "verification_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id"),
    PRIMARY KEY ("id")
);


-- Indices
CREATE INDEX idx_verification_tokens_email ON public.verification_tokens USING btree (email);
CREATE INDEX idx_verification_tokens_expires_at ON public.verification_tokens USING btree (expires_at);
CREATE INDEX idx_verification_tokens_token ON public.verification_tokens USING btree (token);
CREATE INDEX idx_verification_tokens_validation ON public.verification_tokens USING btree (token, expires_at, used_at);
CREATE UNIQUE INDEX verification_tokens_email_token_unique ON public.verification_tokens USING btree (email, token);
CREATE INDEX idx_verification_tokens_token_lookup ON public.verification_tokens USING btree (token);
