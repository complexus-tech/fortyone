ALTER TABLE slack_agent_settings
    ADD COLUMN IF NOT EXISTS assistant_enabled boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS workflow_actions_enabled boolean NOT NULL DEFAULT true;
