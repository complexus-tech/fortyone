import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "../../constants";
import type { NewObjectiveStatus } from "../../actions/statuses/create";
import { createObjectiveStatusAction } from "../../actions/statuses/create";

export const useCreateObjectiveStatusMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const toastId = "create-objective-status";

  const mutation = useMutation({
    mutationFn: (newStatus: NewObjectiveStatus) =>
      createObjectiveStatusAction(newStatus, workspaceSlug),

    onMutate: () => {
      toast.loading("Please wait...", {
        id: toastId,
        description: "Creating status...",
      });
    },
    onError: (error, variables) => {
      toast.error("Failed to create status", {
        description: error.message || "Your changes were not saved",
        id: toastId,
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(workspaceSlug),
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      toast.success("Status created", {
        id: toastId,
        description: "Your status has been created",
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(workspaceSlug),
      });
    },
  });

  return mutation;
};
