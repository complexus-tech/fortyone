import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { aiChatKeys } from "../constants";
import { getAiChatMessages } from "../queries/get-ai-chat-messages";

export const useAiChatMessages = (id: string) => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: aiChatKeys.messages(id),
    queryFn: () => getAiChatMessages(session!, id),
    enabled: Boolean(session && id),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
};
