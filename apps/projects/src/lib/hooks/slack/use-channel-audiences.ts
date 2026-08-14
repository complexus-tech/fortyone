import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { updateSlackChannelAudienceAction } from "@/lib/actions/slack/update-channel-audience";
import { useSession } from "@/lib/auth/client";
import { getSlackChannelAudiences } from "@/lib/queries/slack/get-channel-audiences";
import type {
  SlackChannelAudience,
  UpdateSlackChannelAudienceInput,
} from "@/modules/settings/workspace/integrations/slack/types";

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
  const queryKey = slackKeys.channelAudiences(workspaceSlug);

  return useMutation({
    mutationFn: (input: UpdateSlackChannelAudienceInput) =>
      updateSlackChannelAudienceAction(workspaceSlug, input),
    onError: () => {
      toast.error("Slack channel access", {
        description: "FortyOne could not update this channel. Try again.",
      });
    },
    onSuccess: (response, input) => {
      if (response.error?.message) {
        toast.error("Slack channel access", {
          description: response.error.message,
        });
        return;
      }

      queryClient.setQueryData<SlackChannelAudience[]>(queryKey, (current) =>
        current?.map((audience) =>
          audience.channel.slackChannelId === input.channelId
            ? { ...audience, teamIds: input.teamIds }
            : audience,
        ),
      );
      toast.success("Slack channel access updated");
    },
  });
};
