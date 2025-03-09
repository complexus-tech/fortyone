import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "../constants";
import { bulkDeleteAction } from "../actions/bulk-delete-stories";
import type { Story } from "../types";
import { useBulkRestoreStoryMutation } from "./restore-mutation";

export const useBulkDeleteStoryMutation = () => {
  const queryClient = useQueryClient();

  const { mutateAsync } = useBulkRestoreStoryMutation();

  const mutation = useMutation({
    mutationFn: bulkDeleteAction,
    onMutate: (storyIds) => {
      // Get the relevant query keys
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
      queryClient.invalidateQueries({ queryKey: storyKeys.mine() });
      queryClient.invalidateQueries({ queryKey: storyKeys.sprints() });
      queryClient.invalidateQueries({ queryKey: storyKeys.objectives() });

      // Update each query type individually with proper type checking
      queryClient.setQueriesData<Story[]>(
        { queryKey: storyKeys.lists() },
        (oldData) => {
          if (!oldData || !Array.isArray(oldData)) return oldData;
          return oldData.filter((story) => !storyIds.includes(story.id));
        },
      );

      queryClient.setQueriesData<Story[]>(
        { queryKey: storyKeys.teams() },
        (oldData) => {
          if (!oldData || !Array.isArray(oldData)) return oldData;
          return oldData.filter((story) => !storyIds.includes(story.id));
        },
      );

      queryClient.setQueriesData<Story[]>(
        { queryKey: storyKeys.mine() },
        (oldData) => {
          if (!oldData || !Array.isArray(oldData)) return oldData;
          return oldData.filter((story) => !storyIds.includes(story.id));
        },
      );

      queryClient.setQueriesData<Story[]>(
        { queryKey: storyKeys.sprints() },
        (oldData) => {
          if (!oldData || !Array.isArray(oldData)) return oldData;
          return oldData.filter((story) => !storyIds.includes(story.id));
        },
      );

      queryClient.setQueriesData<Story[]>(
        { queryKey: storyKeys.objectives() },
        (oldData) => {
          if (!oldData || !Array.isArray(oldData)) return oldData;
          return oldData.filter((story) => !storyIds.includes(story.id));
        },
      );

      return storyIds;
    },
    onError: (error, storyIds) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
      queryClient.invalidateQueries({ queryKey: storyKeys.mine() });
      queryClient.invalidateQueries({ queryKey: storyKeys.sprints() });
      queryClient.invalidateQueries({ queryKey: storyKeys.objectives() });

      toast.error("Failed to delete stories", {
        description:
          error.message || "An error occurred while deleting the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(storyIds);
          },
        },
      });
    },
    onSuccess: (_, storyIds) => {
      toast.success("Success", {
        description: `${storyIds.length} Stories deleted successfully`,
        cancel: {
          label: "Undo",
          onClick: () => {
            mutateAsync(storyIds);
          },
        },
      });
    },
    onSettled: (_, __, storyIds) => {
      storyIds.forEach((storyId) => {
        queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId) });
      });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
      queryClient.invalidateQueries({ queryKey: storyKeys.mine() });
      queryClient.invalidateQueries({ queryKey: storyKeys.sprints() });
      queryClient.invalidateQueries({ queryKey: storyKeys.objectives() });
    },
  });

  return mutation;
};
