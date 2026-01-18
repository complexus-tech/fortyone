import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { workspaceKeys } from "@/constants/keys";
import type { Workspace } from "@/types";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { deleteWorkspaceLogoAction } from "@/lib/actions/workspaces/delete-logo";

export const useDeleteWorkspaceLogoMutation = () => {
  const queryClient = useQueryClient();
  const { workspace: currentWorkspace } = useCurrentWorkspace();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: () => deleteWorkspaceLogoAction(workspaceSlug),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.lists() });
      const previousWorkspaces = queryClient.getQueryData<Workspace[]>(
        workspaceKeys.lists(),
      );

      if (previousWorkspaces) {
        queryClient.setQueryData(workspaceKeys.lists(), (old: Workspace[]) => {
          return old.map((workspace) =>
            workspace.id === currentWorkspace?.id
              ? { ...workspace, avatarUrl: null }
              : workspace,
          );
        });
        return { previousWorkspaces };
      }
    },
    onError: (error, variables, context) => {
      if (context?.previousWorkspaces) {
        queryClient.setQueryData(
          workspaceKeys.lists(),
          context.previousWorkspaces,
        );
      }
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
