import type { Metadata } from "next";
import { TeamObjectivesList } from "@/modules/objectives";
import { getTeam } from "@/modules/teams/queries/get-team";
import { auth } from "@/auth";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ teamId: string }>;
}): Promise<Metadata> {
  const { teamId } = await params;
  const session = await auth();
  const teamData = await getTeam(teamId, session!);

  return {
    title: `${teamData.data?.name || "Team"} › Objectives`,
  };
}

export default function Page() {
  return <TeamObjectivesList />;
}
