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
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.detail(id) });
      const previousWorkspace = queryClient.getQueryData<Workspace>(
        workspaceKeys.detail(id),
      );

      if (previousWorkspace) {
        queryClient.setQueryData<Workspace>(workspaceKeys.detail(id), {
          ...previousWorkspace,
          ...input,
        });
      }

      return { previousWorkspace };
    },
    onSuccess: (updatedWorkspace: Workspace) => {
      queryClient.setQueryData(workspaceKeys.detail(id), updatedWorkspace);
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });
};
