import { HydrationBoundary, dehydrate } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { sprintKeys } from "@/constants/keys";
import { SprintsList } from "@/modules/sprints";
import { getTeamSprints } from "@/modules/sprints/queries/get-team-sprints";

export default async function Page({
  params,
}: {
  params: Promise<{ teamId: string }>;
}) {
  const { teamId } = await params;
  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: sprintKeys.team(teamId),
    queryFn: () => getTeamSprints(teamId),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <SprintsList />
    </HydrationBoundary>
  );
}
