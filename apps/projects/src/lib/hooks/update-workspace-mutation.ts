import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { updateWorkspaceAction } from "@/lib/actions/workspaces/update-workspace";
import type { UpdateWorkspaceInput } from "@/lib/actions/workspaces/update-workspace";
import { workspaceKeys } from "@/constants/keys";

export const useUpdateWorkspaceMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
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
    onError: (error, variables) => {
      toast.error("Failed to update workspace", {
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
