import { useEffect, useState } from "react";
import { useQueries } from "@tanstack/react-query";
import { formatISO, parseISO } from "date-fns";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getGroupedStories } from "@/modules/stories/public/queries";
import { storyKeys } from "@/modules/stories/public/keys";
import {
  getStoryAttentionFilters,
  STORY_ATTENTION_VIEWS,
} from "@/shared/story/attention";

export const useWorkAttention = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const [today, setToday] = useState<string | null>(null);
  useEffect(() => {
    const updateDate = () => {
      setToday(formatISO(new Date(), { representation: "date" }));
    };
    updateDate();
    const interval = window.setInterval(updateDate, 60_000);
    window.addEventListener("focus", updateDate);
    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", updateDate);
    };
  }, []);
  const enabled = Boolean(session?.user.id && workspaceSlug && today);
  const queries = useQueries({
    queries: STORY_ATTENTION_VIEWS.map((view) => {
      const filters = today
        ? getStoryAttentionFilters(view, parseISO(today))
        : {};
      const params = {
        ...filters,
        groupBy: "none" as const,
        storiesPerGroup: 1,
      };
      return {
        queryKey: [
          ...storyKeys.mineGrouped(workspaceSlug, params),
          "maya-attention",
          view,
          session?.user.id,
        ],
        queryFn: () => getGroupedStories({ session, workspaceSlug }, params),
        enabled,
        staleTime: 60_000,
        refetchInterval: 60_000,
      };
    }),
  });
  return {
    items: STORY_ATTENTION_VIEWS.map((view, index) => ({
      view,
      count: queries[index].data?.groups.reduce(
        (total, group) => total + group.totalCount,
        0,
      ),
    })),
    isPending: !enabled || queries.some((query) => query.isPending),
    isError: queries.some((query) => query.isError),
    retry: () => {
      queries.forEach((query) => {
        if (query.isError) void query.refetch();
      });
    },
  };
};
