import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { restoreWorkspaceAction } from "@/lib/actions/workspaces/restore-workspace";
import { workspaceKeys } from "@/constants/keys";

export const useRestoreWorkspaceMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => restoreWorkspaceAction(),
    onError: (error, variables) => {
      toast.error("Failed to restore workspace", {
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
      toast.success("Workspace restored", {
        description: "Your workspace has been successfully restored",
      });
      queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
    },
  });

  return mutation;
};
