import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { aiChatKeys } from "../constants";
import { getMemories } from "../queries/get-memory";

export const useMemories = () => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: aiChatKeys.memories(),
    queryFn: () => getMemories(session!),
    enabled: Boolean(session),
  });
};
