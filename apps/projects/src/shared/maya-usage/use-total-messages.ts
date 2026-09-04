import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { mayaUsageKeys } from "./keys";
import { getTotalMessagesForTheMonth } from "./get-total-messages";

export const useTotalMessages = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: mayaUsageKeys.totalMessages(workspaceSlug),
    queryFn: () =>
      getTotalMessagesForTheMonth({ session: session!, workspaceSlug }),
    enabled: Boolean(session && workspaceSlug),
  });
};
