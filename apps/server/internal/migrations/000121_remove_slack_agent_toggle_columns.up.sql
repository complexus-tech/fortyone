ALTER TABLE slack_agent_settings
    DROP COLUMN IF EXISTS assistant_enabled,
    DROP COLUMN IF EXISTS workflow_actions_enabled;
