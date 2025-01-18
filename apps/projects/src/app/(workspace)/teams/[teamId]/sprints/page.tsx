import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { SprintsList } from "@/modules/sprints";
import { getTeamSprints } from "@/modules/sprints/queries/get-team-sprints";
import { sprintKeys } from "@/constants/keys";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
  }>;
}) {
  const params = await props.params;

  const queryClient = getQueryClient();

  const { teamId } = params;
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
