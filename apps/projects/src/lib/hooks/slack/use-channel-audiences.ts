import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { resyncSlackChannelsAction } from "@/lib/actions/slack/resync-channels";
import { updateSlackChannelAudienceAction } from "@/lib/actions/slack/update-channel-audience";
import { getSlackChannelAudiences } from "@/lib/queries/slack/get-channel-audiences";

export const useSlackChannelAudiences = ({ enabled = true } = {}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: enabled && Boolean(session && workspaceSlug),
    queryKey: slackKeys.channelAudiences(workspaceSlug),
    queryFn: () =>
      getSlackChannelAudiences({ session: session!, workspaceSlug }),
  });
};

export const useUpdateSlackChannelAudience = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: ({
      channelId,
      teamIds,
    }: {
      channelId: string;
      teamIds: string[];
    }) =>
      updateSlackChannelAudienceAction(workspaceSlug, channelId, { teamIds }),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Slack channel access", {
          description: response.error.message,
        });
        return;
      }
      queryClient.invalidateQueries({
        queryKey: slackKeys.channelAudiences(workspaceSlug),
      });
      toast.success("Slack channel access updated");
    },
  });
};

export const useResyncSlackChannels = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: () => resyncSlackChannelsAction(workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Slack channels", { description: response.error.message });
        return;
      }
      queryClient.invalidateQueries({
        queryKey: slackKeys.channelAudiences(workspaceSlug),
      });
      toast.success("Slack channels refreshed");
    },
  });
};
