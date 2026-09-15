-- name: LockTeamForDeletion :one
SELECT team.team_id
FROM public.teams AS team
INNER JOIN public.workspace_members AS membership ON membership.workspace_id = team.workspace_id
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role = 'admin'
  AND actor.is_active = TRUE
FOR UPDATE OF team
FOR SHARE OF membership, actor;

-- Keep ownership stable until the objective cascade and polymorphic-reference
-- cleanup have committed; an objective transfer must wait for this decision.
-- name: LockTeamDeletionObjectives :exec
SELECT objective.objective_id FROM public.objectives AS objective
WHERE objective.team_id = sqlc.arg(team_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
ORDER BY objective.objective_id
FOR UPDATE OF objective;

-- Lock intermediate parents so new feedback cannot arrive after attachment
-- capture but before the team cascade.
-- name: LockTeamDeletionBoards :exec
SELECT board.id FROM public.feedback_boards AS board
WHERE board.team_id = sqlc.arg(team_id)
  AND board.workspace_id = sqlc.arg(workspace_id)
ORDER BY board.id
FOR UPDATE OF board;

-- name: LockTeamDeletionFeedback :exec
SELECT item.id FROM public.feedback_items AS item
INNER JOIN public.feedback_boards AS board ON board.id = item.board_id
WHERE board.team_id = sqlc.arg(team_id)
  AND board.workspace_id = sqlc.arg(workspace_id)
  AND item.workspace_id = board.workspace_id
ORDER BY item.id
FOR UPDATE OF item;

-- Entity IDs in document relationships and notifications deliberately have
-- no foreign key. Remove the links while their target rows still exist.
-- name: DeleteTeamDocumentRelationships :exec
DELETE FROM public.document_relationships AS relationship
WHERE relationship.workspace_id = sqlc.arg(workspace_id)
  AND (
      (relationship.entity_type = 'story' AND EXISTS (
          SELECT 1 FROM public.stories AS story
          WHERE story.id = relationship.entity_id
            AND story.team_id = sqlc.arg(team_id)
            AND story.workspace_id = relationship.workspace_id
      ))
      OR (relationship.entity_type = 'objective' AND EXISTS (
          SELECT 1 FROM public.objectives AS objective
          WHERE objective.objective_id = relationship.entity_id
            AND objective.team_id = sqlc.arg(team_id)
            AND objective.workspace_id = relationship.workspace_id
      ))
  );

-- name: DeleteTeamEntityNotifications :exec
DELETE FROM public.notifications AS notification
WHERE notification.workspace_id = sqlc.arg(workspace_id)
  AND (
      (notification.entity_type = 'story' AND EXISTS (
          SELECT 1 FROM public.stories AS story
          WHERE story.id = notification.entity_id
            AND story.team_id = sqlc.arg(team_id)
            AND story.workspace_id = notification.workspace_id
      ))
      OR (notification.entity_type = 'comment' AND EXISTS (
          SELECT 1 FROM public.story_comments AS comment
          INNER JOIN public.stories AS story ON story.id = comment.story_id
          WHERE comment.comment_id = notification.entity_id
            AND story.team_id = sqlc.arg(team_id)
            AND story.workspace_id = notification.workspace_id
      ))
      OR (notification.entity_type = 'objective' AND EXISTS (
          SELECT 1 FROM public.objectives AS objective
          WHERE objective.objective_id = notification.entity_id
            AND objective.team_id = sqlc.arg(team_id)
            AND objective.workspace_id = notification.workspace_id
      ))
      OR (notification.entity_type = 'key_result' AND EXISTS (
          SELECT 1 FROM public.key_results AS result
          INNER JOIN public.objectives AS objective ON objective.objective_id = result.objective_id
          WHERE result.id = notification.entity_id
            AND result.team_id = sqlc.arg(team_id)
            AND objective.workspace_id = notification.workspace_id
      ))
      OR (notification.entity_type = 'feedback' AND EXISTS (
          SELECT 1 FROM public.feedback_items AS item
          INNER JOIN public.feedback_boards AS board ON board.id = item.board_id
          WHERE item.id = notification.entity_id
            AND board.team_id = sqlc.arg(team_id)
            AND board.workspace_id = notification.workspace_id
            AND item.workspace_id = notification.workspace_id
      ))
  );
