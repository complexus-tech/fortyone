-- Add default_workflow_id to teams table
ALTER TABLE teams
ADD COLUMN default_workflow_id UUID,
  ADD CONSTRAINT fk_teams_default_workflow FOREIGN KEY (default_workflow_id) REFERENCES workflows(workflow_id) ON DELETE
SET NULL;