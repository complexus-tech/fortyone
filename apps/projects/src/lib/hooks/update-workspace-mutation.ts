import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateWorkspaceAction } from "@/lib/actions/workspaces/update-workspace";
import type { UpdateWorkspaceInput } from "@/lib/actions/workspaces/update-workspace";
import type { Workspace } from "@/lib/queries/workspaces/get-workspace";
import { workspaceKeys } from "@/constants/keys";

export const useUpdateWorkspaceMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateWorkspaceInput) =>
      updateWorkspaceAction(id, input),
    onSuccess: (updatedWorkspace: Workspace) => {
      queryClient.setQueryData(workspaceKeys.detail(id), updatedWorkspace);
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });
};
