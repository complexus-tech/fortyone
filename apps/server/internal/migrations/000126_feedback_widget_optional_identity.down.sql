UPDATE public.feedback_widget_settings
SET enabled = false,
    updated_at = now()
WHERE enabled = true
  AND (
      signing_secret_encrypted IS NULL
      OR signing_secret_version <= 0
  );

ALTER TABLE public.feedback_widget_settings
    DROP CONSTRAINT feedback_widget_settings_enabled_check;

ALTER TABLE public.feedback_widget_settings
    ADD CONSTRAINT feedback_widget_settings_enabled_check
        CHECK (
            NOT enabled
            OR (
                cardinality(allowed_origins) > 0
                AND signing_secret_encrypted IS NOT NULL
                AND signing_secret_version > 0
            )
        );
