-- name: GetVisibleStoryRelationshipTarget :one
SELECT
    story.id AS entity_id,
    CAST('story' AS text) AS entity_type,
    story.title,
    CAST(team.code || '-' || story.sequence_id AS text) AS reference,
    story.team_id,
    story_status.color AS status_color
FROM public.stories AS story
LEFT JOIN public.statuses AS story_status
    ON story_status.status_id = story.status_id
   AND story_status.team_id = story.team_id
   AND story_status.workspace_id = story.workspace_id
INNER JOIN public.teams AS team
    ON team.team_id = story.team_id
   AND team.workspace_id = story.workspace_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = story.team_id
   AND team_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.workspace_id = story.workspace_id
   AND workspace_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = workspace_membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = story.workspace_id
   AND workspace.deleted_at IS NULL
WHERE story.id = sqlc.arg(entity_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL;

-- name: GetVisibleObjectiveRelationshipTarget :one
SELECT
    objective.objective_id AS entity_id,
    CAST('objective' AS text) AS entity_type,
    objective.name AS title,
    CAST(team.code || '-' || objective.sequence_id AS text) AS reference,
    objective.team_id
FROM public.objectives AS objective
INNER JOIN public.teams AS team
    ON team.team_id = objective.team_id
   AND team.workspace_id = objective.workspace_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.workspace_id = objective.workspace_id
   AND workspace_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = workspace_membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = objective.workspace_id
   AND workspace.deleted_at IS NULL
WHERE objective.objective_id = sqlc.arg(entity_id)
  AND objective.workspace_id = sqlc.arg(workspace_id);

-- Polymorphic entity IDs have no foreign key. Hold a key lock on the target
-- through commit so deletion cannot race its relationship cleanup.
-- name: InsertEditableDocumentRelationship :one
INSERT INTO public.document_relationships (
    document_id,
    workspace_id,
    entity_type,
    entity_id,
    created_by
)
SELECT
    document.document_id,
    document.workspace_id,
    CAST(sqlc.arg(entity_type) AS varchar(20)),
    sqlc.arg(entity_id),
    sqlc.arg(actor_id)
FROM public.documents AS document
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.archived_at IS NULL
  AND membership.role <> CAST('guest' AS public.user_role)
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS editor
          WHERE document.visibility = 'restricted'
            AND editor.document_id = document.document_id
            AND editor.user_id = sqlc.arg(actor_id)
            AND editor.role = 'editor'
      )
  )
  AND (
      (
          CAST(sqlc.arg(entity_type) AS varchar(20)) = 'story'
          AND EXISTS (
              SELECT 1
              FROM public.stories AS target_story
              INNER JOIN public.teams AS target_team
                  ON target_team.team_id = target_story.team_id
                 AND target_team.workspace_id = target_story.workspace_id
              INNER JOIN public.team_members AS target_membership
                  ON target_membership.team_id = target_story.team_id
                 AND target_membership.user_id = sqlc.arg(actor_id)
              WHERE target_story.id = sqlc.arg(entity_id)
                AND target_story.workspace_id = document.workspace_id
                AND target_story.deleted_at IS NULL
              FOR KEY SHARE OF target_story
          )
      )
      OR (
          CAST(sqlc.arg(entity_type) AS varchar(20)) = 'objective'
          AND EXISTS (
              SELECT 1
              FROM public.objectives AS target_objective
              INNER JOIN public.teams AS target_team
                  ON target_team.team_id = target_objective.team_id
                 AND target_team.workspace_id = target_objective.workspace_id
              INNER JOIN public.team_members AS target_membership
                  ON target_membership.team_id = target_objective.team_id
                 AND target_membership.user_id = sqlc.arg(actor_id)
              WHERE target_objective.objective_id = sqlc.arg(entity_id)
                AND target_objective.workspace_id = document.workspace_id
              FOR KEY SHARE OF target_objective
          )
      )
  )
ON CONFLICT (document_id, entity_type, entity_id) DO UPDATE
SET created_by = document_relationships.created_by, source = 'manual'
RETURNING document_id;

-- name: DeleteEditableDocumentRelationship :one
DELETE FROM public.document_relationships AS relationship
USING public.documents AS document, public.workspace_members AS membership,
      public.users AS actor, public.workspaces AS workspace
WHERE relationship.document_id = document.document_id
  AND relationship.document_id = sqlc.arg(document_id)
  AND relationship.workspace_id = sqlc.arg(workspace_id)
  AND relationship.entity_type = CAST(sqlc.arg(entity_type) AS varchar(20))
  AND relationship.entity_id = sqlc.arg(entity_id)
  AND document.archived_at IS NULL
  AND membership.workspace_id = document.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
  AND workspace.workspace_id = document.workspace_id
  AND workspace.deleted_at IS NULL
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS editor
          WHERE document.visibility = 'restricted'
            AND editor.document_id = document.document_id
            AND editor.user_id = sqlc.arg(actor_id)
            AND editor.role = 'editor'
      )
  )
  AND (
      (
          CAST(sqlc.arg(entity_type) AS varchar(20)) = 'story'
          AND EXISTS (
              SELECT 1
              FROM public.stories AS target_story
              INNER JOIN public.teams AS target_team
                  ON target_team.team_id = target_story.team_id
                 AND target_team.workspace_id = target_story.workspace_id
              INNER JOIN public.team_members AS target_membership
                  ON target_membership.team_id = target_story.team_id
                 AND target_membership.user_id = sqlc.arg(actor_id)
              WHERE target_story.id = sqlc.arg(entity_id)
                AND target_story.workspace_id = document.workspace_id
                AND target_story.deleted_at IS NULL
          )
      )
      OR (
          CAST(sqlc.arg(entity_type) AS varchar(20)) = 'objective'
          AND EXISTS (
              SELECT 1
              FROM public.objectives AS target_objective
              INNER JOIN public.teams AS target_team
                  ON target_team.team_id = target_objective.team_id
                 AND target_team.workspace_id = target_objective.workspace_id
              INNER JOIN public.team_members AS target_membership
                  ON target_membership.team_id = target_objective.team_id
                 AND target_membership.user_id = sqlc.arg(actor_id)
              WHERE target_objective.objective_id = sqlc.arg(entity_id)
                AND target_objective.workspace_id = document.workspace_id
          )
      )
  )
RETURNING relationship.entity_id;

-- name: ListVisibleDocumentRelationships :many
SELECT
    relationship.document_id,
    relationship.entity_id,
    relationship.entity_type,
    EXISTS(SELECT 1 FROM public.document_body_references AS reference WHERE reference.document_id=relationship.document_id AND reference.entity_type=relationship.entity_type AND reference.entity_id=relationship.entity_id) AS referenced_in_content,
    COALESCE(story.title, objective.name) AS title,
    CAST(COALESCE(
        story_team.code || '-' || story.sequence_id,
        objective_team.code || '-' || objective.sequence_id
    ) AS text) AS reference,
    COALESCE(story.team_id, objective.team_id) AS team_id,
    story_status.color AS status_color
FROM public.document_relationships AS relationship
INNER JOIN public.documents AS document
    ON document.document_id = relationship.document_id
   AND document.workspace_id = relationship.workspace_id
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
LEFT JOIN public.stories AS story
    ON relationship.entity_type = 'story'
   AND story.id = relationship.entity_id
   AND story.workspace_id = relationship.workspace_id
   AND story.deleted_at IS NULL
LEFT JOIN public.teams AS story_team
    ON story_team.team_id = story.team_id
   AND story_team.workspace_id = relationship.workspace_id
LEFT JOIN public.statuses AS story_status
    ON story_status.status_id = story.status_id
   AND story_status.team_id = story.team_id
   AND story_status.workspace_id = story.workspace_id
LEFT JOIN public.objectives AS objective
    ON relationship.entity_type = 'objective'
   AND objective.objective_id = relationship.entity_id
   AND objective.workspace_id = relationship.workspace_id
LEFT JOIN public.teams AS objective_team
    ON objective_team.team_id = objective.team_id
   AND objective_team.workspace_id = relationship.workspace_id
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.workspace_id = document.workspace_id
   AND workspace_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = workspace_membership.user_id
   AND actor.is_active = TRUE
WHERE relationship.workspace_id = sqlc.arg(workspace_id)
  AND relationship.document_id = ANY(CAST(sqlc.arg(document_ids) AS uuid[]))
  AND (
      document.archived_at IS NULL
      OR (
          CAST(sqlc.arg(include_archived) AS boolean)
          AND document.created_by = sqlc.arg(actor_id)
          AND workspace_membership.role <> CAST('guest' AS public.user_role)
      )
  )
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE document.visibility = 'restricted'
            AND reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
  AND (
      (story.id IS NOT NULL AND story_team.team_id IS NOT NULL)
      OR (objective.objective_id IS NOT NULL AND objective_team.team_id IS NOT NULL)
  )
  AND EXISTS (
      SELECT 1
      FROM public.team_members AS viewer
      WHERE viewer.user_id = sqlc.arg(actor_id)
        AND viewer.team_id = COALESCE(story_team.team_id, objective_team.team_id)
  )
ORDER BY relationship.document_id, relationship.created_at, relationship.entity_id;

-- name: ListAccessibleDocumentsForRelationship :many
SELECT
    document.document_id,
    document.workspace_id,
    document.title,
    document.visibility,
    document.created_by,
    document.updated_by,
    document.created_at,
    document.updated_at,
    document.archived_at,
    CAST(document.public_token IS NOT NULL AS boolean) AS public_enabled,
    (
        workspace_membership.role <> CAST('guest' AS public.user_role)
        AND (
            document.visibility = 'workspace'
            OR document.created_by = sqlc.arg(actor_id)
            OR EXISTS (
                SELECT 1
                FROM public.document_members AS editor
                WHERE document.visibility = 'restricted'
                  AND editor.document_id = document.document_id
                  AND editor.user_id = sqlc.arg(actor_id)
                  AND editor.role = 'editor'
            )
        )
    ) AS can_edit
FROM public.document_relationships AS relationship
INNER JOIN public.documents AS document ON document.document_id = relationship.document_id
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.workspace_id = document.workspace_id
   AND workspace_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = workspace_membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
WHERE relationship.workspace_id = sqlc.arg(workspace_id)
  AND relationship.entity_type = CAST(sqlc.arg(entity_type) AS varchar(20))
  AND relationship.entity_id = sqlc.arg(entity_id)
  AND document.archived_at IS NULL
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE document.visibility = 'restricted'
            AND reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
  AND (
      (
          CAST(sqlc.arg(entity_type) AS varchar(20)) = 'story'
          AND EXISTS (
              SELECT 1
              FROM public.stories AS target_story
              INNER JOIN public.teams AS target_team
                  ON target_team.team_id = target_story.team_id
                 AND target_team.workspace_id = target_story.workspace_id
              INNER JOIN public.team_members AS target_membership
                  ON target_membership.team_id = target_story.team_id
                 AND target_membership.user_id = sqlc.arg(actor_id)
              WHERE target_story.id = sqlc.arg(entity_id)
                AND target_story.workspace_id = document.workspace_id
                AND target_story.deleted_at IS NULL
          )
      )
      OR (
          CAST(sqlc.arg(entity_type) AS varchar(20)) = 'objective'
          AND EXISTS (
              SELECT 1
              FROM public.objectives AS target_objective
              INNER JOIN public.teams AS target_team
                  ON target_team.team_id = target_objective.team_id
                 AND target_team.workspace_id = target_objective.workspace_id
              INNER JOIN public.team_members AS target_membership
                  ON target_membership.team_id = target_objective.team_id
                 AND target_membership.user_id = sqlc.arg(actor_id)
              WHERE target_objective.objective_id = sqlc.arg(entity_id)
                AND target_objective.workspace_id = document.workspace_id
          )
      )
  )
ORDER BY document.updated_at DESC, document.document_id;

-- name: HasDocumentBodyReference :one
SELECT EXISTS(SELECT 1 FROM public.document_body_references WHERE document_id=sqlc.arg(document_id) AND entity_type=sqlc.arg(entity_type) AND entity_id=sqlc.arg(entity_id)) AS present;

-- name: SyncDocumentBodyReferences :exec
SELECT public.sync_document_body_references(sqlc.arg(workspace_id),sqlc.arg(actor_id),sqlc.arg(document_id),CAST(sqlc.arg(reference_data) AS jsonb));

-- name: SyncDocumentBodyMentions :exec
SELECT public.sync_document_body_mentions(sqlc.arg(workspace_id),sqlc.arg(actor_id),sqlc.arg(document_id),CAST(sqlc.arg(mention_data) AS jsonb));
