CREATE TABLE IF NOT EXISTS public.story_collaborators (
    story_id uuid NOT NULL,
    team_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT story_collaborators_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT story_collaborators_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT story_collaborators_team_member_fkey
        FOREIGN KEY (team_id, user_id) REFERENCES public.team_members(team_id, user_id) ON DELETE CASCADE,
    PRIMARY KEY (story_id, user_id)
);

CREATE INDEX idx_story_collaborators_user_id
    ON public.story_collaborators (user_id, story_id);

CREATE TABLE IF NOT EXISTS public.story_watchers (
    story_id uuid NOT NULL,
    team_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT story_watchers_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT story_watchers_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT story_watchers_team_member_fkey
        FOREIGN KEY (team_id, user_id) REFERENCES public.team_members(team_id, user_id) ON DELETE CASCADE,
    PRIMARY KEY (story_id, user_id)
);

CREATE INDEX idx_story_watchers_user_id
    ON public.story_watchers (user_id, story_id);

CREATE TABLE IF NOT EXISTS public.story_notification_mutes (
    story_id uuid NOT NULL,
    team_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT story_notification_mutes_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT story_notification_mutes_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT story_notification_mutes_team_member_fkey
        FOREIGN KEY (team_id, user_id) REFERENCES public.team_members(team_id, user_id) ON DELETE CASCADE,
    PRIMARY KEY (story_id, user_id)
);

CREATE INDEX idx_story_notification_mutes_user_id
    ON public.story_notification_mutes (user_id, story_id);
