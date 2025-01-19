import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { sprintKeys } from "@/constants/keys";
import type { Sprint } from "../types";
import { deleteSprintAction } from "../actions/delete-sprint";

export const useDeleteSprintMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: deleteSprintAction,
    onMutate: async (sprintId) => {
      await queryClient.cancelQueries({ queryKey: sprintKeys.lists() });
      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(),
      );

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(
          sprintKeys.lists(),
          previousSprints.filter((s) => s.id !== sprintId),
        );
      }

      return { previousSprints };
    },
    onError: (_, sprintId) => {
      queryClient.invalidateQueries({ queryKey: sprintKeys.lists() });
      toast.error("Failed to delete sprint", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(sprintId);
          },
        },
      });
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Sprint deleted successfully",
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: sprintKeys.lists() });
    },
  });

  return mutation;
};
