import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { resyncSlackChannelsAction } from "@/lib/actions/slack/resync-channels";

export const useResyncSlackChannels = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: () => resyncSlackChannelsAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Slack", { description: res.error.message });
        return;
      }
      toast.success("Slack channels synced");
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
