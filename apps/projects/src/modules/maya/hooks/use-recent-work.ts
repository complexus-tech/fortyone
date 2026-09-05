import { useQueries, useQuery } from "@tanstack/react-query";
import { ApiError } from "api-client";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getActivities } from "@/lib/queries/activities/get-activities";
import { getStory } from "@/modules/story/public/queries";
import {
  getGroupedStories,
  type GroupedStoryParams,
} from "@/modules/stories/public/queries";
import { storyKeys } from "@/modules/stories/public/keys";
import { useRecentStoryHistory } from "@/shared/story/use-recent-story-history";
import {
  getRecentStories,
  RECENT_WORK_LIMIT,
  type MayaWorkTab,
} from "../utils/recent-work";

const ASSIGNED_TASK_PARAMS: GroupedStoryParams = {
  groupBy: "none",
  assignedToMe: true,
  categories: ["backlog", "unstarted", "started", "paused"],
  orderBy: "updated",
  orderDirection: "desc",
  storiesPerGroup: RECENT_WORK_LIMIT,
  showSubStories: false,
};

export const useRecentWork = (tab: MayaWorkTab = "all") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const visits = useRecentStoryHistory();
  const enabled = Boolean(session?.user.id && workspaceSlug);
  const historyEnabled = enabled && tab === "all";
  const activityQuery = useQuery({
    queryKey: ["activities", workspaceSlug, { limit: 20 }, session?.user.id],
    queryFn: () => getActivities({ session, workspaceSlug }, { limit: 20 }),
    enabled: historyEnabled,
    staleTime: 60_000,
  });
  const activities = activityQuery.data ?? [];
  const recentStories =
    tab === "all" ? getRecentStories(visits, activities) : [];
  const storyIds = recentStories.map(({ storyId }) => storyId);
  const storyQueries = useQueries({
    queries: storyIds.map((id) => ({
      queryKey: [
        ...storyKeys.detail(workspaceSlug, id),
        "maya-recent",
        session?.user.id,
      ],
      queryFn: async () => {
        try {
          return await getStory(id, { session, workspaceSlug });
        } catch (error) {
          // A saved visit can outlive the user's permission to view a task.
          if (error instanceof ApiError && error.status === 403) return null;
          throw error;
        }
      },
      enabled: historyEnabled,
      staleTime: 60_000,
    })),
  });
  const storiesById = new Map(
    storyQueries.flatMap(({ data, isError }) =>
      data && !isError && !data.deletedAt && !data.archivedAt
        ? [[data.id, data] as const]
        : [],
    ),
  );

  const resolvedRecentStories = recentStories.flatMap(({ storyId }) => {
    const story = storiesById.get(storyId);
    return story ? [story] : [];
  });
  const historyPending =
    activityQuery.isPending || storyQueries.some((query) => query.isPending);
  const historyError =
    activityQuery.isError || storyQueries.some((query) => query.isError);
  const isAssignedFallback =
    historyEnabled &&
    !historyPending &&
    !historyError &&
    resolvedRecentStories.length === 0;
  const shouldLoadFilteredStories =
    enabled && (tab !== "all" || isAssignedFallback);
  const filteredParams: GroupedStoryParams =
    tab === "all"
      ? ASSIGNED_TASK_PARAMS
      : {
          groupBy: "none",
          orderBy: "updated",
          orderDirection: "desc",
          storiesPerGroup: RECENT_WORK_LIMIT,
          showSubStories: false,
          ...(tab === "assigned"
            ? { assignedToMe: true }
            : { createdByMe: true }),
        };
  const filteredQuery = useQuery({
    queryKey: [
      ...storyKeys.mineGrouped(workspaceSlug, filteredParams),
      "maya-work",
      session?.user.id,
    ],
    queryFn: () =>
      getGroupedStories({ session, workspaceSlug }, filteredParams),
    enabled: shouldLoadFilteredStories,
    staleTime: 60_000,
  });
  const filteredStories =
    filteredQuery.data?.groups
      .flatMap(({ stories }) => stories)
      .filter((story) => !story.deletedAt && !story.archivedAt)
      .slice(0, RECENT_WORK_LIMIT) ?? [];

  return {
    stories: shouldLoadFilteredStories
      ? filteredStories
      : resolvedRecentStories,
    isAssignedFallback: isAssignedFallback && filteredStories.length > 0,
    isPending:
      enabled &&
      ((historyEnabled && historyPending) ||
        (shouldLoadFilteredStories && filteredQuery.isPending)),
    isError:
      (historyEnabled && historyError) ||
      (shouldLoadFilteredStories && filteredQuery.isError),
    retry: () => {
      if (historyEnabled && activityQuery.isError) void activityQuery.refetch();
      storyQueries.forEach((query) => {
        if (query.isError) void query.refetch();
      });
      if (shouldLoadFilteredStories && filteredQuery.isError)
        void filteredQuery.refetch();
    },
  };
};
