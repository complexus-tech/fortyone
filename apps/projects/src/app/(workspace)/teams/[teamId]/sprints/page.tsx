import { SprintsList } from "@/modules/teams/sprints";
import { getTeamSprints } from "@/modules/teams/sprints/queries/get-sprints";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
  }>;
}) {
  const params = await props.params;

  const { teamId } = params;

  const sprints = await getTeamSprints(teamId);

  return <SprintsList sprints={sprints.data!} />;
}
