import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { InfiniteData } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
} from "@/modules/stories/types";
import type { SearchResponse } from "@/modules/search/types";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),

    onMutate: ({ storyId, payload }) => {
      queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );

      const activeQueries = queryClient.getQueryCache().getAll();

      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyId, payload);
          } else {
            updateListQuery(queryClient, query.queryKey, storyId, payload);
          }
        }
      });

      updateSearchResults(queryClient, storyId, payload);

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          ...payload,
        });
        return { previousStory };
      }
    },

    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(variables.storyId),
          context.previousStory,
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      toast.error("Failed to update story", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res, { storyId, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("story_updated", {
        storyId,
        ...payload,
      });

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
    },
  });

  return mutation;
};

const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const parentStory = queryClient.getQueryData<DetailedStory>(queryKey);
  if (parentStory?.subStories) {
    queryClient.setQueryData<DetailedStory>(queryKey, {
      ...parentStory,
      subStories: parentStory.subStories.map((subStory) =>
        subStory.id === storyId ? { ...subStory, ...payload } : subStory,
      ),
    });
  }
};

const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, storyId, payload);
  } else {
    updateGroupedQuery(queryClient, queryKey, storyId, payload);
  }
};

const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueryData<InfiniteData<GroupStoriesResponse>>(
    queryKey,
    (data) => {
      if (!data?.pages) return data;

      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          stories: page.stories.map((story) =>
            story.id === storyId ? { ...story, ...payload } : story,
          ),
        })),
      };
    },
  );
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueryData<GroupedStoriesResponse>(queryKey, (data) => {
    if (!data) return data;

    return {
      ...data,
      groups: data.groups.map((group) => ({
        ...group,
        stories: group.stories.map((story) =>
          story.id === storyId ? { ...story, ...payload } : story,
        ),
      })),
    };
  });
};

const updateSearchResults = (
  queryClient: ReturnType<typeof useQueryClient>,
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  queryClient
    .getQueriesData<SearchResponse>({ queryKey: ["search"] })
    .forEach(([queryKey, data]) => {
      if (data?.stories) {
        queryClient.setQueryData<SearchResponse>(queryKey, {
          ...data,
          stories: data.stories.map((story) =>
            story.id === storyId ? { ...story, ...payload } : story,
          ),
        });
      }
    });
};
