-- Association rows are locked before their target set is authorized. The
-- caller then reuses AuthorizeSecondaryStoryTargets for live credential,
-- workspace membership, and team membership enforcement inside one tx.

-- name: LockStoryAssociation :one
SELECT
    association.id,
    association.from_story_id,
    association.to_story_id,
    CAST(association.association_type AS text) AS association_type
FROM public.story_associations AS association
WHERE association.id = sqlc.arg(association_id)
  AND association.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE;

-- name: InsertStoryAssociation :one
INSERT INTO public.story_associations (
    id,
    from_story_id,
    to_story_id,
    association_type,
    workspace_id
) VALUES (
    sqlc.arg(association_id),
    sqlc.arg(from_story_id),
    sqlc.arg(to_story_id),
    sqlc.arg(association_type),
    sqlc.arg(workspace_id)
)
RETURNING id, from_story_id, to_story_id, CAST(association_type AS text) AS association_type;

-- name: UpdateStoryAssociation :one
UPDATE public.story_associations AS association
SET
    from_story_id = sqlc.arg(from_story_id),
    to_story_id = sqlc.arg(to_story_id),
    association_type = sqlc.arg(association_type)
WHERE association.id = sqlc.arg(association_id)
  AND association.workspace_id = sqlc.arg(workspace_id)
RETURNING id, from_story_id, to_story_id, CAST(association_type AS text) AS association_type;

-- name: DeleteStoryAssociation :one
DELETE FROM public.story_associations AS association
WHERE association.id = sqlc.arg(association_id)
  AND association.workspace_id = sqlc.arg(workspace_id)
RETURNING id, from_story_id, to_story_id, CAST(association_type AS text) AS association_type;
