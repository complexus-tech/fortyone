import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { memberKeys, teamKeys } from "@/constants/keys";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { TeamManagement } from "@/modules/settings/workspace/teams/management";
import { auth } from "@/auth";
import { getTeamSettings } from "@/modules/teams/queries/get-team-settings";
import { getTeam } from "@/modules/teams/queries/get-team";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}): Promise<Metadata> {
  const { teamId, workspaceSlug } = await params;
  const session = await auth();
  const teamData = await getTeam(teamId, { session: session!, workspaceSlug });

  return {
    title: `Settings › ${teamData.data?.name || "Team"}`,
  };
}

export default async function TeamManagementPage({
  params,
}: {
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}) {
  const { teamId, workspaceSlug } = await params;
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };

  const queryClient = getQueryClient();
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: memberKeys.team(workspaceSlug, teamId),
      queryFn: () => getTeamMembers(teamId, ctx),
    }),
    queryClient.prefetchQuery({
      queryKey: teamKeys.settings(workspaceSlug, teamId),
      queryFn: () => getTeamSettings(teamId, ctx),
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TeamManagement />
    </HydrationBoundary>
  );
}
