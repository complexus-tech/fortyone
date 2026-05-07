import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getSlackRequestLogs } from "@/lib/queries/slack/get-request-logs";

export const useSlackRequestLogs = (limit = 20) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: slackKeys.logs(workspaceSlug, limit),
    queryFn: () => getSlackRequestLogs({ session: session!, workspaceSlug }, limit),
  });
};
