import { SprintsList } from "@/modules/teams/sprints";
import { getTeamSprints } from "@/modules/teams/sprints/queries/get-sprints";

export default async function Page() {
  const sprints = await getTeamSprints();

  return <SprintsList sprints={sprints} />;
}
