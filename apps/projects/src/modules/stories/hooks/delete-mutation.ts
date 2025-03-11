import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "../constants";
import { bulkDeleteAction } from "../actions/bulk-delete-stories";
import type { Story } from "../types";
import { useBulkRestoreStoryMutation } from "./restore-mutation";

export const useBulkDeleteStoryMutation = () => {
  const queryClient = useQueryClient();
  const { storyId } = useParams<{ storyId?: string }>();

  const { mutateAsync } = useBulkRestoreStoryMutation();

  const mutation = useMutation({
    mutationFn: bulkDeleteAction,
    onMutate: (storyIds) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          if (queryKey.toLowerCase().includes("detail")) {
            queryClient.setQueriesData(
              { queryKey: query.queryKey },
              (data: DetailedStory | undefined) => {
                if (data?.subStories) {
                  return {
                    ...data,
                    subStories: data.subStories.filter(
                      (story) => !storyIds.includes(story.id),
                    ),
                  };
                }
              },
            );
          } else {
            queryClient.setQueriesData(
              { queryKey: query.queryKey },
              (data: Story[] = []) => {
                return data.filter((story) => !storyIds.includes(story.id));
              },
            );
          }
        }
      });

      return storyIds;
    },
    onError: (error, storyIds) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
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

      if (storyId) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.detail(storyId),
        });
      }

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
      toast.info("You want to undo this action?", {
        description: `${storyIds.length} stor${
          storyIds.length === 1 ? "y" : "ies"
        } deleted`,
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
