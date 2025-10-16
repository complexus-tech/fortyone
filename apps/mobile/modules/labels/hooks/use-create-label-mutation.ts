import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { labelKeys } from "@/constants/keys";
import type { Label } from "@/types";
import { createLabel } from "../actions/create-label";

export const useCreateLabelMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createLabel,

    onMutate: async (newLabel) => {
      await queryClient.cancelQueries({ queryKey: labelKeys.lists() });
      const previousLabels = queryClient.getQueryData<Label[]>(
        labelKeys.lists()
      );

      if (previousLabels) {
        const optimisticLabel: Label = {
          id: `temp-${Date.now()}`,
          name: newLabel.name,
          color: newLabel.color,
          teamId: newLabel.teamId || null,
          workspaceId: "",
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };

        queryClient.setQueryData<Label[]>(labelKeys.lists(), [
          ...previousLabels,
          optimisticLabel,
        ]);
      }

      return { previousLabels };
    },

    onSuccess: () => {
      toast.success("Label created");
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });
    },

    onError: (error, variables, context) => {
      if (context?.previousLabels) {
        queryClient.setQueryData(labelKeys.lists(), context.previousLabels);
      }

      toast.error("Failed to create label", {
        description: error.message || "Please try again",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
        },
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });
    },
  });

  return mutation;
};
