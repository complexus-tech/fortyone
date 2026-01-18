import type { Metadata } from "next";
import { SprintsList } from "@/modules/sprints";
import { getTeam } from "@/modules/teams/queries/get-team";
import { auth } from "@/auth";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}): Promise<Metadata> {
  const { teamId, workspaceSlug } = await params;
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };
  const teamData = await getTeam(teamId, ctx);

  return {
    title: `${teamData.data?.name || "Team"} › Sprints`,
  };
}

export default function Page() {
  return <SprintsList />;
}
