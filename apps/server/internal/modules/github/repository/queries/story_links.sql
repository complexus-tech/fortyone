-- name: FindGitHubRepositoryByExternalID :one
SELECT gr.id,
       gr.workspace_id,
       w.slug AS workspace_slug,
       gr.full_name,
       gr.owner_login,
       gr.name AS repository_slug,
       gr.default_branch,
       gi.github_installation_id
FROM public.github_repositories AS gr
INNER JOIN public.github_installations AS gi ON gi.id = gr.installation_id
INNER JOIN public.workspaces AS w ON w.workspace_id = gr.workspace_id
WHERE gr.github_repository_id = sqlc.arg(github_repository_id)
  AND gr.is_active = TRUE
  AND gi.is_active = TRUE
  AND gi.suspended_at IS NULL
  AND gi.disconnected_at IS NULL
ORDER BY gr.id ASC
LIMIT 1;

-- name: FindGitHubRepositoryByID :one
SELECT gr.id,
       gr.workspace_id,
       w.slug AS workspace_slug,
       gr.full_name,
       gr.owner_login,
       gr.name AS repository_slug,
       gr.default_branch,
       gi.github_installation_id
FROM public.github_repositories AS gr
INNER JOIN public.github_installations AS gi ON gi.id = gr.installation_id
INNER JOIN public.workspaces AS w ON w.workspace_id = gr.workspace_id
WHERE gr.workspace_id = sqlc.arg(workspace_id)
  AND gr.id = sqlc.arg(repository_id)
  AND gr.is_active = TRUE
  AND gi.is_active = TRUE
  AND gi.suspended_at IS NULL
  AND gi.disconnected_at IS NULL;

-- name: ResolveGitHubStoryReference :one
SELECT s.id AS story_id,
       s.status_id,
       s.team_id,
       t.code AS team_code,
       s.sequence_id,
       s.title
FROM public.stories AS s
INNER JOIN public.teams AS t ON t.team_id = s.team_id
WHERE s.workspace_id = sqlc.arg(workspace_id)
  AND UPPER(t.code) = sqlc.arg(team_code)
  AND s.sequence_id = sqlc.arg(sequence_id)
ORDER BY s.id ASC
LIMIT 1;

-- name: FindGitHubStoryLink :one
SELECT id,
       story_id
FROM public.github_story_links
WHERE repository_id = sqlc.arg(repository_id)
  AND external_type = sqlc.arg(external_type)
  AND COALESCE(github_id, 0) = sqlc.arg(github_id)
  AND COALESCE(ref_name, '') = COALESCE(sqlc.narg(ref_name), '')
ORDER BY created_at ASC, id ASC
LIMIT 1;

-- name: FindGitHubIssueStoryLinkByStoryID :one
SELECT id,
       story_id,
       repository_id,
       github_id,
       github_number,
       url,
       title,
       state,
       last_synced_from,
       last_sync_hash
FROM public.github_story_links
WHERE workspace_id = sqlc.arg(workspace_id)
  AND story_id = sqlc.arg(story_id)
  AND repository_id = sqlc.arg(repository_id)
  AND external_type = 'issue'
ORDER BY created_at DESC, id ASC
LIMIT 1;

-- name: FindGitHubIssueStoryLinkByExternalID :one
SELECT id,
       story_id,
       repository_id,
       github_id,
       github_number,
       url,
       title,
       state,
       last_synced_from,
       last_sync_hash
FROM public.github_story_links
WHERE repository_id = sqlc.arg(repository_id)
  AND external_type = 'issue'
  AND github_id = sqlc.arg(github_id)
ORDER BY created_at ASC, id ASC
LIMIT 1;

-- name: FindGitHubStoryLinksByPRNumber :many
SELECT s.id AS story_id,
       s.status_id,
       s.team_id,
       t.code AS team_code,
       s.sequence_id,
       s.title
FROM public.github_story_links AS link
INNER JOIN public.stories AS s ON s.id = link.story_id
INNER JOIN public.teams AS t ON t.team_id = s.team_id
WHERE link.repository_id = sqlc.arg(repository_id)
  AND link.external_type = 'pull_request'
  AND link.github_number = sqlc.arg(github_number)
ORDER BY link.created_at ASC, link.id ASC;

-- name: GetStoryLinkedGitHubIssues :many
SELECT link.repository_id,
       link.github_number,
       repository.owner_login,
       repository.name AS repository_slug,
       installation.github_installation_id
FROM public.github_story_links AS link
INNER JOIN public.github_repositories AS repository ON repository.id = link.repository_id
INNER JOIN public.github_installations AS installation ON installation.id = repository.installation_id
WHERE link.workspace_id = sqlc.arg(workspace_id)
  AND link.story_id = sqlc.arg(story_id)
  AND link.external_type = 'issue'
  AND link.github_number IS NOT NULL
ORDER BY link.created_at ASC, link.id ASC;

-- name: GetStoryGitHubLinks :many
SELECT link.id,
       link.external_type,
       link.github_number,
       link.url,
       link.title,
       link.state,
       link.review_state,
       link.reviews_approved,
       link.reviews_changes_requested,
       link.check_state,
       repository.full_name AS repository_full_name,
       link.ref_name,
       link.created_at
FROM public.github_story_links AS link
INNER JOIN public.github_repositories AS repository ON repository.id = link.repository_id
WHERE link.workspace_id = sqlc.arg(workspace_id)
  AND link.story_id = sqlc.arg(story_id)
ORDER BY link.created_at DESC, link.id ASC;

-- name: UpsertGitHubStoryLink :exec
INSERT INTO public.github_story_links (
    workspace_id,
    story_id,
    repository_id,
    external_type,
    github_id,
    github_number,
    ref_name,
    url,
    title,
    state,
    metadata,
    last_seen_at,
    updated_at
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(repository_id),
    sqlc.arg(external_type),
    sqlc.arg(github_id),
    sqlc.arg(github_number),
    sqlc.narg(ref_name),
    sqlc.arg(url),
    sqlc.arg(title),
    sqlc.arg(state),
    sqlc.arg(metadata),
    NOW(),
    NOW()
)
ON CONFLICT (
    story_id,
    repository_id,
    external_type,
    (COALESCE(github_id, 0)),
    (COALESCE(ref_name, ''))
) DO UPDATE SET
    url = EXCLUDED.url,
    title = EXCLUDED.title,
    state = EXCLUDED.state,
    metadata = EXCLUDED.metadata,
    last_seen_at = NOW(),
    updated_at = NOW();

-- name: UpdateGitHubStoryLinkSyncState :exec
UPDATE public.github_story_links
SET last_synced_from = sqlc.arg(sync_source),
    last_sync_hash = sqlc.arg(sync_hash),
    updated_at = NOW()
WHERE id = sqlc.arg(link_id);

-- name: UpdateGitHubStoryLinkReviewState :exec
UPDATE public.github_story_links
SET review_state = sqlc.arg(review_state),
    reviews_approved = sqlc.arg(reviews_approved),
    reviews_changes_requested = sqlc.arg(reviews_changes_requested),
    updated_at = NOW()
WHERE story_id = sqlc.arg(story_id)
  AND repository_id = sqlc.arg(repository_id)
  AND external_type = 'pull_request'
  AND github_id = sqlc.arg(github_id);

-- name: UpdateGitHubStoryLinkCheckState :exec
UPDATE public.github_story_links
SET check_state = sqlc.arg(check_state),
    updated_at = NOW()
WHERE story_id = sqlc.arg(story_id)
  AND repository_id = sqlc.arg(repository_id)
  AND external_type = 'pull_request'
  AND github_id = sqlc.arg(github_id);

-- name: DeleteGitHubStoryLink :exec
DELETE FROM public.github_story_links
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(link_id);

-- name: LockStoryURL :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        CAST(CAST(sqlc.arg(story_id) AS uuid) AS text)
            || E'\x1f'
            || CAST(sqlc.arg(url) AS text),
        0
    )
);

-- name: GitHubStoryURLExists :one
SELECT EXISTS (
    SELECT 1
    FROM public.story_links
    WHERE story_id = sqlc.arg(story_id)
      AND url = sqlc.arg(url)
);

-- name: InsertGitHubStoryURL :exec
INSERT INTO public.story_links (title, url, story_id)
VALUES (sqlc.narg(title), sqlc.arg(url), sqlc.arg(story_id));

-- name: GetGitHubStatusCategory :one
SELECT category
FROM public.statuses
WHERE status_id = sqlc.arg(status_id);

-- name: LockGitHubLabelName :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text)
            || E'\x1f'
            || LOWER(CAST(sqlc.arg(label_name) AS text)),
        0
    )
);

-- name: FindGitHubLabelByName :one
SELECT label_id
FROM public.labels
WHERE workspace_id = sqlc.arg(workspace_id)
  AND LOWER(name) = LOWER(sqlc.arg(label_name))
ORDER BY created_at ASC, label_id ASC
LIMIT 1;

-- name: InsertGitHubLabel :one
INSERT INTO public.labels (name, workspace_id, team_id, color, created_at, updated_at)
VALUES (
    sqlc.arg(label_name),
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    '#6B7280',
    NOW(),
    NOW()
)
RETURNING label_id;
