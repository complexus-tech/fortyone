import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { TeamObjectivesList } from "@/modules/objectives";
import { getQueryClient } from "@/app/get-query-client";
import { getTeamObjectives } from "@/modules/objectives/queries/get-team-objectives";
import { objectiveKeys } from "@/modules/objectives/constants";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
  }>;
}) {
  const params = await props.params;

  const { teamId } = params;

  const queryClient = getQueryClient();

  await queryClient.prefetchQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: () => getTeamObjectives(teamId),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TeamObjectivesList />;
    </HydrationBoundary>
  );
}
