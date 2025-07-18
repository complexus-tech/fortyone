import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { aiChatKeys } from "../constants";
import { getAiChats } from "../queries/get-ai-chats";

export const useAiChats = () => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: aiChatKeys.lists(),
    queryFn: () => getAiChats(session!),
    enabled: Boolean(session),
  });
};
