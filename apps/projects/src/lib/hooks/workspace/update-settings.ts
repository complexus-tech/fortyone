import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { workspaceKeys } from "@/constants/keys";
import type { WorkspaceSettings } from "@/types";
import type { UpdateWorkspaceSettings } from "../../actions/workspaces/update";
import { updateWorkspaceSettingsAction } from "../../actions/workspaces/update";

export const useUpdateWorkspaceSettingsMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (payload: UpdateWorkspaceSettings) =>
      updateWorkspaceSettingsAction(payload, workspaceSlug),

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.settings(workspaceSlug) });
      const previousSettings = queryClient.getQueryData<WorkspaceSettings>(
        workspaceKeys.settings(workspaceSlug),
      );

      if (previousSettings) {
        queryClient.setQueryData<WorkspaceSettings>(workspaceKeys.settings(workspaceSlug), {
          ...previousSettings,
          ...updates,
        });
      }

      return { previousSettings };
    },

    onError: (error, variables, context) => {
      if (context?.previousSettings) {
        queryClient.setQueryData(
          workspaceKeys.settings(workspaceSlug),
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
      queryClient.invalidateQueries({ queryKey: workspaceKeys.settings(workspaceSlug) });
    },
  });

  return mutation;
};
