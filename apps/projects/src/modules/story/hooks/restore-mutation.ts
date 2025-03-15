import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import { restoreStoryAction } from "../actions/restore-story";

export const useRestoreStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: restoreStoryAction,

    onMutate: (storyId) => {
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          deletedAt: null,
        });
      }

      return { previousStory };
    },
    onError: (error, storyId, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to restore story", {
        description:
          error.message || "An error occurred while restoring the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(storyId);
          },
        },
      });
    },
    onSuccess: (res, storyId) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId) });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
    },
  });

  return mutation;
};
