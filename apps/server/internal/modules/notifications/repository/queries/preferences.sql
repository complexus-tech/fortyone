-- GetNotificationPreferences creates the typed default document on first read,
-- but only for an active workspace member in a live workspace.
-- name: GetNotificationPreferences :one
WITH authorized_actor AS (
    SELECT actor.user_id, workspace.workspace_id
    FROM public.users AS actor
    INNER JOIN public.workspace_members AS membership
        ON membership.user_id = actor.user_id
       AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
       AND membership.role IN ('admin', 'member', 'guest')
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
), inserted_preferences AS (
    INSERT INTO public.notification_preferences (
        user_id,
        workspace_id,
        preferences
    )
    SELECT
        authorized.user_id,
        authorized.workspace_id,
        CAST(sqlc.arg(default_preferences) AS jsonb)
    FROM authorized_actor AS authorized
    ON CONFLICT (user_id, workspace_id) DO NOTHING
    RETURNING
        preference_id,
        user_id,
        workspace_id,
        preferences,
        created_at,
        updated_at
)
SELECT
    inserted.preference_id,
    inserted.user_id,
    inserted.workspace_id,
    inserted.preferences,
    inserted.created_at,
    inserted.updated_at
FROM inserted_preferences AS inserted
UNION ALL
SELECT
    preference.preference_id,
    preference.user_id,
    preference.workspace_id,
    preference.preferences,
    preference.created_at,
    preference.updated_at
FROM public.notification_preferences AS preference
INNER JOIN authorized_actor AS authorized
    ON authorized.user_id = preference.user_id
   AND authorized.workspace_id = preference.workspace_id
WHERE NOT EXISTS (SELECT 1 FROM inserted_preferences)
LIMIT 1;

-- UpdateNotificationPreference applies one presence-aware channel patch inside
-- PostgreSQL. Concurrent updates to different preference types or channels do
-- not overwrite the rest of the JSON document.
-- name: UpdateNotificationPreference :one
WITH authorized_actor AS (
    SELECT actor.user_id, workspace.workspace_id
    FROM public.users AS actor
    INNER JOIN public.workspace_members AS membership
        ON membership.user_id = actor.user_id
       AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
       AND membership.role IN ('admin', 'member', 'guest')
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
), desired_channels AS (
    SELECT
        authorized.user_id,
        authorized.workspace_id,
        CAST(sqlc.arg(default_preferences) AS jsonb) AS defaults,
        CAST(sqlc.arg(preference_type) AS text) AS preference_type
    FROM authorized_actor AS authorized
), inserted_or_updated AS (
    INSERT INTO public.notification_preferences (
        user_id,
        workspace_id,
        preferences,
        updated_at
    )
    SELECT
        desired.user_id,
        desired.workspace_id,
        desired.defaults || jsonb_build_object(
            desired.preference_type,
            COALESCE(desired.defaults -> desired.preference_type, CAST('{}' AS jsonb))
                || jsonb_build_object(
                    'email', CASE
                        WHEN CAST(sqlc.arg(email_present) AS boolean)
                            THEN CAST(sqlc.arg(email_enabled) AS boolean)
                        ELSE COALESCE(
                            CAST(desired.defaults -> desired.preference_type ->> 'email' AS boolean),
                            TRUE
                        )
                    END,
                    'in_app', CASE
                        WHEN NOT CAST(sqlc.arg(supports_in_app) AS boolean) THEN FALSE
                        WHEN CAST(sqlc.arg(in_app_present) AS boolean)
                            THEN CAST(sqlc.arg(in_app_enabled) AS boolean)
                        ELSE COALESCE(
                            CAST(desired.defaults -> desired.preference_type ->> 'in_app' AS boolean),
                            TRUE
                        )
                    END
                )
        ),
        CAST(sqlc.arg(updated_at) AS timestamptz)
    FROM desired_channels AS desired
    ON CONFLICT (user_id, workspace_id) DO UPDATE
    SET
        preferences = notification_preferences.preferences || jsonb_build_object(
            CAST(sqlc.arg(preference_type) AS text),
            COALESCE(
                notification_preferences.preferences -> CAST(sqlc.arg(preference_type) AS text),
                CAST(sqlc.arg(default_preferences) AS jsonb) -> CAST(sqlc.arg(preference_type) AS text),
                CAST('{}' AS jsonb)
            ) || jsonb_build_object(
                'email', CASE
                    WHEN CAST(sqlc.arg(email_present) AS boolean)
                        THEN CAST(sqlc.arg(email_enabled) AS boolean)
                    ELSE COALESCE(
                        CAST(
                            notification_preferences.preferences
                                -> CAST(sqlc.arg(preference_type) AS text)
                                ->> 'email'
                            AS boolean
                        ),
                        CAST(
                            CAST(sqlc.arg(default_preferences) AS jsonb)
                                -> CAST(sqlc.arg(preference_type) AS text)
                                ->> 'email'
                            AS boolean
                        ),
                        TRUE
                    )
                END,
                'in_app', CASE
                    WHEN NOT CAST(sqlc.arg(supports_in_app) AS boolean) THEN FALSE
                    WHEN CAST(sqlc.arg(in_app_present) AS boolean)
                        THEN CAST(sqlc.arg(in_app_enabled) AS boolean)
                    ELSE COALESCE(
                        CAST(
                            notification_preferences.preferences
                                -> CAST(sqlc.arg(preference_type) AS text)
                                ->> 'in_app'
                            AS boolean
                        ),
                        CAST(
                            CAST(sqlc.arg(default_preferences) AS jsonb)
                                -> CAST(sqlc.arg(preference_type) AS text)
                                ->> 'in_app'
                            AS boolean
                        ),
                        TRUE
                    )
                END
            )
        ),
        updated_at = CAST(sqlc.arg(updated_at) AS timestamptz)
    RETURNING
        preference_id,
        user_id,
        workspace_id,
        preferences,
        created_at,
        updated_at
)
SELECT
    result.preference_id,
    result.user_id,
    result.workspace_id,
    result.preferences,
    result.created_at,
    result.updated_at
FROM inserted_or_updated AS result;
