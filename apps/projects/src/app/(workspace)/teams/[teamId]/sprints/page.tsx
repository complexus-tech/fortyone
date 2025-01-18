import { SprintsList } from "@/modules/sprints";
import { getTeamSprints } from "@/modules/sprints/queries/get-team-sprints";

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
