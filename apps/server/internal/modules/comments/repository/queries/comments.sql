-- name: GetCommentForWorkspace :one
SELECT
    comment.comment_id,
    comment.story_id,
    comment.parent_id,
    comment.commenter_id,
    comment.content,
    comment.created_at,
    comment.updated_at
FROM public.story_comments AS comment
INNER JOIN public.stories AS story ON story.id = comment.story_id
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND story.workspace_id = sqlc.arg(workspace_id);

-- name: CreateCommentForActor :one
WITH scoped_story AS (
    SELECT story.id
    FROM public.stories AS story
    INNER JOIN public.workspace_members AS actor_member
        ON actor_member.workspace_id = story.workspace_id
       AND actor_member.user_id = sqlc.arg(actor_id)
    INNER JOIN public.users AS actor_user
        ON actor_user.user_id = actor_member.user_id
       AND actor_user.is_active = TRUE
    INNER JOIN public.team_members AS actor_team_member
        ON actor_team_member.team_id = story.team_id
       AND actor_team_member.user_id = actor_member.user_id
    WHERE story.id = sqlc.arg(story_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (
          CAST(sqlc.arg(team_access_unrestricted) AS boolean)
          OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
      )
      AND (
          CAST(sqlc.narg(parent_id) AS uuid) IS NULL
          OR EXISTS (
              SELECT 1
              FROM public.story_comments AS parent_comment
              WHERE parent_comment.comment_id = sqlc.narg(parent_id)
                AND parent_comment.story_id = story.id
          )
      )
)
INSERT INTO public.story_comments (
    content,
    story_id,
    commenter_id,
    parent_id
)
SELECT
    sqlc.arg(content),
    scoped_story.id,
    sqlc.arg(actor_id),
    sqlc.narg(parent_id)
FROM scoped_story
RETURNING
    comment_id,
    story_id,
    parent_id,
    commenter_id,
    content,
    created_at,
    updated_at;

-- name: UpdateCommentForAuthor :one
UPDATE public.story_comments AS comment
SET
    content = sqlc.arg(content),
    updated_at = GREATEST(clock_timestamp(), comment.updated_at + INTERVAL '1 microsecond')
FROM public.stories AS story
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND comment.commenter_id = sqlc.arg(actor_id)
  AND story.id = comment.story_id
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(team_access_unrestricted) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS actor_member
      INNER JOIN public.users AS actor_user
          ON actor_user.user_id = actor_member.user_id
         AND actor_user.is_active = TRUE
      INNER JOIN public.team_members AS actor_team_member
          ON actor_team_member.team_id = story.team_id
         AND actor_team_member.user_id = actor_member.user_id
      WHERE actor_member.workspace_id = story.workspace_id
        AND actor_member.user_id = sqlc.arg(actor_id)
  )
RETURNING
    comment.comment_id,
    comment.story_id,
    comment.parent_id,
    comment.commenter_id,
    comment.content,
    comment.created_at,
    comment.updated_at;

-- name: DeleteCommentForAuthor :one
DELETE FROM public.story_comments AS comment
USING public.stories AS story
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND comment.commenter_id = sqlc.arg(actor_id)
  AND story.id = comment.story_id
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(team_access_unrestricted) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS actor_member
      INNER JOIN public.users AS actor_user
          ON actor_user.user_id = actor_member.user_id
         AND actor_user.is_active = TRUE
      INNER JOIN public.team_members AS actor_team_member
          ON actor_team_member.team_id = story.team_id
         AND actor_team_member.user_id = actor_member.user_id
      WHERE actor_member.workspace_id = story.workspace_id
        AND actor_member.user_id = sqlc.arg(actor_id)
  )
RETURNING
    comment.comment_id,
    comment.story_id,
    comment.parent_id,
    comment.commenter_id,
    comment.content,
    comment.created_at,
    CAST(GREATEST(clock_timestamp(), comment.updated_at) AS timestamptz) AS updated_at;

-- name: DeleteCommentMentionsForAuthor :one
WITH scoped_comment AS (
    SELECT comment.comment_id
    FROM public.story_comments AS comment
    INNER JOIN public.stories AS story ON story.id = comment.story_id
    WHERE comment.comment_id = sqlc.arg(comment_id)
      AND comment.commenter_id = sqlc.arg(actor_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (
          CAST(sqlc.arg(team_access_unrestricted) AS boolean)
          OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
      )
      AND EXISTS (
          SELECT 1
          FROM public.workspace_members AS actor_member
          INNER JOIN public.users AS actor_user
              ON actor_user.user_id = actor_member.user_id
             AND actor_user.is_active = TRUE
          INNER JOIN public.team_members AS actor_team_member
              ON actor_team_member.team_id = story.team_id
             AND actor_team_member.user_id = actor_member.user_id
          WHERE actor_member.workspace_id = story.workspace_id
            AND actor_member.user_id = sqlc.arg(actor_id)
      )
),
deleted_mention AS (
    DELETE FROM public.comment_mentions AS mention
    USING scoped_comment
    WHERE mention.comment_id = scoped_comment.comment_id
    RETURNING mention.comment_id
)
SELECT
    EXISTS(SELECT 1 FROM scoped_comment) AS comment_found,
    CAST((SELECT COUNT(*) FROM deleted_mention) AS bigint) AS deleted_count;

-- name: InsertCommentMentionsForAuthor :execrows
WITH requested_user AS (
    SELECT DISTINCT requested.user_id
    FROM unnest(CAST(sqlc.arg(mentioned_user_ids) AS uuid[])) AS requested(user_id)
)
INSERT INTO public.comment_mentions (
    comment_id,
    mentioned_user_id
)
SELECT
    comment.comment_id,
    requested_user.user_id
FROM public.story_comments AS comment
INNER JOIN public.stories AS story ON story.id = comment.story_id
INNER JOIN requested_user ON TRUE
INNER JOIN public.workspace_members AS member
    ON member.workspace_id = story.workspace_id
   AND member.user_id = requested_user.user_id
INNER JOIN public.users AS mentioned_user
    ON mentioned_user.user_id = member.user_id
   AND mentioned_user.is_active = TRUE
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND comment.commenter_id = sqlc.arg(actor_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(team_access_unrestricted) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS actor_member
      INNER JOIN public.users AS actor_user
          ON actor_user.user_id = actor_member.user_id
         AND actor_user.is_active = TRUE
      INNER JOIN public.team_members AS actor_team_member
          ON actor_team_member.team_id = story.team_id
         AND actor_team_member.user_id = actor_member.user_id
      WHERE actor_member.workspace_id = story.workspace_id
        AND actor_member.user_id = sqlc.arg(actor_id)
  );

-- name: AppendCommentMutationEvent :many
WITH created_event AS (
    INSERT INTO public.outbound_webhook_events (
        event_id,
        workspace_id,
        event_type,
        payload_version,
        subject_type,
        subject_id,
        actor_kind,
        actor_id,
        actor_credential_id,
        payload,
        occurred_at,
        created_at
    ) VALUES (
        sqlc.arg(event_id),
        sqlc.arg(workspace_id),
        sqlc.arg(event_type),
        1,
        'comment',
        sqlc.arg(comment_id),
        sqlc.arg(actor_kind),
        sqlc.arg(actor_id),
        sqlc.narg(actor_credential_id),
        sqlc.arg(payload),
        sqlc.arg(occurred_at),
        GREATEST(clock_timestamp(), CAST(sqlc.arg(occurred_at) AS timestamptz))
    )
    RETURNING event_id, workspace_id, created_at
)
INSERT INTO public.outbound_webhook_deliveries (
    delivery_id,
    workspace_id,
    event_id,
    endpoint_id,
    subscription_generation,
    payload_body,
    status,
    attempt_count,
    available_at,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    endpoint.workspace_id,
    created_event.event_id,
    endpoint.endpoint_id,
    endpoint.subscription_generation,
    sqlc.arg(payload_body),
    'pending',
    0,
    created_event.created_at,
    created_event.created_at,
    created_event.created_at
FROM created_event
INNER JOIN public.outbound_webhook_endpoints AS endpoint
    ON endpoint.workspace_id = created_event.workspace_id
INNER JOIN public.outbound_webhook_subscriptions AS subscription
    ON subscription.endpoint_id = endpoint.endpoint_id
   AND subscription.workspace_id = endpoint.workspace_id
INNER JOIN public.principals AS principal
    ON principal.principal_id = endpoint.owner_principal_id
   AND principal.workspace_id = endpoint.workspace_id
LEFT JOIN public.users AS account
    ON account.user_id = principal.subject_user_id
LEFT JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
   AND membership.user_id = principal.subject_user_id
WHERE endpoint.status = 'active'
  AND subscription.event_type = sqlc.arg(event_type)
  AND principal.status = 'active'
  AND principal.kind = 'human_user'
  AND account.is_active = TRUE
  AND membership.user_id IS NOT NULL
RETURNING delivery_id;
