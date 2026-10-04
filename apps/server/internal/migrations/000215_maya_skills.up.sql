CREATE TABLE public.maya_skills (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 80),
    description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 300),
    instructions text NOT NULL CHECK (char_length(btrim(instructions)) BETWEEN 1 AND 12000),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX maya_skills_owner_name_idx
    ON public.maya_skills (workspace_id, user_id, lower(name));
