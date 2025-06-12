import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { updateWorkspaceAction } from "@/lib/actions/workspaces/update-workspace";
import type { UpdateWorkspaceInput } from "@/lib/actions/workspaces/update-workspace";
import { workspaceKeys } from "@/constants/keys";
import type { Workspace } from "@/types";
import { useCurrentWorkspace } from "./workspaces";

export const useUpdateWorkspaceMutation = () => {
  const queryClient = useQueryClient();
  const { workspace: currentWorkspace } = useCurrentWorkspace();

  const mutation = useMutation({
    mutationFn: (input: UpdateWorkspaceInput) => updateWorkspaceAction(input),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.lists() });
      const previousWorkspaces = queryClient.getQueryData<Workspace[]>(
        workspaceKeys.lists(),
      );

      if (previousWorkspaces) {
        queryClient.setQueryData(workspaceKeys.lists(), (old: Workspace[]) => {
          return old.map((workspace) =>
            workspace.id === currentWorkspace?.id
              ? { ...workspace, ...input }
              : workspace,
          );
        });
      }
      return { previousWorkspaces };
    },
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
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });

  return mutation;
};
