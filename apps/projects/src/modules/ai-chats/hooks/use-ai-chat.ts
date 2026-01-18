import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { getAiChat } from "../queries/get-ai-chat";

export const useAiChat = (id: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: aiChatKeys.detail(id),
    queryFn: () => getAiChat({ session: session!, workspaceSlug }, id),
    enabled: Boolean(session && id),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
