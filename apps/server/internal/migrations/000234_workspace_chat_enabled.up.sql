-- Keep the chat switch with the other workspace features. A constant default
-- enables existing settings rows and newly initialized workspaces alike.
ALTER TABLE public.workspace_settings
    ADD COLUMN chat_enabled boolean NOT NULL DEFAULT TRUE;

-- Initialize legacy workspaces once, so later toggles and mutation locks share
-- the same settings row. Never perform this write from the authorization path.
INSERT INTO public.workspace_settings (workspace_id)
SELECT workspace_id FROM public.workspaces
ON CONFLICT (workspace_id) DO NOTHING;

CREATE OR REPLACE FUNCTION public.can_read_chat_conversation(p_workspace_id uuid, p_user_id uuid, p_conversation_id uuid)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS (
        SELECT 1 FROM public.chat_conversations c
        JOIN public.workspaces w ON w.workspace_id = c.workspace_id AND w.deleted_at IS NULL
        JOIN public.workspace_members wm ON wm.workspace_id = w.workspace_id AND wm.user_id = p_user_id
        JOIN public.users u ON u.user_id = wm.user_id AND u.is_active AND NOT u.is_system
        LEFT JOIN public.workspace_settings settings ON settings.workspace_id = w.workspace_id
        WHERE c.id = p_conversation_id AND c.workspace_id = p_workspace_id
          AND COALESCE(settings.chat_enabled, TRUE)
          AND ((c.kind = 'team' AND EXISTS (SELECT 1 FROM public.team_members tm WHERE tm.team_id = c.team_id AND tm.user_id = p_user_id))
            OR (c.kind = 'dm' AND p_user_id IN (c.dm_user_low, c.dm_user_high)))
    );
$$;
