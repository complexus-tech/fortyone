import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { statusKeys } from "@/constants/keys";
import type { NewState } from "../../actions/states/create";
import { createStateAction } from "../../actions/states/create";

export const useCreateStateMutation = () => {
  const queryClient = useQueryClient();
  const { teamId } = useParams<{ teamId: string }>();
  const toastId = "create-state";

  const mutation = useMutation({
    mutationFn: (newState: NewState) => createStateAction(newState),
    onMutate: () => {
      toast.loading("Please wait...", {
        id: toastId,
        description: "Creating state...",
      });
    },

    onError: (error, variables) => {
      toast.error("Failed to create state", {
        description: error.message || "Your changes were not saved",
        id: toastId,
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
      toast.success("State created", {
        id: toastId,
        description: "Your state has been created",
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.team(teamId),
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(),
      });
    },
  });

  return mutation;
};
