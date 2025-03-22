import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { TeamObjectivesList } from "@/modules/objectives";
import { getQueryClient } from "@/app/get-query-client";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getTeamObjectives } from "@/modules/objectives/queries/get-team-objectives";

export default async function Page({
  params,
}: {
  params: Promise<{ teamId: string }>;
}) {
  const { teamId } = await params;
  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: () => getTeamObjectives(teamId),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TeamObjectivesList />
    </HydrationBoundary>
  );
}
