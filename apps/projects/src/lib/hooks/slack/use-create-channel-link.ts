import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { createSlackChannelLinkAction } from "@/lib/actions/slack/create-channel-link";
import type {
  CreateSlackChannelLinkInput,
  SlackIntegration,
} from "@/modules/settings/workspace/integrations/slack/types";

export const useCreateSlackChannelLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (input: CreateSlackChannelLinkInput) =>
      createSlackChannelLinkAction(input, workspaceSlug),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<SlackIntegration>(queryKey);

      queryClient.setQueryData<SlackIntegration>(queryKey, (old) => {
        if (!old) return old;
        const linkedTeam = old.channelLinks.find(
          (link) => link.teamId === input.teamId,
        );
        return {
          ...old,
          channelLinks: [
            ...old.channelLinks,
            {
              id: `draft-${input.slackChannelId}-${input.teamId}`,
              slackChannelId: input.slackChannelId,
              teamId: input.teamId,
              teamCode: linkedTeam?.teamCode ?? "TEAM",
              teamName: linkedTeam?.teamName ?? "Team",
              teamColor: linkedTeam?.teamColor ?? "#6B7280",
              isActive: true,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            },
          ],
        };
      });

      return { previous };
    },
    onError: (_err, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("Slack", { description: "Failed to link channel" });
    },
    onSuccess: (res, _input, context) => {
      if (res.error?.message) {
        if (context?.previous) {
          queryClient.setQueryData(queryKey, context.previous);
        }
        toast.error("Slack", { description: res.error.message });
        return;
      }
      toast.success("Slack channel linked");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
