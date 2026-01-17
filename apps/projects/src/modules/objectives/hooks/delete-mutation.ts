import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams, useRouter } from "next/navigation";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
import { objectiveKeys } from "../constants";
import { deleteObjective } from "../actions/delete-objective";

export const useDeleteObjectiveMutation = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: (objectiveId: string) => deleteObjective(objectiveId, workspaceSlug),
    onError: (error, objectiveId) => {
      toast.error("Failed to delete objective", {
        description:
          error.message || "An error occurred while deleting the objective",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(objectiveId);
          },
        },
      });
    },
    onSuccess: (res, objectiveId) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      // Track the objective deletion event
      analytics.track("objective_deleted", {
        objectiveId,
      });

      toast.success("Success", {
        description: "Objective deleted successfully",
      });
      queryClient.removeQueries({
        queryKey: objectiveKeys.objective(workspaceSlug, objectiveId),
      });
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list(workspaceSlug) });
      if (!teamId) {
        router.replace("/objectives");
      } else {
        router.replace(`/teams/${teamId}/objectives`);
      }
    },
  });

  return mutation;
};
