import type { Metadata } from "next";
import { TeamRoadmapPage } from "@/modules/roadmap/public/client";
import { getTeam } from "@/modules/teams/queries/get-team";
import { auth } from "@/auth";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}): Promise<Metadata> {
  const [{ teamId, workspaceSlug }, session] = await Promise.all([
    params,
    auth(),
  ]);
  const ctx = { session: session!, workspaceSlug };
  const teamData = await getTeam(teamId, ctx);

  return {
    title: `${teamData.data?.name || "Team"} › Objectives`,
  };
}

export default function Page() {
  return <TeamRoadmapPage />;
}
