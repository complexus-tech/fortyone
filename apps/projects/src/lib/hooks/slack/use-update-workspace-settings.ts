import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { updateSlackWorkspaceSettingsAction } from "@/lib/actions/slack/update-workspace-settings";
import type {
  SlackIntegration,
  UpdateSlackWorkspaceSettingsInput,
} from "@/modules/settings/workspace/integrations/slack/types";

export const useUpdateSlackWorkspaceSettings = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (input: UpdateSlackWorkspaceSettingsInput) =>
      updateSlackWorkspaceSettingsAction(input, workspaceSlug),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<SlackIntegration>(queryKey);

      queryClient.setQueryData<SlackIntegration>(queryKey, (old) => {
        if (!old) return old;
        return {
          ...old,
          settings: {
            ...old.settings,
            ...input,
          },
        };
      });

      return { previous };
    },
    onError: (_err, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("Slack", { description: "Failed to update settings" });
    },
    onSuccess: (res, _input, context) => {
      if (res.error?.message) {
        if (context?.previous) {
          queryClient.setQueryData(queryKey, context.previous);
        }
        toast.error("Slack", { description: res.error.message });
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
