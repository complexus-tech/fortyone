import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useChatCapabilities } from "@/lib/hooks/workspace/chat-capabilities";
import { getUnreadSummary } from "./api";
import { chatKeys } from "./constants";

export const useChatUnread = () => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const capabilities = useChatCapabilities();
  const enabled = Boolean(session && capabilities.enabled);

  const query = useQuery({
    queryKey: [...chatKeys.unread(workspaceSlug), session?.user.id],
    queryFn: ({ signal }) => getUnreadSummary({ workspaceSlug }, signal),
    enabled,
    staleTime: 15_000,
  });
  return { ...query, data: enabled ? query.data : undefined };
};
