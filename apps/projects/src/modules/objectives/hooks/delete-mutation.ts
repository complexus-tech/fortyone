import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams, useRouter } from "next/navigation";
import { objectiveKeys } from "../constants";
import { deleteObjective } from "../actions/delete-objective";

export const useDeleteObjectiveMutation = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: deleteObjective,
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
    onSuccess: (_, objectiveId) => {
      toast.success("Success", {
        description: "Objective deleted successfully",
      });
      queryClient.removeQueries({
        queryKey: objectiveKeys.objective(objectiveId),
      });
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list() });
      router.replace(`/teams/${teamId}/objectives`);
    },
  });

  return mutation;
};
