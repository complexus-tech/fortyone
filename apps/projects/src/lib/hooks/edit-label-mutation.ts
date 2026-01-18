import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { labelKeys } from "@/constants/keys";
import type { Label } from "@/types";
import { editLabelAction } from "../actions/labels/edit-label";
import type { UpdateLabel } from "../actions/labels/edit-label";

type EditLabelParams = {
  labelId: string;
  updates: UpdateLabel;
};

export const useEditLabelMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({ labelId, updates }: EditLabelParams) =>
      editLabelAction(labelId, updates, workspaceSlug),

    onMutate: async ({ labelId, updates }) => {
      await queryClient.cancelQueries({ queryKey: labelKeys.lists(workspaceSlug) });
      const previousLabels = queryClient.getQueryData<Label[]>(
        labelKeys.lists(workspaceSlug),
      );

      if (previousLabels) {
        queryClient.setQueryData<Label[]>(
          labelKeys.lists(workspaceSlug),
          previousLabels.map((label) =>
            label.id === labelId ? { ...label, ...updates } : label,
          ),
        );
      }

      return { previousLabels };
    },
    onError: (error, variables, context) => {
      if (context?.previousLabels) {
        queryClient.setQueryData(labelKeys.lists(workspaceSlug), context.previousLabels);
      }
      toast.error("Failed to update label", {
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

      queryClient.invalidateQueries({ queryKey: labelKeys.lists(workspaceSlug) });
    },
  });

  return mutation;
};
