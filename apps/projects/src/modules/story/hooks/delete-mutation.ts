import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { deleteStoryAction } from "../actions/delete-story";
import type { DetailedStory } from "../types";
import { useRestoreStoryMutation } from "./restore-mutation";

export const useDeleteStoryMutation = () => {
  const queryClient = useQueryClient();

  const { mutateAsync } = useRestoreStoryMutation();

  const mutation = useMutation({
    mutationFn: deleteStoryAction,
    onError: (error, storyId) => {
      toast.error("Failed to delete story", {
        description:
          error.message || "An error occurred while deleting the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(storyId);
          },
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
          deletedAt: new Date().toISOString(),
        });
      }
    },
    onSuccess: (storyId) => {
      toast.success("Success", {
        description: "Story deleted successfully",
        cancel: {
          label: "Undo",
          onClick: () => {
            mutateAsync(storyId);
          },
        },
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
