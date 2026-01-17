import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { getAiChats } from "../queries/get-ai-chats";

export const useAiChats = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: aiChatKeys.lists(),
    queryFn: () => getAiChats({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
  });
};
