import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { uploadWorkspaceLogoAction } from "@/lib/actions/workspaces/upload-logo";
import { workspaceKeys } from "@/constants/keys";
import type { Workspace } from "@/types";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";

export const useUploadWorkspaceLogoMutation = () => {
  const queryClient = useQueryClient();
  const { workspace: currentWorkspace } = useCurrentWorkspace();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (file: File) => uploadWorkspaceLogoAction(file, workspaceSlug),
    onMutate: async (file) => {
      await queryClient.cancelQueries({ queryKey: workspaceKeys.lists() });
      const previousWorkspaces = queryClient.getQueryData<Workspace[]>(
        workspaceKeys.lists(),
      );

      if (previousWorkspaces) {
        const tempUrl = URL.createObjectURL(file);
        queryClient.setQueryData(workspaceKeys.lists(), (old: Workspace[]) => {
          return old.map((workspace) =>
            workspace.id === currentWorkspace?.id
              ? { ...workspace, avatarUrl: tempUrl }
              : workspace,
          );
        });
        return { previousWorkspaces, tempUrl };
      }
    },
    onError: (error, variables, context) => {
      if (context?.tempUrl) {
        URL.revokeObjectURL(context.tempUrl);
      }
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
    onSuccess: (res, _, context) => {
      if (context.tempUrl) {
        URL.revokeObjectURL(context.tempUrl);
      }
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });

  return mutation;
};
