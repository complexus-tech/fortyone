import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import { setStoryWatchingAction } from "../actions/set-watching";
import { updateCollaboratorsAction } from "../actions/update-collaborators";

export const useUpdateCollaboratorsMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    scope: { id: `story-collaborators-${workspaceSlug}` },
    mutationFn: async ({
      storyId,
      collaboratorIds,
    }: {
      storyId: string;
      collaboratorIds: string[];
    }) => {
      const response = await updateCollaboratorsAction(
        storyId,
        collaboratorIds,
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },
    onMutate: async ({ storyId, collaboratorIds }) => {
      const queryKey = storyKeys.detail(workspaceSlug, storyId);
      await queryClient.cancelQueries({ queryKey });
      const previousStory = queryClient.getQueryData<DetailedStory>(queryKey);

      if (previousStory) {
        const selected = new Set(collaboratorIds);
        queryClient.setQueryData<DetailedStory>(queryKey, {
          ...previousStory,
          collaboratorIds,
          collaborators: previousStory.collaborators.filter(({ id }) =>
            selected.has(id),
          ),
        });
      }
      return { previousStory };
    },
    onError: (error, { storyId }, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData(
          storyKeys.detail(workspaceSlug, storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to update collaborators", {
        description: error.message || "Your changes were not saved",
      });
    },
    onSuccess: (_response, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.lists(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.activitiesInfinite(workspaceSlug, storyId),
      });
    },
  });
};

export const useSetStoryWatchingMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async ({
      storyId,
      watching,
    }: {
      storyId: string;
      watching: boolean;
    }) => {
      const response = await setStoryWatchingAction(
        storyId,
        watching,
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },
    onError: (error) => {
      toast.error("Failed to update story notifications", {
        description: error.message || "Your preference was not saved",
      });
    },
    onSuccess: (_response, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
    },
  });
};
