import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { deleteWorkspaceAction } from "@/lib/actions/workspaces/delete-workspace";
import { workspaceKeys } from "@/constants/keys";
import type { Workspace } from "@/types";

export const useDeleteWorkspaceMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => deleteWorkspaceAction(),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.lists() });
      const previousWorkspaces = queryClient.getQueryData<Workspace[]>(
        workspaceKeys.lists(),
      );

      return { previousWorkspaces };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        workspaceKeys.lists(),
        context?.previousWorkspaces,
      );
      toast.error("Failed to delete workspace", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });

  return mutation;
};
