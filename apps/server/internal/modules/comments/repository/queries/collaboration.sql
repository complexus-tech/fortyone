-- name: GetStoryCommentReadState :one
WITH RECURSIVE scoped_story AS (
    SELECT story.id
    FROM public.stories AS story
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS member
        ON member.workspace_id = story.workspace_id AND member.user_id = actor.user_id
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
           OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
      AND story.id = sqlc.arg(story_id)
), comment_tree AS (
    SELECT comment.comment_id, comment.comment_id AS root_id, comment.parent_id,
        comment.created_at, comment.commenter_id
    FROM public.story_comments AS comment
    INNER JOIN scoped_story ON scoped_story.id = comment.story_id
    WHERE comment.parent_id IS NULL
    UNION ALL
    SELECT comment.comment_id, parent.root_id, comment.parent_id,
        comment.created_at, comment.commenter_id
    FROM public.story_comments AS comment
    INNER JOIN scoped_story ON scoped_story.id = comment.story_id
    INNER JOIN comment_tree AS parent ON parent.comment_id = comment.parent_id
), thread_reads AS (
    SELECT marker.thread_id, marker.last_read_at
    FROM public.story_comment_thread_reads AS marker
    INNER JOIN public.story_comments AS root ON root.comment_id = marker.thread_id AND root.parent_id IS NULL
    INNER JOIN scoped_story ON scoped_story.id = root.story_id
    WHERE marker.user_id = sqlc.arg(actor_id)
)
SELECT marker.last_read_at, CAST(statement_timestamp() AS timestamptz) AS observed_at,
    latest.comment_id AS latest_comment_id,
    CAST(COALESCE((SELECT jsonb_object_agg(CAST(thread_id AS text), last_read_at) FROM thread_reads), CAST('{}' AS jsonb)) AS jsonb) AS thread_reads,
    CAST((SELECT count(*) FROM comment_tree AS comment
          LEFT JOIN thread_reads AS thread_marker ON thread_marker.thread_id = comment.root_id
          WHERE comment.commenter_id <> sqlc.arg(actor_id)
            AND CASE WHEN comment.parent_id IS NULL
                THEN marker.last_read_at IS NULL OR comment.created_at > marker.last_read_at
                ELSE thread_marker.last_read_at IS NULL OR comment.created_at > thread_marker.last_read_at
            END) AS bigint) AS unread_count
FROM scoped_story
LEFT JOIN public.story_comment_reads AS marker
    ON marker.story_id = scoped_story.id AND marker.user_id = sqlc.arg(actor_id)
LEFT JOIN public.story_comments AS latest ON latest.comment_id = (
    SELECT comment.comment_id FROM public.story_comments AS comment
    WHERE comment.story_id = scoped_story.id
    ORDER BY comment.created_at DESC, comment.comment_id DESC
    LIMIT 1
);

-- name: ListCommentCollaboration :many
WITH scoped_story AS (
    SELECT story.id
    FROM public.stories AS story
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS member
        ON member.workspace_id = story.workspace_id AND member.user_id = actor.user_id
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
           OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
      AND story.id = sqlc.arg(story_id)
)
SELECT comment.comment_id, resolution.resolved_at, resolution.resolved_by,
       reaction.emoji, CAST(count(reaction.user_id) AS bigint) AS reaction_count,
       CAST(COALESCE(bool_or(reaction.user_id = sqlc.arg(actor_id)), FALSE) AS boolean) AS reacted
FROM scoped_story
INNER JOIN public.story_comments AS comment ON comment.story_id = scoped_story.id
LEFT JOIN public.story_comment_resolutions AS resolution ON resolution.comment_id = comment.comment_id
LEFT JOIN public.story_comment_reactions AS reaction ON reaction.comment_id = comment.comment_id
WHERE (CAST(sqlc.narg(comment_id) AS uuid) IS NULL OR comment.comment_id = sqlc.narg(comment_id))
  AND (CAST(sqlc.narg(comment_id) AS uuid) IS NOT NULL OR resolution.resolved_at IS NOT NULL OR reaction.user_id IS NOT NULL)
GROUP BY comment.comment_id, resolution.resolved_at, resolution.resolved_by, reaction.emoji
ORDER BY comment.comment_id, reaction.emoji;

-- name: SetCommentReaction :one
WITH scoped_story AS (
    SELECT story.id
    FROM public.stories AS story
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS member
        ON member.workspace_id = story.workspace_id AND member.user_id = actor.user_id
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
           OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
), scoped_comment AS (
    SELECT comment.comment_id, comment.story_id
    FROM public.story_comments AS comment
    INNER JOIN scoped_story ON scoped_story.id = comment.story_id
    WHERE comment.comment_id = sqlc.arg(comment_id)
), added AS (
    INSERT INTO public.story_comment_reactions (comment_id, user_id, emoji)
    SELECT comment_id, sqlc.arg(actor_id), sqlc.arg(emoji)
    FROM scoped_comment WHERE CAST(sqlc.arg(active) AS boolean)
    ON CONFLICT (comment_id, user_id, emoji) DO NOTHING
), removed AS (
    DELETE FROM public.story_comment_reactions AS reaction
    USING scoped_comment
    WHERE reaction.comment_id = scoped_comment.comment_id
      AND reaction.user_id = sqlc.arg(actor_id) AND reaction.emoji = sqlc.arg(emoji)
      AND NOT CAST(sqlc.arg(active) AS boolean)
)
SELECT story_id FROM scoped_comment;

-- name: SetCommentResolution :one
WITH scoped_story AS (
    SELECT story.id
    FROM public.stories AS story
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS member
        ON member.workspace_id = story.workspace_id AND member.user_id = actor.user_id
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
           OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
), scoped_comment AS (
    SELECT comment.comment_id, comment.story_id
    FROM public.story_comments AS comment
    INNER JOIN scoped_story ON scoped_story.id = comment.story_id
    WHERE comment.comment_id = sqlc.arg(comment_id) AND comment.parent_id IS NULL
), changed AS (
    INSERT INTO public.story_comment_resolutions (comment_id, resolved_at, resolved_by)
    SELECT comment_id,
        CASE WHEN CAST(sqlc.arg(resolved) AS boolean) THEN clock_timestamp() ELSE NULL END,
        CASE WHEN CAST(sqlc.arg(resolved) AS boolean) THEN CAST(sqlc.arg(actor_id) AS uuid) ELSE NULL END
    FROM scoped_comment
    ON CONFLICT (comment_id) DO UPDATE
    SET resolved_at = CASE WHEN CAST(sqlc.arg(resolved) AS boolean)
            THEN COALESCE(story_comment_resolutions.resolved_at, EXCLUDED.resolved_at) ELSE NULL END,
        resolved_by = CASE WHEN CAST(sqlc.arg(resolved) AS boolean)
            THEN CASE WHEN story_comment_resolutions.resolved_at IS NOT NULL
                 THEN story_comment_resolutions.resolved_by ELSE EXCLUDED.resolved_by END ELSE NULL END
)
SELECT story_id FROM scoped_comment;

-- name: MarkStoryCommentsRead :one
WITH scoped_story AS (
    SELECT story.id
    FROM public.stories AS story
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS member
        ON member.workspace_id = story.workspace_id AND member.user_id = actor.user_id
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
           OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
      AND story.id = sqlc.arg(story_id)
)
INSERT INTO public.story_comment_reads (story_id, user_id, last_read_at)
SELECT scoped_story.id, sqlc.arg(actor_id), LEAST(CAST(sqlc.arg(observed_at) AS timestamptz), statement_timestamp())
FROM scoped_story
ON CONFLICT (story_id, user_id) DO UPDATE
SET last_read_at = GREATEST(story_comment_reads.last_read_at, EXCLUDED.last_read_at)
RETURNING last_read_at;

-- name: MarkStoryCommentThreadRead :one
WITH scoped_root AS (
    SELECT root.comment_id
    FROM public.story_comments AS root
    INNER JOIN public.stories AS story ON story.id = root.story_id
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE
    INNER JOIN public.workspace_members AS member
        ON member.workspace_id = story.workspace_id AND member.user_id = actor.user_id
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.id = sqlc.arg(story_id)
      AND story.deleted_at IS NULL
      AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
           OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
      AND root.comment_id = sqlc.arg(thread_id)
      AND root.parent_id IS NULL
)
INSERT INTO public.story_comment_thread_reads (thread_id, user_id, last_read_at)
SELECT scoped_root.comment_id, sqlc.arg(actor_id), LEAST(CAST(sqlc.arg(observed_at) AS timestamptz), statement_timestamp())
FROM scoped_root
ON CONFLICT (thread_id, user_id) DO UPDATE
SET last_read_at = GREATEST(story_comment_thread_reads.last_read_at, EXCLUDED.last_read_at)
RETURNING last_read_at;
