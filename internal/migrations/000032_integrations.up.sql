-- 000032_integrations.up.sql
CREATE TABLE "public"."integrations" (
    "integration_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "type" varchar(255) NOT NULL,
    "config" jsonb,
    "user_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "integrations_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id"),
    PRIMARY KEY ("integration_id")
);
