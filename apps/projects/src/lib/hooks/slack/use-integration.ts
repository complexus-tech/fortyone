import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getSlackIntegration } from "@/lib/queries/slack/get-integration";

export const useSlackIntegration = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: slackKeys.integration(workspaceSlug),
    queryFn: () => getSlackIntegration({ session: session!, workspaceSlug }),
  });
};
