ALTER TABLE public.feedback_widget_settings
    DROP CONSTRAINT feedback_widget_settings_enabled_check;

ALTER TABLE public.feedback_widget_settings
    ADD CONSTRAINT feedback_widget_settings_enabled_check
        CHECK (NOT enabled OR cardinality(allowed_origins) > 0);
