-- name: BindIntegrationRequestProviderThread :one
WITH bound AS (
    INSERT INTO public.integration_request_threads (
        workspace_id,
        integration_request_id,
        provider,
        external_workspace_id,
        installation_generation,
        external_channel_id,
        external_thread_id,
        external_source_message_id,
        source_url
    )
    SELECT
        sqlc.arg(workspace_id),
        request.id,
        sqlc.arg(provider),
        sqlc.arg(external_workspace_id),
        sqlc.narg(installation_generation),
        sqlc.arg(external_channel_id),
        sqlc.arg(external_thread_id),
        NULLIF(CAST(sqlc.arg(external_source_message_id) AS text), ''),
        sqlc.narg(source_url)
    FROM public.integration_requests AS request
    WHERE request.id = sqlc.arg(request_id)
      AND request.workspace_id = sqlc.arg(workspace_id)
      AND request.provider = sqlc.arg(provider)
    ON CONFLICT (integration_request_id, provider)
    DO UPDATE SET
        external_workspace_id = EXCLUDED.external_workspace_id,
        installation_generation = EXCLUDED.installation_generation,
        external_channel_id = EXCLUDED.external_channel_id,
        external_thread_id = EXCLUDED.external_thread_id,
        external_source_message_id = COALESCE(EXCLUDED.external_source_message_id, integration_request_threads.external_source_message_id),
        source_url = COALESCE(EXCLUDED.source_url, integration_request_threads.source_url),
        updated_at = NOW()
    RETURNING *
)
SELECT
    bound.id,
    bound.workspace_id,
    bound.integration_request_id,
    request.team_id,
    request.accepted_story_id,
    bound.provider,
    bound.external_workspace_id,
    bound.installation_generation,
    bound.external_channel_id,
    bound.external_thread_id,
    bound.external_source_message_id,
    COALESCE(bound.source_url, request.source_url) AS source_url,
    request.title AS request_title,
    bound.created_at,
    bound.updated_at
FROM bound
INNER JOIN public.integration_requests AS request
    ON request.id = bound.integration_request_id;

-- name: HasAuthorizedIntegrationRequestProviderThread :one
SELECT EXISTS (
    SELECT 1
    FROM public.integration_request_threads AS request_thread
    INNER JOIN public.integration_requests AS request
        ON request.id = request_thread.integration_request_id
    WHERE request_thread.workspace_id = sqlc.arg(workspace_id)
      AND request_thread.provider = sqlc.arg(provider)
      AND request_thread.external_workspace_id = sqlc.arg(external_workspace_id)
      AND request_thread.installation_generation = sqlc.arg(installation_generation)
      AND request_thread.external_channel_id = sqlc.arg(external_channel_id)
      AND request_thread.external_thread_id = sqlc.arg(external_thread_id)
      AND (
          EXISTS (
              SELECT 1
              FROM public.team_members AS request_team_member
              WHERE request_team_member.team_id = request.team_id
                AND request_team_member.user_id = sqlc.arg(actor_id)
          )
          OR EXISTS (
              SELECT 1
              FROM public.workspace_members AS request_workspace_member
              WHERE request_workspace_member.workspace_id = request.workspace_id
                AND request_workspace_member.user_id = sqlc.arg(actor_id)
                AND request_workspace_member.role = 'admin'
          )
      )
);

-- name: HasCurrentIntegrationRequestProviderThread :one
SELECT EXISTS (
    SELECT 1
    FROM public.integration_request_threads AS request_thread
    WHERE request_thread.workspace_id = sqlc.arg(workspace_id)
      AND request_thread.provider = sqlc.arg(provider)
      AND request_thread.external_workspace_id = sqlc.arg(external_workspace_id)
      AND request_thread.installation_generation = sqlc.arg(installation_generation)
      AND request_thread.external_channel_id = sqlc.arg(external_channel_id)
      AND request_thread.external_thread_id = sqlc.arg(external_thread_id)
);

-- name: FindIntegrationRequestProviderThread :one
SELECT
    request_thread.id,
    request_thread.workspace_id,
    request_thread.integration_request_id,
    request.team_id,
    request.accepted_story_id,
    request_thread.provider,
    request_thread.external_workspace_id,
    request_thread.installation_generation,
    request_thread.external_channel_id,
    request_thread.external_thread_id,
    request_thread.external_source_message_id,
    COALESCE(request_thread.source_url, request.source_url) AS source_url,
    request.title AS request_title,
    request_thread.created_at,
    request_thread.updated_at
FROM public.integration_request_threads AS request_thread
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
WHERE request_thread.workspace_id = sqlc.arg(workspace_id)
  AND request_thread.integration_request_id = sqlc.arg(request_id)
  AND request_thread.provider = sqlc.arg(provider);

-- name: GetAuthorizedIntegrationRequestProviderThread :one
SELECT
    request_thread.id,
    request_thread.workspace_id,
    request_thread.integration_request_id,
    request.team_id,
    request.accepted_story_id,
    request_thread.provider,
    request_thread.external_workspace_id,
    request_thread.installation_generation,
    request_thread.external_channel_id,
    request_thread.external_thread_id,
    request_thread.external_source_message_id,
    COALESCE(request_thread.source_url, request.source_url) AS source_url,
    request.title AS request_title,
    request_thread.created_at,
    request_thread.updated_at
FROM public.integration_request_threads AS request_thread
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
WHERE request_thread.workspace_id = sqlc.arg(workspace_id)
  AND request_thread.integration_request_id = sqlc.arg(request_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  );

-- name: ListAuthorizedIntegrationRequestProviderThreadsForStory :many
SELECT
    request_thread.id,
    request_thread.workspace_id,
    request_thread.integration_request_id,
    request.team_id,
    request.accepted_story_id,
    request_thread.provider,
    request_thread.external_workspace_id,
    request_thread.installation_generation,
    request_thread.external_channel_id,
    request_thread.external_thread_id,
    request_thread.external_source_message_id,
    COALESCE(request_thread.source_url, request.source_url) AS source_url,
    request.title AS request_title,
    request_thread.created_at,
    request_thread.updated_at
FROM public.integration_request_threads AS request_thread
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
WHERE request_thread.workspace_id = sqlc.arg(workspace_id)
  AND request.accepted_story_id = sqlc.arg(story_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  )
ORDER BY request_thread.created_at ASC, request_thread.id ASC;

-- name: LockAuthorizedIntegrationRequestProviderThread :one
SELECT
    request_thread.id,
    request_thread.workspace_id,
    request_thread.integration_request_id,
    request.team_id,
    request.accepted_story_id,
    request_thread.provider,
    request_thread.external_workspace_id,
    request_thread.installation_generation,
    request_thread.external_channel_id,
    request_thread.external_thread_id,
    request_thread.external_source_message_id,
    COALESCE(request_thread.source_url, request.source_url) AS source_url,
    request.title AS request_title,
    request_thread.created_at,
    request_thread.updated_at
FROM public.integration_request_threads AS request_thread
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
WHERE request_thread.workspace_id = sqlc.arg(workspace_id)
  AND request_thread.integration_request_id = sqlc.arg(request_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  )
FOR UPDATE OF request_thread;

-- name: FindAuthorizedProviderThreadForComment :one
SELECT
    request_thread.id,
    request_thread.workspace_id,
    request_thread.integration_request_id,
    request.team_id,
    request.accepted_story_id,
    request_thread.provider,
    request_thread.external_workspace_id,
    request_thread.installation_generation,
    request_thread.external_channel_id,
    request_thread.external_thread_id,
    request_thread.external_source_message_id,
    COALESCE(request_thread.source_url, request.source_url) AS source_url,
    request.title AS request_title,
    request_thread.created_at,
    request_thread.updated_at
FROM public.integration_request_threads AS request_thread
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
INNER JOIN public.integration_request_comments AS request_comment
    ON request_comment.thread_id = request_thread.id
WHERE request_comment.id = sqlc.arg(comment_id)
  AND request_thread.integration_request_id = sqlc.arg(request_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  );

-- name: IsCurrentIntegrationRequestProviderThreadKnown :one
SELECT EXISTS (
    SELECT 1
    FROM public.integration_request_threads AS request_thread
    INNER JOIN public.integration_requests AS request
        ON request.id = request_thread.integration_request_id
    WHERE request_thread.provider = sqlc.arg(provider)
      AND request_thread.external_workspace_id = sqlc.arg(external_workspace_id)
      AND request_thread.external_channel_id = sqlc.arg(external_channel_id)
      AND request_thread.external_thread_id = sqlc.arg(external_thread_id)
      AND request_thread.installation_generation = sqlc.arg(installation_generation)
);
