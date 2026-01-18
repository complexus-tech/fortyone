import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { getTotalMessagesForTheMonth } from "../queries/get-total-messages";

export const useTotalMessages = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: aiChatKeys.totalMessages(),
    queryFn: () => getTotalMessagesForTheMonth({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
  });
};
