import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { DetailedStory } from "@/modules/story/types";
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
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          queryClient.setQueriesData(
            { queryKey: query.queryKey },
            (data: Story[]) => {
              return data.filter((story) => !storyIds.includes(story.id));
            },
          );
        }
      });

      return storyIds;
    },
    onError: (error, storyIds) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
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
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      storyIds.forEach((storyId) => {
        queryClient.setQueriesData(
          { queryKey: storyKeys.detail(storyId) },
          (oldData: DetailedStory) => {
            return {
              ...oldData,
              deletedAt: new Date().toISOString(),
            };
          },
        );
      });

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
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
  });

  return mutation;
};
