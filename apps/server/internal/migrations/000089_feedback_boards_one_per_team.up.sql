CREATE UNIQUE INDEX feedback_boards_workspace_team_unique
    ON public.feedback_boards (workspace_id, team_id);
