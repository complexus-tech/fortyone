import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { InfiniteData } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
} from "@/modules/stories/types";
import { deleteStoryAction } from "../actions/delete-story";
import type { DetailedStory } from "../types";
import { useRestoreStoryMutation } from "./restore-mutation";

const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: InfiniteData<GroupStoriesResponse> | undefined) => {
      if (!data?.pages) return data;
      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          stories: page.stories.filter((story) => story.id !== storyId),
        })),
      };
    },
  );
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: GroupedStoriesResponse | undefined) => {
      if (!data) return data;
      return {
        ...data,
        groups: data.groups.map((group) => ({
          ...group,
          stories: group.stories.filter((story) => story.id !== storyId),
          totalCount: Math.max(0, group.totalCount - 1),
          loadedCount: Math.max(0, group.loadedCount - 1),
        })),
      };
    },
  );
};

const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, storyId);
  } else {
    updateGroupedQuery(queryClient, queryKey, storyId);
  }
};

export const useDeleteStoryMutation = () => {
  const queryClient = useQueryClient();
  const { mutateAsync } = useRestoreStoryMutation();

  const mutation = useMutation({
    mutationFn: deleteStoryAction,

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

      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail") &&
          query.isActive()
        ) {
          updateListQuery(queryClient, query.queryKey, storyId);
        }
      });

      return { previousStory };
    },

    onError: (error, storyId, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(storyId),
          context.previousStory,
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });

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

    onSuccess: (res, storyId) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
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
  });

  return mutation;
};
