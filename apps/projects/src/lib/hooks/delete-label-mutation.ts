import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { labelKeys } from "@/constants/keys";
import type { Label } from "@/types";
import { deleteLabelAction } from "../actions/labels/delete-label";

export const useDeleteLabelMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (labelId: string) => deleteLabelAction(labelId),
    onMutate: async (labelId) => {
      await queryClient.cancelQueries({ queryKey: labelKeys.lists() });
      const previousLabels = queryClient.getQueryData<Label[]>(
        labelKeys.lists(),
      );

      if (previousLabels) {
        queryClient.setQueryData<Label[]>(
          labelKeys.lists(),
          previousLabels.filter((label) => label.id !== labelId),
        );
      }

      return { previousLabels };
    },
    onError: (_, variables, context) => {
      if (context?.previousLabels) {
        queryClient.setQueryData(labelKeys.lists(), context.previousLabels);
      }
      toast.error("Failed to delete label", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });
    },
  });

  return mutation;
};
