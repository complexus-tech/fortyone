import {
  useIsMutating,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
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

export const useUpdateSlackChannelAudience = (channelId?: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.channelAudiences(workspaceSlug);
  const mutationKey = [...queryKey, "update"] as const;
  const pendingChannelUpdates = useIsMutating({
    mutationKey,
    predicate: (candidate) => {
      if (!channelId) return true;
      const variables = candidate.state.variables as
        | UpdateSlackChannelAudienceInput
        | undefined;
      return variables?.channelId === channelId;
    },
  });

  const mutation = useMutation({
    mutationFn: async (input: UpdateSlackChannelAudienceInput) => {
      const response = await updateSlackChannelAudienceAction(
        workspaceSlug,
        input,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey });
      const previousAudience = queryClient
        .getQueryData<SlackChannelAudience[]>(queryKey)
        ?.find(
          (audience) => audience.channel.slackChannelId === input.channelId,
        );

      queryClient.setQueryData<SlackChannelAudience[]>(queryKey, (current) =>
        current?.map((audience) =>
          audience.channel.slackChannelId === input.channelId
            ? {
                ...audience,
                isConfigured: input.isConfigured,
                teamIds: input.teamIds,
              }
            : audience,
        ),
      );

      return { previousAudience };
    },
    onError: (error, input, context) => {
      const previousAudience = context?.previousAudience;
      if (previousAudience) {
        queryClient.setQueryData<SlackChannelAudience[]>(queryKey, (current) =>
          current?.map((audience) =>
            audience.channel.slackChannelId === input.channelId
              ? previousAudience
              : audience,
          ),
        );
      }
      toast.error("Slack channel access", {
        description: error.message,
      });
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey }),
    mutationKey,
    scope: channelId
      ? { id: `slack-channel-audience:${workspaceSlug}:${channelId}` }
      : undefined,
  });

  return {
    ...mutation,
    isPending: mutation.isPending || pendingChannelUpdates > 0,
  };
};
