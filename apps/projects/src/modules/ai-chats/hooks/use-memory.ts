import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { aiChatKeys } from "../constants";
import { getMemory } from "../queries/get-memory";

export const useMemory = () => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: aiChatKeys.memory(),
    queryFn: () => getMemory(session!),
    enabled: Boolean(session),
  });
};
