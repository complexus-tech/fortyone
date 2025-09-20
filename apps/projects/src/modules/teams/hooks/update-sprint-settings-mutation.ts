import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { sprintKeys, teamKeys } from "@/constants/keys";
import { updateSprintSettingsAction } from "../actions/update-sprint-settings";
import type { UpdateSprintSettingsInput, TeamSettings } from "../types";

export const useUpdateSprintSettingsMutation = (teamId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (input: UpdateSprintSettingsInput) =>
      updateSprintSettingsAction(teamId, input),
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
          sprintSettings: {
            ...previousSettings.sprintSettings,
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
        description: error.message || "Failed to update sprint settings",
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
      queryClient.invalidateQueries({
        queryKey: teamKeys.lists(),
      });
      queryClient.invalidateQueries({
        queryKey: sprintKeys.running(),
      });
    },
  });

  return mutation;
};
