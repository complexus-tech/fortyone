import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { InfiniteData } from "@tanstack/react-query";
import { toast } from "sonner";
import { memberKeys } from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { deriveAutoSchedulingStatus } from "@/lib/auto-scheduling";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
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
  patchStories,
  parseGroupQueryKey,
  updateStoryInGroups,
} from "@/modules/stories/utils/optimistic";
import type { DetailedStory, StoryUpdate } from "../types";
import { updateStoryAction } from "../actions/update-story";

type UpdateStoryVariables = {
  storyId: string;
  payload: StoryUpdate;
};

type UpdateStoryContext = {
  previousStory?: DetailedStory;
};

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();
  const { getTermDisplay } = useTerminology();

  const mutation = useMutation<
    ApiResponse<null>,
    Error,
    UpdateStoryVariables,
    UpdateStoryContext
  >({
    mutationFn: async ({ storyId, payload }) => {
      const response = await updateStoryAction(storyId, payload, workspaceSlug);
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },

    onMutate: ({ storyId, payload }) => {
      const storyPayload = { ...payload };
      delete storyPayload.reconcileDescriptionMedia;

      queryClient.cancelQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(workspaceSlug, storyId),
      );
      const optimisticPayload = buildOptimisticStoryPayload(
        queryClient,
        workspaceSlug,
        storyPayload,
        previousStory,
      );

      const activeQueries = queryClient.getQueryCache().findAll({
        queryKey: storyKeys.all(workspaceSlug),
      });

      activeQueries.forEach((query) => {
        if (query.isActive()) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          if (query.queryKey.includes("detail")) {
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
      toast.error(`Failed to update ${getTermDisplay("storyTerm")}`, {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (_res, { storyId, payload }) => {
      const storyPayload = { ...payload };
      delete storyPayload.reconcileDescriptionMedia;
      analytics.track("story_updated", {
        storyId,
        ...storyPayload,
      });

      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
        refetchType: "inactive",
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.activitiesInfinite(workspaceSlug, storyId),
        refetchType: "all",
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });
    },
  });

  return mutation;
};

export const buildOptimisticStoryPayload = (
  queryClient: ReturnType<typeof useQueryClient>,
  workspaceSlug: string,
  payload: Partial<DetailedStory>,
  currentStory?: DetailedStory,
): Partial<DetailedStory> => {
  const optimisticPayload = { ...payload };
  const changesSchedulingInputs =
    "autoSchedulingEnabled" in payload ||
    "autoSchedulingLocked" in payload ||
    "assigneeId" in payload ||
    "estimatedDurationMinutes" in payload;

  if (changesSchedulingInputs && currentStory) {
    const mergedStory = { ...currentStory, ...payload };
    optimisticPayload.autoSchedulingStatus = deriveAutoSchedulingStatus({
      assigneeId: mergedStory.assigneeId,
      autoSchedulingEnabled: mergedStory.autoSchedulingEnabled,
      autoSchedulingLocked: mergedStory.autoSchedulingLocked,
      estimatedDurationMinutes: mergedStory.estimatedDurationMinutes,
    });
    optimisticPayload.autoSchedulingReason = null;
  }

  if ("assigneeId" in payload) {
    optimisticPayload.assignee = payload.assigneeId
      ? resolveAssigneeSummary(queryClient, workspaceSlug, payload.assigneeId)
      : null;
  }

  return optimisticPayload;
};

const resolveAssigneeSummary = (
  queryClient: ReturnType<typeof useQueryClient>,
  workspaceSlug: string,
  assigneeId: string,
): UserSummary | null => {
  const memberQueries = queryClient.getQueriesData<
    | Member
    | Member[]
    | MembersPage
    | InfiniteData<MembersPage>
    | ApiResponse<Member>
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
    const subStories = patchStories(parentStory.subStories, storyId, payload);
    if (subStories === parentStory.subStories) return;

    queryClient.setQueryData<DetailedStory>(queryKey, {
      ...parentStory,
      subStories,
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
      return patchStories(data, storyId, payload);
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

/** @internal Exported for focused cache-identity tests. */
export const updateInfiniteQuery = (
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
      if (
        !data.pages.some(
          (page) =>
            Array.isArray(page.stories) &&
            page.stories.some((story) => story.id === storyId),
        )
      ) {
        return data;
      }

      const pages = data.pages.map((page) => {
        if (!Array.isArray(page.stories)) return page;

        const stories = patchStories(page.stories, storyId, payload);
        if (stories === page.stories) return page;

        movedStory = stories.findLast((story) => story.id === storyId);
        return { ...page, stories };
      });

      return {
        ...data,
        pages,
      };
    }

    // moved: filter out and capture
    if (
      !data.pages.some(
        (page) =>
          Array.isArray(page.stories) &&
          page.stories.some((story) => story.id === storyId),
      )
    ) {
      return data;
    }

    const pages = data.pages.map((page) => {
      if (
        !Array.isArray(page.stories) ||
        !page.stories.some((story) => story.id === storyId)
      ) {
        return page;
      }

      const stories = page.stories.filter((story) => {
        if (story.id !== storyId) return true;

        movedStory = { ...story, ...payload };
        return false;
      });

      return { ...page, stories };
    });

    return {
      ...data,
      pages,
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

    const groups = updateStoryInGroups(
      data.groups,
      storyId,
      data.meta.groupBy,
      payload,
    );
    if (groups === data.groups) return data;

    return {
      ...data,
      groups,
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
        const stories = patchStories(data.stories, storyId, payload);
        if (stories === data.stories) return;

        queryClient.setQueryData<SearchResponse>(queryKey, {
          ...data,
          stories,
        });
      }
    });
};
