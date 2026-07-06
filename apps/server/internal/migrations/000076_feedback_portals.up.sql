CREATE TABLE public.feedback_portals (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    description text NOT NULL DEFAULT '',
    is_public boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_portals_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_portals_workspace_unique
    ON public.feedback_portals (workspace_id);

CREATE TABLE public.feedback_boards (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    portal_id uuid NOT NULL,
    team_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    color text NOT NULL DEFAULT 'green',
    order_index integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_boards_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_boards_portal_id_fkey FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_boards_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT feedback_boards_slug_not_empty CHECK (length(trim(slug)) > 0),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_boards_portal_slug_unique
    ON public.feedback_boards (portal_id, slug);

CREATE INDEX idx_feedback_boards_portal_order
    ON public.feedback_boards (portal_id, order_index ASC, created_at ASC);

CREATE INDEX idx_feedback_boards_team
    ON public.feedback_boards (team_id);

CREATE TABLE public.feedback_items (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    portal_id uuid NOT NULL,
    board_id uuid NOT NULL,
    author_id uuid,
    title text NOT NULL,
    description text NOT NULL DEFAULT '',
    slug text NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    roadmap_summary text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_items_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_items_portal_id_fkey FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_items_board_id_fkey FOREIGN KEY (board_id) REFERENCES public.feedback_boards(id) ON DELETE CASCADE,
    CONSTRAINT feedback_items_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT feedback_items_status_check CHECK (status IN ('pending', 'reviewing', 'planned', 'in_progress', 'completed', 'closed')),
    CONSTRAINT feedback_items_slug_not_empty CHECK (length(trim(slug)) > 0),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_items_portal_slug_unique
    ON public.feedback_items (portal_id, slug);

CREATE INDEX idx_feedback_items_portal_status
    ON public.feedback_items (portal_id, status, created_at DESC);

CREATE INDEX idx_feedback_items_board_status
    ON public.feedback_items (board_id, status, created_at DESC);

CREATE INDEX idx_feedback_items_search
    ON public.feedback_items USING gin (to_tsvector('english', title || ' ' || description || ' ' || slug));

CREATE TABLE public.feedback_votes (
    workspace_id uuid NOT NULL,
    item_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_votes_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_votes_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_votes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, user_id)
);

CREATE TABLE public.feedback_comments (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    item_id uuid NOT NULL,
    author_id uuid,
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_comments_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_comments_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_comments_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_feedback_comments_item_created
    ON public.feedback_comments (item_id, created_at ASC);

CREATE TABLE public.feedback_story_links (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    item_id uuid NOT NULL,
    story_id uuid NOT NULL,
    relationship text NOT NULL DEFAULT 'linked',
    created_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_story_links_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_story_links_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_story_links_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT feedback_story_links_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT feedback_story_links_relationship_check CHECK (relationship IN ('created_from', 'linked', 'solves')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX feedback_story_links_item_story_unique
    ON public.feedback_story_links (item_id, story_id);

CREATE INDEX idx_feedback_story_links_story
    ON public.feedback_story_links (story_id);

CREATE TABLE public.feedback_updates (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    portal_id uuid NOT NULL,
    author_id uuid,
    title text NOT NULL,
    body text NOT NULL,
    status text NOT NULL DEFAULT 'draft',
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_updates_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_updates_portal_id_fkey FOREIGN KEY (portal_id) REFERENCES public.feedback_portals(id) ON DELETE CASCADE,
    CONSTRAINT feedback_updates_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT feedback_updates_status_check CHECK (status IN ('draft', 'published')),
    PRIMARY KEY (id)
);

CREATE TABLE public.feedback_update_items (
    update_id uuid NOT NULL,
    item_id uuid NOT NULL,
    CONSTRAINT feedback_update_items_update_id_fkey FOREIGN KEY (update_id) REFERENCES public.feedback_updates(id) ON DELETE CASCADE,
    CONSTRAINT feedback_update_items_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    PRIMARY KEY (update_id, item_id)
);

INSERT INTO public.feedback_portals (workspace_id, description)
SELECT
    workspace_id,
    'Collect public feedback, prioritize requests, and publish roadmap progress.'
FROM public.workspaces
WHERE deleted_at IS NULL
ON CONFLICT (workspace_id) DO NOTHING;
