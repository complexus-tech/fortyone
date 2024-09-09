import { SprintsList } from "@/modules/teams/sprints";
import { getTeamSprints } from "@/modules/teams/sprints/queries/get-sprints";

export default async function Page({
  params: { teamId },
}: {
  params: {
    teamId: string;
  };
}) {
  const sprints = await getTeamSprints(teamId);

  return <SprintsList sprints={sprints} />;
}
