import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { getMemories } from "../queries/get-memory";

export const useMemories = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: aiChatKeys.memories(),
    queryFn: () => getMemories({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
  });
};
