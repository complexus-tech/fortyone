import type { ReactNode } from "react";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { auth } from "@/auth";
import { getQueryClient } from "@/app/get-query-client";
import { feedbackKeys } from "@/constants/keys";
import { getTeamFeedbackPage } from "@/modules/team-feedback/queries/get-team-feedback";
import { TeamFeedbackShell } from "@/modules/team-feedback/shell";
import type { TeamFeedbackPage } from "@/modules/team-feedback/types";

export default async function Layout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}) {
  const queryClient = getQueryClient();
  const [session, { teamId, workspaceSlug }] = await Promise.all([
    auth(),
    params,
  ]);
  const ctx = { session: session!, workspaceSlug };

  await queryClient.prefetchInfiniteQuery({
    queryKey: [
      ...feedbackKeys.team(workspaceSlug, teamId, "active"),
      "infinite",
    ] as const,
    queryFn: ({ pageParam }) =>
      getTeamFeedbackPage(
        teamId,
        ctx,
        "active",
        typeof pageParam === "number" ? pageParam : 1,
      ),
    getNextPageParam: (lastPage: TeamFeedbackPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    initialPageParam: 1,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TeamFeedbackShell teamId={teamId}>{children}</TeamFeedbackShell>
    </HydrationBoundary>
  );
}
