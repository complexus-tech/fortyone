import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { deleteSlackChannelLinkAction } from "@/lib/actions/slack/delete-channel-link";
import type { SlackIntegration } from "@/modules/settings/workspace/integrations/slack/types";

export const useDeleteSlackChannelLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (linkId: string) =>
      deleteSlackChannelLinkAction(linkId, workspaceSlug),
    onMutate: async (linkId) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<SlackIntegration>(queryKey);

      queryClient.setQueryData<SlackIntegration>(queryKey, (old) => {
        if (!old) return old;
        return {
          ...old,
          channelLinks: old.channelLinks.filter((item) => item.id !== linkId),
        };
      });

      return { previous };
    },
    onError: (_err, _linkId, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("Slack", { description: "Failed to unlink channel" });
    },
    onSuccess: (res, _linkId, context) => {
      if (res.error?.message) {
        if (context?.previous) {
          queryClient.setQueryData(queryKey, context.previous);
        }
        toast.error("Slack", { description: res.error.message });
        return;
      }
      toast.success("Slack channel unlinked");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
