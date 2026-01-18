import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { getAiChatMessages } from "../queries/get-ai-chat-messages";

export const useAiChatMessages = (id: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: aiChatKeys.messages(id),
    queryFn: () => getAiChatMessages({ session: session!, workspaceSlug }, id),
    enabled: Boolean(session && id),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
