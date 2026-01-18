import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { userKeys } from "@/constants/keys";
import { updateAutomationPreferencesAction } from "@/lib/actions/users/update-automation-preferences";
import type {
  AutomationPreferences,
  UpdateAutomationPreferences,
} from "@/types";

export const useUpdateAutomationPreferencesMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (payload: UpdateAutomationPreferences) =>
      updateAutomationPreferencesAction(payload, workspaceSlug),

    onMutate: async (payload) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({
        queryKey: userKeys.automationPreferences(workspaceSlug),
      });

      // Get the previous data
      const previousPreferences =
        queryClient.getQueryData<AutomationPreferences>(
          userKeys.automationPreferences(workspaceSlug),
        );

      // Optimistically update to the new value
      if (previousPreferences) {
        queryClient.setQueryData<AutomationPreferences>(
          userKeys.automationPreferences(workspaceSlug),
          {
            ...previousPreferences,
            ...payload,
          },
        );
      }

      return { previousPreferences };
    },

    onError: (error, variables, context) => {
      // Revert to the previous value
      if (context?.previousPreferences) {
        queryClient.setQueryData(
          userKeys.automationPreferences(workspaceSlug),
          context.previousPreferences,
        );
      }

      toast.error("Failed to update automation preferences", {
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

      queryClient.invalidateQueries({
        queryKey: userKeys.automationPreferences(workspaceSlug),
      });
    },
  });

  return mutation;
};
