import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import {
  getTeamIntegrationRequests,
  getTeamIntegrationRequestsPage,
} from "../queries/get-team-requests";
import type { IntegrationRequestStatus } from "../types";

export const useTeamIntegrationRequests = (
  teamId: string,
  status: IntegrationRequestStatus = "pending",
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: integrationRequestKeys.team(workspaceSlug, teamId, status),
    queryFn: () =>
      getTeamIntegrationRequests(
        teamId,
        { session: session!, workspaceSlug },
        status,
      ),
    enabled: Boolean(teamId && session),
  });
};

export const useTeamIntegrationRequestsInfinite = (
  teamId: string,
  status: IntegrationRequestStatus = "pending",
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useInfiniteQuery({
    queryKey: [
      ...integrationRequestKeys.team(workspaceSlug, teamId, status),
      "infinite",
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamIntegrationRequestsPage(
        teamId,
        { session: session!, workspaceSlug },
        status,
        pageParam,
      ),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
    enabled: Boolean(teamId && session),
  });
};
