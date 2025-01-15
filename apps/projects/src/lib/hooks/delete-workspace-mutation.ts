import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteWorkspaceAction } from "@/lib/actions/workspaces/delete-workspace";
import { workspaceKeys } from "@/constants/keys";

export const useDeleteWorkspaceMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteWorkspaceAction(id),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: workspaceKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });
};
