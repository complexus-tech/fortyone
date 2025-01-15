import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteWorkspaceAction } from "@/lib/actions/workspaces/delete-workspace";
import { workspaceKeys } from "@/constants/keys";
import type { Workspace } from "@/types";

export const useDeleteWorkspaceMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteWorkspaceAction(id),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.lists() });
      const previousWorkspaces = queryClient.getQueryData<Workspace[]>(
        workspaceKeys.lists(),
      );

      if (previousWorkspaces) {
        queryClient.setQueryData<Workspace[]>(
          workspaceKeys.lists(),
          previousWorkspaces.filter((workspace) => workspace.id !== id),
        );
      }

      return { previousWorkspaces };
    },
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: workspaceKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });
};
