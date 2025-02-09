-- Remove default_workflow_id from teams table
ALTER TABLE teams DROP CONSTRAINT IF EXISTS fk_teams_default_workflow,
  DROP COLUMN IF EXISTS default_workflow_id;