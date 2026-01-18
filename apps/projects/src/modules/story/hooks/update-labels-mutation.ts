import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { InfiniteData } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
} from "@/modules/stories/types";
import { labelKeys } from "@/constants/keys";
import type { DetailedStory } from "../types";
import { updateLabelsAction } from "../actions/update-labels";

// Helper function to update detail queries
const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  labels: string[],
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: DetailedStory | undefined) => {
      if (data?.subStories) {
        return {
          ...data,
          subStories: data.subStories.map((story) =>
            story.id === storyId ? { ...story, labels } : story,
          ),
        };
      }
      return data;
    },
  );
};

// Helper function to update infinite queries
const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  labels: string[],
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: InfiniteData<GroupStoriesResponse> | undefined) => {
      if (!data?.pages) return data;
      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          stories: page.stories.map((story) =>
            story.id === storyId ? { ...story, labels } : story,
          ),
        })),
      };
    },
  );
};

// Helper function to update grouped queries
const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  labels: string[],
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: GroupedStoriesResponse | undefined) => {
      if (!data) return data;
      return {
        ...data,
        groups: data.groups.map((group) => ({
          ...group,
          stories: group.stories.map((story) =>
            story.id === storyId ? { ...story, labels } : story,
          ),
        })),
      };
    },
  );
};

// Helper function to update list queries (grouped or infinite)
const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  labels: string[],
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, storyId, labels);
  } else {
    updateGroupedQuery(queryClient, queryKey, storyId, labels);
  }
};

export const useUpdateLabelsMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({ storyId, labels }: { storyId: string; labels: string[] }) =>
      updateLabelsAction(storyId, labels, workspaceSlug),

    onMutate: ({ storyId, labels }) => {
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(workspaceSlug, storyId),
      );
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          {
            ...previousStory,
            labels,
          },
        );
      }

      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyId, labels);
          } else {
            updateListQuery(queryClient, query.queryKey, storyId, labels);
          }
        }
      });

      return { previousStory };
    },

    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, variables.storyId),
          context.previousStory,
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });

      toast.error("Failed to update labels", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
      queryClient.invalidateQueries({ queryKey: labelKeys.lists(workspaceSlug) });
    },
  });

  return mutation;
};
