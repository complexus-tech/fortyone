import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { DetailedStory } from "../types";
import { restoreStoryAction } from "../actions/restore-story";

export const useRestoreStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: restoreStoryAction,
    onError: (_, storyId) => {
      toast.error("Failed to restore story", {
        description: "An error occurred while restoring the story",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(storyId),
        },
      });
    },
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
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Story restored successfully",
      });
    },
    onSettled: (storyId) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId!) });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
    },
  });

  return mutation;
};
