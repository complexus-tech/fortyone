import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateWorkspaceAction } from "@/lib/actions/workspaces/update-workspace";
import type { UpdateWorkspaceInput } from "@/lib/actions/workspaces/update-workspace";
import type { Workspace } from "@/types";
import { workspaceKeys } from "@/constants/keys";

export const useUpdateWorkspaceMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateWorkspaceInput) => updateWorkspaceAction(input),
    // onMutate: async (input) => {
    //   await queryClient.cancelQueries({ queryKey: workspaceKeys.detail() });
    //   const previousWorkspace = queryClient.getQueryData<Workspace>(
    //     workspaceKeys.detail(),
    //   );

    //   if (previousWorkspace) {
    //     queryClient.setQueryData<Workspace>(workspaceKeys.detail(), {
    //       ...previousWorkspace,
    //       ...input,
    //     });
    //   }

    //   return { previousWorkspace };
    // },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });
};
