CREATE TABLE IF NOT EXISTS "public"."key_result_contributors" (
    "key_result_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "key_result_contributors_key_result_id_fkey" FOREIGN KEY ("key_result_id") REFERENCES "public"."key_results"("id") ON DELETE CASCADE,
    CONSTRAINT "key_result_contributors_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("key_result_id", "user_id")
);

CREATE INDEX "idx_key_result_contributors_user_id" ON "public"."key_result_contributors" USING btree ("user_id");
