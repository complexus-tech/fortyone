import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { memberKeys } from "@/constants/keys";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { TeamManagement } from "@/modules/settings/workspace/teams/management";

export default async function TeamManagementPage({
  params,
}: {
  params: Promise<{ teamId: string }>;
}) {
  const { teamId } = await params;

  const queryClient = getQueryClient();
  await queryClient.prefetchQuery({
    queryKey: memberKeys.team(teamId),
    queryFn: () => getTeamMembers(teamId),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TeamManagement />
    </HydrationBoundary>
  );
}
