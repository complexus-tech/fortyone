import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { updateSlackAgentSettingsAction } from "@/lib/actions/slack/update-agent-settings";
import { getSlackAgentSettings } from "@/lib/queries/slack/get-agent-settings";
import type { SlackAgentSettings } from "@/modules/settings/workspace/integrations/slack/types";

export const useSlackAgentSettings = ({ enabled = true } = {}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: enabled && Boolean(session && workspaceSlug),
    queryKey: slackKeys.agentSettings(workspaceSlug),
    queryFn: () => getSlackAgentSettings({ session: session!, workspaceSlug }),
  });
};

export const useUpdateSlackAgentSettings = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: (input: SlackAgentSettings) =>
      updateSlackAgentSettingsAction(workspaceSlug, input),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Slack agent settings", {
          description: response.error.message,
        });
        return;
      }
      if (response.data) {
        queryClient.setQueryData(
          slackKeys.agentSettings(workspaceSlug),
          response.data,
        );
      }
      toast.success("Slack agent settings updated");
    },
  });
};
