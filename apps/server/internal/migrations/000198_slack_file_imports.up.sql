ALTER TABLE public.attachments
    ADD COLUMN slack_file_import_id uuid UNIQUE;

ALTER TABLE public.messaging_story_mutation_confirmations
    DROP CONSTRAINT messaging_story_mutation_confirmations_operation_check;

ALTER TABLE public.messaging_story_mutation_confirmations
    ADD CONSTRAINT messaging_story_mutation_confirmations_operation_check
        CHECK (operation IN (
            'create_story',
            'create_stories',
            'update_story',
            'add_story_comment',
            'add_story_relationship',
            'attach_story_file'
        ));

CREATE TABLE public.slack_file_imports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    slack_workspace_id uuid NOT NULL REFERENCES public.slack_workspaces(id) ON DELETE CASCADE,
    installation_generation uuid NOT NULL,
    slack_team_id text NOT NULL,
    slack_user_id text NOT NULL,
    slack_channel_id text NOT NULL,
    slack_thread_ts text NOT NULL DEFAULT '',
    slack_message_ts text NOT NULL DEFAULT '',
    slack_file_id text NOT NULL,
    idempotency_key text NOT NULL,
    story_id uuid NOT NULL REFERENCES public.stories(id) ON DELETE CASCADE,
    actor_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    attachment_id uuid REFERENCES public.attachments(attachment_id) ON DELETE SET NULL,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    lease_until timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_file_imports_status_check
        CHECK (status IN ('pending', 'processing', 'complete', 'failed', 'cancelled')),
    CONSTRAINT slack_file_imports_attempt_count_check CHECK (attempt_count >= 0),
    CONSTRAINT slack_file_imports_identity_key UNIQUE (workspace_id, story_id, slack_file_id),
    CONSTRAINT slack_file_imports_idempotency_key UNIQUE (workspace_id, idempotency_key)
);

CREATE INDEX slack_file_imports_recovery_idx
    ON public.slack_file_imports (updated_at, id)
    WHERE status IN ('pending', 'failed', 'processing');
