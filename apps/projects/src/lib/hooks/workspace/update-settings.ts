import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { workspaceKeys } from "@/constants/keys";
import type { WorkspaceSettings } from "@/types";
import type { UpdateWorkspaceSettings } from "../../actions/workspaces/update";
import { updateWorkspaceSettingsAction } from "../../actions/workspaces/update";

export const useUpdateWorkspaceSettingsMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (payload: UpdateWorkspaceSettings) =>
      updateWorkspaceSettingsAction(payload),

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.settings() });
      const previousSettings = queryClient.getQueryData<WorkspaceSettings>(
        workspaceKeys.settings(),
      );

      if (previousSettings) {
        queryClient.setQueryData<WorkspaceSettings>(workspaceKeys.settings(), {
          ...previousSettings,
          ...updates,
        });
      }

      return { previousSettings };
    },

    onError: (error, variables, context) => {
      if (context?.previousSettings) {
        queryClient.setQueryData(
          workspaceKeys.settings(),
          context.previousSettings,
        );
      }
      toast.error("Failed to update workspace settings", {
        description: error.message || "Your changes were not saved",
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
      queryClient.invalidateQueries({ queryKey: workspaceKeys.settings() });
    },
  });

  return mutation;
};
