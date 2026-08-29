-- name: RecordOutboundGitHubComment :exec
INSERT INTO public.github_comment_links (
    workspace_id,
    story_id,
    repository_id,
    local_comment_id,
    github_comment_id,
    source,
    created_by_user_id
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(repository_id),
    sqlc.narg(local_comment_id),
    sqlc.arg(github_comment_id),
    'fortyone',
    sqlc.arg(created_by_user_id)
)
ON CONFLICT (repository_id, github_comment_id) DO UPDATE SET
    local_comment_id = COALESCE(EXCLUDED.local_comment_id, github_comment_links.local_comment_id),
    source = 'fortyone',
    created_by_user_id = EXCLUDED.created_by_user_id;

-- name: ReserveInboundGitHubComment :execrows
INSERT INTO public.github_comment_links (
    workspace_id,
    story_id,
    repository_id,
    github_comment_id,
    source,
    created_by_user_id
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(repository_id),
    sqlc.arg(github_comment_id),
    'github',
    sqlc.arg(created_by_user_id)
)
ON CONFLICT (repository_id, github_comment_id) DO NOTHING;

-- name: CompleteInboundGitHubComment :exec
UPDATE public.github_comment_links
SET local_comment_id = sqlc.arg(local_comment_id)
WHERE repository_id = sqlc.arg(repository_id)
  AND github_comment_id = sqlc.arg(github_comment_id)
  AND source = 'github';

-- name: DeleteGitHubCommentLink :exec
DELETE FROM public.github_comment_links
WHERE repository_id = sqlc.arg(repository_id)
  AND github_comment_id = sqlc.arg(github_comment_id);

-- name: IsOutboundGitHubComment :one
SELECT EXISTS (
    SELECT 1
    FROM public.github_comment_links
    WHERE repository_id = sqlc.arg(repository_id)
      AND github_comment_id = sqlc.arg(github_comment_id)
      AND source = 'fortyone'
);

