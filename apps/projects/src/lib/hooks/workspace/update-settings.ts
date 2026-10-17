import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { chatCapabilityKeys, workspaceKeys } from "@/constants/keys";
import type { WorkspaceSettings } from "@/types";
import type { UpdateWorkspaceSettings } from "../../actions/workspaces/update";
import { updateWorkspaceSettingsAction } from "../../actions/workspaces/update";

export const useUpdateWorkspaceSettingsMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: async (payload: UpdateWorkspaceSettings) => {
      const response = await updateWorkspaceSettingsAction(
        payload,
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },

    onMutate: async (updates) => {
      await queryClient.cancelQueries({
        queryKey: workspaceKeys.settings(workspaceSlug),
      });
      const previousSettings = queryClient.getQueryData<WorkspaceSettings>(
        workspaceKeys.settings(workspaceSlug),
      );

      if (previousSettings) {
        queryClient.setQueryData<WorkspaceSettings>(
          workspaceKeys.settings(workspaceSlug),
          {
            ...previousSettings,
            ...updates,
          },
        );
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

    onSettled: (_data, _error, updates) => {
      void queryClient.invalidateQueries({
        queryKey: workspaceKeys.settings(workspaceSlug),
      });
      if (updates.chatEnabled !== undefined) {
        // Capabilities remain the authoritative chat and attachment boundary.
        // Recheck even after a failed response: the server may have committed.
        void queryClient.invalidateQueries({
          queryKey: chatCapabilityKeys.workspace(workspaceSlug),
        });
      }
    },
  });

  return mutation;
};
