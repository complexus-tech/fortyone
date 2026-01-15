-- 000014_key_results.up.sql
CREATE TABLE "public"."key_results" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "objective_id" uuid NOT NULL,
    "name" varchar(255) NOT NULL,
    "measurement_type" "public"."measurement_type" NOT NULL,
    "start_value" numeric,
    "target_value" numeric,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "current_value" numeric,
    "created_by" uuid,
    "lead" uuid,
    "start_date" date NOT NULL,
    "end_date" date NOT NULL,
    CONSTRAINT "key_results_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users"("user_id"),
    CONSTRAINT "key_results_lead_fkey" FOREIGN KEY ("lead") REFERENCES "public"."users"("user_id"),
    CONSTRAINT "key_results_objective_id_fkey" FOREIGN KEY ("objective_id") REFERENCES "public"."objectives"("objective_id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE INDEX idx_key_results_lead ON public.key_results USING btree (lead);
CREATE INDEX idx_key_results_start_date ON public.key_results USING btree (start_date);
CREATE INDEX idx_key_results_end_date ON public.key_results USING btree (end_date);
