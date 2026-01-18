import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { labelKeys } from "@/constants/keys";
import type { Label } from "@/types";
import type { NewLabel } from "../actions/labels/create-label";
import { createLabelAction } from "../actions/labels/create-label";

export const useCreateLabelMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (newLabel: NewLabel) =>
      createLabelAction(newLabel, workspaceSlug),
    onError: (error, variables) => {
      toast.error("Failed to create label", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: labelKeys.lists(workspaceSlug),
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      const newLabel = res.data;
      const previousLabels = queryClient.getQueryData<Label[]>(
        labelKeys.lists(workspaceSlug),
      );
      if (previousLabels && newLabel) {
        queryClient.setQueryData<Label[]>(labelKeys.lists(workspaceSlug), [
          ...previousLabels,
          newLabel,
        ]);
      }
      queryClient.invalidateQueries({ queryKey: labelKeys.lists(workspaceSlug) });
    },
  });

  return mutation;
};
