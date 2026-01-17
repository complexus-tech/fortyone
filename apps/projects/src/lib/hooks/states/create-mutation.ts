import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { useWorkspacePath } from "@/hooks";
import { statusKeys } from "@/constants/keys";
import type { NewState } from "../../actions/states/create";
import { createStateAction } from "../../actions/states/create";
import { State } from "@/types/states";

export const useCreateStateMutation = () => {
  const queryClient = useQueryClient();
  const { teamId } = useParams<{ teamId: string }>();
  const { workspaceSlug } = useWorkspacePath();
  const toastId = "create-state";

  const mutation = useMutation({
    mutationFn: (newState: NewState) =>
      createStateAction(newState, workspaceSlug),
    onMutate: (newState) => {
      // Optimistically add the new state to the cache
      const previousStates = queryClient.getQueryData<State[]>(
        statusKeys.team(workspaceSlug, teamId),
      );

      const tempId = `temp-state-${Date.now()}`;
      const optimisticState: State = {
        id: tempId, // Temporary ID for optimistic update
        name: newState.name,
        category: newState.category,
        color: newState.color,
        isDefault: false,
        orderIndex: 9999, // Will be set by backend
        teamId: newState.teamId,
        workspaceId: "", // Will be set by backend
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      if (previousStates) {
        queryClient.setQueryData<State[]>(statusKeys.team(workspaceSlug, teamId), [
          ...previousStates,
          optimisticState,
        ]);
      }

      toast.loading("Please wait...", {
        id: toastId,
        description: "Creating state...",
      });

      return { previousStates, optimisticState, tempId };
    },

    onError: (error, variables, context) => {
      // Rollback optimistic update on error
      if (context?.previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.team(workspaceSlug, teamId),
          context.previousStates,
        );
      }

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
    onSuccess: (res, variables, context) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      const createdState = res.data!;

      // Replace the optimistic state with the real state from server
      const currentStates = queryClient.getQueryData<State[]>(
        statusKeys.team(workspaceSlug, teamId),
      );
      if (currentStates) {
        const updatedStates = currentStates.map((state) =>
          state.id === context?.tempId ? createdState : state,
        );
        queryClient.setQueryData<State[]>(
          statusKeys.team(workspaceSlug, teamId),
          updatedStates,
        );
      }

      toast.success("State created", {
        id: toastId,
        description: "Your state has been created",
      });

      // Invalidate other related queries
      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(workspaceSlug),
      });
    },
  });

  return mutation;
};
