import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { resyncSlackChannelsAction } from "@/lib/actions/slack/resync-channels";

export const useResyncSlackChannels = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: () => resyncSlackChannelsAction(workspaceSlug),
    onSuccess: async (response) => {
      if (response.error?.message) {
        toast.error("Slack", { description: response.error.message });
        return;
      }

      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: slackKeys.integration(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: slackKeys.channelAudiences(workspaceSlug),
        }),
      ]);
      toast.success("Slack channels synced");
    },
  });
};
