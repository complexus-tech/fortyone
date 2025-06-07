import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { memberKeys, teamKeys } from "@/constants/keys";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { TeamManagement } from "@/modules/settings/workspace/teams/management";
import { auth } from "@/auth";
import { getTeamSettings } from "@/modules/teams/queries/get-team-settings";

export default async function TeamManagementPage({
  params,
}: {
  params: Promise<{ teamId: string }>;
}) {
  const { teamId } = await params;
  const session = await auth();

  const queryClient = getQueryClient();
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: memberKeys.team(teamId),
      queryFn: () => getTeamMembers(teamId, session!),
    }),
    queryClient.prefetchQuery({
      queryKey: teamKeys.settings(teamId),
      queryFn: () => getTeamSettings(teamId, session!),
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TeamManagement />
    </HydrationBoundary>
  );
}
