import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { TeamObjectivesList } from "@/modules/objectives";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
  }>;
}) {
  const params = await props.params;

  const { teamId } = params;

  const objectives = await getObjectives();
  const teamObjectives = objectives.filter(
    (objective) => objective.teamId === teamId,
  );

  return <TeamObjectivesList objectives={teamObjectives} />;
}
