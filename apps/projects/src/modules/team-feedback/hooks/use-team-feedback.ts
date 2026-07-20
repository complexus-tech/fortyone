import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import {
  getTeamFeedback,
  getTeamFeedbackPage,
} from "../queries/get-team-feedback";
import type { TeamFeedbackListStatus } from "../types";

export const useTeamFeedback = (
  teamId: string,
  status: TeamFeedbackListStatus = "active",
  search = "",
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: [...feedbackKeys.team(workspaceSlug, teamId, status), search],
    queryFn: () =>
      getTeamFeedback(
        teamId,
        { session: session!, workspaceSlug },
        status,
        search,
      ),
    enabled: Boolean(teamId && session),
  });
};

export const useTeamFeedbackInfinite = (
  teamId: string,
  status: TeamFeedbackListStatus = "active",
  search = "",
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useInfiniteQuery({
    queryKey: [
      ...feedbackKeys.team(workspaceSlug, teamId, status),
      search,
      "infinite",
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamFeedbackPage(
        teamId,
        { session: session!, workspaceSlug },
        status,
        search,
        pageParam,
      ),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(teamId && session),
  });
};
