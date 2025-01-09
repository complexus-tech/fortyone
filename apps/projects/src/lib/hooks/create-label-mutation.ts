import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { labelKeys } from "@/constants/keys";
import type { Label } from "@/types";
import type { NewLabel } from "../actions/labels/create-label";
import { createLabelAction } from "../actions/labels/create-label";

export const useCreateLabelMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (newLabel: NewLabel) => createLabelAction(newLabel),
    onError: (_, variables) => {
      toast.error("Failed to create label", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => { mutation.mutate(variables); },
        },
      });
      queryClient.invalidateQueries({
        queryKey: labelKeys.lists(),
      });
    },
    onSettled: (newLabel) => {
      const previousLabels = queryClient.getQueryData<Label[]>(
        labelKeys.lists(),
      );
      if (previousLabels) {
        queryClient.setQueryData<Label[]>(labelKeys.lists(), [
          ...previousLabels,
          newLabel!,
        ]);
      }
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });
    },
  });

  return mutation;
};
