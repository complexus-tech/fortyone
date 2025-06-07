import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { teamKeys } from "@/constants/keys";
import { updateStoryAutomationSettingsAction } from "../actions/update-story-automation-settings";
import type {
  UpdateStoryAutomationSettingsInput,
  TeamSettings,
} from "../types";

export const useUpdateStoryAutomationSettingsMutation = (teamId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (input: UpdateStoryAutomationSettingsInput) =>
      updateStoryAutomationSettingsAction(teamId, input),
    onMutate: async (input) => {
      await queryClient.cancelQueries({
        queryKey: teamKeys.settings(teamId),
      });
      const previousSettings = queryClient.getQueryData<TeamSettings>(
        teamKeys.settings(teamId),
      );

      if (previousSettings) {
        queryClient.setQueryData<TeamSettings>(teamKeys.settings(teamId), {
          ...previousSettings,
          storyAutomationSettings: {
            ...previousSettings.storyAutomationSettings,
            ...input,
          },
        });
      }

      return { previousSettings };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        teamKeys.settings(teamId),
        context?.previousSettings,
      );
      toast.error("Error", {
        description:
          error.message || "Failed to update story automation settings",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({
        queryKey: teamKeys.settings(teamId),
      });
    },
  });

  return mutation;
};
