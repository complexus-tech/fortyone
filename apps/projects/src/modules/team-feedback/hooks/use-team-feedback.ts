import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import {
  getTeamFeedback,
  getTeamFeedbackPage,
} from "../queries/get-team-feedback";
import { getTeamFeedbackMergeCandidates } from "../queries/get-feedback";
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

export const useTeamFeedbackMergeCandidates = (
  sourceItemId: string,
  search: string,
  enabled: boolean,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: feedbackKeys.mergeCandidates(
      workspaceSlug,
      sourceItemId,
      normalizedSearch,
    ),
    queryFn: () =>
      getTeamFeedbackMergeCandidates(sourceItemId, normalizedSearch, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(enabled && session && sourceItemId),
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
