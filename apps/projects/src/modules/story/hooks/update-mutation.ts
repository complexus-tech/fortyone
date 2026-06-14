import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { InfiniteData } from "@tanstack/react-query";
import { toast } from "sonner";
import { memberKeys } from "@/constants/keys";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
  GroupStoryParams,
  Story,
} from "@/modules/stories/types";
import type { SearchResponse } from "@/modules/search/types";
import type { ApiResponse, Member, MembersPage, UserSummary } from "@/types";
import {
  computeTargetKey,
  moveStoryBetweenGroups,
  parseGroupQueryKey,
} from "@/modules/stories/utils/optimistic";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload, workspaceSlug),

    onMutate: ({ storyId, payload }) => {
      const optimisticPayload = buildOptimisticStoryPayload(
        queryClient,
        workspaceSlug,
        payload,
      );

      queryClient.cancelQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(workspaceSlug, storyId),
      );

      const activeQueries = queryClient.getQueryCache().getAll();

      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(
              queryClient,
              query.queryKey,
              storyId,
              optimisticPayload,
            );
          } else {
            updateListQuery(
              queryClient,
              query.queryKey,
              storyId,
              optimisticPayload,
            );
          }
        }
      });

      updateSearchResults(queryClient, storyId, optimisticPayload);

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          {
            ...previousStory,
            ...optimisticPayload,
          },
        );
        return { previousStory };
      }
    },

    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, variables.storyId),
          context.previousStory,
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
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

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
      queryClient.invalidateQueries({
        queryKey: storyKeys.activitiesInfinite(workspaceSlug, storyId),
        refetchType: "all",
      });
    },
  });

  return mutation;
};

export const buildOptimisticStoryPayload = (
  queryClient: ReturnType<typeof useQueryClient>,
  workspaceSlug: string,
  payload: Partial<DetailedStory>,
): Partial<DetailedStory> => {
  if (!("assigneeId" in payload)) {
    return payload;
  }

  return {
    ...payload,
    assignee: payload.assigneeId
      ? resolveAssigneeSummary(queryClient, workspaceSlug, payload.assigneeId)
      : null,
  };
};

const resolveAssigneeSummary = (
  queryClient: ReturnType<typeof useQueryClient>,
  workspaceSlug: string,
  assigneeId: string,
): UserSummary | null => {
  const memberQueries = queryClient.getQueriesData<
    Member | Member[] | MembersPage | InfiniteData<MembersPage> | ApiResponse<Member>
  >({
    queryKey: memberKeys.all(workspaceSlug),
  });

  for (const [, data] of memberQueries) {
    const member = findMemberInCachedData(data, assigneeId);
    if (member) {
      return toUserSummary(member);
    }
  }

  return null;
};

const findMemberInCachedData = (
  data:
    | Member
    | Member[]
    | MembersPage
    | InfiniteData<MembersPage>
    | ApiResponse<Member>
    | undefined,
  assigneeId: string,
): Member | undefined => {
  if (!data) {
    return undefined;
  }
  if (Array.isArray(data)) {
    return data.find((member) => member.id === assigneeId);
  }
  if (isApiResponse(data)) {
    return data.data?.id === assigneeId ? data.data : undefined;
  }
  if ("pages" in data) {
    for (const page of data.pages) {
      const member = page.members.find(({ id }) => id === assigneeId);
      if (member) {
        return member;
      }
    }
    return undefined;
  }
  if ("members" in data) {
    return data.members.find((member) => member.id === assigneeId);
  }
  return data.id === assigneeId ? data : undefined;
};

const isApiResponse = (
  data: Member | MembersPage | InfiniteData<MembersPage> | ApiResponse<Member>,
): data is ApiResponse<Member> => "data" in data;

const toUserSummary = (member: Member): UserSummary => ({
  id: member.id,
  username: member.username,
  fullName: member.fullName,
  avatarUrl: member.avatarUrl,
  isActive: member.isActive,
  isSystem: member.isSystem,
});

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
  if (Array.isArray(queryData)) {
    queryClient.setQueryData<Story[]>(queryKey, (data) => {
      if (!Array.isArray(data)) return data;
      return data.map((story) =>
        story.id === storyId ? { ...story, ...payload } : story,
      );
    });
    return;
  }

  const isInfiniteQuery = Boolean(
    queryData && typeof queryData === "object" && "pages" in queryData,
  );

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
  // Store the real story object once we encounter it
  let movedStory: Story | undefined;

  const patchInfiniteQuery = (
    key: readonly unknown[],
    updater: (
      data: InfiniteData<GroupStoriesResponse> | undefined,
    ) => InfiniteData<GroupStoriesResponse> | undefined,
  ) =>
    queryClient.setQueryData<InfiniteData<GroupStoriesResponse>>(key, updater);

  // patch current query (remove/update)
  patchInfiniteQuery(queryKey, (data) => {
    if (!data?.pages) return data;
    const { groupKey, params } = parseGroupQueryKey(queryKey);
    const target = computeTargetKey(params.groupBy ?? "none", payload);

    if (!target || target === groupKey) {
      return {
        ...data,
        pages: data.pages.map((p) => {
          if (!Array.isArray(p.stories)) return p;
          return {
            ...p,
            stories: p.stories.map((s) => {
              if (s.id === storyId) {
                movedStory = { ...s, ...payload };
                return movedStory;
              }
              return s;
            }),
          };
        }),
      };
    }

    // moved: filter out and capture
    return {
      ...data,
      pages: data.pages.map((p) => {
        if (!Array.isArray(p.stories)) return p;
        return {
          ...p,
          stories: p.stories.filter((s) => {
            if (s.id === storyId) {
              movedStory = { ...s, ...payload };
              return false;
            }
            return true;
          }),
        };
      }),
    };
  });

  const { params: currentParams, workspaceSlug: keyWorkspaceSlug } =
    parseGroupQueryKey(queryKey);
  const targetKeyValue = computeTargetKey(
    currentParams.groupBy ?? "none",
    payload,
  );

  if (!targetKeyValue) return;

  const targetParams: Partial<GroupStoryParams> = {
    ...currentParams,
    groupKey: targetKeyValue,
  };

  const targetQueryKey = storyKeys.groupStories(
    keyWorkspaceSlug,
    targetKeyValue,
    targetParams,
  );

  patchInfiniteQuery(targetQueryKey, (data) => {
    if (!movedStory) return data;

    // If no pages yet, create minimal first page
    if (!data?.pages || data.pages.length === 0) {
      return {
        pages: [
          {
            groupKey: targetKeyValue,
            stories: [movedStory],
            pagination: {
              page: 1,
              pageSize: 1,
              hasMore: false,
              nextPage: 2,
            },
            filters: {},
          },
        ],
        pageParams: [1],
      } as InfiniteData<GroupStoriesResponse>;
    }

    const firstPage = data.pages[0];
    const firstPageStories = Array.isArray(firstPage.stories)
      ? firstPage.stories
      : [];

    // Avoid duplicates
    if (firstPageStories.some((s) => s.id === storyId)) return data;

    return {
      ...data,
      pages: [
        { ...firstPage, stories: [movedStory, ...firstPageStories] },
        ...data.pages.slice(1),
      ],
    };
  });
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueryData<GroupedStoriesResponse>(queryKey, (data) => {
    if (!data || !Array.isArray(data.groups)) return data;

    const target = computeTargetKey(data.meta.groupBy, payload);
    return {
      ...data,
      groups: moveStoryBetweenGroups(data.groups, storyId, target, payload),
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
      if (Array.isArray(data?.stories)) {
        queryClient.setQueryData<SearchResponse>(queryKey, {
          ...data,
          stories: data.stories.map((story) =>
            story.id === storyId ? { ...story, ...payload } : story,
          ),
        });
      }
    });
};
