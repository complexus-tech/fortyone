ALTER TABLE public.slack_channels
    ADD COLUMN is_assistant_configured boolean NOT NULL DEFAULT false;

UPDATE public.slack_channels channel_record
SET is_assistant_configured = true
WHERE EXISTS (
    SELECT 1
    FROM public.slack_channel_team_access access
    JOIN public.teams team
      ON team.team_id = access.team_id
     AND team.workspace_id = access.workspace_id
     AND team.is_private = false
    WHERE access.workspace_id = channel_record.workspace_id
      AND access.slack_workspace_id = channel_record.slack_workspace_id
      AND access.slack_channel_id = channel_record.slack_channel_id
);
